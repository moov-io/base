package sql

import (
	"database/sql"
	"sync"
	"time"
)

// ConnectionStats is a snapshot of pool pressure used by MeasureStats.
//
// database/sql-backed pools populate this from db.Stats(). Drivers that pool
// outside database/sql (for example pgxpool under OpenDBFromPool with
// MaxIdleConns=0) should RegisterStatsProvider so metrics reflect the real
// pool instead of the empty sql.DB handoff layer.
type ConnectionStats struct {
	Idle              int
	InUse             int
	Open              int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxIdleClosed     int64
	MaxIdleTimeClosed int64
	MaxLifetimeClosed int64

	// Extended fields. Zero when the underlying pool does not expose them.
	MaxOpen              int
	Constructing         int
	NewConnsCount        int64
	CanceledAcquireCount int64
}

// StatsProvider returns a ConnectionStats snapshot for a *sql.DB.
type StatsProvider func() ConnectionStats

// statsProviders maps *sql.DB → StatsProvider for pools that do not surface
// meaningful values through database/sql.DB.Stats.
var statsProviders sync.Map

// RegisterStatsProvider attaches a custom stats source for db. MeasureStats
// and MonitorSQLDriver will use it instead of db.Stats(). A nil db or
// provider is ignored. Re-registering replaces the previous provider.
func RegisterStatsProvider(db *sql.DB, provider StatsProvider) {
	if db == nil || provider == nil {
		return
	}
	statsProviders.Store(db, provider)
}

// UnregisterStatsProvider removes a previously registered provider.
// Safe to call when none is registered.
func UnregisterStatsProvider(db *sql.DB) {
	if db == nil {
		return
	}
	statsProviders.Delete(db)
}

// StatsFor returns pool stats for db, preferring a registered provider.
func StatsFor(db *sql.DB) ConnectionStats {
	if db == nil {
		return ConnectionStats{}
	}
	if v, ok := statsProviders.Load(db); ok {
		if provider, ok := v.(StatsProvider); ok && provider != nil {
			return provider()
		}
	}
	return connectionStatsFromDB(db.Stats())
}

func connectionStatsFromDB(s sql.DBStats) ConnectionStats {
	return ConnectionStats{
		Idle:              s.Idle,
		InUse:             s.InUse,
		Open:              s.OpenConnections,
		WaitCount:         s.WaitCount,
		WaitDuration:      s.WaitDuration,
		MaxIdleClosed:     s.MaxIdleClosed,
		MaxIdleTimeClosed: s.MaxIdleTimeClosed,
		MaxLifetimeClosed: s.MaxLifetimeClosed,
		MaxOpen:           s.MaxOpenConnections,
	}
}
