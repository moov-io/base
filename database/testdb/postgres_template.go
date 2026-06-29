package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/moov-io/base"
	"github.com/moov-io/base/database"
	"github.com/moov-io/base/log"
)

// NewPostgresDatabaseFromTemplate creates an isolated test database by cloning
// a template database that has already been migrated. The template is created
// once per unique migration set (content-addressed by SHA256) and reused across
// all tests in the process. This avoids running migrations for every test.
//
// The template database is marked IS_TEMPLATE=true and ALLOW_CONNECTIONS=false
// so no test can accidentally connect to it. Each test gets its own clone with
// full isolation.
func NewPostgresDatabaseFromTemplate(
	tb testing.TB,
	logger log.Logger,
	cfg database.DatabaseConfig,
	migrations fs.FS,
) (database.DatabaseConfig, func()) {
	tb.Helper()

	if cfg.Postgres == nil {
		tb.Fatal("NewPostgresDatabaseFromTemplate: postgres config not defined")
	}
	if migrations == nil {
		tb.Fatal("NewPostgresDatabaseFromTemplate: migrations FS is nil")
	}

	names, contents, err := migrationFiles(migrations, ".up.postgres.sql")
	if err != nil {
		tb.Fatal(fmt.Errorf("reading postgres migrations: %w", err))
	}

	hash := migrationHash(names, contents)
	templateName := fmt.Sprintf("tmpl_%s", hash[:16])

	// Connect to an admin database to manage templates.
	// Try "moov" first (the standard Moov docker-compose DB), then "postgres".
	adminDb := openAdminDb(tb, cfg.Postgres)

	// Ensure the service database exists. Some tests (e.g. TestEnvironment)
	// call service.NewEnvironment directly without going through
	// CreateTestDatabase, relying on the service database existing as a
	// side effect of prior test setup. The old CreateTestDatabase created
	// it via openOrCreateDatabase; we preserve that behavior here.
	ensureServiceDatabase(tb, adminDb, cfg.DatabaseName)

	ensurePostgresTemplate(tb, logger, adminDb, cfg, migrations, hash, templateName)

	// Create the per-test database from the template. The template creation lock
	// is intentionally released before this point: the lock protects the template
	// definition, not every clone created from it.
	testDbName := "test" + base.ID()[0:26]
	_, err = adminDb.Exec(fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s",
		pgx.Identifier{testDbName}.Sanitize(),
		pgx.Identifier{templateName}.Sanitize()))
	if err != nil {
		tb.Fatal(fmt.Errorf("creating test database from template: %w", err))
	}

	dropFn := func() {
		_, err := adminDb.Exec(fmt.Sprintf("DROP DATABASE %s", pgx.Identifier{testDbName}.Sanitize()))
		if err != nil {
			tb.Logf("cleanup: drop database %s: %v", testDbName, err)
		}
		if err := adminDb.Close(); err != nil {
			tb.Logf("cleanup: close admin database: %v", err)
		}
	}

	cfg.DatabaseName = testDbName
	return cfg, dropFn
}

func ensurePostgresTemplate(
	tb testing.TB,
	logger log.Logger,
	adminDb *sql.DB,
	cfg database.DatabaseConfig,
	migrations fs.FS,
	hash string,
	templateName string,
) {
	tb.Helper()

	ctx := context.Background()
	conn, err := adminDb.Conn(ctx)
	if err != nil {
		tb.Fatal(fmt.Errorf("acquiring admin connection: %w", err))
	}
	defer func() {
		if err := conn.Close(); err != nil {
			tb.Logf("cleanup: close admin connection: %v", err)
		}
	}()

	// Serialize template creation across processes using a session-level advisory
	// lock keyed by the migration hash. The lock and unlock must use the same
	// physical connection.
	lockKey := hashToLockKey(hash)
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", lockKey); err != nil {
		tb.Fatal(fmt.Errorf("acquiring advisory lock: %w", err))
	}
	defer func() {
		if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", lockKey); err != nil {
			tb.Logf("cleanup: release advisory lock: %v", err)
		}
	}()

	var exists bool
	err = conn.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1 AND datistemplate = true)",
		templateName,
	).Scan(&exists)
	if err != nil {
		tb.Fatal(fmt.Errorf("checking template existence: %w", err))
	}

	if exists {
		return
	}

	_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", pgx.Identifier{templateName}.Sanitize()))
	if err != nil {
		tb.Fatal(fmt.Errorf("creating template database: %w", err))
	}

	templateCfg := database.DatabaseConfig{
		DatabaseName: templateName,
		Postgres:     cfg.Postgres,
	}
	if err := database.RunMigrations(logger, templateCfg, database.WithEmbeddedMigrations(migrations)); err != nil {
		if _, dropErr := conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", pgx.Identifier{templateName}.Sanitize())); dropErr != nil {
			tb.Logf("cleanup: drop template database %s: %v", templateName, dropErr)
		}
		tb.Fatal(fmt.Errorf("running migrations on template: %w", err))
	}

	_, err = conn.ExecContext(ctx, fmt.Sprintf("ALTER DATABASE %s IS_TEMPLATE true", pgx.Identifier{templateName}.Sanitize()))
	if err != nil {
		tb.Fatal(fmt.Errorf("marking template: %w", err))
	}
	_, err = conn.ExecContext(ctx, fmt.Sprintf("ALTER DATABASE %s ALLOW_CONNECTIONS false", pgx.Identifier{templateName}.Sanitize()))
	if err != nil {
		tb.Fatal(fmt.Errorf("disallowing connections to template: %w", err))
	}
}

// openAdminDb connects to an admin database for managing templates.
// Tries "moov" first, then falls back to "postgres".
func openAdminDb(tb testing.TB, pgCfg *database.PostgresConfig) *sql.DB {
	tb.Helper()

	for _, adminDb := range []string{"moov", "postgres"} {
		db, err := sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s/%s",
			pgCfg.User, pgCfg.Password, pgCfg.Address, adminDb))
		if err != nil {
			continue
		}
		if err := db.Ping(); err != nil {
			db.Close()
			continue
		}
		return db
	}

	tb.Fatal("could not connect to admin database (tried 'moov' and 'postgres')")
	return nil
}

// hashToLockKey converts the first 8 bytes of a hex hash string into an int64
// for use with pg_advisory_lock.
func hashToLockKey(hash string) int64 {
	var key int64
	for i := 0; i < 8 && i < len(hash); i++ {
		key = key<<8 | int64(hash[i])
	}
	return key
}

// ensureServiceDatabase creates the service database if it doesn't exist.
// This preserves the side-effect behavior of the old CreateTestDatabase,
// which created the service database via openOrCreateDatabase. Some tests
// call service.NewEnvironment directly without CreateTestDatabase and rely
// on the service database existing.
func ensureServiceDatabase(tb testing.TB, adminDb *sql.DB, dbName string) {
	tb.Helper()
	if dbName == "" {
		return
	}
	var exists bool
	err := adminDb.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		dbName,
	).Scan(&exists)
	if err != nil {
		tb.Fatal(fmt.Errorf("checking service database existence: %w", err))
	}
	if !exists {
		_, err = adminDb.Exec(fmt.Sprintf("CREATE DATABASE %s", pgx.Identifier{dbName}.Sanitize()))
		if err != nil {
			if checkErr := adminDb.QueryRow(
				"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
				dbName,
			).Scan(&exists); checkErr != nil || !exists {
				tb.Fatal(fmt.Errorf("creating service database %s: %w", dbName, err))
			}
		}
	}
}
