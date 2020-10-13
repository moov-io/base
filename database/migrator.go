package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	migmysql "github.com/golang-migrate/migrate/v4/database/mysql"
	migsqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/pkger"
	"github.com/markbates/pkger"

	"github.com/moov-io/base/log"
)

func RunMigrations(log log.Logger, db *sql.DB, config DatabaseConfig) error {
	if _, err := os.Stat(config.migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory=\"%s\" does not exist", config.migrationsDir)
	}

	log.Info().Logf("Running Migrations")
	_ = pkger.Include(config.migrationsDir)

	driver, err := GetDriver(db, config)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("pkger://%s", config.migrationsDir),
		config.DatabaseName,
		driver,
	)
	if err != nil {
		return log.Fatal().LogErrorf("Error running migration - %w", err)
	}

	err = m.Up()
	switch err {
	case nil:
	case migrate.ErrNoChange:
		log.Info().Logf("Database already at version")
	default:
		return log.Fatal().LogErrorf("Error running migrations - %w", err)
	}

	log.Info().Logf("Migrations complete")

	return nil
}

func GetDriver(db *sql.DB, config DatabaseConfig) (database.Driver, error) {
	if config.MySql != nil {
		return MySqlDriver(db)
	} else if config.SqlLite != nil {
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
