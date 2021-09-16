// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package database

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMySQLConfig(t *testing.T) {
	cfg := &MySQLConfig{
		Address:  "tcp(localhost:3306)",
		User:     "app",
		Password: "secret",
		SSLCA:    "/etc/ssl/certs/dummy.crt",
		Connections: ConnectionsConfig{
			MaxOpen: 100,
		},
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(cfg)
	require.NoError(t, err)
	require.Contains(t, buf.String(), `"Password":"s*****t"`)
}
