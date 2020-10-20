package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/moov-io/base/log"
)

// New establishes a database connection according to the type and environmental
// variables for that specific database.
func New(ctx context.Context, logger log.Logger, config DatabaseConfig) (*sql.DB, error) {
	if config.MySQL != nil {
		return mysqlConnection(logger, config.MySQL.User, config.MySQL.Password, config.MySQL.Address, config.DatabaseName).Connect(ctx)
	} else if config.SQLite != nil {
		return sqliteConnection(logger, config.SQLite.Path).Connect(ctx)
	}

	return nil, fmt.Errorf("database config not defined")
}

func NewAndMigrate(ctx context.Context, logger log.Logger, config DatabaseConfig) (*sql.DB, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	// run migrations first
	db, err := RunMigrations(logger, config)
	if err != nil {
		return nil, err
	}

	// In SQLite, we can reuse the same connection
	if config.SQLite != nil {
		return db, nil
	}

	// create DB connection for our service
	db, err = New(ctx, logger, config)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// UniqueViolation returns true when the provided error matches a database error
// for duplicate entries (violating a unique table constraint).
func UniqueViolation(err error) bool {
	return MySQLUniqueViolation(err) || SqliteUniqueViolation(err)
}
