// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

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
		preppedDb, err := mysqlConnection(logger, config.MySQL, config.DatabaseName)
		if err != nil {
			return nil, fmt.Errorf("configuring mysql connection: %v", err)
		}

		db, err := preppedDb.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("connecting to mysql: %w", err)
		}

		return ApplyConnectionsConfig(db, &config.MySQL.Connections, logger), nil

	} else if config.Spanner != nil {
		db, err := spannerConnection(logger, *config.Spanner, config.DatabaseName)
		if err != nil {
			return nil, fmt.Errorf("connecting to spanner: %w", err)
		}
		return db, nil
	} else if config.Postgres != nil {
		db, err := postgresConnection(ctx, logger, *config.Postgres, config.DatabaseName)
		if err != nil {
			return nil, fmt.Errorf("connecting to postgres: %w", err)
		}
		return ApplyConnectionsConfig(db, &config.Postgres.Connections, logger), nil
	}

	return nil, fmt.Errorf("database config not defined")
}

func NewAndMigrate(ctx context.Context, logger log.Logger, config DatabaseConfig, opts ...MigrateOption) (*sql.DB, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// run migrations first
	if err := RunMigrations(logger, config, opts...); err != nil {
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
	return MySQLUniqueViolation(err) || SpannerUniqueViolation(err) || PostgresUniqueViolation(err)
}

func DataTooLong(err error) bool {
	return MySQLDataTooLong(err)
}

func ApplyConnectionsConfig(db *sql.DB, connections *ConnectionsConfig, logger log.Logger) *sql.DB {
	if connections.MaxOpen > 0 {
		logger.Logf("setting SQL max open connections to %d", connections.MaxOpen)
		db.SetMaxOpenConns(connections.MaxOpen)
	}

	if connections.MaxIdle > 0 {
		logger.Logf("setting SQL max idle connections to %d", connections.MaxIdle)
		db.SetMaxIdleConns(connections.MaxIdle)
	}

	// Due to a known issue https://github.com/golang/go/issues/45993#issuecomment-1427873850,
	// maxIdleTime must be specified before MaxLifetime or else it will not be honored.
	if connections.MaxIdleTime > 0 {
		logger.Logf("setting SQL max idle time to %v", connections.MaxIdleTime)
		db.SetConnMaxIdleTime(connections.MaxIdleTime)
	}

	if connections.MaxLifetime > 0 {
		logger.Logf("setting SQL max lifetime to %v", connections.MaxLifetime)
		db.SetConnMaxLifetime(connections.MaxLifetime)
	}

	return db
}
