// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	migmysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/moov-io/base/log"
)

var migrationMutex sync.Mutex

func RunMigrations(logger log.Logger, config DatabaseConfig, opts ...MigrateOption) error {
	logger.Info().Log("Running Migrations")

	// apply all of our optional arguments
	o := &migrateOptions{}
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return err
		}
	}

	source, driver, err := getDriver(logger, config, o)
	if err != nil {
		return err
	}
	defer driver.Close()

	migrationMutex.Lock()
	m, err := migrate.NewWithInstance(
		source.name,
		source,
		config.DatabaseName,
		driver,
	)
	if err != nil {
		return logger.Fatal().LogErrorf("Error running migration: %w", err).Err()
	}

	if o.timeout != nil {
		m.LockTimeout = *o.timeout
	}

	currentVersion, dirty, err := m.Version()
	if err != nil {
		if err != migrate.ErrNilVersion {
			return logger.Fatal().LogErrorf("Error getting current DB version: %w", err).Err()
		}
		// set sane values
		currentVersion = 0
		dirty = false
	}
	err = m.Up()
	migrationMutex.Unlock()

	switch err {
	case nil:
	case migrate.ErrNoChange:
		logger.Info().Logf("Database already at version %d (dirty: %b)", currentVersion, dirty)
	default:
		return logger.Fatal().LogErrorf("Error running migrations (current: %d, dirty: %b): %w", currentVersion, dirty, err).Err()
	}

	newVersion, newDirty, err := m.Version()
	if err != nil {
		if err != migrate.ErrNilVersion {
			return logger.Fatal().LogErrorf("Error getting new DB version: %w", err).Err()
		}
		// set sane values
		newVersion = 0
		newDirty = false
	}
	logger.Info().Logf("Migrations complete: %d (%b) -> %d (%b)", currentVersion, dirty, newVersion, newDirty)

	return nil
}

// Deprecated: Here to not break compatibility since it was once public.
func GetDriver(logger log.Logger, config DatabaseConfig) (source.Driver, database.Driver, error) {
	return getDriver(logger, config, &migrateOptions{})
}

func getDriver(logger log.Logger, config DatabaseConfig, opts *migrateOptions) (*SourceDriver, database.Driver, error) {
	var err error

	if config.MySQL != nil {
		if opts.source == nil {
			src, err := NewPkgerSource("mysql", true)
			if err != nil {
				return nil, nil, err
			}
			opts.source = &SourceDriver{
				name:   "pkger-mysql",
				Driver: src,
			}
		}

		if opts.driver == nil {
			db, err := New(context.Background(), logger, config)
			if err != nil {
				return nil, nil, err
			}

			opts.driver, err = MySQLDriver(db)
			if err != nil {
				return nil, nil, err
			}
		}

	} else if config.Spanner != nil {
		if opts.source == nil {
			src, err := NewPkgerSource("spanner", false)
			if err != nil {
				return nil, nil, err
			}
			opts.source = &SourceDriver{
				name:   "pkger-spanner",
				Driver: src,
			}
		}

		if opts.driver == nil {
			opts.driver, err = SpannerDriver(config)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	if opts.source == nil || opts.driver == nil {
		return nil, nil, fmt.Errorf("database config not defined")
	}

	return opts.source, opts.driver, nil
}

func MySQLDriver(db *sql.DB) (database.Driver, error) {
	return migmysql.WithInstance(db, &migmysql.Config{})
}

func SpannerDriver(config DatabaseConfig) (database.Driver, error) {
	return SpannerMigrationDriver(*config.Spanner, config.DatabaseName)
}

type MigrateOption func(o *migrateOptions) error

type SourceDriver struct {
	name string
	source.Driver
}

type migrateOptions struct {
	source *SourceDriver
	driver database.Driver

	timeout *time.Duration
}

func WithEmbeddedMigrations(f fs.FS) MigrateOption {
	return func(o *migrateOptions) error {
		src, err := iofs.New(f, "migrations")
		if err != nil {
			return err
		}
		o.source = &SourceDriver{
			name:   "embedded",
			Driver: src,
		}
		return nil
	}
}

func WithTimeout(dur time.Duration) MigrateOption {
	return func(o *migrateOptions) error {
		o.timeout = &dur
		return nil
	}
}
