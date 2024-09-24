package database

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/moov-io/base/log"
)

func postgresConnection(logger log.Logger, config PostgresConfig, databaseName string) (*sql.DB, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s/%s", config.User, config.Password, config.Address, databaseName)

	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("connecting to alloydb: %w", err)
	}

	return db, nil
}
