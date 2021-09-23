// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	"github.com/moov-io/base/docker"

	kitprom "github.com/go-kit/kit/metrics/prometheus"
	gomysql "github.com/go-sql-driver/mysql"
	dc "github.com/ory/dockertest/v3/docker"
	stdprom "github.com/prometheus/client_golang/prometheus"

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

type discardLogger struct{}

func (l discardLogger) Print(v ...interface{}) {}

func init() {
	gomysql.SetLogger(discardLogger{})
}

type mysql struct {
	dsn    string
	logger log.Logger
	tls    *tls.Config

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

func mysqlConnection(logger log.Logger, mysqlConfig *MySQLConfig, databaseName string) (*mysql, error) {
	timeout := "30s"
	if v := os.Getenv("MYSQL_TIMEOUT"); v != "" {
		timeout = v
	}
	params := fmt.Sprintf(`timeout=%s&charset=utf8mb4&parseTime=true&sql_mode="ALLOW_INVALID_DATES,STRICT_ALL_TABLES"`, timeout)

	var tlsConfig *tls.Config

	if mysqlConfig.UseTLS {
		// If any custom options are set then we need to create a custom TLS configuration. Otherwise we can just set
		// tls=true in the DSN
		if mysqlConfig.InsecureSkipVerify || mysqlConfig.TLSCAFile != "" || mysqlConfig.VerifyCAFile {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: mysqlConfig.InsecureSkipVerify,
			}

			if mysqlConfig.TLSCAFile != "" {
				rootCertPool := x509.NewCertPool()
				certPem, err := ioutil.ReadFile(mysqlConfig.TLSCAFile)
				if err != nil {
					return nil, err
				}

				block, _ := pem.Decode(certPem)
				caCert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					return nil, err
				}

				if appendOK := rootCertPool.AppendCertsFromPEM(certPem); !appendOK {
					return nil, err
				}
				tlsConfig.RootCAs = rootCertPool

				if mysqlConfig.VerifyCAFile {
					tlsConfig.VerifyConnection = func(state tls.ConnectionState) error {
						if !(state.PeerCertificates[len(state.PeerCertificates)-1].Equal(caCert)) {
							return errors.New("server certificate chain does not start with CA cert")
						}
						return nil
					}
				}
			}

			const TLS_CONFIG_NAME = "custom"

			gomysql.RegisterTLSConfig(TLS_CONFIG_NAME, tlsConfig)
			params = params + fmt.Sprintf("&tls=%s", TLS_CONFIG_NAME)

		} else {
			params = params + "&tls=true"
		}
	}

	dsn := fmt.Sprintf("%s:%s@%s/%s?%s", mysqlConfig.User, mysqlConfig.Password, mysqlConfig.Address, databaseName, params)

	return &mysql{
		dsn:         dsn,
		logger:      logger,
		tls:         tlsConfig,
		connections: mysqlConnections,
	}, nil
}

// TestMySQLDB is a wrapper around sql.DB for MySQL connections designed for tests to provide
// a clean database for each testcase.  Callers should cleanup with Close() when finished.
type TestMySQLDB struct {
	*sql.DB
	name     string
	shutdown func() // context shutdown func
	t        *testing.T
}

func (r *TestMySQLDB) Close() error {
	r.shutdown()

	// Verify all connections are closed before closing DB
	if conns := r.DB.Stats().OpenConnections; conns != 0 {
		require.FailNow(r.t, ErrOpenConnections{
			Database:       "mysql",
			NumConnections: conns,
		}.Error())
	}

	_, err := r.DB.Exec(fmt.Sprintf("drop database %s", r.name))
	if err != nil {
		return err
	}

	if err := r.DB.Close(); err != nil {
		return err
	}

	return nil
}

var sharedMySQLConfig *MySQLConfig
var mySQLTestDBSetup sync.Once

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

	mySQLTestDBSetup.Do(func() {
		var err error
		sharedMySQLConfig, err = findOrLaunchMySQLContainer()
		require.NoError(t, err)
	})

	dbName, err := createTemporaryDatabase(t, sharedMySQLConfig)
	require.NoError(t, err)

	dbConfig := &DatabaseConfig{
		DatabaseName: dbName,
		MySQL:        sharedMySQLConfig,
	}

	logger := log.NewNopLogger()
	ctx, cancelFunc := context.WithCancel(context.Background())
	db, err := NewAndMigrate(ctx, logger, *dbConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Don't allow idle connections so we can verify all are closed at the end of testing
	db.SetMaxIdleConns(0)

	return &TestMySQLDB{
		DB:       db,
		name:     dbName,
		shutdown: cancelFunc,
		t:        t,
	}
}

// We connect as root to MySQL server and create database with random name to
// run our migrations on it later.
func createTemporaryDatabase(t *testing.T, config *MySQLConfig) (string, error) {
	dsn := fmt.Sprintf("%s:%s@%s/", "root", config.Password, config.Address)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return "", err
	}
	defer db.Close()

	maxIdx := len(t.Name()) - 1
	if maxIdx > 20 {
		maxIdx = 20
	}

	// Set dbName to something like `TestCreateTemporaryD-Jun-25-08:30:07`
	dbName := fmt.Sprintf(
		"%s %s",
		t.Name()[:maxIdx],
		time.Now().Local().Format(time.Stamp),
	)
	dbName = strings.ReplaceAll(dbName, " ", "-")

	_, err = db.ExecContext(context.Background(), fmt.Sprintf("create database `%s`", dbName))
	if err != nil {
		return "", err
	}

	_, err = db.ExecContext(context.Background(), fmt.Sprintf("grant all on `%s`.* to '%s'@'%%'", dbName, config.User))
	if err != nil {
		return "", err
	}

	return dbName, nil
}

func findOrLaunchMySQLContainer() (*MySQLConfig, error) {
	var containerName = "moov-mysql-test-container"
	var resource *dockertest.Resource
	var err error

	config := &MySQLConfig{
		User:     "moov",
		Password: "secret",
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, err
	}

	_, err = pool.RunWithOptions(&dockertest.RunOptions{
		Name:       containerName,
		Repository: "moov/mysql-volumeless",
		Tag:        "8.0",
		Env: []string{
			fmt.Sprintf("MYSQL_USER=%s", config.User),
			fmt.Sprintf("MYSQL_PASSWORD=%s", config.Password),
			fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", config.Password),
		},
	})

	if err != nil && !errors.Is(err, dc.ErrContainerAlreadyExists) {
		return nil, err
	}

	// look for running container
	resource, found := pool.ContainerByName(containerName)
	if !found {
		return nil, errors.New("failed to launch (or find) MySQL container")
	}

	config.Address = fmt.Sprintf("tcp(localhost:%s)", resource.GetPort("3306/tcp"))

	dbURL := fmt.Sprintf("%s:%s@%s/",
		config.User,
		config.Password,
		config.Address,
	)

	err = pool.Retry(func() error {
		db, err := sql.Open("mysql", dbURL)
		if err != nil {
			return err
		}
		defer db.Close()
		return db.Ping()
	})
	if err != nil {
		resource.Close()
		return nil, err
	}

	return config, nil
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
