// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package database_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/moov-io/base/database"
	"github.com/stretchr/testify/require"
)

func TestMySQLConfig(t *testing.T) {
	cfg := &database.MySQLConfig{
		Address:  "tcp(localhost:3306)",
		User:     "app",
		Password: "secret",
		Connections: database.ConnectionsConfig{
			MaxOpen: 100,
		},
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(cfg)
	require.NoError(t, err)
	require.Contains(t, buf.String(), `"Password":"s*****t"`)
}
