package testdb

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/moov-io/base/database"
)

func NewMySQLDatabase(t *testing.T, cfg database.DatabaseConfig) error {
	t.Helper()
	if cfg.MySQL == nil {
		return fmt.Errorf("mysql config not defined")
	}

	rootDb, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s/", cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Address))
	if err != nil {
		return err
	}

	if err := rootDb.Ping(); err != nil {
		return err
	}

	_, err = rootDb.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DatabaseName))
	if err != nil {
		return err
	}

	t.Cleanup(func() {
		rootDb.Exec(fmt.Sprintf("DROP DATABASE %s", cfg.DatabaseName))
		rootDb.Close()
	})

	return nil
}
