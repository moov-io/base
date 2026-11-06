package database

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// acquireStats collects the connects pgx runs while poolConnector.Connect waits.
type acquireStats struct {
	// pgxpool connects in its own goroutine, which keeps running if the caller
	// gives up, so TraceConnectEnd can race recordAcquire. Each acquire has its
	// own acquireStats, so this only blocks those two, never other requests.
	mu              sync.Mutex
	connectAttempts int
	connectDuration time.Duration
	// Attempts run one at a time, so one start time covers the attempt in flight.
	connectStart time.Time
}

func (s *acquireStats) startConnect(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connectAttempts++
	s.connectStart = now
}

func (s *acquireStats) endConnect(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.connectStart.IsZero() {
		s.connectDuration += now.Sub(s.connectStart)
		s.connectStart = time.Time{}
	}
}

// snapshot includes a connect still in flight, which happens when the caller
// gives up before pgx finishes the handshake.
func (s *acquireStats) snapshot(now time.Time) (int, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	duration := s.connectDuration
	if !s.connectStart.IsZero() {
		duration += now.Sub(s.connectStart)
	}
	return s.connectAttempts, duration
}

type acquireStatsKey struct{}

func withAcquireStats(ctx context.Context) (context.Context, *acquireStats) {
	stats := &acquireStats{}
	return context.WithValue(ctx, acquireStatsKey{}, stats), stats
}

// recordAcquire adds connection timing to the span that requested the connection.
// When a span covers several acquires (retries), the last one wins; span events
// would keep each, but cost an extra event per query. Acquire time includes pinging
// an idle connection, so a slow acquire without a new connection is not always a full pool.
func recordAcquire(ctx context.Context, start time.Time, stats *acquireStats) {
	now := time.Now()
	attempts, connectDuration := stats.snapshot(now)

	// Always set every key so a later acquire on the same span replaces all of an earlier one's values.
	attrs := []attribute.KeyValue{
		attribute.Int64("db.acquire.duration_ms", now.Sub(start).Milliseconds()),
		// True even when every attempt failed: the request tried to open a connection.
		attribute.Bool("db.acquire.new_conn", attempts > 0),
		attribute.Int64("db.acquire.connect_ms", connectDuration.Milliseconds()),
		attribute.Int("db.acquire.connect_attempts", attempts),
	}
	trace.SpanFromContext(ctx).SetAttributes(attrs...)
}

// connectTracer times each new connection pgx opens. pgx requires
// ConnConfig.Tracer to be a QueryTracer, so the query hooks are no-ops.
type connectTracer struct{}

var _ pgx.ConnectTracer = connectTracer{}

func (connectTracer) TraceConnectStart(ctx context.Context, _ pgx.TraceConnectStartData) context.Context {
	// Background connects (MinIdle refills) have no acquireStats to report to.
	if stats, ok := ctx.Value(acquireStatsKey{}).(*acquireStats); ok {
		stats.startConnect(time.Now())
	}
	return ctx
}

func (connectTracer) TraceConnectEnd(ctx context.Context, _ pgx.TraceConnectEndData) {
	if stats, ok := ctx.Value(acquireStatsKey{}).(*acquireStats); ok {
		stats.endConnect(time.Now())
	}
}

func (connectTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	return ctx
}

func (connectTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}
