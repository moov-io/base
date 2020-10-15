package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	migmysql "github.com/golang-migrate/migrate/v4/database/mysql"
	migsqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/pkger"
	"github.com/markbates/pkger"

	"github.com/moov-io/base/log"
)

func RunMigrations(logger log.Logger, db *sql.DB, config DatabaseConfig) error {
	logger.Info().Log("Running Migrations")

	_ = pkger.Include("/migrations/")

	driver, err := GetDriver(db, config)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"pkger:///migrations/",
		config.DatabaseName,
		driver,
	)
	if err != nil {
		return logger.Fatal().LogErrorf("Error running migration: %w", err).Err()
	}

	err = m.Up()
	switch err {
	case nil:
	case migrate.ErrNoChange:
		logger.Info().Logf("Database already at version")
	default:
		return logger.Fatal().LogErrorf("Error running migrations: %w", err).Err()
	}

	logger.Info().Logf("Migrations complete")

	return nil
}

func GetDriver(db *sql.DB, config DatabaseConfig) (database.Driver, error) {
	if config.MySql != nil {
		return MySqlDriver(db)
	} else if config.SQLite != nil {
		return Sqlite3Driver(db)
	}

	return nil, fmt.Errorf("database config not defined")
}

func MySqlDriver(db *sql.DB) (database.Driver, error) {
	return migmysql.WithInstance(db, &migmysql.Config{})
}

func Sqlite3Driver(db *sql.DB) (database.Driver, error) {
	return migsqlite3.WithInstance(db, &migsqlite3.Config{})
}
