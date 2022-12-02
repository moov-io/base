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
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	kitprom "github.com/go-kit/kit/metrics/prometheus"
	gomysql "github.com/go-sql-driver/mysql"
	stdprom "github.com/prometheus/client_golang/prometheus"

	"github.com/moov-io/base/log"
)

var (
	metricsMu = &sync.Mutex{}

	mysqlConnections = kitprom.NewGaugeFrom(stdprom.GaugeOpts{
		Name: "mysql_connections",
		Help: "How many MySQL connections and what status they're in.",
	}, []string{"state"})

	mysqlConnectionsCounters = kitprom.NewGaugeFrom(stdprom.GaugeOpts{
		Name: "mysql_connections_counters",
		Help: `Counters specific to the sql connections.
			wait_count: The total number of connections waited for.
			wait_duration: The total time blocked waiting for a new connection.
			max_idle_closed: The total number of connections closed due to SetMaxIdleConns.
			max_idle_time_closed: The total number of connections closed due to SetConnMaxIdleTime.
			max_lifetime_closed: The total number of connections closed due to SetConnMaxLifetime.
		`,
	}, []string{"counter"})

	// mySQLErrDuplicateKey is the error code for duplicate entries
	// https://dev.mysql.com/doc/mysql-errors/8.0/en/server-error-reference.html#error_er_dup_entry
	mySQLErrDuplicateKey uint16 = 1062
	mysqlErrDataTooLong  uint16 = 1406

	maxActiveMySQLConnections = func() int {
		if v := os.Getenv("MYSQL_MAX_CONNECTIONS"); v != "" {
			if n, _ := strconv.ParseInt(v, 10, 32); n > 0 {
				return int(n)
			}
		}
		return 16
	}()
)

func RecordMySQLStats(db *sql.DB) error {
	stats := db.Stats()

	metricsMu.Lock()
	defer metricsMu.Unlock()

	mysqlConnections.With("state", "idle").Set(float64(stats.Idle))
	mysqlConnections.With("state", "inuse").Set(float64(stats.InUse))
	mysqlConnections.With("state", "open").Set(float64(stats.OpenConnections))

	mysqlConnectionsCounters.With("counter", "wait_count").Set(float64(stats.WaitCount))
	mysqlConnectionsCounters.With("counter", "wait_ms").Set(float64(stats.WaitDuration.Milliseconds()))
	mysqlConnectionsCounters.With("counter", "max_idle_closed").Set(float64(stats.MaxIdleClosed))
	mysqlConnectionsCounters.With("counter", "max_idle_time_closed").Set(float64(stats.MaxIdleTimeClosed))
	mysqlConnectionsCounters.With("counter", "max_lifetime_closed").Set(float64(stats.MaxLifetimeClosed))

	return nil
}

type discardLogger struct{}

func (l discardLogger) Print(v ...interface{}) {}

func init() {
	gomysql.SetLogger(discardLogger{})
}

type mysql struct {
	dsn    string
	logger log.Logger
	tls    *tls.Config
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
	go func(db *sql.DB) {
		t := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				RecordMySQLStats(db)
			}
		}
	}(db)

	return db, nil
}

func mysqlConnection(logger log.Logger, mysqlConfig *MySQLConfig, databaseName string) (*mysql, error) {
	timeout := "30s"
	if v := os.Getenv("MYSQL_TIMEOUT"); v != "" {
		timeout = v
	}
	params := fmt.Sprintf(`timeout=%s&charset=utf8mb4&parseTime=true&sql_mode="ALLOW_INVALID_DATES,STRICT_ALL_TABLES"&multiStatements=true`, timeout)

	var tlsConfig *tls.Config

	if mysqlConfig.UseTLS {
		logger.Log("using TLS for MySQL connection")
		// If any custom options are set then we need to create a custom TLS configuration. Otherwise we can just set
		// tls=true in the DSN
		if mysqlConfig.InsecureSkipVerify || mysqlConfig.TLSCAFile != "" || mysqlConfig.VerifyCAFile {
			logger.Log("creating custom TLS configuration for MySQL connection")
			tlsConfig = &tls.Config{
				InsecureSkipVerify: mysqlConfig.InsecureSkipVerify, //nolint:gosec
			}

			if mysqlConfig.TLSCAFile != "" {
				logger.Logf("reading and adding MySQL CA file from %s", mysqlConfig.TLSCAFile)
				rootCertPool := x509.NewCertPool()
				certPem, err := os.ReadFile(mysqlConfig.TLSCAFile)
				if err != nil {
					return nil, err
				}

				block, _ := pem.Decode(certPem)
				_, err = x509.ParseCertificate(block.Bytes)
				if err != nil {
					return nil, err
				}

				if appendOK := rootCertPool.AppendCertsFromPEM(certPem); !appendOK {
					return nil, errors.New("failed to append certificate PEM to root cert pool")
				}
				tlsConfig.RootCAs = rootCertPool

				clientCerts, err := LoadTLSClientCertsFromConfig(logger, mysqlConfig)
				if err != nil {
					return nil, errors.New("failed to load client certificate(s)")
				}

				if mysqlConfig.VerifyCAFile {
					tlsConfig.VerifyConnection = func(state tls.ConnectionState) error {
						logger.Logf("verifying MySQL server certificate using CA from file %s", mysqlConfig.TLSCAFile)
						_, err := state.PeerCertificates[0].Verify(x509.VerifyOptions{Roots: rootCertPool})
						if err != nil {
							return logger.Error().LogError(err).Err()
						}
						return nil
					}
				}

				tlsConfig.Certificates = clientCerts
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
		dsn:    dsn,
		logger: logger,
		tls:    tlsConfig,
	}, nil
}

// MySQLUniqueViolation returns true when the provided error matches the MySQL code
// for duplicate entries (violating a unique table constraint).
func MySQLUniqueViolation(err error) bool {
	match := strings.Contains(err.Error(), fmt.Sprintf("Error %d", mySQLErrDuplicateKey))
	if e, ok := err.(*gomysql.MySQLError); ok {
		return match || e.Number == mySQLErrDuplicateKey
	}
	return match
}

// MySQLDataTooLong returns true when the provided error matches the MySQL code
// for data too long for column (when trying to insert a value that is greater than
// the defined max size of the column).
func MySQLDataTooLong(err error) bool {
	match := strings.Contains(err.Error(), fmt.Sprintf("Error %d", mysqlErrDataTooLong))
	if e, ok := err.(*gomysql.MySQLError); ok {
		return match || e.Number == mysqlErrDataTooLong
	}
	return match
}
