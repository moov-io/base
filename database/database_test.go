// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package database_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	gomysql "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/moov-io/base/database"

	"github.com/stretchr/testify/require"
)

func TestUniqueViolation(t *testing.T) {
	// mysql
	mysqlErr := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1062 (23000): Duplicate entry '282f6ffcd9ba5b029afbf2b739ee826e22d9df3b' for key 'PRIMARY'`)
	if !database.UniqueViolation(mysqlErr) {
		t.Error("should have matched mysql unique violation")
	}
	gomysqlErr := &gomysql.MySQLError{
		Number: 1062,
	}
	if !database.UniqueViolation(gomysqlErr) {
		t.Error("should have matched go mysql driver unique violation")
	}

	// postgres
	psqlErr := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": ERROR: duplicate key value violates unique constraint "depository" (SQLSTATE 23505)`)
	if !database.UniqueViolation(psqlErr) {
		t.Error("should have matched postgres unique violation")
	}
	pgconnErr := &pgconn.PgError{
		Code: "23505",
	}
	if !database.UniqueViolation(pgconnErr) {
		t.Error("should have matched PgError unique violation")
	}

	// no violation
	noViolationErr := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1061 (23000): Something went wrong`)
	if database.UniqueViolation(noViolationErr) {
		t.Error("should not have matched unique violation")
	}
	gomysqlErr.Number = 1061
	if database.UniqueViolation(gomysqlErr) {
		t.Error("should not have matched go mysql driver unique violation")
	}
	pgconnErr.Code = "23504"
	if database.UniqueViolation(pgconnErr) {
		t.Error("should not have matched PgError unique violation")
	}
}

func TestDeadlockFound(t *testing.T) {
	// mysql
	mysqlErr := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1213 (40001): Deadlock found when trying to get lock; try restarting transaction`)
	if !database.DeadlockFound(mysqlErr) {
		t.Error("should have matched mysql deadlock found")
	}
	gomysqlErr := &gomysql.MySQLError{
		Number: 1213,
	}
	if !database.DeadlockFound(gomysqlErr) {
		t.Error("should have matched go mysql driver deadlock found")
	}

	// postgres
	psqlErr := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": ERROR: deadlock detected (SQLSTATE 40P01)`)
	if !database.DeadlockFound(psqlErr) {
		t.Error("should have matched postgres deadlock found")
	}
	pgconnErr := &pgconn.PgError{
		Code: "40P01",
	}
	if !database.DeadlockFound(pgconnErr) {
		t.Error("should have matched PgError deadlock found")
	}

	// no deadlock found
	noDeadlockErr := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1061 (23000): Something went wrong`)
	if database.DeadlockFound(noDeadlockErr) {
		t.Error("should not have matched deadlock found")
	}

	gomysqlErr.Number = 1231
	if database.DeadlockFound(gomysqlErr) {
		t.Error("should not have matched go mysql driver deadlock found")
	}
	pgconnErr.Code = "40P02"
	if database.DeadlockFound(pgconnErr) {
		t.Error("should not have matched PgError deadlock found")
	}
}

func TestDataTooLong(t *testing.T) {
	// mysql
	mysqlErr := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1406 (22001): Data too long for column 'depository' at row 1`)
	if !database.DataTooLong(mysqlErr) {
		t.Error("should have matched mysql data too long")
	}
	gomysqlErr := &gomysql.MySQLError{
		Number: 1406,
	}
	if !database.DataTooLong(gomysqlErr) {
		t.Error("should have matched go mysql driver data too long")
	}

	// no data too long
	noDataTooLongErr := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1062 (23000): Something went wrong`)
	if database.DataTooLong(noDataTooLongErr) {
		t.Error("should not have mysql matched data too long")
	}
	gomysqlErr.Number = 1062
	if database.DataTooLong(gomysqlErr) {
		t.Error("should not have matched go mysql driver data too long")
	}
}

func TestConnectionsConfigOrder(t *testing.T) {
	bs, err := os.ReadFile("database.go")
	require.NoError(t, err)

	// SetConnMaxIdleTime must be specified first
	// See: https://github.com/golang/go/issues/45993#issuecomment-1427873850
	maxIdleTimeIdx := bytes.Index(bs, []byte("db.SetConnMaxIdleTime"))
	maxLifetimeIdx := bytes.Index(bs, []byte("db.SetConnMaxLifetime"))

	if maxIdleTimeIdx > maxLifetimeIdx {
		t.Error(".SetConnMaxIdleTime must come first")
	}
}
