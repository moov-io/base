package database_test

import (
	"context"
	"testing"

	"github.com/moov-io/base/database"
	"github.com/moov-io/base/database/testdb"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func Test_OpenConnection(t *testing.T) {

	// Switches the spanner driver into using the emulator and bypassing the auth checks.
	testdb.SetSpannerEmulator(nil)

	cfg := database.DatabaseConfig{
		DatabaseName: "my-database",
		Spanner: &database.SpannerConfig{
			Project:  "my-project",
			Instance: "my-instance",
		},
	}

	db, err := database.New(context.Background(), log.NewDefaultLogger(), cfg)
	require.NoError(t, err)
	defer db.Close()
}

func Test_Migration(t *testing.T) {

	// Switches the spanner driver into using the emulator and bypassing the auth checks.
	testdb.SetSpannerEmulator(nil)

	cfg, err := testdb.NewSpannerDatabase("mydb")
	require.NoError(t, err)

	err = database.RunMigrations(log.NewDefaultLogger(), cfg)
	require.NoError(t, err)
}

func Test_MigrateAndRun(t *testing.T) {

	// Switches the spanner driver into using the emulator and bypassing the auth checks.
	testdb.SetSpannerEmulator(nil)

	cfg, err := testdb.NewSpannerDatabase("mydb")
	require.NoError(t, err)

	err = database.RunMigrations(log.NewDefaultLogger(), cfg)
	require.NoError(t, err)

	db, err := database.New(context.Background(), log.NewDefaultLogger(), cfg)
	require.NoError(t, err)
	defer db.Close()

	rows, err := db.Query("SELECT * FROM MigrationTest")
	require.NoError(t, err)
	defer rows.Close()
	require.NoError(t, rows.Err())
}
