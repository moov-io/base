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
	Retries      *RetryConfig
}

type SpannerConfig struct {
	Project  string
	Instance string

	DisableCleanStatements bool
}

type PostgresConfig struct {
	Address     string
	User        string
	Password    string `json:"-"`
	Connections ConnectionsConfig
	TLS         *PostgresTLSConfig
	Alloy       *PostgresAlloyConfig
}

type PostgresTLSConfig struct {
	Mode string

	CACertFile     string
	ClientKeyFile  string
	ClientCertFile string
}

type PostgresAlloyConfig struct {
	InstanceURI string
	UseIAM      bool
	UsePSC      bool
}

type MySQLConfig struct {
	Address        string
	User           string
	Password       string `json:"-"`
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

type RetryConfig struct {
	MaxAttempts int
	MinDuration time.Duration
	MaxDuration time.Duration
}

// DefaultPostgresConnectionsConfig returns connection pool defaults tuned for
// database failover recovery (e.g. AlloyDB maintenance switchovers).
//
// These are applied by ResolvePostgresConnectionsConfig / ApplyPostgresPoolConfig
// whenever a field on ConnectionsConfig is zero. pgxpool always uses a finite
// MaxConns (unlike database/sql, where MaxOpen=0 means unlimited), so leaving
// MaxOpen unset would otherwise silently fall back to max(4, NumCPU()).
// Explicit defaults keep pool size and eviction policy predictable across
// services that never set Connections.
//
// Short MaxLifetime / MaxIdleTime help the background reaper drop stale
// connections after a primary change; acquire-time liveness is separate
// (pgxpool ShouldPing / PingTimeout).
func DefaultPostgresConnectionsConfig() ConnectionsConfig {
	return ConnectionsConfig{
		MaxOpen:     25,
		MaxIdle:     5,
		MaxLifetime: 5 * time.Minute,
		MaxIdleTime: 2 * time.Minute,
	}
}
