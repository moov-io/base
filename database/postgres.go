package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/moov-io/base/log"
)

const (
	// PostgreSQL Error Codes
	// https://www.postgresql.org/docs/current/errcodes-appendix.html
	postgresErrUniqueViolation = "23505"
)

func postgresConnection(logger log.Logger, config PostgresConfig, databaseName string) (*sql.DB, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s/%s", config.User, config.Password, config.Address, databaseName)

	params := ""

	if config.UseTLS {
		params += "sslmode=verify-full"

		if config.TLSCAFile == "" {
			return nil, fmt.Errorf("missing TLS CA file")
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
