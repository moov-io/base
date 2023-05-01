package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/spanner"
	"github.com/golang-migrate/migrate/v4/database"
	migspanner "github.com/golang-migrate/migrate/v4/database/spanner"
	"github.com/googleapis/gax-go/v2/apierror"
	_ "github.com/googleapis/go-sql-spanner"
	"google.golang.org/grpc/codes"

	"github.com/moov-io/base/log"
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

// SpannerUniqueViolation returns true when the provided error matches the Spanner code
// for duplicate entries (violating a unique table constraint).
// Refer to https://cloud.google.com/spanner/docs/error-codes for Spanner error definitions,
// and https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto for error codes
func SpannerUniqueViolation(err error) bool {
	match := strings.Contains(err.Error(), "Failed to insert row with primary key")

	var apiErr *apierror.APIError
	if errors.As(err, &apiErr) {
		return match || spanner.ErrCode(apiErr) == codes.AlreadyExists
	}
	return match
}
