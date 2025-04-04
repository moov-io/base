// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package database_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/moov-io/base/database"

	"github.com/stretchr/testify/require"
)

func TestUniqueViolation(t *testing.T) {
	mysqlErr := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1062 (23000): Duplicate entry '282f6ffcd9ba5b029afbf2b739ee826e22d9df3b' for key 'PRIMARY'`)
	if !database.UniqueViolation(mysqlErr) {
		t.Error("should have matched unique violation")
	}

	psqlErr := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": ERROR: duplicate key value violates unique constraint "depository" (SQLSTATE 23505)`)
	if !database.UniqueViolation(psqlErr) {
		t.Error("should have matched unique violation")
	}
}

func TestDeadlockFound(t *testing.T) {
	mysqlErr := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1213 (40001): Deadlock found when trying to get lock; try restarting transaction`)
	if !database.DeadlockFound(mysqlErr) {
		t.Error("should have matched deadlock found")
	}

	psqlErr := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": ERROR: deadlock detected (SQLSTATE 40P01)`)
	if !database.DeadlockFound(psqlErr) {
		t.Error("should have matched deadlock found")
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
