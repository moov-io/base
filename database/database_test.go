// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package database

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/moov-io/base/docker"
)

func Test_NewAndMigration_MySql(t *testing.T) {
	if !docker.Enabled() {
		t.SkipNow()
	}

	mySQLConfig, err := findOrLaunchMySQLContainer()
	require.NoError(t, err)

	databaseName, err := createTemporaryDatabase(t, mySQLConfig)
	require.NoError(t, err)

	config := DatabaseConfig{
		DatabaseName: databaseName,
		MySQL:        mySQLConfig,
	}

	db, err := NewAndMigrate(context.Background(), nil, config)
	require.NoError(t, err)
	db.Close()
}

func TestUniqueViolation__Sqlite(t *testing.T) {
	err := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1062: Duplicate entry '282f6ffcd9ba5b029afbf2b739ee826e22d9df3b' for key 'PRIMARY'`)
	if !UniqueViolation(err) {
		t.Error("should have matched unique violation")
	}
}
