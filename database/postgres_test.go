package database_test

import (
	"context"
	"testing"

	"github.com/moov-io/base/database"
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

	require.NoError(t, db.Ping())
}
