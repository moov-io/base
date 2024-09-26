package database

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/moov-io/base/log"
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
		return nil, fmt.Errorf("connecting to alloydb: %w", err)
	}

	return db, nil
}
