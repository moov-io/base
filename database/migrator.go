// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	migmysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source"

	"github.com/moov-io/base/log"
)

var migrationMutex sync.Mutex

func RunMigrations(logger log.Logger, config DatabaseConfig) error {
	logger.Info().Log("Running Migrations")

	source, driver, err := GetDriver(logger, config)
	if err != nil {
		return err
	}

	defer driver.Close()

	migrationMutex.Lock()
	m, err := migrate.NewWithInstance(
		"filtering-pkger",
		source,
		config.DatabaseName,
		driver,
	)
	if err != nil {
		return logger.Fatal().LogErrorf("Error running migration: %w", err).Err()
	}

	err = m.Up()
	migrationMutex.Unlock()

	switch err {
	case nil:
	case migrate.ErrNoChange:
		logger.Info().Log("Database already at version")
	default:
		return logger.Fatal().LogErrorf("Error running migrations: %w", err).Err()
	}

	logger.Info().Log("Migrations complete")

	return nil
}

func GetDriver(logger log.Logger, config DatabaseConfig) (source.Driver, database.Driver, error) {
	if config.MySQL != nil {
		src, err := NewPkgerSource("mysql", true)
		if err != nil {
			return nil, nil, err
		}

		db, err := New(context.Background(), logger, config)
		if err != nil {
			return nil, nil, err
		}
		defer db.Close()

		drv, err := MySQLDriver(db)
		if err != nil {
			return nil, nil, err
		}

		return src, drv, nil

	} else if config.Spanner != nil {
		src, err := NewPkgerSource("spanner", false)
		if err != nil {
			return nil, nil, err
		}

		drv, err := SpannerDriver(config)
		if err != nil {
			return nil, nil, err
		}

		return src, drv, nil
	}

	return nil, nil, fmt.Errorf("database config not defined")
}

func MySQLDriver(db *sql.DB) (database.Driver, error) {
	return migmysql.WithInstance(db, &migmysql.Config{})
}

func SpannerDriver(config DatabaseConfig) (database.Driver, error) {
	return SpannerMigrationDriver(*config.Spanner, config.DatabaseName)
}
