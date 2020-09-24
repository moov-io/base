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
	} else if config.SqlLite != nil {
		return sqliteConnection(logger, config.SqlLite.Path).Connect(ctx)
	}

	return nil, fmt.Errorf("database config not defined")
}

func NewAndMigrate(config DatabaseConfig, logger log.Logger, ctx context.Context) (*sql.DB, func(), error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	if ctx == nil {
		ctx = context.Background()
	}

	db, err := New(ctx, logger, config)
	if err != nil {
		return nil, func() {}, err
	}

	shutdown := func() {
		db.Close()
	}

	if config.migrationsDir != "" {
		if err = RunMigrations(logger, db, config); err != nil {
			return nil, shutdown, err
		}
	}

	return db, shutdown, nil
}

// UniqueViolation returns true when the provided error matches a database error
// for duplicate entries (violating a unique table constraint).
func UniqueViolation(err error) bool {
	return MySQLUniqueViolation(err) || SqliteUniqueViolation(err)
}
