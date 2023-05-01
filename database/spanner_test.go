package database_test

import (
	"context"
	"testing"

	"cloud.google.com/go/spanner"
	"github.com/googleapis/gax-go/v2/apierror"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/moov-io/base/database"
	"github.com/moov-io/base/database/testdb"
	"github.com/moov-io/base/log"
)

func Test_OpenConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

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
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	// Switches the spanner driver into using the emulator and bypassing the auth checks.
	testdb.SetSpannerEmulator(nil)

	cfg, err := testdb.NewSpannerDatabase("mydb")
	require.NoError(t, err)

	err = database.RunMigrations(log.NewDefaultLogger(), cfg)
	require.NoError(t, err)
}

func Test_MigrateAndRun(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

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

func TestSpannerUniqueViolation(t *testing.T) {
	errMsg := "Failed to insert row with primary key ({pk#primary_key:\"282f6ffcd9ba5b029afbf2b739ee826e22d9df3b\"}) due to previously existing row"
	// Test backwards-compatible parsing of spanner.Error (soon to be deprecated) from Spanner client
	oldSpannerErr := spanner.ToSpannerError(status.New(codes.AlreadyExists, errMsg).Err())
	if !database.SpannerUniqueViolation(oldSpannerErr) {
		t.Error("should have matched unique violation")
	}

	// Test new apirerror.APIError response from Spanner client
	newSpannerErr, parseErr := apierror.FromError(status.New(codes.AlreadyExists, errMsg).Err())
	require.True(t, parseErr)
	if !database.SpannerUniqueViolation(newSpannerErr) {
		t.Error("should have matched unique violation")
	}
}
