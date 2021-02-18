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
		db, err := mysqlConnection(logger, config.MySQL.User, config.MySQL.Password, config.MySQL.Address, config.DatabaseName).Connect(ctx)
		if err != nil {
			return nil, err
		}

		return ApplyConnectionsConfig(db, &config.MySQL.Connections), nil
	} else if config.SQLite != nil {
		return sqliteConnection(logger, config.SQLite.Path).Connect(ctx)
	}

	return nil, fmt.Errorf("database config not defined")
}

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
	return MySQLUniqueViolation(err) || SQLiteUniqueViolation(err)
}

func ApplyConnectionsConfig(db *sql.DB, connections *ConnectionsConfig) *sql.DB {
	if connections.MaxOpen > 0 {
		db.SetMaxOpenConns(connections.MaxOpen)
	}

	if connections.MaxIdle > 0 {
		db.SetMaxOpenConns(connections.MaxIdle)
	}

	if connections.MaxLifetime > 0 {
		db.SetConnMaxLifetime(connections.MaxLifetime)
	}

	if connections.MaxIdleTime > 0 {
		db.SetConnMaxIdleTime(connections.MaxIdleTime)
	}

	return db
}
