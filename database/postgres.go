package database

import (
	"cloud.google.com/go/alloydbconn"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/moov-io/base/log"
	"net"
)

const (
	// PostgreSQL Error Codes
	// https://www.postgresql.org/docs/current/errcodes-appendix.html
	postgresErrUniqueViolation = "23505"
)

func postgresConnection(ctx context.Context, logger log.Logger, config PostgresConfig, databaseName string) (*sql.DB, error) {
	var connStr string
	if config.UseAlloyDBConnector {
		c, err := getAlloyDBConnectorConnStr(ctx, config, databaseName)
		if err != nil {
			return nil, fmt.Errorf("creating alloydb connection: %w", err)
		}
		connStr = c
	} else {
		c, err := getPostgresConnStr(config, databaseName)
		if err != nil {
			return nil, fmt.Errorf("creating postgres connection: %w", err)
		}
		connStr = c
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	return db, nil
}

func getPostgresConnStr(config PostgresConfig, databaseName string) (string, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s/%s", config.User, config.Password, config.Address, databaseName)

	params := ""

	if config.UseTLS {
		params += "sslmode=verify-full"

		if config.TLSCAFile == "" {
			return "", fmt.Errorf("missing TLS CA file")
		}
		params += "&sslrootcert=" + config.TLSCAFile

		if config.TLSClientCertFile != "" {
			params += "&sslcert=" + config.TLSClientCertFile
		}

		if config.TLSClientKeyFile != "" {
			params += "&sslkey=" + config.TLSClientKeyFile
		}
	}

	connStr := fmt.Sprintf("%s?%s", url, params)
	return connStr, nil
}

func getAlloyDBConnectorConnStr(ctx context.Context, config PostgresConfig, databaseName string) (string, error) {
	var dialer *alloydbconn.Dialer
	var dsn string

	if config.UseAlloyDBIAM {
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

	connConfig.DialFunc = func(ctx context.Context, _ string, _ string) (net.Conn, error) {
		return dialer.Dial(ctx, config.AlloyDBInstanceURI)
	}

	connStr := stdlib.RegisterConnConfig(connConfig)
	return connStr, nil
}

func PostgresUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		if pgError.Code == postgresErrUniqueViolation {
			return true
		}
	}
	return false
}
