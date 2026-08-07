package sql

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/moov-io/base/log"
	"github.com/moov-io/base/telemetry"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	statusLock = &sync.Mutex{}

	sqlConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sql_connections",
		Help: "How many MySQL connections and what status they're in.",
	}, []string{"state", "id"})

	sqlConnectionsCounters = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sql_connections_counters",
		Help: `Counters specific to the sql connections.
			wait_count: The total number of connections waited for.
			wait_duration: The total time blocked waiting for a new connection.
			max_idle_closed: The total number of connections closed due to SetMaxIdleConns.
			max_idle_time_closed: The total number of connections closed due to SetConnMaxIdleTime.
			max_lifetime_closed: The total number of connections closed due to SetConnMaxLifetime.
		`,
	}, []string{"counter", "id"})

	sqlQueries = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sql_queries",
		Help:    `Histogram that measures the time in milliseconds queries take`,
		Buckets: []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	}, []string{"operation", "id"})

	sqlErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sql_errors",
		Help: `Histogram that measures the time in milliseconds queries take`,
	}, []string{"id"})

	// Adding in aliases for the usual error cases
	ErrNoRows   = sql.ErrNoRows
	ErrConnDone = sql.ErrConnDone
	ErrTxDone   = sql.ErrTxDone
)

func MonitorSQLDriver(db *sql.DB, id string) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	// Setup metrics after the database is setup
	go func(db *sql.DB, id string) {
		t := time.NewTicker(60 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				MeasureStats(db, id)
			}
		}
	}(db, id)

	return cancel
}

func MeasureStats(db *sql.DB, id string) error {
	statusLock.Lock()
	defer statusLock.Unlock()

	stats := db.Stats()

	sqlConnections.With(prometheus.Labels{"state": "idle", "id": id}).Set(float64(stats.Idle))
	sqlConnections.With(prometheus.Labels{"state": "inuse", "id": id}).Set(float64(stats.InUse))
	sqlConnections.With(prometheus.Labels{"state": "open", "id": id}).Set(float64(stats.OpenConnections))

	sqlConnectionsCounters.With(prometheus.Labels{"counter": "wait_count", "id": id}).Set(float64(stats.WaitCount))
	sqlConnectionsCounters.With(prometheus.Labels{"counter": "wait_ms", "id": id}).Set(float64(stats.WaitDuration.Milliseconds()))
	sqlConnectionsCounters.With(prometheus.Labels{"counter": "max_idle_closed", "id": id}).Set(float64(stats.MaxIdleClosed))
	sqlConnectionsCounters.With(prometheus.Labels{"counter": "max_idle_time_closed", "id": id}).Set(float64(stats.MaxIdleTimeClosed))
	sqlConnectionsCounters.With(prometheus.Labels{"counter": "max_lifetime_closed", "id": id}).Set(float64(stats.MaxLifetimeClosed))

	return nil
}

type LazyLogger func() log.Logger

func MeasureQuery(logger LazyLogger, slowQueryThresholdMs int64, id string, op string, qry string, args int) func() int64 {
	s := time.Now().UnixMilli()

	once := sync.Once{}

	return func() int64 {
		d := int64(-1)

		once.Do(func() {
			d = time.Now().UnixMilli() - s

			sqlQueries.With(prometheus.Labels{"id": id, "operation": op}).Observe(float64(d))

			if d >= slowQueryThresholdMs {
				logger().Warn().With(log.Fields{
					"query":         log.String(CleanQuery(qry)),
					"query_id":      log.String(id),
					"query_op":      log.String(op),
					"query_time_ms": log.Int64(d),
					"query_args":    log.Int(args),
				}).Log("slow query detected")
			}

			// Lazy loggers could self reference, so lets nil it out.
			logger = nil
		})

		return d
	}
}

func MeasureError(id string, err error) error {
	if err != nil && err != ErrNoRows {
		sqlErrors.With(prometheus.Labels{"id": id}).Inc()
	}
	return err
}

func CleanQuery(s string) string {
	cleaner := strings.ReplaceAll(s, "\n", " ")
	cleaner = strings.ReplaceAll(cleaner, "\t", " ")
	cleaner = strings.Trim(cleaner, "\n\t ")

	for {
		spaces := strings.ReplaceAll(cleaner, "  ", " ")

		// Check if it didn't change after the last replace
		if spaces == cleaner {
			break
		}

		cleaner = spaces
	}

	return cleaner
}

func span(ctx context.Context, id string, op string, query string, args int) (context.Context, func()) {
	start := time.Now()

	ctx, span := telemetry.StartSpan(ctx, "sql "+op,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("sql.query", CleanQuery(query)),
			attribute.String("sql.query_id", id),
			attribute.String("sql.query_op", op),
			attribute.Int("sql.query_args", args),
		),
	)

	return ctx, func() {
		took := time.Since(start)
		span.SetAttributes(attribute.Int64("sql.query_time_ms", took.Milliseconds()))

		span.End()
	}
}
