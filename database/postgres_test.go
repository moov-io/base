package database_test

import (
	"context"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/base/database"
	"github.com/moov-io/base/database/testdb"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func TestPostgres_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	config := database.DatabaseConfig{
		DatabaseName: "moov",
		Postgres: &database.PostgresConfig{
			Address:  "localhost:5432",
			User:     "moov",
			Password: "moov",
		},
	}

	db, err := database.New(context.Background(), log.NewTestLogger(), config)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()
}

func TestPostgres_TLS(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	config := database.DatabaseConfig{
		DatabaseName: "moov",
		Postgres: &database.PostgresConfig{
			Address:           "localhost:5432",
			User:              "moov",
			Password:          "moov",
			UseTLS:            true,
			TLSCAFile:         "../testcerts/root.crt",
			TLSClientCertFile: "../testcerts/client.crt",
			TLSClientKeyFile:  "../testcerts/client.key",
		},
	}

	db, err := database.New(context.Background(), log.NewTestLogger(), config)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()
}

func Test_Postgres_Embedded_Migration(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	// create a test postgres db
	config := database.DatabaseConfig{
		DatabaseName: "postgres" + base.ID(),
		Postgres: &database.PostgresConfig{
			Address:  "localhost:5432",
			User:     "moov",
			Password: "moov",
		},
	}

	err := testdb.NewPostgresDatabase(t, config)
	require.NoError(t, err)

	db, err := database.NewAndMigrate(context.Background(), log.NewDefaultLogger(), config, database.WithEmbeddedMigrations(base.PostgresMigrations))
	require.NoError(t, err)
	defer db.Close()
}

func Test_Postgres_UniqueViolation(t *testing.T) {
	// create a test postgres db
	config := database.DatabaseConfig{
		DatabaseName: "postgres" + base.ID(),
		Postgres: &database.PostgresConfig{
			Address:  "localhost:5432",
			User:     "moov",
			Password: "moov",
		},
	}

	err := testdb.NewPostgresDatabase(t, config)
	require.NoError(t, err)

	db, err := database.New(context.Background(), log.NewDefaultLogger(), config)
	require.NoError(t, err)

	createQry := `CREATE TABLE names (id SERIAL PRIMARY KEY, name VARCHAR(255));`
	_, err = db.Exec(createQry)
	require.NoError(t, err)

	insertQry := `INSERT INTO names (id, name) VALUES ($1, $2);`
	_, err = db.Exec(insertQry, 1, "James")
	require.NoError(t, err)

	_, err = db.Exec(insertQry, 1, "James")
	require.Error(t, err)
	require.True(t, database.UniqueViolation(err))
}
