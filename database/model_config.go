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

// PostgresConfig configures a Postgres or AlloyDB connection opened by New.
//
// Pool behavior (pgxpool under *sql.DB):
//   - Connections settings are applied to the underlying pgxpool, not database/sql.
//   - Zero-valued fields on Connections are filled from DefaultPostgresConnectionsConfig
//     (see that function for operator-facing defaults and upgrade notes).
//   - Connections.MaxIdle has no pgxpool equivalent and is ignored (logged when set).
//   - Do not call sql.DB SetMaxIdleConns with a non-zero value on the returned DB;
//     OpenDBFromPool requires MaxIdleConns=0 so connections return to pgxpool.
//   - db.Stats() is not meaningful for capacity monitoring (sql.DB does not hold
//     the pool). New starts a ticker that records pgxpool.Stat into the
//     postgres_connections* Prometheus metrics (same idea as mysql_connections).
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

// ConnectionsConfig tunes the database connection pool.
//
// For MySQL these map to database/sql setters via ApplyConnectionsConfig.
// For Postgres/AlloyDB they map to pgxpool via ApplyPostgresPoolConfig:
//
//	MaxOpen     -> pgxpool MaxConns (required; pgxpool has no "unlimited")
//	MaxLifetime  -> pgxpool MaxConnLifetime
//	MaxIdleTime -> pgxpool MaxConnIdleTime
//	MaxIdle     -> ignored (no pgxpool equivalent; logged when > 0)
//
// Zero means "unset". For Postgres, unset fields are filled from
// DefaultPostgresConnectionsConfig — unlike database/sql, 0 does not mean unlimited.
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

// DefaultPostgresConnectionsConfig returns library defaults applied when a
// Postgres/AlloyDB ConnectionsConfig field is zero.
//
// # Operator notes
//
// These defaults change runtime behavior versus the previous database/sql +
// pgx driver path, where unset fields meant unlimited:
//
//   - MaxOpen (25): pgxpool always has a finite MaxConns. Unset MaxOpen no
//     longer means unlimited — it becomes 25. High-concurrency services must
//     set PostgresConfig.Connections.MaxOpen explicitly if they need more.
//   - MaxLifetime (5m): connections are recycled after this age (pgxpool
//     adds jitter). Unset previously meant no max lifetime. Shorter lifetime
//     helps drop stale sockets after AlloyDB failover; increase if reconnect
//     cost (for example AlloyDB IAM) is more expensive than churn.
//   - MaxIdleTime (2m): idle connections are closed after this duration.
//     Unset previously meant keep idle forever. Quiet processes will
//     reconnect after idle gaps; raise this if cold-connect latency matters.
//   - MaxIdle: not defaulted and not applied for Postgres (no pgxpool
//     equivalent). Prefer MaxOpen to bound total connections.
//
// Override any value via PostgresConfig.Connections. There is no sentinel for
// "unlimited" lifetime or idle time under pgxpool — leave a field zero only
// when the default above is acceptable.
func DefaultPostgresConnectionsConfig() ConnectionsConfig {
	return ConnectionsConfig{
		MaxOpen:     25,
		MaxLifetime: 5 * time.Minute,
		MaxIdleTime: 2 * time.Minute,
	}
}
