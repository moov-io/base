package sql_test

import (
	"context"
	gosql "database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/moov-io/base/log"
	"github.com/moov-io/base/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SQL_Connect(t *testing.T) {
	a := assert.New(t)

	db, _ := ConnectTestDB(t)
	a.NotNil(db)
}

func Test_SQL_Prepare(t *testing.T) {
	a := assert.New(t)

	db, _ := ConnectTestDB(t)
	a.NotNil(db)

	sql := "INSERT INTO moov.test(id, value) VALUES (?, ?)"
	istmt, err := db.PrepareContext(context.Background(), sql)
	a.NoError(err)
	t.Cleanup(func() { a.NoError(istmt.Close()) })

	first := uuid.NewString()
	res, err := istmt.Exec(first, uuid.NewString())
	a.NoError(err)
	n, err := res.RowsAffected()
	a.NoError(err)
	a.Equal(int64(1), n)

	second := uuid.NewString()
	res, err = istmt.ExecContext(context.Background(), second, uuid.NewString())
	a.NoError(err)
	n, err = res.RowsAffected()
	a.NoError(err)
	a.Equal(int64(1), n)

	sql = "SELECT * FROM moov.test WHERE id = ? LIMIT 1"
	sstmt, err := db.Prepare(sql)
	a.NoError(err)
	t.Cleanup(func() { a.NoError(sstmt.Close()) })

	rows, err := sstmt.Query(first)
	a.NoError(err)
	a.NoError(rows.Err())
	t.Cleanup(func() { a.NoError(rows.Close()) })

	row := sstmt.QueryRow(first)
	a.NoError(row.Err())

	rows2, err := sstmt.QueryContext(context.Background(), second)
	a.NoError(err)
	a.NoError(rows2.Err())
	t.Cleanup(func() { a.NoError(rows2.Close()) })

	row = sstmt.QueryRowContext(context.Background(), second)
	a.NoError(row.Err())
}

func Test_SQL_Exec(t *testing.T) {
	a := assert.New(t)

	db, _ := ConnectTestDB(t)
	a.NotNil(db)

	sql := "INSERT INTO moov.test(id, value) VALUES (?, ?)"

	_, err := db.Exec(sql, uuid.NewString(), uuid.NewString())
	a.NoError(err)

	_, err = db.ExecContext(context.Background(), sql, uuid.NewString(), uuid.NewString())
	a.NoError(err)

	tx, err := db.Begin()
	a.NoError(err)

	_, err = tx.Exec(sql, uuid.NewString(), uuid.NewString())
	a.NoError(err)

	_, err = tx.ExecContext(context.Background(), sql, uuid.NewString(), uuid.NewString())
	a.NoError(err)

	err = tx.Rollback()
	a.NoError(err)
}

func Test_SQL_Query(t *testing.T) {
	a := assert.New(t)

	db, _ := ConnectTestDB(t)
	a.NotNil(db)

	id := AddRecord(t, db)

	sql := "SELECT * FROM moov.test WHERE id = ? LIMIT 1"

	r, err := db.Query(sql, id)
	a.NoError(err)
	a.NoError(r.Err())
	defer r.Close()

	r, err = db.QueryContext(context.Background(), sql, id)
	a.NoError(err)
	a.NoError(r.Err())
	defer r.Close()

	row := db.QueryRow(sql, id)
	a.NoError(row.Err())

	row = db.QueryRowContext(context.Background(), sql, id)
	a.NoError(row.Err())
}

func Test_SQL_Query_Tx(t *testing.T) {
	a := assert.New(t)

	db, _ := ConnectTestDB(t)
	a.NotNil(db)

	id := AddRecord(t, db)

	sql := "SELECT * FROM moov.test WHERE id = ? LIMIT 1"

	tx, err := db.BeginTx(context.Background(), &gosql.TxOptions{})
	a.NoError(err)

	r, err := tx.Query(sql, id)
	a.NoError(err)
	a.NoError(r.Err())
	defer r.Close()
	r.Close()

	r, err = tx.QueryContext(context.Background(), sql, id)
	a.NoError(err)
	a.NoError(r.Err())
	defer r.Close()
	r.Close()

	err = tx.Commit()
	a.NoError(err)
}

func Test_SQL_Query_Tx_Row(t *testing.T) {
	a := assert.New(t)

	db, logBuilder := ConnectTestDB(t)
	a.NotNil(db)

	id := AddRecord(t, db)

	// to be able to run multiple queries we have to dump the scanned value
	dump := ""

	sql := "SELECT * FROM moov.test WHERE id = ? LIMIT 1"

	tx, err := db.BeginTx(context.Background(), &gosql.TxOptions{})
	a.NoError(err)

	row := tx.QueryRow(sql, id)
	row.Scan(&dump)
	a.NoError(row.Err())

	row = tx.QueryRow(sql, id)
	row.Scan(&dump)
	a.NoError(row.Err())

	err = tx.Commit()
	a.NoError(err)

	logs := logBuilder.String()
	a.Contains(logs, "0_query")
	a.Contains(logs, "0_query_op")
	a.Contains(logs, "0_query_time_ms")
	a.Contains(logs, "1_query")
	a.Contains(logs, "1_query_op")
	a.Contains(logs, "1_query_time_ms")
}

func Test_SQL_Query_Tx_RowCtx(t *testing.T) {
	a := assert.New(t)

	db, _ := ConnectTestDB(t)

	db.SetSlowQueryThreshold(0 * time.Millisecond)

	a.NotNil(db)

	id := AddRecord(t, db)

	sql := "SELECT * FROM moov.test WHERE id = ? LIMIT 1"

	tx, err := db.BeginTx(context.Background(), &gosql.TxOptions{})
	a.NoError(err)

	row := tx.QueryRowContext(context.Background(), sql, id)
	a.NoError(row.Err())

	err = tx.Commit()
	a.NoError(err)
}

func Test_SQL_Create(t *testing.T) {
	a := assert.New(t)

	db, err := sql.ObserveDB(&gosql.DB{}, log.NewNopLogger(), "test1")
	a.NoError(err)
	a.NotNil(db)
}

func Test_SQL_Monitor(t *testing.T) {
	a := assert.New(t)

	a.NoError(sql.MeasureStats(&gosql.DB{}, "test1"))
}

func Test_SQL_Monitor_Query(t *testing.T) {
	done := sql.MeasureQuery(LazyNopLogger, time.Minute.Milliseconds(), "1", "tx", "select * from test", 0)
	done()

	t.Run("slow query", func(t *testing.T) {
		threshold := time.Second.Milliseconds()
		require.Equal(t, int64(1000), threshold)

		buf, logger := log.NewBufferLogger()
		lazyLogger := func() log.Logger {
			return logger
		}

		done = sql.MeasureQuery(lazyLogger, threshold, "2", "exec", "delete from things;", 0)
		time.Sleep(250 * time.Millisecond)
		done()

		fmt.Printf("\n\n%s\n", buf.String())
		buf.Reset()

		done = sql.MeasureQuery(lazyLogger, threshold, "2", "exec", "delete from things;", 0)
		time.Sleep(900 * time.Millisecond)
		done()

		done = sql.MeasureQuery(lazyLogger, threshold, "2", "exec", "delete from things;", 0)
		time.Sleep(2 * time.Second)
		done()

		fmt.Printf("\n\n%s\n", buf.String())
	})
}

func Test_SQL_Monitor_Error(t *testing.T) {
	sql.MeasureError("1", errors.New("error!"))
}

func LazyNopLogger() log.Logger {
	return log.NewNopLogger()
}

func ConnectTestDB(t *testing.T) (*sql.DB, *log.BufferedLogger) {
	t.Helper()
	open := func() (*gosql.DB, error) {
		db, err := gosql.Open("mysql", "moov:moov@tcp(localhost:3306)/")
		if err != nil {
			return nil, err
		}

		if err := db.Ping(); err != nil {
			db.Close()
			return nil, err
		}

		return db, nil
	}

	db, err := open()
	for i := 0; err != nil && i < 22; i++ {
		time.Sleep(time.Second * 1)
		db, err = open()
	}
	if err != nil {
		t.Fatal(err)
	}

	lines, logger := log.NewBufferLogger()

	odb, err := sql.ObserveDB(db, logger, "test")
	if err != nil {
		t.Fatal(err)
	}

	odb.SetSlowQueryThreshold(0)

	t.Cleanup(func() {
		db.Close()
	})

	createTable := `
		CREATE TABLE IF NOT EXISTS moov.test (
			id     VARCHAR(36) NOT NULL,
			value  VARCHAR(255),

			CONSTRAINT connection_pk PRIMARY KEY (id)
		)
	`

	_, err = odb.Exec(createTable)
	if err != nil {
		t.Fatal(err)
	}

	return odb, lines
}

func AddRecord(t *testing.T, db *sql.DB) string {
	t.Helper()
	// Add a record
	id := uuid.NewString()
	sql := "INSERT INTO moov.test(id, value) VALUES (?, ?)"
	_, err := db.Exec(sql, id, uuid.NewString())
	if err != nil {
		assert.NoError(t, err)
	}

	return id
}

func Test_CleanQuery(t *testing.T) {
	query := `
		SELECT *
			FROM sometable
			WHERE sometable.field   =  ?
	`

	cleaned := sql.CleanQuery(query)

	assert.Equal(t, `SELECT * FROM sometable WHERE sometable.field = ?`, cleaned)
}
