package database

import (
	"context"
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

func RunMigrations(logger log.Logger, config Config) error {
	db, err := New(context.Background(), logger, config)
	if err != nil {
		return err
	}
	defer db.Close()

	logger.Info().Log("Running Migrations")

	_ = pkger.Include("/migrations/")

	driver, err := GetDriver(db, config)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"pkger:///migrations/",
		string(config.Type),
		driver,
	)
	if err != nil {
		return logger.Fatal().LogErrorf("Error running migration: %w", err).Err()
	}

	err = m.Up()
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

func GetDriver(db *sql.DB, config Config) (database.Driver, error) {
	switch config.Type {
	case TypeMySQL:
		return migmysql.WithInstance(db, &migmysql.Config{})
	case TypeSQLite:
		return migsqlite3.WithInstance(db, &migsqlite3.Config{})
	}

	return nil, fmt.Errorf("database config not defined")
}
