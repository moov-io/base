package database

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/moov-io/base/log"
)

func TestMySQL__Basic(t *testing.T) {
	db := CreateTestMySQLDB(t)
	defer db.Close()

	err := db.DB.Ping()
	require.NoError(t, err)

	require.Equal(t, 0, db.DB.Stats().OpenConnections)
}

func TestMySQL_Teardown(t *testing.T) {
	for i := 0; i < 3; i++ {
		db := CreateTestMySQLDB(t)
		defer db.Close()

		row := db.QueryRow("SELECT COUNT(*) FROM tests")
		var count int
		require.NoError(t, row.Scan(&count))
		require.Equal(t, 0, count)

		insertQuery := "insert into tests (id) values (100),(200),(300);"
		_, err := db.Exec(insertQuery)
		require.NoError(t, err)
	}

}

func TestMySQL_BadConnection(t *testing.T) {
	// create a phony MySQL
	m := mysqlConnection(log.NewNopLogger(), MySQLConfig{
		Name:     "fake",
		Address:  "9000",
		User:     "moov",
		Password: "secret",
	})

	ctx, cancelFunc := context.WithCancel(context.Background())

	conn, err := m.Connect(ctx)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	require.Nil(t, conn)
	require.Error(t, err)

	cancelFunc()
}

func TestMySQLUniqueViolation(t *testing.T) {
	err := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1062: Duplicate entry '282f6ffcd9ba5b029afbf2b739ee826e22d9df3b' for key 'PRIMARY'`)
	if !UniqueViolation(err) {
		t.Error("should have matched unique violation")
	}
}
