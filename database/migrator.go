package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	migmysql "github.com/golang-migrate/migrate/v4/database/mysql"
	migsqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source"

	"github.com/moov-io/base/log"
)

var migrationMutex sync.Mutex

func RunMigrations(logger log.Logger, config DatabaseConfig) error {
	db, err := New(context.Background(), logger, config)
	if err != nil {
		return err
	}
	defer db.Close()

	logger.Info().Log("Running Migrations")

	source, driver, err := GetDriver(db, config)
	if err != nil {
		return err
	}

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

func GetDriver(db *sql.DB, config DatabaseConfig) (source.Driver, database.Driver, error) {
	if config.MySQL != nil {
		src, err := NewPkgerSource("mysql")
		if err != nil {
			return nil, nil, err
		}

		drv, err := MySQLDriver(db)
		if err != nil {
			return nil, nil, err
		}

		return src, drv, nil
	} else if config.SQLite != nil {
		src, err := NewPkgerSource("sqlite")
		if err != nil {
			return nil, nil, err
		}

		drv, err := SQLite3Driver(db)
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

func SQLite3Driver(db *sql.DB) (database.Driver, error) {
	return migsqlite3.WithInstance(db, &migsqlite3.Config{})
}
