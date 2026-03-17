package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"cloud.google.com/go/alloydbconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	var connStr string
	if config.Alloy != nil {
		c, err := getAlloyDBConnectorConnStr(ctx, config, databaseName)
		if err != nil {
			return nil, logger.LogErrorf("creating alloydb connection: %w", err).Err()
		}
		connStr = c
	} else {
		c, err := getPostgresConnStr(config, databaseName)
		if err != nil {
			return nil, logger.LogErrorf("creating postgres connection: %w", err).Err()
		}
		connStr = c
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, logger.LogErrorf("opening database: %w", err).Err()
	}

	err = db.Ping()
	if err != nil {
		_ = db.Close()
		return nil, logger.LogErrorf("connecting to database: %w", err).Err()
	}

	return db, nil
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

func getAlloyDBConnectorConnStr(ctx context.Context, config PostgresConfig, databaseName string) (string, error) {
	if config.Alloy == nil {
		return "", fmt.Errorf("missing alloy config")
	}

	var dialer *alloydbconn.Dialer
	var dsn string

	if config.Alloy.UseIAM {
		d, err := alloydbconn.NewDialer(ctx, alloydbconn.WithIAMAuthN())
		if err != nil {
			return "", fmt.Errorf("creating alloydb dialer: %v", err)
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
			return "", fmt.Errorf("creating alloydb dialer: %v", err)
		}
		dialer = d
		dsn = fmt.Sprintf(
			// sslmode is disabled because the alloy db connection dialer will handle it
			"user=%s password=%s dbname=%s sslmode=disable",
			config.User, config.Password, databaseName,
		)
	}

	// TODO
	//cleanup := func() error { return d.Close() }

	connConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return "", fmt.Errorf("failed to parse pgx config: %v", err)
	}

	var connOptions []alloydbconn.DialOption
	if config.Alloy.UsePSC {
		connOptions = append(connOptions, alloydbconn.WithPSC())
	}

	connConfig.DialFunc = func(ctx context.Context, _ string, _ string) (net.Conn, error) {
		return dialer.Dial(ctx, config.Alloy.InstanceURI, connOptions...)
	}

	connStr := stdlib.RegisterConnConfig(connConfig)
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

	// PostgreSQL error codes indicating the server is shutting down or unavailable
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "57P01": // admin_shutdown
			return true
		case "57P02": // crash_shutdown
			return true
		case "57P03": // cannot_connect_now
			return true
		case "08000": // connection_exception
			return true
		case "08001": // sqlclient_unable_to_establish_sqlconnection
			return true
		case "08003": // connection_does_not_exist
			return true
		case "08004": // sqlserver_rejected_establishment_of_sqlconnection
			return true
		case "08006": // connection_failure
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

	// pgx wraps connection errors with these messages
	msg := err.Error()
	if strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "unexpected EOF") ||
		strings.Contains(msg, "conn closed") {
		return true
	}

	return false
}

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
			backoff := time.Duration(attempt+1) * 200 * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
	return err
}
