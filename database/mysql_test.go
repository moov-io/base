// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func TestMySQL__basic(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	// create a phony MySQL
	mysqlConfig := &MySQLConfig{
		User:     "moov",
		Password: "moov",
		Address:  "tcp(127.0.0.1:3306)",
	}
	m, err := mysqlConnection(log.NewNopLogger(), mysqlConfig, "moov")
	require.NoError(t, err)

	ctx, cancelFunc := context.WithCancel(context.Background())

	conn, err := m.Connect(ctx)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	require.NotNil(t, conn)
	require.NoError(t, err)

	// Inspect the global and session SQL modes
	// See: https://dev.mysql.com/doc/refman/8.0/en/sql-mode.html#sql-mode-setting
	sqlModes := readSQLModes(t, conn, "SELECT @@SESSION.sql_mode;")
	require.Contains(t, sqlModes, "ALLOW_INVALID_DATES")
	require.Contains(t, sqlModes, "STRICT_ALL_TABLES")

	require.Equal(t, 1, conn.Stats().OpenConnections)

	cancelFunc()
}

func TestMySQLUniqueViolation(t *testing.T) {
	err := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1062: Duplicate entry '282f6ffcd9ba5b029afbf2b739ee826e22d9df3b' for key 'PRIMARY'`)
	if !UniqueViolation(err) {
		t.Error("should have matched unique violation")
	}
}

func TestMySQLUniqueViolation_WithStateValue(t *testing.T) {
	err := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1062 (23000): Duplicate entry '282f6ffcd9ba5b029afbf2b739ee826e22d9df3b' for key 'PRIMARY'`)
	if !UniqueViolation(err) {
		t.Error("should have matched unique violation")
	}
}

func TestMySQLDataTooLong(t *testing.T) {
	err := errors.New("Error 1406: Data too long")
	if !MySQLDataTooLong(err) {
		t.Error("should have matched")
	}
}

func readSQLModes(t *testing.T, db *sql.DB, query string) string {
	stmt, err := db.Prepare(query)
	require.NoError(t, err)
	defer stmt.Close()

	row := stmt.QueryRow()
	require.NoError(t, row.Err())

	var sqlModes string
	require.NoError(t, row.Scan(&sqlModes))
	return sqlModes
}
