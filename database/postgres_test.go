package database_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moov-io/base"
	"github.com/moov-io/base/database"
	"github.com/moov-io/base/database/testdb"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

var (
	alloydbInstanceURI    = os.Getenv("ALLOYDB_INSTANCE_URI")
	alloydbDBName         = os.Getenv("ALLOYDB_DBNAME")
	alloydbIAMUser        = os.Getenv("ALLOYDB_IAM_USER")
	alloydbNativeUser     = os.Getenv("ALLOYDB_NATIVE_USER")
	alloydbNativePassword = os.Getenv("ALLOYDB_NATIVE_PASSWORD")
)

func TestPostgres_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	config := database.DatabaseConfig{
		DatabaseName: "moov",
		Postgres: &database.PostgresConfig{
			Address:  "127.0.0.1:5432",
			User:     "moov",
			Password: "moov",
			Connections: database.ConnectionsConfig{
				MaxOpen:     4,
				MaxIdle:     4,
				MaxLifetime: time.Minute * 2,
				MaxIdleTime: time.Minute * 2,
			},
		},
	}

	db, err := database.New(context.Background(), log.NewTestLogger(), config)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()
}

func TestPostgres_TLS(t *testing.T) {
	t.Skip()

	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	config := database.DatabaseConfig{
		DatabaseName: "moov",
		Postgres: &database.PostgresConfig{
			Address:  "127.0.0.1:5432",
			User:     "moov",
			Password: "moov",
			TLS: &database.PostgresTLSConfig{
				CACertFile:     filepath.Join("..", "testcerts", "root.crt"),
				ClientCertFile: filepath.Join("..", "testcerts", "client.crt"),
				ClientKeyFile:  filepath.Join("..", "testcerts", "client.key"),
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

	if alloydbInstanceURI == "" || alloydbDBName == "" || alloydbNativeUser == "" || alloydbNativePassword == "" {
		t.Skip("missing required environment variables")
	}

	config := database.DatabaseConfig{
		DatabaseName: alloydbDBName,
		Postgres: &database.PostgresConfig{
			User:     alloydbNativeUser,
			Password: alloydbNativePassword,
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

	if alloydbInstanceURI == "" || alloydbDBName == "" || alloydbIAMUser == "" {
		t.Skip("missing required environment variables")
	}

	config := database.DatabaseConfig{
		DatabaseName: alloydbDBName,
		Postgres: &database.PostgresConfig{
			User: alloydbIAMUser,
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
			Address:  "127.0.0.1:5432",
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

func Test_Postgres_Alloy_Migrations(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	if alloydbInstanceURI == "" || alloydbDBName == "" || alloydbNativeUser == "" || alloydbNativePassword == "" {
		t.Skip("missing required environment variables")
	}

	config := database.DatabaseConfig{
		DatabaseName: alloydbDBName,
		Postgres: &database.PostgresConfig{
			User:     alloydbNativeUser,
			Password: alloydbNativePassword,
			Alloy: &database.PostgresAlloyConfig{
				InstanceURI: alloydbInstanceURI,
				UseIAM:      false,
				UsePSC:      true,
			},
		},
	}

	// migrating database given by ALLOYDB_DBNAME env var

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
			Address:  "127.0.0.1:5432",
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
