package sql

import (
	"context"
	gosql "database/sql"
	"fmt"
	"time"

	"github.com/moov-io/base/log"
)

type Tx struct {
	*gosql.Tx

	logger log.Logger

	id   string
	done func() int64
	ctx  context.Context

	slowQueryThresholdMs int64

	queries []ranQuery
}

type ranQuery struct {
	op   string
	qry  string
	dur  int64
	args int
}

func (w *Tx) lazyLogger() log.Logger {
	return w.logger
}

func (w *Tx) Context() context.Context {
	return w.ctx
}

func (w *Tx) start(op string, query string, args int) func() int64 {
	_, end := span(w.ctx, w.id, op, query, args)

	s := time.Now().UnixMilli()
	return func() int64 {
		end()
		d := time.Now().UnixMilli() - s

		w.queries = append(w.queries, ranQuery{
			op:   op,
			qry:  query,
			dur:  d,
			args: args,
		})

		return d
	}
}

func (w *Tx) error(err error) error {
	return MeasureError(w.id, err)
}

func (w *Tx) Commit() error {
	defer w.finished()
	return w.error(w.Tx.Commit())
}

func (w *Tx) Rollback() error {
	defer w.finished()
	return w.error(w.Tx.Rollback())
}

func (w *Tx) finished() {
	w.logger = w.logger.With(log.Fields{
		"query_id":  log.String(w.id),
		"query_cnt": log.Int(len(w.queries)),
	})

	for i, q := range w.queries {
		if i < 7 {
			pre := fmt.Sprintf("%d_", i)
			w.logger = w.logger.With(log.Fields{
				pre + "query":         log.String(CleanQuery(q.qry)),
				pre + "query_op":      log.String(q.op),
				pre + "query_time_ms": log.Int64(q.dur),
				pre + "query_args":    log.Int(q.args),
			})
		}
	}

	w.done()
}

func (w *Tx) ExecContext(ctx context.Context, query string, args ...any) (gosql.Result, error) {
	done := w.start("exec", query, len(args))
	defer done()

	r, err := w.Tx.ExecContext(ctx, query, args...)
	return r, w.error(err)
}

func (w *Tx) Exec(query string, args ...any) (gosql.Result, error) {
	done := w.start("exec", query, len(args))
	defer done()

	r, err := w.Tx.Exec(query, args...)
	return r, w.error(err)
}

func (w *Tx) QueryContext(ctx context.Context, query string, args ...any) (*gosql.Rows, error) {
	done := w.start("query", query, len(args))
	defer done()

	r, err := w.Tx.QueryContext(ctx, query, args...) //nolint:sqlclosecheck
	return r, w.error(err)
}

func (w *Tx) Query(query string, args ...any) (*gosql.Rows, error) {
	done := w.start("query", query, len(args))
	defer done()

	r, err := w.Tx.Query(query, args...) //nolint:sqlclosecheck
	return r, w.error(err)
}

func (w *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *gosql.Row {
	done := w.start("query-row", query, len(args))
	defer done()

	r := w.Tx.QueryRowContext(ctx, query, args...)
	w.error(r.Err())

	return r
}

func (w *Tx) QueryRow(query string, args ...any) *gosql.Row {
	done := w.start("query-row", query, len(args))
	defer done()

	r := w.Tx.QueryRow(query, args...)
	w.error(r.Err())

	return r
}

func (w *Tx) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	done := w.start("prepare", query, 0)
	defer done()

	return newTxStmt(ctx, w.logger, w.Tx, query, w.id, w.slowQueryThresholdMs)
}
