package database

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// startSpan returns a recording span context and a func that ends the span and returns its attributes.
func startSpan(t *testing.T) (context.Context, func() map[attribute.Key]attribute.Value) {
	t.Helper()
	ctx, end := startSpanWithEvents(t)
	return ctx, func() map[attribute.Key]attribute.Value {
		attrs, _ := end()
		return attrs
	}
}

func startSpanWithEvents(t *testing.T) (context.Context, func() (map[attribute.Key]attribute.Value, []sdktrace.Event)) {
	t.Helper()
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	ctx, span := tp.Tracer("test").Start(context.Background(), "sql query")
	return ctx, func() (map[attribute.Key]attribute.Value, []sdktrace.Event) {
		span.End()
		ended := recorder.Ended()
		require.Len(t, ended, 1)
		return attrMap(ended[0].Attributes()), ended[0].Events()
	}
}

func attrMap(kvs []attribute.KeyValue) map[attribute.Key]attribute.Value {
	attrs := map[attribute.Key]attribute.Value{}
	for _, kv := range kvs {
		attrs[kv.Key] = kv.Value
	}
	return attrs
}

// requireAcquireAttrs checks the attributes are present, since a missing bool reads as false.
func requireAcquireAttrs(t *testing.T, attrs map[attribute.Key]attribute.Value, newConn bool) {
	t.Helper()
	for _, key := range []attribute.Key{"db.acquire.duration_ms", "db.acquire.new_conn", "db.acquire.connect_ms", "db.acquire.connect_attempts"} {
		require.Contains(t, attrs, key)
	}
	require.Equal(t, newConn, attrs["db.acquire.new_conn"].AsBool())
	if !newConn {
		require.Zero(t, attrs["db.acquire.connect_ms"].AsInt64())
		require.Zero(t, attrs["db.acquire.connect_attempts"].AsInt64())
	}
}

// hungPostgres returns the address of a server that accepts TCP and never answers the startup message.
func hungPostgres(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			t.Cleanup(func() { conn.Close() })
		}
	}()
	return ln.Addr().String()
}

func TestConnectTracer_AddsToAcquireStats(t *testing.T) {
	ctx, stats := withAcquireStats(context.Background())

	tracer := connectTracer{}
	connectCtx := tracer.TraceConnectStart(ctx, pgx.TraceConnectStartData{})
	time.Sleep(5 * time.Millisecond)
	tracer.TraceConnectEnd(connectCtx, pgx.TraceConnectEndData{})

	attempts, duration := stats.snapshot(time.Now())
	require.Equal(t, 1, attempts)
	require.GreaterOrEqual(t, duration, 5*time.Millisecond)
}

func TestConnectTracer_IgnoresBackgroundConnects(t *testing.T) {
	tracer := connectTracer{}
	connectCtx := tracer.TraceConnectStart(context.Background(), pgx.TraceConnectStartData{})
	require.NotPanics(t, func() { tracer.TraceConnectEnd(connectCtx, pgx.TraceConnectEndData{}) })
}

func TestAcquireStats_SnapshotIncludesConnectInFlight(t *testing.T) {
	stats := &acquireStats{}
	start := time.Now()
	stats.startConnect(start)

	attempts, duration := stats.snapshot(start.Add(200 * time.Millisecond))
	require.Equal(t, 1, attempts)
	require.Equal(t, 200*time.Millisecond, duration)
}

func TestAcquireStats_ConcurrentEndAndSnapshot(t *testing.T) {
	stats := &acquireStats{}
	stats.startConnect(time.Now())

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); stats.endConnect(time.Now()) }()
	go func() { defer wg.Done(); stats.snapshot(time.Now()) }()
	wg.Wait()

	attempts, _ := stats.snapshot(time.Now())
	require.Equal(t, 1, attempts)
}

func TestRecordAcquire_ReusedConnection(t *testing.T) {
	ctx, end := startSpan(t)
	ctx, stats := withAcquireStats(ctx)

	recordAcquire(ctx, time.Now().Add(-3*time.Millisecond), stats)

	attrs := end()
	requireAcquireAttrs(t, attrs, false)
	require.GreaterOrEqual(t, attrs["db.acquire.duration_ms"].AsInt64(), int64(3))
}

func TestRecordAcquire_NewConnection(t *testing.T) {
	ctx, end := startSpan(t)
	ctx, stats := withAcquireStats(ctx)
	now := time.Now()
	stats.startConnect(now.Add(-60 * time.Millisecond))
	stats.endConnect(now.Add(-40 * time.Millisecond))
	stats.startConnect(now.Add(-40 * time.Millisecond))
	stats.endConnect(now.Add(-10 * time.Millisecond))

	recordAcquire(ctx, now.Add(-65*time.Millisecond), stats)

	attrs := end()
	requireAcquireAttrs(t, attrs, true)
	require.GreaterOrEqual(t, attrs["db.acquire.duration_ms"].AsInt64(), int64(65))
	require.Equal(t, int64(50), attrs["db.acquire.connect_ms"].AsInt64())
	require.Equal(t, int64(2), attrs["db.acquire.connect_attempts"].AsInt64())
}

func TestRecordAcquire_EventPerAcquire(t *testing.T) {
	ctx, end := startSpanWithEvents(t)

	slowCtx, slow := withAcquireStats(ctx)
	slow.startConnect(time.Now().Add(-time.Second))
	slow.endConnect(time.Now())
	recordAcquire(slowCtx, time.Now().Add(-time.Second), slow)

	fastCtx, fast := withAcquireStats(ctx)
	recordAcquire(fastCtx, time.Now(), fast)

	attrs, events := end()
	requireAcquireAttrs(t, attrs, false)
	require.Len(t, events, 2)
	require.Equal(t, "db.acquire", events[0].Name)
	require.True(t, attrMap(events[0].Attributes)["db.acquire.new_conn"].AsBool())
	require.False(t, attrMap(events[1].Attributes)["db.acquire.new_conn"].AsBool())
}

func TestRecordAcquire_NoSpan(t *testing.T) {
	ctx, stats := withAcquireStats(context.Background())
	require.NotPanics(t, func() { recordAcquire(ctx, time.Now(), stats) })
}

func TestOpenDBFromPool_CanceledDuringConnectCountsAttempt(t *testing.T) {
	cfg := PostgresConfig{Address: hungPostgres(t), User: "user", TLS: &PostgresTLSConfig{Mode: "disable"}}
	poolConfig, _, err := buildPgxPoolConfig(context.Background(), cfg, "db")
	require.NoError(t, err)
	poolConfig.ConnConfig.Tracer = connectTracer{}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	require.NoError(t, err)
	db := openDBFromPool(pool, nil)
	t.Cleanup(func() { db.Close() })

	spanCtx, end := startSpan(t)
	ctx, cancel := context.WithTimeout(spanCtx, 100*time.Millisecond)
	defer cancel()
	require.Error(t, db.PingContext(ctx))

	attrs := end()
	requireAcquireAttrs(t, attrs, true)
	require.Equal(t, int64(1), attrs["db.acquire.connect_attempts"].AsInt64())
	require.GreaterOrEqual(t, attrs["db.acquire.connect_ms"].AsInt64(), int64(90))
}

func TestPostgres_AcquireSpanAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	minIdle := 0
	db, err := New(context.Background(), log.NewTestLogger(), DatabaseConfig{
		DatabaseName: "moov",
		Postgres: &PostgresConfig{
			Address:     "127.0.0.1:5432",
			User:        "moov",
			Password:    "moov",
			TLS:         &PostgresTLSConfig{Mode: "disable"},
			Connections: ConnectionsConfig{MinIdle: &minIdle},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	// New pings once, which leaves one idle connection in the pool.
	ctx, end := startSpan(t)
	require.NoError(t, db.PingContext(ctx))
	requireAcquireAttrs(t, end(), false)

	// Holding that connection makes the next request open a new one.
	held, err := db.Conn(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { held.Close() })

	ctx, end = startSpan(t)
	require.NoError(t, db.PingContext(ctx))
	attrs := end()
	requireAcquireAttrs(t, attrs, true)
	require.Equal(t, int64(1), attrs["db.acquire.connect_attempts"].AsInt64())
	require.GreaterOrEqual(t, attrs["db.acquire.duration_ms"].AsInt64(), attrs["db.acquire.connect_ms"].AsInt64())
}
