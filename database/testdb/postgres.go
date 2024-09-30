package testdb

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/moov-io/base/database"
)

func NewPostgresDatabase(t *testing.T, cfg database.DatabaseConfig) error {
	t.Helper()
	if cfg.Postgres == nil {
		return fmt.Errorf("postgres config not defined")
	}

	db, err := sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s", cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Address))
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DatabaseName))
	if err != nil {
		return err
	}

	t.Cleanup(func() {
		db.Exec(fmt.Sprintf("DROP DATABASE %s", cfg.DatabaseName))
		db.Close()
	})

	return nil
}
