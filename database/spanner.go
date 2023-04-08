package database

import (
	"database/sql"
	"fmt"

	"github.com/moov-io/base/log"

	"github.com/golang-migrate/migrate/v4/database"
	migspanner "github.com/golang-migrate/migrate/v4/database/spanner"
	_ "github.com/googleapis/go-sql-spanner"
)

func spannerConnection(logger log.Logger, cfg SpannerConfig, databaseName string) (*sql.DB, error) {
	db, err := sql.Open("spanner", fmt.Sprintf("projects/%s/instances/%s/databases/%s", cfg.Project, cfg.Instance, databaseName))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func SpannerMigrationDriver(cfg SpannerConfig, databaseName string) (database.Driver, error) {
	s := migspanner.Spanner{}
	return s.Open(fmt.Sprintf("spanner://projects/%s/instances/%s/databases/%s?x-migrations-table=spanner_schema_migrations&x-clean-statements=true", cfg.Project, cfg.Instance, databaseName))
}
