package database

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4/source"
	mpkger "github.com/golang-migrate/migrate/v4/source/pkger"
	"github.com/markbates/pkger"
	"github.com/markbates/pkger/pkging/mem"
)

const MIGRATIONS_DIR = "/migrations/"

func NewPkgerSource(database string) (source.Driver, error) {
	database = strings.ToLower(database)

	hereInfo, err := pkger.Current()
	if err != nil {
		return nil, err
	}

	pmem, err := mem.New(hereInfo)
	if err != nil {
		return nil, err
	}

	pmem.MkdirAll(MIGRATIONS_DIR, 0755)

	pkger.Walk(MIGRATIONS_DIR, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		splits := strings.Split(info.Name(), ".")
		slen := len(splits)
		if slen < 3 {
			return fmt.Errorf("doesn't follow format of {version}_{title}.up{.db}?.sql - %s", info.Name())
		}

		if splits[:-1] != "sql" {
			return fmt.Errorf("must end in .sql")
		}

		run := false
		switch splits[slen-2] {
		case "up":
			run = true
		case "down":
			run = true
		case database:
			run = true
		default:
			run = false
		}

		if !run {
			return nil
		}

		cur, err := pkger.Open(MIGRATIONS_DIR + info.Name())
		if err != nil {
			return err
		}

		nw, err := pmem.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(nw, cur)
		if err != nil {
			return err
		}

		nw.Close()

		return nil
	})

	drv, err := mpkger.WithInstance(pmem, MIGRATIONS_DIR)
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate driver - %w", err)
	}

	return drv, nil
}
