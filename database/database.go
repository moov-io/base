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
	if config.MySql != nil {
		return mysqlConnection(logger, config.MySql.User, config.MySql.Password, config.MySql.Address, config.DatabaseName).Connect(ctx)
	} else if config.SQLite != nil {
		return sqliteConnection(logger, config.SQLite.Path).Connect(ctx)
	}

	return nil, fmt.Errorf("database config not defined")
}

// TODO remove shutdown method to make it consistent with New(...) *sql.DB, error
func NewAndMigrate(ctx context.Context, logger log.Logger, config DatabaseConfig) (*sql.DB, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// run migrations first
	if err := RunMigrations(logger, config); err != nil {
		return nil, err
	}

	// create DB connection for our service
	db, err := New(ctx, logger, config)
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
