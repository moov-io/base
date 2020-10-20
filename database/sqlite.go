package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	kitprom "github.com/go-kit/kit/metrics/prometheus"
	"github.com/mattn/go-sqlite3"
	stdprom "github.com/prometheus/client_golang/prometheus"

	"github.com/moov-io/base/log"
)

var (
	sqliteConnections = kitprom.NewGaugeFrom(stdprom.GaugeOpts{
		Name: "sqlite_connections",
		Help: "How many sqlite connections and what status they're in.",
	}, []string{"state"})

	sqliteVersionLogOnce sync.Once
)

type sqlite struct {
	path string

	connections *kitprom.Gauge
	logger      log.Logger

	err error
}

func (s *sqlite) Connect(ctx context.Context) (*sql.DB, error) {
	if s.err != nil {
		return nil, fmt.Errorf("sqlite had error %v", s.err)
	}

	sqliteVersionLogOnce.Do(func() {
		if v, _, _ := sqlite3.Version(); v != "" {
			s.logger.Info().Logf(fmt.Sprintf("sqlite version %s", v))
		}
	})

	db, err := sql.Open("sqlite3", s.path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return db, err
	}

	// Spin up metrics only after everything works
	go func() {
		t := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				stats := db.Stats()
				s.connections.With("state", "idle").Set(float64(stats.Idle))
				s.connections.With("state", "inuse").Set(float64(stats.InUse))
				s.connections.With("state", "open").Set(float64(stats.OpenConnections))
			}

		}
	}()

	return db, err
}

func sqliteConnection(logger log.Logger, path string) *sqlite {
	return &sqlite{
		path:        path,
		logger:      logger,
		connections: sqliteConnections,
	}
}

// TestSQLiteDB is a wrapper around sql.DB for SQLite connections designed for tests to provide
// a clean database for each testcase.  Callers should cleanup with Close() when finished.
type TestSQLiteDB struct {
	*sql.DB
	dir      string // temp dir created for sqlite files
	shutdown func() // context shutdown func
}

func (r *TestSQLiteDB) Close() error {
	r.shutdown()

	// Verify all connections are closed before closing DB
	if conns := r.DB.Stats().Idle; conns != 0 {
		panic(fmt.Sprintf("found %d open sqlite connection(s)", conns))
	}
	if err := r.DB.Close(); err != nil {
		return err
	}
	return os.RemoveAll(r.dir)
}

// CreateTestSQLiteDB returns a TestSQLiteDB which can be used in tests
// as a clean sqlite database. All migrations are ran on the db before.
//
// Callers should call close on the returned *TestSQLiteDB.
func CreateTestSQLiteDB(t *testing.T) *TestSQLiteDB {
	dir, err := ioutil.TempDir("", "sqlite-test")
	if err != nil {
		t.Fatalf("sqlite test: %v", err)
	}

	dbPath := filepath.Join(dir, "tests.db")

	config := &DatabaseConfig{SQLite: &SQLiteConfig{
		Path: dbPath,
	}}

	logger := log.NewNopLogger()
	ctx, cancelFunc := context.WithCancel(context.Background())

	db, err := NewAndMigrate(ctx, logger, *config)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}

	// Don't allow idle connections so we can verify all are closed at the end of testing
	db.SetMaxIdleConns(0)

	return &TestSQLiteDB{DB: db, dir: dir, shutdown: cancelFunc}
}

// SQLiteUniqueViolation returns true when the provided error matches the SQLite error
// for duplicate entries (violating a unique table constraint).
func SQLiteUniqueViolation(err error) bool {
	match := strings.Contains(err.Error(), "UNIQUE constraint failed")
	if e, ok := err.(sqlite3.Error); ok {
		return match || e.Code == sqlite3.ErrConstraint
	}
	return match
}
