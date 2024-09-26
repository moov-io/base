package base

import "embed"

//go:embed configs/config.default.yml
var ConfigDefaults embed.FS

//go:embed migrations/*.up.sql migrations/*.up.mysql.sql
var MySQLMigrations embed.FS

//go:embed migrations/*.up.spanner.sql
var SpannerMigrations embed.FS

//go:embed migrations/*.up.postgres.sql
var PostgresMigrations embed.FS
