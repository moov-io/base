package database

type DatabaseConfig struct {
	MySql        *MySqlConfig
	SQLite      *SQLiteConfig
	DatabaseName string
}

type MySqlConfig struct {
	Address  string
	User     string
	Password string
}

type SQLiteConfig struct {
	Path string
}

var InMemorySqliteConfig = DatabaseConfig{
	DatabaseName: "sqlite",
	SQLite: &SQLiteConfig{
		Path: ":memory:",
	},
}
