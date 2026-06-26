package testdb

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moov-io/base/database"
	"github.com/moov-io/base/log"
)

//go:embed all:testdata
var testEmbedFS embed.FS

// testMigrationsFS wraps the embedded files so that migrationFiles can call
// fs.ReadDir(migrations, "migrations"). We embed testdata/ which contains
// a migrations/ subdirectory, matching the convention used by
// database.WithEmbeddedMigrations (iofs.New(f, "migrations")).
var testMigrationsFS fs.FS

func init() {
	sub, err := fs.Sub(testEmbedFS, "testdata")
	if err != nil {
		panic(err)
	}
	testMigrationsFS = sub
}

func TestMigrationFiles_postgres(t *testing.T) {
	names, contents, err := migrationFiles(testMigrationsFS, ".up.postgres.sql")
	require.NoError(t, err)
	assert.Len(t, names, 2)
	assert.Equal(t, "001_create_users.up.postgres.sql", names[0])
	assert.Equal(t, "002_add_email.up.postgres.sql", names[1])
	assert.Contains(t, contents[0], "CREATE TABLE users (id TEXT PRIMARY KEY)")
	assert.Contains(t, contents[1], "ALTER TABLE users ADD COLUMN email TEXT")
}

func TestMigrationFiles_spanner(t *testing.T) {
	names, contents, err := migrationFiles(testMigrationsFS, ".up.spanner.sql")
	require.NoError(t, err)
	assert.Len(t, names, 2)
	assert.Equal(t, "001_create_items.up.spanner.sql", names[0])
	assert.Equal(t, "002_add_price.up.spanner.sql", names[1])
	assert.Contains(t, contents[0], "CREATE TABLE items")
	assert.Contains(t, contents[1], "ADD COLUMN Price")
}

func TestMigrationFiles_wrongSuffix(t *testing.T) {
	names, _, err := migrationFiles(testMigrationsFS, ".up.mysql.sql")
	require.NoError(t, err)
	assert.Empty(t, names)
}

func TestMigrationHash_stable(t *testing.T) {
	names := []string{"001_a.up.sql", "002_b.up.sql"}
	contents := []string{"-- a\n", "-- b\n"}
	h1 := migrationHash(names, contents)
	h2 := migrationHash(names, contents)
	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 64)
}

func TestMigrationHash_changesOnContent(t *testing.T) {
	names := []string{"001_a.up.sql"}
	h1 := migrationHash(names, []string{"-- a\n"})
	h2 := migrationHash(names, []string{"-- b\n"})
	assert.NotEqual(t, h1, h2)
}

func TestMigrationHash_changesOnName(t *testing.T) {
	contents := []string{"-- a\n"}
	h1 := migrationHash([]string{"001_a.up.sql"}, contents)
	h2 := migrationHash([]string{"001_c.up.sql"}, contents)
	assert.NotEqual(t, h1, h2)
}

func TestMigrationVersion(t *testing.T) {
	tests := []struct {
		name    string
		names   []string
		want    int
		wantErr bool
	}{
		{"empty", nil, 0, false},
		{"single", []string{"001_a.up.sql"}, 1, false},
		{"multi", []string{"001_a.up.sql", "010_b.up.sql", "003_c.up.sql"}, 3, false},
		{"large", []string{"001_a.up.sql", "999_final.up.sql"}, 999, false},
		{"bad", []string{"abc.up.sql"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := migrationVersion(tt.names)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestHashToLockKey(t *testing.T) {
	assert.NotZero(t, hashToLockKey("0123456789abcdef"))
	assert.Equal(t, int64('a'), hashToLockKey("a"))
}

func TestEnsureServiceDatabase(t *testing.T) {
	t.Run("empty name", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		ensureServiceDatabase(t, db, "")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("already exists", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs("svc").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		ensureServiceDatabase(t, db, "svc")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create succeeds", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs("svc").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
		mock.ExpectExec("CREATE DATABASE").
			WillReturnResult(sqlmock.NewResult(0, 1))

		ensureServiceDatabase(t, db, "svc")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create loses race", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs("svc").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
		mock.ExpectExec("CREATE DATABASE").
			WillReturnError(assert.AnError)
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs("svc").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		ensureServiceDatabase(t, db, "svc")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCreateSpannerDatabaseFromMigrations_nilMigrations(t *testing.T) {
	cfg, dropFn, err := CreateSpannerDatabaseFromMigrations(log.NewNopLogger(), database.DatabaseConfig{}, nil)
	require.Error(t, err)
	assert.Nil(t, dropFn)
	assert.Empty(t, cfg.DatabaseName)
}
