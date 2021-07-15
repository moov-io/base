//go:build !cgo
// +build !cgo

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
)

type sqlite struct{}

func (s *sqlite) Connect(ctx context.Context) (*sql.DB, error) {
	return nil, errors.New("sqlite: Connect CGO disabled")
}

func sqliteConnection(logger log.Logger, path string) (*sqlite, error) {
	return nil, errors.New("sqlite: CGO disabled")
}

func SQLiteUniqueViolation(err error) bool {
	return false
}

type TestSQLiteDB struct{}

func (r *TestSQLiteDB) Close() error {
	return errors.New("CGO disabled")
}

func CreateTestSQLiteDB(t *testing.T) *TestSQLiteDB {
	t.Skip("sqlite: CGO disabled")
	return nil
}
