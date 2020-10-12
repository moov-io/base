// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/moov-io/base/docker"
	"github.com/moov-io/base/log"
	"github.com/ory/dockertest"
)

var (
	mysqlInstance *MySQLTestDB = &MySQLTestDB{}
)

func init() {
	// after the above mysqlInstance is initialized with a zero-value
	// write-lock our mysqlInstance so tests trying to read (RLock()
	// are prevented until the underlying container is started.
	mysqlInstance.mu.Lock()

	go func() {
		// ugly hack to stall .Wait() and allow tests to startup
		time.Sleep(1 * time.Second)

		// Block until all tests are finished
		mysqlInstance.waiters.Wait()

		// Shutdown our root connection and container
		mysqlInstance.conn.Close()
		mysqlInstance.container.Close()

		// // Verify all connections are closed before closing DB
		// if conns := r.DB.Stats().OpenConnections; conns != 0 {
		// 	panic(fmt.Sprintf("found %d open MySQL connections", conns))
		// }
	}()
}

type MySQLTestDB struct {
	conn      *sql.DB
	container *dockertest.Resource
	mu        sync.RWMutex   // blocks conn and initial startup
	waiters   sync.WaitGroup // used to count each running test and wait for shutdown
	setup     sync.Once
	shutdown  func() // context shutdown func
}

func TestMySQLConnection(t *testing.T) *sql.DB {
	mysqlInstance.setup(t) // call into setup code, let the winner start up our MySQL container

	// block until our test is ready to read
	mysqlInstance.mu.RLock()
	defer mysqlInstance.mu.RLock()

	// create our connection, register a cleanup hook, and return
	conn := mysqlInstance.createConnection(t)
	t.Cleanup(func() {
		conn.Close()
		mysqlInstance.waiters.Done()
	})
	return conn
}

func (db *MySQLTestDB) setup(t *testing.T) {
	// Don't start MySQL containers in -short mode
	if testing.Short() {
		return
	}

	// Skip this setup if Docker isn't enabled
	if !docker.Enabled() {
		t.Skip("Docker not enabled")
	}

	db.waiters.Add(1)
	db.setup.Do(func() {
		// init container
		// mysqlInstance.Unlock()
	})
}

func (db *MySQLTestDB) createConnection(t *testing.T) *sql.DB {
	// setup a new database from t.Name()
	return nil
}

type mysqlInstance struct {
	conn      *sql.DB
	container *dockertest.Resource
	shutdown  context.CancelFunc
}

// createMySQLContainer returns a mysqlInstance that can be used across tests
// by creating unique databases in each. All migrations are ran on the database
// before handing them off to test instances.
func createMySQLContainer(t *testing.T) *mysqlInstance {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal(err)
	}
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "8",
		Env: []string{
			"MYSQL_USER=moov",
			"MYSQL_PASSWORD=secret",
			"MYSQL_ROOT_PASSWORD=secret",
			"MYSQL_DATABASE=paygate",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = pool.Retry(func() error {
		db, err := sql.Open("mysql", fmt.Sprintf("moov:secret@tcp(localhost:%s)/paygate", resource.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		defer db.Close()
		return db.Ping()
	})
	if err != nil {
		resource.Close()
		t.Fatal(err)
	}

	logger := log.NewNopLogger()
	address := fmt.Sprintf("tcp(localhost:%s)", resource.GetPort("3306/tcp"))

	ctx, cancelFunc := context.WithCancel(context.Background())

	db, err := mysqlConnection(logger, "moov", "secret", address, "paygate").Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Don't allow idle connections so we can verify all are closed at the end of testing
	db.SetMaxIdleConns(0)

	return &TestMySQLDB{DB: db, container: resource, shutdown: cancelFunc}
}
