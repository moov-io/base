package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/moov-io/base/docker"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func TestMySQL__basic(t *testing.T) {
	db := CreateTestMySQLDB(t)
	defer db.Close()

	err := db.DB.Ping()
	require.NoError(t, err)

	require.Equal(t, 0, db.DB.Stats().OpenConnections)

	// create a phony MySQL
	m := mysqlConnection(log.NewNopLogger(), "user", "pass", "127.0.0.1:3006", "db")

	ctx, cancelFunc := context.WithCancel(context.Background())

	conn, err := m.Connect(ctx)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	require.Nil(t, conn)
	require.Error(t, err)

	cancelFunc()
}

func TestMySQLUniqueViolation(t *testing.T) {
	err := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1062: Duplicate entry '282f6ffcd9ba5b029afbf2b739ee826e22d9df3b' for key 'PRIMARY'`)
	if !UniqueViolation(err) {
		t.Error("should have matched unique violation")
	}
}

func TestCreateTemporaryDatabase(t *testing.T) {
	if !docker.Enabled() {
		t.Skip("Docker not enabled")
	}

	config, err := findOrLaunchMySQLContainer()
	require.NoError(t, err)

	name, err := createTemporaryDatabase(t, config)
	require.NoError(t, err)
	require.Contains(t, name, "Test")
}

func TestMySQLModes(t *testing.T) {
	db := CreateTestMySQLDB(t)
	defer db.Close()

	// Inspect the global and session SQL modes
	// See: https://dev.mysql.com/doc/refman/8.0/en/sql-mode.html#sql-mode-setting
	sqlModes := readSQLModes(t, db.DB, "SELECT @@SESSION.sql_mode;")
	require.Contains(t, sqlModes, "ALLOW_INVALID_DATES")
	require.Contains(t, sqlModes, "STRICT_ALL_TABLES")
}

func readSQLModes(t *testing.T, db *sql.DB, query string) string {
	stmt, err := db.Prepare(query)
	require.NoError(t, err)

	row := stmt.QueryRow()
	require.NoError(t, row.Err())

	var sqlModes string
	require.NoError(t, row.Scan(&sqlModes))
	return sqlModes
}
