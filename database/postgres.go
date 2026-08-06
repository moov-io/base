package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
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
	defaultPostgresPingTimeout = time.Second
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

	logger.Logf("setting pgx pool MaxConns to %d", applied.MaxOpen)
	poolConfig.MaxConns = int32(applied.MaxOpen)

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
	if connections.MaxIdle <= 0 {
		connections.MaxIdle = defaults.MaxIdle
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
