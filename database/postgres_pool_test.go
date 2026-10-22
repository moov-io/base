package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"math"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func TestResolvePostgresConnectionsConfig_Defaults(t *testing.T) {
	defaults := DefaultPostgresConnectionsConfig()
	got := ResolvePostgresConnectionsConfig(ConnectionsConfig{})

	require.Equal(t, defaults.MaxOpen, got.MaxOpen)
	require.Equal(t, defaults.MaxLifetime, got.MaxLifetime)
	require.Equal(t, defaults.MaxIdleTime, got.MaxIdleTime)
	require.Equal(t, defaults.MinIdle, got.MinIdle)
	// MaxIdle is not defaulted — no pgxpool equivalent.
	require.Equal(t, 0, got.MaxIdle)
}

func TestResolvePostgresConnectionsConfig_Overrides(t *testing.T) {
	in := ConnectionsConfig{
		MaxOpen:     10,
		MaxIdle:     7,
		MinIdle:     intPtr(2),
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
	require.Equal(t, time.Duration(0), poolConfig.MaxConnLifetimeJitter)

	ApplyPostgresPoolConfig(log.NewTestLogger(), poolConfig, ConnectionsConfig{})

	defaults := DefaultPostgresConnectionsConfig()
	require.Equal(t, int32(defaults.MaxOpen), poolConfig.MaxConns)
	require.Equal(t, defaults.MaxLifetime, poolConfig.MaxConnLifetime)
	require.Equal(t, defaults.MaxIdleTime, poolConfig.MaxConnIdleTime)
	// pgxpool does not default jitter; we fill it so lifetime recycle is staggered.
	require.Equal(t, defaultPostgresMaxConnLifetimeJitter, poolConfig.MaxConnLifetimeJitter)
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
	// Capture MinConns so we can prove MaxIdle is not mapped.
	minConnsBefore := poolConfig.MinConns

	ApplyPostgresPoolConfig(log.NewTestLogger(), poolConfig, in)

	require.Equal(t, int32(12), poolConfig.MaxConns)
	require.Equal(t, 20*time.Minute, poolConfig.MaxConnLifetime)
	require.Equal(t, 90*time.Second, poolConfig.MaxConnIdleTime)
	require.Equal(t, minConnsBefore, poolConfig.MinConns)
	// MinIdleConns gets the MinIdle default, not MaxIdle.
	require.Equal(t, int32(2), poolConfig.MinIdleConns)
	// 10% of 20m = 2m > fixed 30s default.
	require.Equal(t, 2*time.Minute, poolConfig.MaxConnLifetimeJitter)
}

func TestApplyPostgresPoolConfig_MinIdle(t *testing.T) {
	cases := []struct {
		name string
		dsn  string
		in   ConnectionsConfig
		want int32
	}{
		{name: "unset uses default", in: ConnectionsConfig{}, want: 2},
		{name: "explicit zero disables", in: ConnectionsConfig{MinIdle: intPtr(0)}, want: 0},
		{name: "negative treated as zero", in: ConnectionsConfig{MinIdle: intPtr(-1)}, want: 0},
		{name: "set", in: ConnectionsConfig{MinIdle: intPtr(4)}, want: 4},
		{name: "capped at MaxOpen", in: ConnectionsConfig{MaxOpen: 3, MinIdle: intPtr(4)}, want: 3},
		{name: "replaces DSN value", dsn: "&pool_min_idle_conns=5", in: ConnectionsConfig{MinIdle: intPtr(2)}, want: 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			poolConfig, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/db?sslmode=disable" + tc.dsn)
			require.NoError(t, err)

			ApplyPostgresPoolConfig(log.NewTestLogger(), poolConfig, tc.in)

			require.Equal(t, tc.want, poolConfig.MinIdleConns)
		})
	}
}

func intPtr(v int) *int { return &v }

func TestApplyPostgresPoolConfig_PreservesConfiguredJitter(t *testing.T) {
	poolConfig, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.NoError(t, err)
	poolConfig.MaxConnLifetimeJitter = 45 * time.Second

	ApplyPostgresPoolConfig(log.NewTestLogger(), poolConfig, ConnectionsConfig{})

	require.Equal(t, 45*time.Second, poolConfig.MaxConnLifetimeJitter)
}

func TestDefaultPostgresPingTimeout(t *testing.T) {
	// Document the acquire worst-case: (MaxOpen+1) * PingTimeout should stay
	// well under a long request budget for blackholed peers.
	require.Equal(t, 300*time.Millisecond, defaultPostgresPingTimeout)
	maxOpen := DefaultPostgresConnectionsConfig().MaxOpen
	worstCase := time.Duration(maxOpen+1) * defaultPostgresPingTimeout
	require.LessOrEqual(t, worstCase, 10*time.Second)
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

func TestClosePgxPool_EmptyPool(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	require.NoError(t, err)

	start := time.Now()
	require.NoError(t, closePgxPool(pool, defaultPostgresPoolCloseWait))
	require.Less(t, time.Since(start), 2*time.Second)

	_, err = pool.Acquire(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "closed")
}

func TestClosePgxPool_NilSafe(t *testing.T) {
	require.NoError(t, closePgxPool(nil, time.Second))
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

func TestPoolDBStats_FromOpenDBFromPool(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.NoError(t, err)
	cfg.MaxConns = 9

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	require.NoError(t, err)

	db := openDBFromPool(pool, nil)

	// sql.DB handoff layer stays empty.
	require.Equal(t, 0, db.Stats().OpenConnections)
	require.Equal(t, 0, db.Stats().MaxOpenConnections)

	// PoolDBStats reflects the real pgxpool.
	got, ok := PoolDBStats(db)
	require.True(t, ok)
	require.Equal(t, 9, got.MaxOpenConnections)
	require.Equal(t, int(pool.Stat().TotalConns()), got.OpenConnections)
	require.Equal(t, int(pool.Stat().IdleConns()), got.Idle)
	require.Equal(t, int(pool.Stat().AcquiredConns()), got.InUse)

	require.NoError(t, db.Close())

	_, ok = PoolDBStats(db)
	require.False(t, ok)
}

func TestPoolDBStats_UnknownDB(t *testing.T) {
	_, ok := PoolDBStats(nil)
	require.False(t, ok)

	_, ok = PoolDBStats(&sql.DB{})
	require.False(t, ok)
}

func TestPgxPoolStatToDBStats_NilSafe(t *testing.T) {
	require.Equal(t, sql.DBStats{}, pgxPoolStatToDBStats(nil))
}

type fakeConnector struct {
	errs  []error
	calls int
}

func (f *fakeConnector) Connect(context.Context) (driver.Conn, error) {
	f.calls++
	if len(f.errs) == 0 {
		return nil, nil
	}
	err := f.errs[0]
	f.errs = f.errs[1:]
	return nil, err
}

func (f *fakeConnector) Driver() driver.Driver { return nil }

func TestPoolConnector_RetriesConnectErrorOnce(t *testing.T) {
	// A real *pgconn.ConnectError: nothing listens on port 1.
	_, connectErr := pgconn.Connect(context.Background(), "postgres://user@127.0.0.1:1/db?sslmode=disable&connect_timeout=1")
	var ce *pgconn.ConnectError
	require.ErrorAs(t, connectErr, &ce)

	otherErr := errors.New("other")
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	cases := []struct {
		name      string
		ctx       context.Context
		errs      []error
		wantErr   error
		wantCalls int
	}{
		{name: "success", ctx: context.Background(), wantCalls: 1},
		{name: "connect error then success", ctx: context.Background(), errs: []error{connectErr}, wantCalls: 2},
		{name: "wrapped connect error then success", ctx: context.Background(), errs: []error{fmt.Errorf("acquire: %w", connectErr)}, wantCalls: 2},
		{name: "connect error twice", ctx: context.Background(), errs: []error{connectErr, connectErr}, wantErr: connectErr, wantCalls: 2},
		{name: "other error is not retried", ctx: context.Background(), errs: []error{otherErr}, wantErr: otherErr, wantCalls: 1},
		{name: "canceled context is not retried", ctx: canceled, errs: []error{connectErr}, wantErr: connectErr, wantCalls: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeConnector{errs: tc.errs}
			c := &poolConnector{Connector: fake}

			_, err := c.Connect(tc.ctx)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.wantCalls, fake.calls)
		})
	}
}

func TestBuildPgxPoolConfig_DefaultConnectTimeout(t *testing.T) {
	cfg := PostgresConfig{Address: "127.0.0.1:1", User: "user", TLS: &PostgresTLSConfig{Mode: "disable"}}

	poolConfig, dialer, err := buildPgxPoolConfig(context.Background(), cfg, "db")
	require.NoError(t, err)
	require.Nil(t, dialer)
	require.Equal(t, defaultPostgresConnectTimeout, poolConfig.ConnConfig.ConnectTimeout)
}

func TestOpenDBFromPool_HungHandshakeTimesOutAndRetries(t *testing.T) {
	// A server that accepts TCP and never answers the startup message.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })

	var accepted atomic.Int32
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			accepted.Add(1)
			t.Cleanup(func() { conn.Close() })
		}
	}()

	cfg := PostgresConfig{Address: ln.Addr().String(), User: "user", TLS: &PostgresTLSConfig{Mode: "disable"}}
	poolConfig, _, err := buildPgxPoolConfig(context.Background(), cfg, "db")
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	require.NoError(t, err)
	db := openDBFromPool(pool, nil)
	t.Cleanup(func() { db.Close() })

	start := time.Now()
	err = db.PingContext(context.Background())
	elapsed := time.Since(start)

	var connectErr *pgconn.ConnectError
	require.ErrorAs(t, err, &connectErr)
	require.Equal(t, int32(2), accepted.Load(), "one attempt and one retry")
	require.GreaterOrEqual(t, elapsed, 2*defaultPostgresConnectTimeout)
	require.Less(t, elapsed, 2*defaultPostgresConnectTimeout+500*time.Millisecond)
}
