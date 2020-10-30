package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	dc "github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	kitprom "github.com/go-kit/kit/metrics/prometheus"
	gomysql "github.com/go-sql-driver/mysql"
	"github.com/ory/dockertest/v3"
	stdprom "github.com/prometheus/client_golang/prometheus"

	"github.com/moov-io/base/docker"
	"github.com/moov-io/base/log"
)

var (
	mysqlConnections = kitprom.NewGaugeFrom(stdprom.GaugeOpts{
		Name: "mysql_connections",
		Help: "How many MySQL connections and what status they're in.",
	}, []string{"state"})

	// mySQLErrDuplicateKey is the error code for duplicate entries
	// https://dev.mysql.com/doc/refman/8.0/en/server-error-reference.html#error_er_dup_entry
	mySQLErrDuplicateKey uint16 = 1062

	maxActiveMySQLConnections = func() int {
		if v := os.Getenv("MYSQL_MAX_CONNECTIONS"); v != "" {
			if n, _ := strconv.ParseInt(v, 10, 32); n > 0 {
				return int(n)
			}
		}
		return 16
	}()
)

type mysql struct {
	dsn    string
	logger log.Logger

	connections *kitprom.Gauge
}

func (my *mysql) Connect(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("mysql", my.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxActiveMySQLConnections)

	// Check out DB is up and working
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Setup metrics after the database is setup
	go func() {
		t := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				stats := db.Stats()
				my.connections.With("state", "idle").Set(float64(stats.Idle))
				my.connections.With("state", "inuse").Set(float64(stats.InUse))
				my.connections.With("state", "open").Set(float64(stats.OpenConnections))
			}
		}
	}()

	return db, nil
}

func mysqlConnection(logger log.Logger, config MySQLConfig) *mysql {
	timeout := "30s"
	if v := os.Getenv("MYSQL_TIMEOUT"); v != "" {
		timeout = v
	}
	params := fmt.Sprintf("timeout=%s&charset=utf8mb4&parseTime=true&sql_mode=ALLOW_INVALID_DATES", timeout)
	dsn := fmt.Sprintf("%s:%s@%s/%s?%s", config.User, config.Password, config.Address, config.Name, params)
	return &mysql{
		dsn:         dsn,
		logger:      logger,
		connections: mysqlConnections,
	}
}

var onceRunMigrations sync.Once

// CreateTestMySQLDB returns a TestMySQLDB which can be used in tests
// as a clean mysql database. All migrations are ran on the db before.
//
// Callers should call close on the returned *TestMySQLDB.
func CreateTestMySQLDB(t *testing.T) *TestMySQLDB {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	if !docker.Enabled() {
		t.Skip("Docker not enabled")
	}

	config := Config{
		Type: TypeMySQL,
		MySQL: MySQLConfig{
			Name:     "test",
			User:     "moov",
			Password: "secret",
		},
	}

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	container, err := getMySQLDockerInstance(pool, "mysql-test-db", &config.MySQL)
	require.NoError(t, err)

	config.MySQL.Address = fmt.Sprintf("tcp(localhost:%s)", container.GetPort("3306/tcp"))
	dbURL := fmt.Sprintf("%s:%s@%s/%s",
		config.MySQL.User,
		config.MySQL.Password,
		config.MySQL.Address,
		config.MySQL.Name,
	)

	var db *sql.DB
	err = pool.Retry(func() error {
		db, err = sql.Open("mysql", dbURL)
		require.NoError(t, err)

		return db.Ping()
	})
	if err != nil {
		container.Close()
		require.FailNow(t, err.Error())
	}
	// Don't allow idle connections so we can verify all are closed at the end of testing
	db.SetMaxIdleConns(0)

	// Run DB migrations
	onceRunMigrations.Do(func() {
		err = RunMigrations(log.NewNopLogger(), config)
		require.NoError(t, err)
	})

	result := &TestMySQLDB{
		DB:        db,
		container: nil,
		t:         t,
		logger:    log.NewDefaultLogger(),
	}
	// Teardown to ensure we're working with a clean DB
	result.Teardown()

	return result
}

func getMySQLDockerInstance(pool *dockertest.Pool, containerName string, config *MySQLConfig) (*dockertest.Resource, error) {
	resource, ok := pool.ContainerByName(containerName)
	if ok {
		return resource, nil
	}

	return pool.RunWithOptions(&dockertest.RunOptions{
		Name:       containerName,
		Repository: "mysql",
		Tag:        "8",
		Env: []string{
			fmt.Sprintf("MYSQL_USER=%s", config.User),
			fmt.Sprintf("MYSQL_PASSWORD=%s", config.Password),
			"MYSQL_ROOT_PASSWORD=secret",
			fmt.Sprintf("MYSQL_DATABASE=%s", config.Name),
		},
	}, func(dockerConfig *dc.HostConfig) {
		dockerConfig.AutoRemove = true
	},
	)
}

// TestMySQLDB is a wrapper around sql.DB for MySQL connections designed for tests to provide
// a clean database for each testcase.  Callers should cleanup with Close() when finished.
type TestMySQLDB struct {
	*sql.DB
	container *dockertest.Resource
	t         *testing.T
	logger    log.Logger
}

func (r *TestMySQLDB) Close() error {
	// Verify all connections are closed before closing DB
	if conns := r.DB.Stats().OpenConnections; conns != 0 {
		require.FailNow(r.t, ErrOpenConnections{
			Database:       "mysql",
			NumConnections: conns,
		}.Error())
	}

	if err := r.DB.Close(); err != nil {
		return err
	}

	return nil
}

func (r *TestMySQLDB) Teardown() {
	// List of tables we don't want to touch in this teardown
	blockList := map[string]bool{
		"schema_migrations": true,
	}

	// Temporarily disable foreign key requirements
	_, err := r.DB.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	require.NoError(r.t, err)
	defer r.DB.Exec("SET FOREIGN_KEY_CHECKS = 1;")

	// Delete all rows from all tables
	query := "select TABLE_NAME from information_schema.tables where table_schema = (select database());"
	stmt, err := r.DB.Prepare(query)
	require.NoError(r.t, err)
	rows, err := stmt.Query()
	require.NoError(r.t, err)

	for rows.Next() {
		var table string
		require.NoError(r.t, rows.Scan(&table))

		if _, ok := blockList[table]; ok {
			continue
		}

		result, err := r.DB.Exec(fmt.Sprintf("delete from %s", table))
		require.NoError(r.t, err)

		numRowsDeleted, err := result.RowsAffected()
		require.NoError(r.t, err)

		r.logger.Logf("Deleted %d rows from %s", numRowsDeleted, table)
	}
}

// MySQLUniqueViolation returns true when the provided error matches the MySQL code
// for duplicate entries (violating a unique table constraint).
func MySQLUniqueViolation(err error) bool {
	match := strings.Contains(err.Error(), fmt.Sprintf("Error %d: Duplicate entry", mySQLErrDuplicateKey))
	if e, ok := err.(*gomysql.MySQLError); ok {
		return match || e.Number == mySQLErrDuplicateKey
	}
	return match
}
