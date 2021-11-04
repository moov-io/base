//go:build cgo
// +build cgo

// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewAndMigration_SQLite(t *testing.T) {
	dir, err := ioutil.TempDir("", "sqlite-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	config := &DatabaseConfig{SQLite: &SQLiteConfig{
		Path: filepath.Join(dir, "tests.db"),
	}}

	db, err := NewAndMigrate(context.Background(), nil, *config)
	require.NoError(t, err)
	defer db.Close()

	rows, err := db.Query("select * from tests")
	require.NoError(t, err)
	require.NoError(t, rows.Err())
}

func TestUniqueViolation(t *testing.T) {
	err := errors.New(`problem upserting depository="7d676c65eccd48090ff238a0d5e35eb6126c23f2", userId="80cfe1311d9eb7659d02cba9ee6cb04ed3739a85": UNIQUE constraint failed: depositories.depository_id`)
	if !UniqueViolation(err) {
		t.Error("should have matched unique violation")
	}
}
