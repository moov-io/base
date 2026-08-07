package sql_test

import (
	gosql "database/sql"
	"testing"
	"time"

	"github.com/moov-io/base/sql"
	"github.com/stretchr/testify/require"
)

func TestStatsFor_DefaultsToDBStats(t *testing.T) {
	db := &gosql.DB{}
	stats := sql.StatsFor(db)

	// Zero DB has empty stats; important that we don't panic and don't
	// require a registered provider.
	require.Equal(t, 0, stats.Open)
	require.Equal(t, 0, stats.Idle)
	require.Equal(t, 0, stats.InUse)
}

func TestStatsFor_NilDB(t *testing.T) {
	require.Equal(t, sql.ConnectionStats{}, sql.StatsFor(nil))
}

func TestRegisterStatsProvider_OverridesDBStats(t *testing.T) {
	db := &gosql.DB{}
	defer sql.UnregisterStatsProvider(db)

	want := sql.ConnectionStats{
		Idle:                 3,
		InUse:                2,
		Open:                 5,
		WaitCount:            9,
		WaitDuration:         12 * time.Millisecond,
		MaxIdleClosed:        1,
		MaxIdleTimeClosed:    2,
		MaxLifetimeClosed:    3,
		MaxOpen:              25,
		Constructing:         1,
		NewConnsCount:        40,
		CanceledAcquireCount: 7,
	}

	sql.RegisterStatsProvider(db, func() sql.ConnectionStats {
		return want
	})

	require.Equal(t, want, sql.StatsFor(db))
	require.NoError(t, sql.MeasureStats(db, "provider-test"))
}

func TestRegisterStatsProvider_NilArgsIgnored(t *testing.T) {
	db := &gosql.DB{}

	sql.RegisterStatsProvider(nil, func() sql.ConnectionStats { return sql.ConnectionStats{Open: 1} })
	sql.RegisterStatsProvider(db, nil)

	// Still falls back to db.Stats — no provider stored.
	require.Equal(t, 0, sql.StatsFor(db).Open)
}

func TestUnregisterStatsProvider_RestoresDBStats(t *testing.T) {
	db := &gosql.DB{}

	sql.RegisterStatsProvider(db, func() sql.ConnectionStats {
		return sql.ConnectionStats{Open: 99, Idle: 98}
	})
	require.Equal(t, 99, sql.StatsFor(db).Open)

	sql.UnregisterStatsProvider(db)
	require.Equal(t, 0, sql.StatsFor(db).Open)

	// Idempotent.
	sql.UnregisterStatsProvider(db)
	sql.UnregisterStatsProvider(nil)
}

func TestRegisterStatsProvider_Replace(t *testing.T) {
	db := &gosql.DB{}
	defer sql.UnregisterStatsProvider(db)

	sql.RegisterStatsProvider(db, func() sql.ConnectionStats {
		return sql.ConnectionStats{Open: 1}
	})
	sql.RegisterStatsProvider(db, func() sql.ConnectionStats {
		return sql.ConnectionStats{Open: 2}
	})

	require.Equal(t, 2, sql.StatsFor(db).Open)
}
