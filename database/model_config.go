package database

import (
	"path/filepath"
	"strings"
)

type Config interface {
	DBName() string
}

type MySQLConfig struct {
	Name     string
	Address  string
	User     string
	Password string
}

func (c MySQLConfig) DBName() string {
	return c.Name
}

type SQLiteConfig struct {
	Path string
}

func (c SQLiteConfig) DBName() string {
	s := filepath.Base(c.Path)
	return strings.TrimSuffix(s, filepath.Ext(s))
}
