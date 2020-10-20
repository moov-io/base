package database

type DatabaseConfig struct {
	MySQL        *MySQLConfig
	SQLite       *SQLiteConfig
	DatabaseName string
}

type MySQLConfig struct {
	Address  string
	User     string
	Password string
}

type SQLiteConfig struct {
	Path string
}
