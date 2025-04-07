package sql

import (
	"context"
	gosql "database/sql"
	"time"

	"github.com/moov-io/base/log"
)

type DB struct {
	*gosql.DB

	logger               log.Logger
	slowQueryThresholdMs int64

	id        string
	stopTimer context.CancelFunc
}

func ObserveDB(innerDB *gosql.DB, logger log.Logger, id string) (*DB, error) {
	cancel := MonitorSQLDriver(innerDB, id)

	return &DB{
		DB:        innerDB,
		id:        id,
		stopTimer: cancel,
		logger:    logger,

		slowQueryThresholdMs: (time.Second * 2).Milliseconds(),
	}, nil
}

func (w *DB) lazyLogger() log.Logger {
	return w.logger
}

func (w *DB) start(op string, qry string, args int) func() int64 {
	return MeasureQuery(w.lazyLogger, w.slowQueryThresholdMs, w.id, op, qry, args)
}

func (w *DB) error(err error) error {
	return MeasureError(w.id, err)
}

func (w *DB) Close() error {
	return w.DB.Close()
}

func (w *DB) SetSlowQueryThreshold(d time.Duration) {
	w.slowQueryThresholdMs = d.Milliseconds()
}

func (w *DB) Prepare(query string) (*Stmt, error) {
	done := w.start("prepare", query, 0)
	defer done()

	return newStmt(context.Background(), w.logger, w.DB, query, w.id, w.slowQueryThresholdMs)
}

func (w *DB) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	done := w.start("prepare", query, 0)
	defer done()

	return newStmt(ctx, w.logger, w.DB, query, w.id, w.slowQueryThresholdMs)
}

func (w *DB) Exec(query string, args ...interface{}) (gosql.Result, error) {
	done := w.start("exec", query, len(args))
	defer done()

	r, err := w.DB.Exec(query, args...)
	return r, w.error(err)
}

func (w *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (gosql.Result, error) {
	done := w.start("exec", query, len(args))
	ctx, end := span(ctx, w.id, "exec", query, len(args))
	defer func() {
		end()
		done()
	}()

	r, err := w.DB.ExecContext(ctx, query, args...)
	return r, w.error(err)
}

func (w *DB) Query(query string, args ...interface{}) (*gosql.Rows, error) {
	done := w.start("query", query, len(args))
	defer done()

	//nolint:sqlclosecheck
	r, err := w.DB.Query(query, args...)
	return r, w.error(err)
}

func (w *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*gosql.Rows, error) {
	done := w.start("query", query, len(args))
	ctx, end := span(ctx, w.id, "query", query, len(args))
	defer func() {
		end()
		done()
	}()

	r, err := w.DB.QueryContext(ctx, query, args...) //nolint:sqlclosecheck
	return r, w.error(err)
}

func (w *DB) QueryRow(query string, args ...interface{}) *gosql.Row {
	done := w.start("query-row", query, len(args))
	defer done()

	r := w.DB.QueryRow(query, args...)
	w.error(r.Err())

	return r
}

func (w *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *gosql.Row {
	done := w.start("query-row", query, len(args))
	ctx, end := span(ctx, w.id, "query-row", query, len(args))
	defer func() {
		end()
		done()
	}()

	r := w.DB.QueryRowContext(ctx, query, args...)
	w.error(r.Err())

	return r
}

func (w *DB) Begin() (*Tx, error) {
	t, err := w.DB.Begin()
	if err != nil {
		return nil, w.error(err)
	}

	tx := &Tx{
		Tx:                   t,
		logger:               w.logger,
		id:                   w.id,
		ctx:                  context.Background(),
		slowQueryThresholdMs: w.slowQueryThresholdMs,
	}

	tx.done = MeasureQuery(tx.lazyLogger, w.slowQueryThresholdMs, tx.id, "tx", "Transaction", 0)

	return tx, nil
}

type TxOptions = gosql.TxOptions

func (w *DB) BeginTx(ctx context.Context, opts *TxOptions) (*Tx, error) {
	ctx, end := span(ctx, w.id, "tx", "BEGIN TRANSACTION", 0)

	t, err := w.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, w.error(err)
	}

	tx := &Tx{
		Tx:                   t,
		logger:               w.logger,
		id:                   w.id,
		ctx:                  ctx,
		slowQueryThresholdMs: w.slowQueryThresholdMs,
	}

	done := MeasureQuery(tx.lazyLogger, w.slowQueryThresholdMs, tx.id, "tx", "Transaction", 0)

	tx.done = func() int64 {
		end()
		return done()
	}

	return tx, nil
}
