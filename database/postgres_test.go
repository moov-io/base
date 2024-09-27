package database_test

import (
	"context"
	"os"
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
			Address:  "localhost:5432",
			User:     "moov",
			Password: "moov",
			TLS: &database.PostgresTLSConfig{
				CACertFile:     "../testcerts/root.crt",
				ClientCertFile: "../testcerts/client.crt",
				ClientKeyFile:  "../testcerts/client.key",
			},
		},
	}

	db, err := database.New(context.Background(), log.NewTestLogger(), config)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()
}

func TestProstres_Alloy(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	alloydbInstanceURI := os.Getenv("ALLOYDB_INSTANCE_URI")
	alloydbDBName := os.Getenv("ALLOYDB_DBNAME")
	alloydbUser := os.Getenv("ALLOYDB_USER")
	alloydbPassword := os.Getenv("ALLOYDB_PASSWORD")

	if alloydbInstanceURI == "" || alloydbDBName == "" || alloydbUser == "" || alloydbPassword == "" {
		t.Skip("missing required environment variables")
	}

	config := database.DatabaseConfig{
		DatabaseName: alloydbDBName,
		Postgres: &database.PostgresConfig{
			User:     alloydbUser,
			Password: alloydbPassword,
			Alloy: &database.PostgresAlloyConfig{
				InstanceURI: alloydbInstanceURI,
				UseIAM:      false,
				UsePSC:      true,
			},
		},
	}

	db, err := database.New(context.Background(), log.NewTestLogger(), config)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()
}

func TestProstres_Alloy_IAM(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	alloydbInstanceURI := os.Getenv("ALLOYDB_INSTANCE_URI")
	alloydbDBName := os.Getenv("ALLOYDB_DBNAME")
	alloydbUser := os.Getenv("ALLOYDB_USER")

	if alloydbInstanceURI == "" || alloydbDBName == "" || alloydbUser == "" {
		t.Skip("missing required environment variables")
	}

	config := database.DatabaseConfig{
		DatabaseName: alloydbDBName,
		Postgres: &database.PostgresConfig{
			User: alloydbUser,
			Alloy: &database.PostgresAlloyConfig{
				InstanceURI: alloydbInstanceURI,
				UseIAM:      true,
				UsePSC:      true,
			},
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
