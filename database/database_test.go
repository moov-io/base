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
	err := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1062: Duplicate entry '282f6ffcd9ba5b029afbf2b739ee826e22d9df3b' for key 'PRIMARY'`)
	if !database.UniqueViolation(err) {
		t.Error("should have matched unique violation")
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
