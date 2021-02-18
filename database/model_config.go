package database

import "time"

type DatabaseConfig struct {
	MySQL        *MySQLConfig
	SQLite       *SQLiteConfig
	DatabaseName string
}

type MySQLConfig struct {
	Address     string
	User        string
	Password    string
	Connections ConnectionsConfig
}

type SQLiteConfig struct {
	Path string
}

type ConnectionsConfig struct {
	MaxOpen     int
	MaxIdle     int
	MaxLifetime time.Duration
	MaxIdleTime time.Duration
}
