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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/moov-io/base/log"
)

const (
	// PostgreSQL Error Codes
	// https://www.postgresql.org/docs/current/errcodes-appendix.html
	postgresErrUniqueViolation = "23505"
	postgresErrDeadlockFound   = "40P01"

	// Bound ShouldPing wait so acquire retries during failover stay inside
	// typical request budgets. AlloyDB disconnects usually fail fast (TCP RST);
	// this caps hung/TIME_WAIT peers.
	//
	// Keep this low: Pool.Acquire retries up to maxConns+1 times, and each
	// failed ping can burn the full timeout. At MaxOpen=25 and 1s that is a
	// theoretical ~26s blackhole path; 300ms keeps a full-pool burn nearer a
	// typical request budget while still covering real ping RTT.
	defaultPostgresPingTimeout = 300 * time.Millisecond

	// pgxpool defaults MaxConnLifetimeJitter to 0. Without jitter, every
	// connection created around the same time can expire together.
	defaultPostgresMaxConnLifetimeJitter = 30 * time.Second

	// database/sql.DB.Close calls connector.Close without waiting for in-use
	// driver connections (open Stmts/Rows/Tx). pgxpool.Close blocks until every
	// acquired conn is returned, which deadlocks in that situation. Bound the
	// wait so process shutdown and tests cannot hang forever.
	defaultPostgresPoolCloseWait = 5 * time.Second
)

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

	// MaxConnLifetimeJitter defaults to 0 in pgxpool (not automatic). Only fill
	// when unset so a DSN/pool_max_conn_lifetime_jitter still wins.
	if poolConfig.MaxConnLifetimeJitter <= 0 {
		jitter := defaultPostgresMaxConnLifetimeJitter
		// Prefer ~10% of lifetime when that is larger than the fixed default.
		if tenth := applied.MaxLifetime / 10; tenth > jitter {
			jitter = tenth
		}
		logger.Logf("setting pgx pool MaxConnLifetimeJitter to %v", jitter)
		poolConfig.MaxConnLifetimeJitter = jitter
	}
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
//
// The pool is registered so PoolDBStats can map pgxpool.Stat into sql.DBStats
// for observers (e.g. go-libs observability/sql MeasureStats) that call
// db.Stats(). With MaxIdleConns=0, database/sql.DB.Stats is not meaningful.
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
	c.db = db
	registerPostgresPool(db, pool)
	return db
}

// postgresPools maps *sql.DB from openDBFromPool → underlying pgxpool.
var postgresPools sync.Map

func registerPostgresPool(db *sql.DB, pool *pgxpool.Pool) {
	if db == nil || pool == nil {
		return
	}
	postgresPools.Store(db, pool)
}

func unregisterPostgresPool(db *sql.DB) {
	if db == nil {
		return
	}
	postgresPools.Delete(db)
}

// PoolDBStats returns sql.DBStats derived from the underlying pgxpool when db
// was opened by this package's Postgres/AlloyDB path.
//
// ok is false for MySQL/Spanner DBs, unknown *sql.DB values, or after Close.
// Callers that scrape connection pressure (for example OTel db-metrics spans)
// should prefer this over db.Stats() for Postgres from New.
//
// Mapping from pgxpool.Stat:
//
//	MaxOpenConnections ← MaxConns
//	OpenConnections    ← TotalConns
//	InUse              ← AcquiredConns
//	Idle               ← IdleConns
//	WaitCount          ← EmptyAcquireCount
//	WaitDuration       ← EmptyAcquireWaitTime
//	MaxIdleTimeClosed  ← MaxIdleDestroyCount
//	MaxLifetimeClosed  ← MaxLifetimeDestroyCount
//	MaxIdleClosed      ← 0 (no pgxpool equivalent)
func PoolDBStats(db *sql.DB) (sql.DBStats, bool) {
	if db == nil {
		return sql.DBStats{}, false
	}
	v, ok := postgresPools.Load(db)
	if !ok {
		return sql.DBStats{}, false
	}
	pool, ok := v.(*pgxpool.Pool)
	if !ok || pool == nil {
		return sql.DBStats{}, false
	}
	return pgxPoolStatToDBStats(pool.Stat()), true
}

func pgxPoolStatToDBStats(s *pgxpool.Stat) sql.DBStats {
	if s == nil {
		return sql.DBStats{}
	}
	return sql.DBStats{
		MaxOpenConnections: int(s.MaxConns()),
		OpenConnections:    int(s.TotalConns()),
		InUse:              int(s.AcquiredConns()),
		Idle:               int(s.IdleConns()),
		WaitCount:          s.EmptyAcquireCount(),
		WaitDuration:       s.EmptyAcquireWaitTime(),
		MaxIdleTimeClosed:  s.MaxIdleDestroyCount(),
		MaxLifetimeClosed:  s.MaxLifetimeDestroyCount(),
	}
}

// poolConnector delegates to pgx stdlib's pool connector and implements
// io.Closer so database/sql.DB.Close shuts down the underlying pgxpool.
type poolConnector struct {
	driver.Connector
	pool   *pgxpool.Pool
	dialer *alloydbconn.Dialer
	db     *sql.DB

	closeOnce sync.Once
	closeErr  error
}

func (c *poolConnector) Close() error {
	c.closeOnce.Do(func() {
		if c.db != nil {
			unregisterPostgresPool(c.db)
		}
		if c.pool != nil {
			c.closeErr = closePgxPool(c.pool, defaultPostgresPoolCloseWait)
		}
		if err := closeAlloyDialer(c.dialer); err != nil && c.closeErr == nil {
			c.closeErr = err
		}
	})
	return c.closeErr
}

// closePgxPool closes pool, returning an error if acquired connections prevent
// shutdown within wait. The close continues in the background on timeout so
// connections released later are still reaped.
func closePgxPool(pool *pgxpool.Pool, wait time.Duration) error {
	if pool == nil {
		return nil
	}
	done := make(chan struct{})
	go func() {
		pool.Close()
		close(done)
	}()
	if wait <= 0 {
		<-done
		return nil
	}
	select {
	case <-done:
		return nil
	case <-time.After(wait):
		return fmt.Errorf("pgxpool close timed out after %s; open Stmts/Rows/Tx may still hold connections", wait)
	}
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
