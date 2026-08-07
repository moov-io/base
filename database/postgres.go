package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/alloydbconn"
	kitprom "github.com/go-kit/kit/metrics/prometheus"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/moov-io/base/log"
	stdprom "github.com/prometheus/client_golang/prometheus"
)

const (
	// PostgreSQL Error Codes
	// https://www.postgresql.org/docs/current/errcodes-appendix.html
	postgresErrUniqueViolation = "23505"
	postgresErrDeadlockFound   = "40P01"

	// Bound ShouldPing wait so acquire retries during failover stay inside
	// typical request budgets. AlloyDB disconnects usually fail fast (TCP RST);
	// this caps hung/TIME_WAIT peers.
	defaultPostgresPingTimeout = time.Second
)

var (
	postgresMetricsMu = &sync.Mutex{}

	// Same pattern as mysql_connections: scrape pgxpool.Stat on a ticker.
	// db.Stats() is not useful here — OpenDBFromPool keeps sql.DB MaxIdleConns=0
	// so the database/sql layer does not retain connections.
	postgresConnections = kitprom.NewGaugeFrom(stdprom.GaugeOpts{
		Name: "postgres_connections",
		Help: "How many Postgres/pgxpool connections and what status they're in.",
	}, []string{"state"})

	postgresConnectionsCounters = kitprom.NewGaugeFrom(stdprom.GaugeOpts{
		Name: "postgres_connections_counters",
		Help: `Counters specific to the Postgres/pgxpool connections.
			wait_count: Successful acquires that waited because the pool was empty.
			wait_ms: Cumulative time spent waiting for those acquires (milliseconds).
			max_idle_time_closed: Connections destroyed due to MaxConnIdleTime.
			max_lifetime_closed: Connections destroyed due to MaxConnLifetime.
			max_open: Configured MaxConns.
			new_conns: Cumulative new connections opened.
			canceled_acquire: Cumulative acquires canceled by context.
		`,
	}, []string{"counter"})
)

// RecordPostgresPoolStats writes pgxpool.Stat into postgres_connections* gauges.
// Mirrors RecordMySQLStats for the Postgres/AlloyDB path.
func RecordPostgresPoolStats(pool *pgxpool.Pool) error {
	if pool == nil {
		return nil
	}

	s := pool.Stat()

	postgresMetricsMu.Lock()
	defer postgresMetricsMu.Unlock()

	postgresConnections.With("state", "idle").Set(float64(s.IdleConns()))
	postgresConnections.With("state", "inuse").Set(float64(s.AcquiredConns()))
	postgresConnections.With("state", "open").Set(float64(s.TotalConns()))
	postgresConnections.With("state", "constructing").Set(float64(s.ConstructingConns()))

	postgresConnectionsCounters.With("counter", "wait_count").Set(float64(s.EmptyAcquireCount()))
	postgresConnectionsCounters.With("counter", "wait_ms").Set(float64(s.EmptyAcquireWaitTime().Milliseconds()))
	postgresConnectionsCounters.With("counter", "max_idle_time_closed").Set(float64(s.MaxIdleDestroyCount()))
	postgresConnectionsCounters.With("counter", "max_lifetime_closed").Set(float64(s.MaxLifetimeDestroyCount()))
	postgresConnectionsCounters.With("counter", "max_open").Set(float64(s.MaxConns()))
	postgresConnectionsCounters.With("counter", "new_conns").Set(float64(s.NewConnsCount()))
	postgresConnectionsCounters.With("counter", "canceled_acquire").Set(float64(s.CanceledAcquireCount()))

	return nil
}

func monitorPostgresPool(ctx context.Context, pool *pgxpool.Pool) {
	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				_ = RecordPostgresPoolStats(pool)
			}
		}
	}()
}

func postgresConnection(ctx context.Context, logger log.Logger, config PostgresConfig, databaseName string) (*sql.DB, error) {
	poolConfig, dialer, err := buildPgxPoolConfig(ctx, config, databaseName)
	if err != nil {
		return nil, logger.LogErrorf("building pgx pool config: %w", err).Err()
	}

	// Apply connection limits to pgxpool (not database/sql). OpenDBFromPool
	// requires sql.DB MaxIdleConns=0; sql.DB setters do not configure the
	// underlying pool and SetMaxIdleConns(n>0) actively breaks it.
	ApplyPostgresPoolConfig(logger, poolConfig, config.Connections)

	// Ping connections that have been idle for more than 200ms before handing
	// them to the caller. This catches dead connections left by an AlloyDB
	// switchover before a query is attempted, without adding overhead on
	// hot connections used moments ago.
	// HealthCheckPeriod (the background reaper) does NOT ping — it only evicts
	// connections that have exceeded their age thresholds. ShouldPing is the
	// mechanism that actually tests liveness at acquire time.
	poolConfig.ShouldPing = func(_ context.Context, p pgxpool.ShouldPingParams) bool {
		return p.IdleDuration > 200*time.Millisecond
	}

	if poolConfig.PingTimeout <= 0 {
		poolConfig.PingTimeout = defaultPostgresPingTimeout
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = closeAlloyDialer(dialer)
		return nil, logger.LogErrorf("creating pgx pool: %w", err).Err()
	}

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		_ = closeAlloyDialer(dialer)
		return nil, logger.LogErrorf("connecting to database: %w", err).Err()
	}

	// OpenDBFromPool does not close the pool when *sql.DB is closed. Wrap the
	// connector so db.Close() shuts down the pool (and AlloyDB dialer).
	db := openDBFromPool(pool, dialer)

	// Collect pool metrics from pgxpool.Stat (not db.Stats()).
	monitorPostgresPool(ctx, pool)

	return db, nil
}

// ApplyPostgresPoolConfig fills zero-valued fields in connections with
// DefaultPostgresConnectionsConfig, then maps them onto poolConfig.
//
// Unlike database/sql (where MaxOpen=0 means unlimited), pgxpool always has a
// finite MaxConns. Leaving MaxOpen unset previously fell through to pgxpool's
// default of max(4, NumCPU()), which silently shrinks pools for services that
// never configured Connections. We instead apply explicit library defaults so
// behavior is predictable and logged.
//
// MaxIdle has no pgxpool equivalent — pgxpool caps total connections via MaxConns
// rather than idle count. When set, MaxIdle is logged and ignored so operators
// aren't misled into thinking it took effect.
func ApplyPostgresPoolConfig(logger log.Logger, poolConfig *pgxpool.Config, connections ConnectionsConfig) {
	if poolConfig == nil {
		return
	}

	applied := ResolvePostgresConnectionsConfig(connections)

	// pgxpool.MaxConns is int32; clamp to avoid gosec G115 on int->int32.
	maxOpen := applied.MaxOpen
	if maxOpen > math.MaxInt32 {
		logger.Logf("clamping pgx pool MaxConns from %d to %d", maxOpen, math.MaxInt32)
		maxOpen = math.MaxInt32
	}
	logger.Logf("setting pgx pool MaxConns to %d", maxOpen)
	poolConfig.MaxConns = int32(maxOpen)

	if applied.MaxIdle > 0 {
		logger.Logf("ignoring ConnectionsConfig.MaxIdle=%d: pgxpool has no MaxIdle equivalent", applied.MaxIdle)
	}

	logger.Logf("setting pgx pool MaxConnIdleTime to %v", applied.MaxIdleTime)
	poolConfig.MaxConnIdleTime = applied.MaxIdleTime

	logger.Logf("setting pgx pool MaxConnLifetime to %v", applied.MaxLifetime)
	poolConfig.MaxConnLifetime = applied.MaxLifetime
}

// ResolvePostgresConnectionsConfig returns connections with zero-valued fields
// replaced by DefaultPostgresConnectionsConfig.
func ResolvePostgresConnectionsConfig(connections ConnectionsConfig) ConnectionsConfig {
	defaults := DefaultPostgresConnectionsConfig()
	if connections.MaxOpen <= 0 {
		connections.MaxOpen = defaults.MaxOpen
	}
	if connections.MaxLifetime <= 0 {
		connections.MaxLifetime = defaults.MaxLifetime
	}
	if connections.MaxIdleTime <= 0 {
		connections.MaxIdleTime = defaults.MaxIdleTime
	}
	return connections
}

// openDBFromPool wraps pgxpool in a *sql.DB whose Close also closes the pool
// and optional AlloyDB dialer. stdlib.OpenDBFromPool alone leaks both.
func openDBFromPool(pool *pgxpool.Pool, dialer *alloydbconn.Dialer) *sql.DB {
	c := &poolConnector{
		Connector: stdlib.GetPoolConnector(pool),
		pool:      pool,
		dialer:    dialer,
	}
	db := sql.OpenDB(c)
	// Required when using a pgxpool-backed connector: non-zero idle conns on
	// sql.DB prevent connections from being released back to the pool.
	db.SetMaxIdleConns(0)
	return db
}

// poolConnector delegates to pgx stdlib's pool connector and implements
// io.Closer so database/sql.DB.Close shuts down the underlying pgxpool.
type poolConnector struct {
	driver.Connector
	pool   *pgxpool.Pool
	dialer *alloydbconn.Dialer

	closeOnce sync.Once
	closeErr  error
}

func (c *poolConnector) Close() error {
	c.closeOnce.Do(func() {
		if c.pool != nil {
			c.pool.Close()
		}
		c.closeErr = closeAlloyDialer(c.dialer)
	})
	return c.closeErr
}

func closeAlloyDialer(dialer *alloydbconn.Dialer) error {
	if dialer == nil {
		return nil
	}
	return dialer.Close()
}

func buildPgxPoolConfig(ctx context.Context, config PostgresConfig, databaseName string) (*pgxpool.Config, *alloydbconn.Dialer, error) {
	if config.Alloy != nil {
		return buildAlloyDBPoolConfig(ctx, config, databaseName)
	}

	connStr, err := getPostgresConnStr(config, databaseName)
	if err != nil {
		return nil, nil, err
	}
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, nil, err
	}
	return poolConfig, nil, nil
}

func getPostgresConnStr(config PostgresConfig, databaseName string) (string, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s/%s", config.User, config.Password, config.Address, databaseName)

	params := ""

	if config.TLS != nil {
		if len(config.TLS.Mode) < 1 {
			config.TLS.Mode = "verify-full"
		}

		params += "sslmode=" + config.TLS.Mode

		if len(config.TLS.CACertFile) > 0 {
			params += "&sslrootcert=" + config.TLS.CACertFile
		}

		if len(config.TLS.ClientCertFile) > 0 {
			params += "&sslcert=" + config.TLS.ClientCertFile
		}

		if len(config.TLS.ClientKeyFile) > 0 {
			params += "&sslkey=" + config.TLS.ClientKeyFile
		}
	}

	connStr := fmt.Sprintf("%s?%s", url, params)
	return connStr, nil
}

func buildAlloyDBPoolConfig(ctx context.Context, config PostgresConfig, databaseName string) (*pgxpool.Config, *alloydbconn.Dialer, error) {
	if config.Alloy == nil {
		return nil, nil, fmt.Errorf("missing alloy config")
	}

	var dialer *alloydbconn.Dialer
	var dsn string

	if config.Alloy.UseIAM {
		d, err := alloydbconn.NewDialer(ctx, alloydbconn.WithIAMAuthN())
		if err != nil {
			return nil, nil, fmt.Errorf("creating alloydb dialer: %w", err)
		}
		dialer = d
		dsn = fmt.Sprintf(
			// sslmode is disabled because the alloy db connection dialer will handle it
			// no password is used with IAM
			"user=%s dbname=%s sslmode=disable",
			config.User, databaseName,
		)
	} else {
		d, err := alloydbconn.NewDialer(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("creating alloydb dialer: %w", err)
		}
		dialer = d
		dsn = fmt.Sprintf(
			// sslmode is disabled because the alloy db connection dialer will handle it
			"user=%s password=%s dbname=%s sslmode=disable",
			config.User, config.Password, databaseName,
		)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		_ = closeAlloyDialer(dialer)
		return nil, nil, fmt.Errorf("failed to parse pgx pool config: %w", err)
	}

	var connOptions []alloydbconn.DialOption
	if config.Alloy.UsePSC {
		connOptions = append(connOptions, alloydbconn.WithPSC())
	}

	poolConfig.ConnConfig.DialFunc = func(ctx context.Context, _ string, _ string) (net.Conn, error) {
		return dialer.Dial(ctx, config.Alloy.InstanceURI, connOptions...)
	}

	return poolConfig, dialer, nil
}

// PostgresUniqueViolation returns true when the provided error matches the Postgres code
// for unique violation.
func PostgresUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == postgresErrUniqueViolation {
		return true
	}

	return strings.Contains(err.Error(), postgresErrUniqueViolation)
}

// PostgresDeadlockFound returns true when the provided error matches the Postgres code
// for deadlock found.
func PostgresDeadlockFound(err error) bool {
	if err == nil {
		return false
	}

	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == postgresErrDeadlockFound {
		return true
	}

	return strings.Contains(err.Error(), postgresErrDeadlockFound)
}
