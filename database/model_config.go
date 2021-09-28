// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package database

import (
	"encoding/json"
	"time"

	"github.com/moov-io/base/mask"
)

type DatabaseConfig struct {
	MySQL        *MySQLConfig
	SQLite       *SQLiteConfig
	DatabaseName string
}

type MySQLConfig struct {
	Address            string
	User               string
	Password           string
	Connections        ConnectionsConfig
	UseTLS             bool
	TLSCAFile          string
	InsecureSkipVerify bool
	VerifyCAFile       bool
}

func (m *MySQLConfig) MarshalJSON() ([]byte, error) {
	type Aux struct {
		Address            string
		User               string
		Password           string
		Connections        ConnectionsConfig
		UseTLS             bool
		TLSCAFile          string
		InsecureSkipVerify bool
		VerifyCAFile       bool
	}
	return json.Marshal(Aux{
		Address:            m.Address,
		User:               m.User,
		Password:           mask.Password(m.Password),
		Connections:        m.Connections,
		UseTLS:             m.UseTLS,
		TLSCAFile:          m.TLSCAFile,
		InsecureSkipVerify: m.InsecureSkipVerify,
		VerifyCAFile:       m.VerifyCAFile,
	})
}

type SQLiteConfig struct {
	Path string
}

type ConnectionsConfig struct {
	MaxOpen     int
	MaxIdle     int
	MaxLifetime time.Duration
	MaxIdleTime time.Duration
}
