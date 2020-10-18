package database

import (
	"context"
	"errors"
	"testing"

	"github.com/moov-io/base/docker"
)

func Test_NewAndMigration_SQLite3(t *testing.T) {
	db, err := NewAndMigrate(context.Background(), nil, InMemorySqliteConfig)
	if err != nil {
		t.FailNow()
	}
	db.Close()
}

func Test_NewAndMigration_MySql(t *testing.T) {
	if !docker.Enabled() {
		t.SkipNow()
	}

	config, container, err := RunMySQLDockerInstance(&DatabaseConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer container.Close()

	db, err := NewAndMigrate(context.Background(), nil, *config)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestUniqueViolation(t *testing.T) {
	err := errors.New(`problem upserting depository="282f6ffcd9ba5b029afbf2b739ee826e22d9df3b", userId="f25f48968da47ef1adb5b6531a1c2197295678ce": Error 1062: Duplicate entry '282f6ffcd9ba5b029afbf2b739ee826e22d9df3b' for key 'PRIMARY'`)
	if !UniqueViolation(err) {
		t.Error("should have matched unique violation")
	}

	err = errors.New(`problem upserting depository="7d676c65eccd48090ff238a0d5e35eb6126c23f2", userId="80cfe1311d9eb7659d02cba9ee6cb04ed3739a85": UNIQUE constraint failed: depositories.depository_id`)
	if !UniqueViolation(err) {
		t.Error("should have matched unique violation")
	}
}