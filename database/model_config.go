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
	Spanner      *SpannerConfig
	Postgres     *PostgresConfig
	DatabaseName string
}

type SpannerConfig struct {
	Project  string
	Instance string

	DisableCleanStatements bool
}

type PostgresConfig struct {
	Address             string
	User                string
	Password            string
	UseTLS              bool
	TLSCAFile           string
	TLSClientKeyFile    string
	TLSClientCertFile   string
	UseAlloyDBConnector bool
	AlloyDBInstanceURI  string
	UseAlloyDBIAM       bool
}

type MySQLConfig struct {
	Address        string
	User           string
	Password       string
	Connections    ConnectionsConfig
	UseTLS         bool
	TLSCAFile      string
	VerifyCAFile   bool
	TLSClientCerts []TLSClientCertConfig

	// InsecureSkipVerify is a dangerous option which should be used with extreme caution.
	// This setting disables multiple security checks performed with TLS connections.
	InsecureSkipVerify bool
}

type TLSClientCertConfig struct {
	CertFilePath string
	KeyFilePath  string
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

type ConnectionsConfig struct {
	MaxOpen     int
	MaxIdle     int
	MaxLifetime time.Duration
	MaxIdleTime time.Duration
}
