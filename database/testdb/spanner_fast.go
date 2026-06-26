package testdb

import (
	"context"
	"fmt"
	"io/fs"
	"testing"

	"cloud.google.com/go/spanner"
	spannerdb "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"cloud.google.com/go/spanner/spansql"
	"github.com/moov-io/base/database"
	"github.com/moov-io/base/log"
)

// NewSpannerDatabaseFromMigrations creates a Spanner test database with all
// migrations applied in a single batched DDL update, instead of running each
// migration file individually through golang-migrate. The
// spanner_schema_migrations table is pre-populated with the latest version so
// subsequent RunMigrations calls are no-ops (ErrNoChange).
//
// If the batched DDL fails or cannot be parsed, the function falls back to
// creating a fresh empty database and lets the caller run migrations normally.
func NewSpannerDatabaseFromMigrations(
	tb testing.TB,
	logger log.Logger,
	cfg database.DatabaseConfig,
	migrations fs.FS,
) (database.DatabaseConfig, func()) {
	tb.Helper()
	cfg2, dropFn, err := CreateSpannerDatabaseFromMigrations(logger, cfg, migrations)
	if err != nil {
		tb.Fatal(err)
	}
	return cfg2, dropFn
}

// CreateSpannerDatabaseFromMigrations is the error-returning variant of
// NewSpannerDatabaseFromMigrations. It is safe to call from contexts that
// cannot use tb.Fatal (e.g. inside sync.Once.Do).
func CreateSpannerDatabaseFromMigrations(
	logger log.Logger,
	cfg database.DatabaseConfig,
	migrations fs.FS,
) (database.DatabaseConfig, func(), error) {
	if migrations == nil {
		return cfg, nil, fmt.Errorf("migrations FS is nil")
	}

	cfg, err := NewSpannerDatabase(cfg.DatabaseName, cfg.Spanner)
	if err != nil {
		return cfg, nil, fmt.Errorf("creating spanner database: %w", err)
	}

	dropFn := func() { dropSpannerDBByCfg(cfg) }

	names, contents, err := migrationFiles(migrations, ".up.spanner.sql")
	if err != nil {
		dropSpannerDBByCfg(cfg)
		return cfg, nil, fmt.Errorf("reading spanner migrations: %w", err)
	}

	if len(names) == 0 {
		return cfg, dropFn, nil
	}

	version, err := migrationVersion(names)
	if err != nil {
		dropSpannerDBByCfg(cfg)
		return cfg, nil, err
	}

	var allStmts []string
	for i, content := range contents {
		ddl, err := spansql.ParseDDL(names[i], content)
		if err != nil {
			logger.Info().Logf("spanner fast create: DDL parse failed for %s, falling back: %v", names[i], err)
			return fallbackSpannerE(cfg)
		}
		for _, stmt := range ddl.List {
			allStmts = append(allStmts, stmt.SQL())
		}
	}

	schemaMigrationsDDL := `CREATE TABLE spanner_schema_migrations (
    Version INT64 NOT NULL,
    Dirty    BOOL NOT NULL
) PRIMARY KEY(Version)`
	allStmts = append([]string{schemaMigrationsDDL}, allStmts...)

	ctx := context.Background()
	adminClient, err := spannerdb.NewDatabaseAdminClient(ctx)
	if err != nil {
		dropSpannerDBByCfg(cfg)
		return cfg, nil, fmt.Errorf("creating spanner admin client: %w", err)
	}
	defer adminClient.Close()

	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s",
		cfg.Spanner.Project, cfg.Spanner.Instance, cfg.DatabaseName)

	op, err := adminClient.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
		Database:   dbPath,
		Statements: allStmts,
	})
	if err != nil {
		logger.Info().Logf("spanner fast create: batched DDL request failed, falling back: %v", err)
		return fallbackSpannerE(cfg)
	}
	if err := op.Wait(ctx); err != nil {
		logger.Info().Logf("spanner fast create: batched DDL operation failed, falling back: %v", err)
		return fallbackSpannerE(cfg)
	}

	dataClient, err := spanner.NewClient(ctx, dbPath)
	if err != nil {
		dropSpannerDBByCfg(cfg)
		return cfg, nil, fmt.Errorf("creating spanner data client for version insert: %w", err)
	}
	_, err = dataClient.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		return txn.BufferWrite([]*spanner.Mutation{
			spanner.Delete("spanner_schema_migrations", spanner.AllKeys()),
			spanner.Insert("spanner_schema_migrations",
				[]string{"Version", "Dirty"},
				[]interface{}{version, false},
			),
		})
	})
	dataClient.Close()
	if err != nil {
		dropSpannerDBByCfg(cfg)
		return cfg, nil, fmt.Errorf("inserting migration version: %w", err)
	}

	logger.Info().Logf("spanner fast create: applied %d migration files as %d DDL statements (version %d)", len(names), len(allStmts)-1, version)
	return cfg, dropFn, nil
}

// fallbackSpannerE drops the current database, creates a fresh empty one, and
// returns it so the caller can run migrations normally via LoadDatabase.
func fallbackSpannerE(cfg database.DatabaseConfig) (database.DatabaseConfig, func(), error) {
	dropSpannerDBByCfg(cfg)
	cfg2, err := NewSpannerDatabase(cfg.DatabaseName, cfg.Spanner)
	if err != nil {
		return cfg, nil, fmt.Errorf("fallback: creating fresh spanner database: %w", err)
	}
	return cfg2, func() { dropSpannerDBByCfg(cfg2) }, nil
}

// dropSpannerDBByCfg drops a Spanner database described by a DatabaseConfig.
func dropSpannerDBByCfg(cfg database.DatabaseConfig) {
	ctx := context.Background()
	adminClient, err := spannerdb.NewDatabaseAdminClient(ctx)
	if err != nil {
		return
	}
	defer adminClient.Close()

	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s",
		cfg.Spanner.Project, cfg.Spanner.Instance, cfg.DatabaseName)

	_ = adminClient.DropDatabase(ctx, &databasepb.DropDatabaseRequest{
		Database: dbPath,
	})
}
