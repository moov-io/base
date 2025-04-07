package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"

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
		params += "sslmode=verify-full"

		if config.TLS.CACertFile == "" {
			return "", fmt.Errorf("missing TLS CA file")
		}
		params += "&sslrootcert=" + config.TLS.CACertFile

		if config.TLS.ClientCertFile != "" {
			params += "&sslcert=" + config.TLS.ClientCertFile
		}

		if config.TLS.ClientKeyFile != "" {
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
