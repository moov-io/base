package database

type Type string

const (
	TypeMySQL  Type = "mysql"
	TypeSQLite Type = "sqlite"
)

type Config struct {
	Type   Type
	MySQL  MySQLConfig
	SQLite SQLiteConfig
}

type MySQLConfig struct {
	Name     string
	Address  string
	User     string
	Password string
}

type SQLiteConfig struct {
	Path string
}
