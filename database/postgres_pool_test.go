package database

import (
	"context"
	"database/sql"
	"math"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moov-io/base/log"
	moovsql "github.com/moov-io/base/sql"
	"github.com/stretchr/testify/require"
)

func TestResolvePostgresConnectionsConfig_Defaults(t *testing.T) {
	defaults := DefaultPostgresConnectionsConfig()
	got := ResolvePostgresConnectionsConfig(ConnectionsConfig{})

	require.Equal(t, defaults.MaxOpen, got.MaxOpen)
	require.Equal(t, defaults.MaxLifetime, got.MaxLifetime)
	require.Equal(t, defaults.MaxIdleTime, got.MaxIdleTime)
	// MaxIdle is not defaulted — no pgxpool equivalent.
	require.Equal(t, 0, got.MaxIdle)
}

func TestResolvePostgresConnectionsConfig_Overrides(t *testing.T) {
	in := ConnectionsConfig{
		MaxOpen:     10,
		MaxIdle:     7,
		MaxLifetime: 15 * time.Minute,
		MaxIdleTime: 3 * time.Minute,
	}
	got := ResolvePostgresConnectionsConfig(in)
	require.Equal(t, in, got)
}

func TestResolvePostgresConnectionsConfig_PartialZeros(t *testing.T) {
	defaults := DefaultPostgresConnectionsConfig()
	got := ResolvePostgresConnectionsConfig(ConnectionsConfig{
		MaxOpen: 40,
		// MaxLifetime / MaxIdleTime zero → defaults
		MaxIdle: 3,
	})

	require.Equal(t, 40, got.MaxOpen)
	require.Equal(t, defaults.MaxLifetime, got.MaxLifetime)
	require.Equal(t, defaults.MaxIdleTime, got.MaxIdleTime)
	require.Equal(t, 3, got.MaxIdle)
}

func TestResolvePostgresConnectionsConfig_NegativeTreatedAsUnset(t *testing.T) {
	defaults := DefaultPostgresConnectionsConfig()
	got := ResolvePostgresConnectionsConfig(ConnectionsConfig{
		MaxOpen:     -1,
		MaxLifetime: -time.Second,
		MaxIdleTime: -time.Second,
	})
	require.Equal(t, defaults.MaxOpen, got.MaxOpen)
	require.Equal(t, defaults.MaxLifetime, got.MaxLifetime)
	require.Equal(t, defaults.MaxIdleTime, got.MaxIdleTime)
}

func TestApplyPostgresPoolConfig_Defaults(t *testing.T) {
	poolConfig, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.NoError(t, err)

	// Start from pgxpool's own defaults so we can see our overrides.
	require.NotEqual(t, int32(DefaultPostgresConnectionsConfig().MaxOpen), poolConfig.MaxConns)

	ApplyPostgresPoolConfig(log.NewTestLogger(), poolConfig, ConnectionsConfig{})

	defaults := DefaultPostgresConnectionsConfig()
	require.Equal(t, int32(defaults.MaxOpen), poolConfig.MaxConns)
	require.Equal(t, defaults.MaxLifetime, poolConfig.MaxConnLifetime)
	require.Equal(t, defaults.MaxIdleTime, poolConfig.MaxConnIdleTime)
}

func TestApplyPostgresPoolConfig_Overrides(t *testing.T) {
	poolConfig, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.NoError(t, err)

	in := ConnectionsConfig{
		MaxOpen:     12,
		MaxLifetime: 20 * time.Minute,
		MaxIdleTime: 90 * time.Second,
		MaxIdle:     99, // ignored for pgxpool
	}
	// Capture MinConns / MinIdleConns so we can prove MaxIdle is not mapped.
	minConnsBefore := poolConfig.MinConns
	minIdleBefore := poolConfig.MinIdleConns

	ApplyPostgresPoolConfig(log.NewTestLogger(), poolConfig, in)

	require.Equal(t, int32(12), poolConfig.MaxConns)
	require.Equal(t, 20*time.Minute, poolConfig.MaxConnLifetime)
	require.Equal(t, 90*time.Second, poolConfig.MaxConnIdleTime)
	require.Equal(t, minConnsBefore, poolConfig.MinConns)
	require.Equal(t, minIdleBefore, poolConfig.MinIdleConns)
}

func TestApplyPostgresPoolConfig_ClampsMaxOpenToInt32(t *testing.T) {
	poolConfig, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.NoError(t, err)

	ApplyPostgresPoolConfig(log.NewTestLogger(), poolConfig, ConnectionsConfig{
		MaxOpen: math.MaxInt32 + 1,
	})

	require.Equal(t, int32(math.MaxInt32), poolConfig.MaxConns)
}

func TestApplyPostgresPoolConfig_NilPoolConfig(t *testing.T) {
	require.NotPanics(t, func() {
		ApplyPostgresPoolConfig(log.NewTestLogger(), nil, ConnectionsConfig{MaxOpen: 5})
	})
}

func TestOpenDBFromPool_CloseClosesPool(t *testing.T) {
	// NewWithConfig does not dial until a connection is needed; use an
	// unreachable address so this stays a pure unit test.
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.NoError(t, err)
	cfg.MaxConns = 2

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)

	db := openDBFromPool(pool, nil)
	require.NoError(t, db.Close())

	// Closing the sql.DB must close the underlying pgxpool.
	_, err = pool.Acquire(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "closed")
}

func TestOpenDBFromPool_CloseIdempotent(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	require.NoError(t, err)

	db := openDBFromPool(pool, nil)
	require.NoError(t, db.Close())
	require.NoError(t, db.Close())

	// Pool stays closed after repeated DB.Close.
	_, err = pool.Acquire(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "closed")
}

func TestOpenDBFromPool_SetsMaxIdleConnsZero(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	require.NoError(t, err)
	defer pool.Close()

	db := openDBFromPool(pool, nil)
	defer db.Close()

	// With MaxIdleConns=0, database/sql should not retain idle connections
	// in its own pool (OpenConns drops to 0 when unused).
	require.Equal(t, 0, db.Stats().OpenConnections)
	require.Equal(t, 0, db.Stats().Idle)
}

func TestOpenDBFromPool_RegistersStatsProvider(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.NoError(t, err)
	cfg.MaxConns = 7

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	require.NoError(t, err)

	db := openDBFromPool(pool, nil)

	// db.Stats() is the useless sql.DB handoff view.
	require.Equal(t, 0, db.Stats().OpenConnections)

	// StatsFor / PoolStat must surface the real pgxpool.
	stat, ok := PoolStat(db)
	require.True(t, ok)
	require.Equal(t, int32(7), stat.MaxConns())

	got := moovsql.StatsFor(db)
	require.Equal(t, 7, got.MaxOpen)
	require.Equal(t, int(stat.TotalConns()), got.Open)
	require.Equal(t, int(stat.IdleConns()), got.Idle)
	require.Equal(t, int(stat.AcquiredConns()), got.InUse)
	require.NoError(t, moovsql.MeasureStats(db, "pgx-pool-test"))

	require.NoError(t, db.Close())

	// After Close, provider and pool registry are cleared.
	_, ok = PoolStat(db)
	require.False(t, ok)
	require.Equal(t, 0, moovsql.StatsFor(db).MaxOpen)
}

func TestConnectionStatsFromPgx_NilSafe(t *testing.T) {
	require.Equal(t, moovsql.ConnectionStats{}, connectionStatsFromPgx(nil))
}

func TestConnectionStatsFromPgx_MapsFields(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.NoError(t, err)
	cfg.MaxConns = 4

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	require.NoError(t, err)
	defer pool.Close()

	raw := pool.Stat()
	got := connectionStatsFromPgx(raw)

	require.Equal(t, int(raw.IdleConns()), got.Idle)
	require.Equal(t, int(raw.AcquiredConns()), got.InUse)
	require.Equal(t, int(raw.TotalConns()), got.Open)
	require.Equal(t, raw.EmptyAcquireCount(), got.WaitCount)
	require.Equal(t, raw.EmptyAcquireWaitTime(), got.WaitDuration)
	require.Equal(t, raw.MaxIdleDestroyCount(), got.MaxIdleTimeClosed)
	require.Equal(t, raw.MaxLifetimeDestroyCount(), got.MaxLifetimeClosed)
	require.Equal(t, int(raw.MaxConns()), got.MaxOpen)
	require.Equal(t, int(raw.ConstructingConns()), got.Constructing)
	require.Equal(t, raw.NewConnsCount(), got.NewConnsCount)
	require.Equal(t, raw.CanceledAcquireCount(), got.CanceledAcquireCount)
	// No pgxpool equivalent.
	require.Equal(t, int64(0), got.MaxIdleClosed)
}

func TestPoolStat_NilAndUnknown(t *testing.T) {
	_, ok := PoolStat(nil)
	require.False(t, ok)

	_, ok = PoolStat(&sql.DB{})
	require.False(t, ok)
}
