package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"time"

	"cloud.google.com/go/alloydbconn"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/moov-io/base/log"
)

const (
	// PostgreSQL Error Codes
	// https://www.postgresql.org/docs/current/errcodes-appendix.html
	postgresErrUniqueViolation = "23505"
	postgresErrDeadlockFound   = "40P01"
)

func postgresConnection(ctx context.Context, logger log.Logger, config PostgresConfig, databaseName string) (*sql.DB, error) {
	poolConfig, err := buildPgxPoolConfig(ctx, config, databaseName)
	if err != nil {
		return nil, logger.LogErrorf("building pgx pool config: %w", err).Err()
	}

	// HealthCheckPeriod makes pgxpool ping idle connections in the background.
	// Dead connections (e.g. from an AlloyDB switchover) are evicted before
	// the application ever sees them.
	poolConfig.HealthCheckPeriod = 1 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, logger.LogErrorf("creating pgx pool: %w", err).Err()
	}

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, logger.LogErrorf("connecting to database: %w", err).Err()
	}

	// Wrap the pgxpool in a *sql.DB so the rest of the codebase doesn't change.
	// pgxpool manages the real pool (with health checks); database/sql pool
	// settings are applied on top via ApplyPostgresConnectionsConfig.
	db := stdlib.OpenDBFromPool(pool)

	return db, nil
}

func buildPgxPoolConfig(ctx context.Context, config PostgresConfig, databaseName string) (*pgxpool.Config, error) {
	if config.Alloy != nil {
		return buildAlloyDBPoolConfig(ctx, config, databaseName)
	}

	connStr, err := getPostgresConnStr(config, databaseName)
	if err != nil {
		return nil, err
	}
	return pgxpool.ParseConfig(connStr)
}

func buildAlloyDBPoolConfig(ctx context.Context, config PostgresConfig, databaseName string) (*pgxpool.Config, error) {
	if config.Alloy == nil {
		return nil, fmt.Errorf("missing alloy config")
	}

	var dialer *alloydbconn.Dialer
	var dsn string

	if config.Alloy.UseIAM {
		d, err := alloydbconn.NewDialer(ctx, alloydbconn.WithIAMAuthN())
		if err != nil {
			return nil, fmt.Errorf("creating alloydb dialer: %v", err)
		}
		dialer = d
		dsn = fmt.Sprintf(
			// sslmode is disabled because the alloy db connection dialer will handle it
			// no password is used with IAM
			"user=%s dbname=%s sslmode=disable",
			config.User, databaseName,
		)
	} else {
		d, err := alloydbconn.NewDialer(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating alloydb dialer: %v", err)
		}
		dialer = d
		dsn = fmt.Sprintf(
			// sslmode is disabled because the alloy db connection dialer will handle it
			"user=%s password=%s dbname=%s sslmode=disable",
			config.User, config.Password, databaseName,
		)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx pool config: %v", err)
	}

	var connOptions []alloydbconn.DialOption
	if config.Alloy.UsePSC {
		connOptions = append(connOptions, alloydbconn.WithPSC())
	}

	poolConfig.ConnConfig.DialFunc = func(ctx context.Context, _ string, _ string) (net.Conn, error) {
		return dialer.Dial(ctx, config.Alloy.InstanceURI, connOptions...)
	}

	return poolConfig, nil
}

func getPostgresConnStr(config PostgresConfig, databaseName string) (string, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s/%s", config.User, config.Password, config.Address, databaseName)

	params := ""

	if config.TLS != nil {
		if len(config.TLS.Mode) < 1 {
			config.TLS.Mode = "verify-full"
		}

		params += "sslmode=" + config.TLS.Mode

		if len(config.TLS.CACertFile) > 0 {
			params += "&sslrootcert=" + config.TLS.CACertFile
		}

		if len(config.TLS.ClientCertFile) > 0 {
			params += "&sslcert=" + config.TLS.ClientCertFile
		}

		if len(config.TLS.ClientKeyFile) > 0 {
			params += "&sslkey=" + config.TLS.ClientKeyFile
		}
	}

	connStr := fmt.Sprintf("%s?%s", url, params)
	return connStr, nil
}

// PostgresUniqueViolation returns true when the provided error matches the Postgres code
// for unique violation.
func PostgresUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == postgresErrUniqueViolation {
		return true
	}

	return strings.Contains(err.Error(), postgresErrUniqueViolation)
}

// PostgresDeadlockFound returns true when the provided error matches the Postgres code
// for deadlock found.
func PostgresDeadlockFound(err error) bool {
	if err == nil {
		return false
	}

	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == postgresErrDeadlockFound {
		return true
	}

	return strings.Contains(err.Error(), postgresErrDeadlockFound)
}

// IsRetryablePostgresError returns true if the error is a transient connection-level
// error that is safe to retry. This covers the errors seen during AlloyDB maintenance
// switchovers and other transient network failures.
func IsRetryablePostgresError(err error) bool {
	if err == nil {
		return false
	}

	// PostgreSQL error codes that are safe to retry because the server guarantees
	// the transaction was rolled back before the error was returned.
	//
	// 57P01 admin_shutdown, 57P02 crash_shutdown: fast shutdown rolls back all
	// active transactions before disconnecting clients. See:
	// https://www.postgresql.org/docs/current/server-shutdown.html
	//
	// 57P03 cannot_connect_now: server is still starting up; the operation never
	// reached a transaction.
	//
	// 40001 serialization_failure, 40P01 deadlock_detected: the server explicitly
	// rolled back the transaction and expects the client to retry. See:
	// https://www.postgresql.org/docs/current/mvcc-serialization-failure-handling.html
	//
	// 53300 too_many_connections: rejected at connect time; no transaction started.
	//
	// 57014 query_canceled: the server rolled back the transaction before returning
	// this error.
	//
	// Note: 08xxx (connection_exception class) codes are intentionally omitted.
	// pgx surfaces connection-level failures as TCP/network errors, not as
	// *pgconn.PgError, so those cases are handled below.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "57P01", "57P02", "57P03", // admin_shutdown, crash_shutdown, cannot_connect_now
			"40001", "40P01", // serialization_failure, deadlock_detected
			"53300",          // too_many_connections
			"57014":          // query_canceled
			return true
		}
		return false
	}

	// Network-level errors: connection reset, broken pipe, EOF, etc.
	// These occur when the TCP connection is severed during a switchover.
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return false // don't retry if the caller's context timed out
	}

	// pgx sometimes wraps TCP-level disconnection errors as plain strings rather
	// than preserving the original net.OpError type. These are safe to retry
	// because they occur on the write path — the server never received the query.
	// "conn closed" is pgx's own sentinel for a connection already closed before use.
	// Note: "connection refused" and "unexpected EOF" are intentionally excluded —
	// the former does not appear in pgx's codebase, the latter is already caught
	// above by errors.Is(err, io.ErrUnexpectedEOF).
	msg := err.Error()
	if strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "conn closed") {
		return true
	}

	return false
}

// retryJitterMax is the upper bound for the random delay between retries.
// Full jitter in [0, retryJitterMax) spreads concurrent retries across the
// fleet rather than letting them all slam the database at the same instant.
// AlloyDB planned maintenance switchovers typically cause less than one second
// of downtime, so three attempts with up to 100ms between each gives a worst-
// case retry window of ~200ms — well within typical service SLAs while still
// spanning the switchover blip. The caller's context deadline is the escape
// hatch for services with tighter latency budgets.
const retryJitterMax = 100 * time.Millisecond

// RetryPostgres executes fn up to maxAttempts times, retrying on transient
// connection errors. This is intended for use around individual database
// operations to survive brief outages like AlloyDB maintenance switchovers.
func RetryPostgres(ctx context.Context, maxAttempts int, fn func() error) error {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		if !IsRetryablePostgresError(err) {
			return err
		}
		if attempt < maxAttempts-1 {
			delay := time.Duration(rand.Int63n(int64(retryJitterMax)))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return err
}
