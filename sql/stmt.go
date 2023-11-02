package sql

import (
	"context"
	gosql "database/sql"

	"github.com/moov-io/base/log"
)

type Stmt struct {
	logger log.Logger

	id string

	slowQueryThresholdMs int64

	query string
	ss    *gosql.Stmt
}

func newStmt(ctx context.Context, logger log.Logger, db *gosql.DB, query, id string, slowQueryThresholdMs int64) (*Stmt, error) {
	// This statement is closed by (*Stmt).Close() and the responsibility of callers.
	// We want to keep the *gosql.Stmt alive
	ss, err := db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return newWrappedStmt(logger, ss, query, id, slowQueryThresholdMs)
}

func newTxStmt(ctx context.Context, logger log.Logger, tx *gosql.Tx, query, id string, slowQueryThresholdMs int64) (*Stmt, error) {
	// This statement is closed by (*Stmt).Close() and the responsibility of callers.
	// We want to keep the *gosql.Stmt alive
	ss, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return newWrappedStmt(logger, ss, query, id, slowQueryThresholdMs)
}

func newWrappedStmt(logger log.Logger, ss *gosql.Stmt, query, id string, slowQueryThresholdMs int64) (*Stmt, error) {
	return &Stmt{
		logger: logger,
		id:     id,
		query:  query,
		ss:     ss,

		slowQueryThresholdMs: slowQueryThresholdMs,
	}, nil
}

func (s *Stmt) lazyLogger() log.Logger {
	return s.logger
}

func (s *Stmt) start(op string, qry string, args int) func() int64 {
	return MeasureQuery(s.lazyLogger, s.slowQueryThresholdMs, s.id, op, qry, args)
}

func (s *Stmt) error(err error) error {
	return MeasureError(s.id, err)
}

func (s *Stmt) Close() error {
	if s != nil && s.ss != nil {
		return s.ss.Close()
	}
	return nil
}

func (s *Stmt) Exec(args ...any) (gosql.Result, error) {
	done := s.start("exec", s.query, len(args))
	defer done()

	r, err := s.ss.Exec(args...)
	return r, s.error(err)
}

func (s *Stmt) ExecContext(ctx context.Context, args ...any) (gosql.Result, error) {
	done := s.start("exec", s.query, len(args))
	ctx, end := span(ctx, s.id, "exec", s.query, len(args))
	defer func() {
		end()
		done()
	}()

	r, err := s.ss.ExecContext(ctx, args...)
	return r, s.error(err)
}

func (s *Stmt) Query(args ...any) (*gosql.Rows, error) {
	done := s.start("query", s.query, len(args))
	defer done()

	r, err := s.ss.Query(args...)
	return r, s.error(err)
}

func (s *Stmt) QueryContext(ctx context.Context, args ...any) (*gosql.Rows, error) {
	done := s.start("query", s.query, len(args))
	ctx, end := span(ctx, s.id, "query", s.query, len(args))
	defer func() {
		end()
		done()
	}()

	r, err := s.ss.QueryContext(ctx, args...)
	return r, s.error(err)
}

func (s *Stmt) QueryRow(args ...any) *gosql.Row {
	done := s.start("query-row", s.query, len(args))
	defer done()

	r := s.ss.QueryRow(args...)
	s.error(r.Err())

	return r
}

func (s *Stmt) QueryRowContext(ctx context.Context, args ...any) *gosql.Row {
	done := s.start("query-row", s.query, len(args))
	ctx, end := span(ctx, s.id, "query-row", s.query, len(args))
	defer func() {
		end()
		done()
	}()

	r := s.ss.QueryRowContext(ctx, args...)
	s.error(r.Err())

	return r
}
