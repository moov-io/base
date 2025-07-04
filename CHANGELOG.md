## Future Releases

Please refer to the [Github Releases](https://github.com/moov-io/base/releases) page for future updates.

## v0.56.0 (Released 2025-06-11)

IMPROVEMENTS

- http: Update `LimitedSkipCount` and `GetSkipAndCount` to support either http.Request or string as input

## v0.55.1 (Released 2025-06-09)

IMPROVEMENTS

- fix: remove mutex locking in `database.RunMigrations` to allow tests to fully benefit from runs using `t.Parallel()`

BUILD

- fix(deps): update opentelemetry-go monorepo to v1.36.0 (#458)

## v0.55.0 (Released 2025-05-06)

ADDITIONS

- feat: add `StructContext` for logging fields of a struct

BUILD

- build(deps): bump github.com/go-jose/go-jose/v4 from 4.0.4 to 4.0.5
- fix(deps): update module github.com/golang-migrate/migrate/v4 to v4.18.3 (#455)

## v0.54.4 (Released 2025-04-22)

BUILD

- build(deps): bump golang.org/x/net from 0.37.0 to 0.38.0
- build: update cloud.google.com/go/alloydbconn to v1.15.1
- build: update cloud.google.com/go/spanner to v1.79.0
- build: update github.com/go-sql-driver/mysql to v1.9.2
- build: update github.com/jackc/pgx/v5 to v5.7.4
- build: update google.golang.org/grpc to v1.72.0

## v0.54.3 (Released 2025-04-09)

IMPROVEMENTS

- feat: catch psql deadlock err

BUILD

- build: update cloud.google.com/go/spanner to v1.79.0
- build: update github.com/go-sql-driver/mysql to v1.9.2
- build: update github.com/googleapis/go-sql-spanner to v1.13.0
- build: update google.golang.org/grpc to v1.71.1

## v0.54.2 (Released 2025-03-17)

BUILD

- build(deps): bump golang.org/x/net from 0.35.0 to 0.36.0
- build: bump cloud.google.com/go/alloydbconn from v1.14.1 to v1.15.0
- build: bump cloud.google.com/go/spanner from v1.75.0 to v1.77.0
- build: bump github.com/go-sql-driver/mysql from v1.8.1 to v1.9.0
- build: bump github.com/googleapis/go-sql-spanner from v1.11.1 to v1.11.2
- build: bump github.com/madflojo/testcerts from v1.3.0 to v1.4.0
- build: bump github.com/prometheus/client_golang from v1.20.5 to v1.21.1
- build: bump github.com/rickar/cal/v2 from v2.1.21 to v2.1.22
- build: bump github.com/spf13/viper from v1.19.0 to v1.20.0

## v0.54.1 (Released 2025-02-17)

v0.54.0 was accidently pushed as a breaking change by forcing upgrades to Go 1.24 - v0.54.1 has been released which does not require Go 1.24

BUILD

- build(deps): bump golang.org/x/crypto from 0.27.0 to 0.31.0
- build(deps): bump github.com/googleapis/go-sql-spanner to v1.11.1

## v0.53.0 (Released 2024-09-30)

IMPROVEMENTS

- database: enable TLS with postgres tests
- database: fix another printf
- feat: Support PostgreSQL databases via config

BUILD

- fix(deps): update module github.com/googleapis/go-sql-spanner to v1.7.2 (#438)
- fix(deps): update module github.com/madflojo/testcerts to v1.3.0 (#435)

## v0.52.1 (Released 2024-09-24)

IMPROVEMENTS

- database: fix printf

## v0.52.0 (Released 2024-09-20)

IMPROVEMENTS

- database: add RunMigrationsContext with tracing
- test: verify ErrorList doesn't obviously panic
- test: verify `yaml:"x-foo"` works

BUILD

- chore(deps): update actions/checkout action to v4
- chore(deps): update actions/setup-go action to v5
- chore(deps): update github/codeql-action action to v3
- fix(deps): update module cloud.google.com/go/spanner to v1.67.0
- fix(deps): update module github.com/googleapis/go-sql-spanner to v1.7.1
- fix(deps): update module github.com/prometheus/client_golang to v1.20.2
- fix(deps): update module github.com/rickar/cal/v2 to v2.1.19
- fix(deps): update module google.golang.org/grpc to v1.67.0
- fix(deps): update opentelemetry-go monorepo to v1.29.0

## v0.51.1 (Released 2024-07-11)

IMPROVEMENTS

- config: include decoder fallback to parse time.Duration values

## v0.51.0 (Released 2024-07-11)

IMPROVEMENTS

- feat: allow config to unmarshal regexes as strings

BUILD

- chore(deps): update mysql docker tag to v9
- fix(deps): update module github.com/googleapis/go-sql-spanner to v1.5.0
- fix(deps): update module github.com/rickar/cal/v2 to v2.1.17
- fix(deps): update module google.golang.org/grpc to v1.65.0
- fix(deps): update opentelemetry-go monorepo to v1.28.0

## v0.50.0 (Released 2024-06-19)

IMPROVEMENTS

- feat: allowing for the x-clean-statements during spanner migrations to be configurable

BUILD

- fix(deps): update module github.com/rickar/cal/v2 to v2.1.16

## v0.49.4 (Released 2024-06-11)

IMPROVEMENTS

- time: if exactly 5pm, should go to next banking day

BUILD

- fix(deps): update module cloud.google.com/go/spanner to v1.63.0
- fix(deps): update module github.com/googleapis/go-sql-spanner to v1.4.0
- fix(deps): update module github.com/madflojo/testcerts to v1.2.0
- fix(deps): update module github.com/spf13/viper to v1.19.0
- fix(deps): update module google.golang.org/grpc to v1.64.0
- fix(deps): update opentelemetry-go monorepo to v1.27.0

## v0.49.3 (Released 2024-05-13)

IMPROVEMENTS

- database: enforce ordering of sql.DB config (SetConnMaxIdleTime before SetConnMaxLifetime)

## v0.49.2 (Released 2024-05-10)

IMPROVEMENTS

- strx: revert Go 1.22 simplification

BUILD

- build: -short test on Windows
- build: run oldstable Go, run go test on windows
- database: fix cert paths in test

## v0.49.1 (Released 2024-05-10)

BUILD

- meta: downgrade Go to 1.21 until Openshift supports newer Go

## v0.49.0 (Released 2024-05-09)

IMPROVEMENTS

- config: merge arbitrary map's together
- feat: AddBankingTime
- strx: simplify implementation in Go 1.22+

BUILD

- chore(deps): update mysql docker tag to v8.4
- chore(deps): update dependency go to v1.22.3
- fix(deps): update module github.com/go-sql-driver/mysql to v1.8.1
- fix(deps): update module cloud.google.com/go/spanner to v1.61.0
- fix(deps): update module github.com/google/uuid to v1.6.0
- fix(deps): update module github.com/stretchr/testify to v1.9.0
- fix(deps): update module github.com/googleapis/gax-go/v2 to v2.12.4
- fix(deps): update module github.com/googleapis/go-sql-spanner to v1.3.1
- fix(deps): update module github.com/prometheus/client_golang to v1.19.1
- fix(deps): update module github.com/golang-migrate/migrate/v4 to v4.17.1
- fix(deps): update module github.com/rickar/cal/v2 to v2.1.14
- fix(deps): update module github.com/rickar/cal/v2 to v2.1.15
- fix(deps): update module google.golang.org/grpc to v1.63.2
- fix(deps): update opentelemetry-go monorepo to v1.26.0

## v0.48.5 (Released 2024-01-11)

IMPROVEMENTS

- http: Update default count value to 200

BUILD

- fix(deps): update module cloud.google.com/go/spanner to v1.55.0

## v0.48.4 (Released 2024-01-09)

IMPROVEMENTS

- fix: proper printf verbs

BUILD

- build(deps): bump golang.org/x/crypto from 0.16.0 to 0.17.0
- fix(deps): update module github.com/golang-migrate/migrate/v4 to v4.17.0
- fix(deps): update module github.com/prometheus/client_golang to v1.18.0
- fix(deps): update module github.com/spf13/viper to v1.18.2
- fix(deps): update module google.golang.org/grpc to v1.60.1

## v0.48.3 (Released 2023-12-13)

IMPROVEMENTS

- chore: add better logging around running migrations
- chore: check for some rare null pointers
- test: checking ErrorList conditions
- test: verify upcoming holidays calculate correctly

BUILD

- fix(deps): update google.golang.org/grpc to v1.60.0
- fix(deps): update module cloud.google.com/go/spanner to v1.53.1
- fix(deps): update opentelemetry-go monorepo to v1.20.0

## v0.48.2 (Released 2023-11-10)

IMPROVEMENTS

- time: use generic logic for Friday-observed banking days

BUILD

- fix(deps): update module github.com/gorilla/mux to v1.8.1

## v0.48.1 (Released 2023-11-10)

IMPROVEMENTS

- fix: Veteran's day is not observed today
- test: check future Saturday holidays

## v0.48.0 (Released 2023-11-03)

ADDITIONS

- feat: add `build` package to log runtime information
- feat: add `sql` package to instrument SQL statements
- feat: add `telemetry` package to instrument OpenTracing in applications

BUILD

- chore(deps): update mysql docker tag to v8.2
- fix(deps): update module github.com/google/uuid to v1.4.0

## v0.47.1 (Released 2023-10-26)

IMPROVEMENTS

- fix(logger): prevent nil pointer dereference when calling `log.Stringer(nil)`

BUILD

- fix(deps): update module github.com/madflojo/testcerts to v1.1.1
- fix(deps): update module cloud.google.com/go/spanner to v1.51.0
- fix(deps): update module google.golang.org/grpc to v1.59.0

## v0.47.0 (Released 2023-09-26)

ADDITIONS

- http: Add GetOrderBy(r *http.Request) to get 'orderBy' vars from request

BUILD

- fix(deps): update module cloud.google.com/go/spanner to v1.49.0
- fix(deps): update module github.com/go-kit/kit to v0.13.0
- fix(deps): update module github.com/google/uuid to v1.3.1
- fix(deps): update module google.golang.org/grpc to v1.58.1

## v0.46.0 (Released 2023-08-21)

BUILD

- admin: remove `NewServer`
- chore(deps): update mysql docker tag to v8.1
- fix(deps): update module cloud.google.com/go/spanner to v1.48.0
- fix(deps): update module google.golang.org/grpc to v1.57.0

## v0.45.1 (Released 2023-07-21)

IMPROVEMENTS

- log: make NewBufferLogger() safe to read/write across goroutines

BUILD

- build: use latest stable Go release
- fix(deps): update module github.com/googleapis/gax-go/v2 to v2.12.0
- fix(deps): update module google.golang.org/grpc to v1.56.2

## v0.45.0 (Released 2023-07-05)

ADDITIONS

- log: Added the ability to specify the format for log messages via `MOOV_LOG_FORMAT` with values of `nop`, `json`, or `logfmt`
- fix(deps): update module github.com/googleapis/go-sql-spanner to v1.1.0

## v0.44.0 (Released 2023-05-31)

ADDITIONS

- config: allows specifying a `embed.FS` to pull the default configuration from
- database: allow specifying a `embed.FS` to pull migrations from

## v0.43.0 (Released 2023-05-22)

IMPROVEMENTS

- config: clarify the naming of the mapstructure.DecoderConfig
- feat: detect mysql deadlock err

BUILD

- fix(deps): update module cloud.google.com/go/spanner to v1.46.0
- fix(deps): update module github.com/googleapis/gax-go/v2 to v2.9.0
- fix(deps): update module github.com/stretchr/testify to v1.8.3

## v0.42.0 (Released 2023-05-04)

IMPROVEMENTS

- database: idempotently creates the spanner databases.
- database: allows for specific creation in addition to random spanner database creation

## v0.41.0 (Released 2023-05-04)

ADDITIONS

- database: add `SpannerUniqueViolation` helper for mapping Spanner DB duplicate error
- database: adjusted `UniqueViolation` helper to check for either MySQL or Spanner errors

BUILD

- fix(deps): update module cloud.google.com/go/spanner to v1.45.1
- fix(deps): update module github.com/go-sql-driver/mysql to v1.7.1
- fix(deps): update module github.com/prometheus/client_golang to v1.15.1

## v0.40.2 (Released 2023-04-07)

IMPROVEMENTS

- config: overwrite slices instead of merging

## v0.40.1 (Released 2023-04-07)

IMPROVEMENTS

- database: fixed issues with spanner migrations and comments

## v0.40.0 (Released 2023-04-07)

ADDITIONS

- Adding in spanner support for databases

BUILD

- build: update github.com/stretchr/testify to v1.8.2
- fix(deps): update module github.com/madflojo/testcerts to v1.1.0
- fix(deps): update module github.com/rickar/cal/v2 to v2.1.13

## v0.39.0 (Released 2023-01-26)

ADDITIONS

- http: add LimitedSkipCount helper

## v0.38.2 (Released 2023-01-26)

IMPROVEMENTS

- Increase maximum 'skip' value in GetSkipAndCount(r *http.Request) to math.MaxInt32

BUILD

- fix(deps): update module github.com/rickar/cal/v2 to v2.1.10
- fix(deps): update module github.com/spf13/viper to v1.15.0

## v0.38.1 (Released 2022-12-19)

IMPROVEMENTS

- idempotent: remove outdated package

## v0.38.0 (Released 2022-12-12)

ADDITIONS

- admin: add constructor for Admin server that doesn't panic, add timeout setters
- randx: add new package

## v0.37.0 (Released 2022-12-06)

BREAKING CHANGES

- database: remove SQLite as a database option

## v0.36.4 (Released 2022-12-05)

BUILD

- fix(deps): update module github.com/go-sql-driver/mysql to v1.7.0

## v0.36.3 (Released 2022-12-02)

IMPROVEMENTS

- Fix MySQLUniqueViolation check to look for error dupe code more broadly
- Fix MySQLDataTooLong check to look for error data length code more broadly

## v0.36.2 (Released 2022-11-14)

BUILD

- fix(deps): update module github.com/hashicorp/golang-lru to v0.6.0
- fix(deps): update module github.com/mattn/go-sqlite3 to v1.14.16
- fix(deps): update module github.com/prometheus/client_golang to v1.14.0
- fix(deps): update module github.com/rickar/cal/v2 to v2.1.8
- fix(deps): update module github.com/spf13/viper to v1.14.0

## v0.36.1 (Released 2022-10-24)

BUILD

- fix(deps): update module github.com/fsnotify/fsnotify to v1.6.0
- fix(deps): update module github.com/gobuffalo/here to v0.6.7
- fix(deps): update module github.com/matttproud/golang_protobuf_extensions to v1.0.2
- fix(deps): update module github.com/prometheus/client_model to v0.3.0
- fix(deps): update module github.com/spf13/afero to v1.9.2
- fix(deps): update module github.com/stretchr/testify to v1.8.1
- fix(deps): update module go.uber.org/atomic to v1.10.0
- fix(deps): update module golang.org/x/sys to v0.1.0
- fix(deps): update module golang.org/x/text to v0.4.0

## v0.36.0 (Released 2022-10-11)

IMPROVEMENTS

- build: add GetHoliday() onto Time

BUILD

- build: require Go +1.19 in Actions
- fix(deps): update module github.com/rickar/cal/v2 to v2.1.7

## v0.35.0 (Released 2022-09-12)

IMPROVEMENTS

- Added TLS client certs to TLS config for database connection
- TLS client certs can be included via configuration

BUILD

- fix(deps): update module github.com/spf13/viper to v1.13.0

## v0.34.1 (Released 2022-08-30)

BUILD

- fix(deps): update module github.com/mattn/go-sqlite3 to v1.14.15
- fix(deps): update module github.com/rickar/cal/v2 to v2.1.6

## v0.34.0 (Released 2022-08-11)

IMPROVEMENTS

- Add Int64OrNil method for logging
- docs: describe liveness/readiness probes and metrics endpoints
- test: ensure ID() length

BUILD

- build: update deprecated io/ioutil functions
- chore(deps): update module go to 1.19
- fix(deps): update module github.com/prometheus/client_golang to v1.13.0

## v0.33.0 (Released 2022-07-11)

BREAKING CHANGES

- fix: quit converting times to UTC in `Time`

## v0.32.0 (Released 2022-07-06)

ADDITIONS

- database: add `MySQLDataTooLong` helper for detecting "data too long" errors (Code: 1406)

BUILD

- fix(deps): update module github.com/stretchr/testify to v1.8.0

## v0.31.1 (Released 2022-06-15)

IMPROVEMENTS

- fix: recover from panics during logging, log those if we can

## v0.31.0 (Released 2022-06-15)

ADDITIONS

- feat: add AddBusinessDay to Time (#235)

IMPROVEMENTS

- time: update test cases to make sure Juneteenth holiday is handled appropriately (#234)

BUILD

- fix(deps): update module github.com/rickar/cal/v2 to v2.1.5 (#234)

## v0.30.0 (Released 2022-06-02)

ADDITIONS

- feat: add IsBusinessDay to Time
- log: add a helper for stdout logging during verbose test runs

IMPROVEMENTS

- do not ignore error when walk dir in pkger

BUILD

- fix(deps): update module github.com/spf13/viper to v1.12.0

## v0.29.3 (Released 2022-05-19)

IMPROVEMENTS

- k8s: update to check more paths

BUILD

- build: update codeql action
- fix(deps): update module github.com/go-kit/log to v0.2.1
- fix: mysql/Dockerfile to reduce vulnerabilities

## v0.29.2 (Released 2022-05-19)

BUILD

- fix(deps): update module github.com/mattn/go-sqlite3 to v1.14.13
- fix(deps): update module github.com/prometheus/client_golang to v1.12.2

## v0.29.0 (Released 2022-05-09)

REMOVALS

- database: remove test containers based on dockertest (aka `database.CreateTestMySQLDB`)

BUILD

- fix(deps): update module github.com/stretchr/testify to v1.7.1
- fix(deps): update module github.com/spf13/viper to v1.11.0
- fix(deps): update module github.com/golang-migrate/migrate/v4 to v4.15.2

## v0.28.1 (Released 2022-03-07)

ADDITIONS

- time: add IsHoliday()

## v0.28.0 (Released 2021-01-09)

BUILD

- fix(deps): update module github.com/mattn/go-sqlite3 to v1.14.12
- fix(deps): update module github.com/prometheus/client_golang to v1.12.1
- time: update github.com/rickar/cal to v2 release

## v0.27.5 (Released 2021-01-09)

IMPROVEMENTS

- Adding in a lock around writing to the metrics

## v0.27.4 (Released 2021-01-09)

IMPROVEMENTS

- Cleanup and simplify the recording of metrics

## v0.27.3 (Released 2021-01-09)

IMPROVEMENTS

- Adding logging to the error that could come back from verify ca
- Adding in additional metrics to track for database connections
- Adding in test cases to make sure the metrics get recorded

## v0.27.2 (Released 2021-01-09)

BUG FIXES

- database: Fix ApplyConnectionsConfig

BUILD

- fix(deps): update module github.com/spf13/viper to v1.10.1
- fix(deps): update module github.com/ory/dockertest/v3 to v3.8.1

## v0.27.1 (Released 2021-12-14)

IMPROVEMENTS

- log: add nil pointer check in adding log contexts
- database: close sql test resources

BUILD

- fix(deps): update github.com/mitchellh/mapstructure to v1.4.3

## v0.27.0 (Released 2021-11-05)

BREAKING CHANGES

- config: fail loading if there are unused (extra) fields

IMPROVEMENTS

- config: verify blank strings replace populated strings
- database: verify TLS connections work as expected
- database: check sql rows error

BUILD

- build: enable gosec, fix go-kit depreciations
- fix(deps): update module github.com/go-kit/kit to v0.12.0
- fix(deps): update module github.com/mattn/go-sqlite3 to v1.14.9

## v0.26.1 (Released 2021-11-01)

ADDITIONS

- database: add `multiStatement=true` to DSN to allow multiple statements in a migration

## v0.26.0 (Released 2021-09-28)

IMPROVEMENTS

- log: allow fetching of the values and setting up the log values to be sorted by keys
- log: add Details method to return a map of context values

## v0.25.0 (Released 2021-09-28)

ADDITIONS

- database: add TLS support to MySQL

BUILD

- fix(deps): update module github.com/golang-migrate/migrate/v4 to v4.15.0
- fix(deps): update module github.com/ory/dockertest/v3 to v3.8.0
- fix(deps): update module github.com/spf13/viper to v1.9.0

## v0.24.0 (Released 2021-09-10)

ADDITIONS

- log: add `Int64(..)` valuer

## v0.23.0 (Released 2021-08-11)

IMPROVEMENTS

- log: added more valuer types (`uint32`, `uint64`, `float32`)

## v0.22.0 (Released 2021-08-09)

IMPROVEMENTS

- log: added debug log level

## v0.21.1 (Released 2021-07-21)

BUG FIXES

- time: fix `AddBankingDay` calculation around weekend holidays

## v0.21.0 (Released 2021-07-15)

IMPROVEMENTS

- database: support disabling cgo by removing sqlite support
- Set meaningful names for databases used in tests (#172)

BUILD

- fix(deps): update module github.com/go-kit/kit to v0.11.0
- fix(deps): update module github.com/google/uuid to v1.3.0
- fix(deps): update module github.com/mattn/go-sqlite3 to v1.14.8
- fix(deps): update module github.com/spf13/viper to v1.8.1

## v0.20.0 (Released 2021-06-21)

ADDITIONS

- database: mask mysql password in JSON marshaling

BUILD

- fix(deps): update module github.com/prometheus/client_golang to v1.11.0
- fix(deps): update module github.com/ory/dockertest/v3 to v3.7.0
- build: update gotilla/websocket and spf13/viper

## v0.19.0 (Released 2021-05-11)

ADDITIONS

- Add database Transaction functions

## v0.18.3 (Released 2021-04-30)

IMPROVEMENTS

- config: include which file is missing

BUILD

- fix(deps): update module github.com/go-sql-driver/mysql to v1.6.0
- fix(deps): update module github.com/mattn/go-sqlite3 to v1.14.7
- fix(deps): update module github.com/ory/dockertest/v3 to v3.6.5

## v0.18.2 (Released 2021-04-09)

ADDITIONS

- add subrouter function to Server (#160)

## v0.18.1 (Released 2021-04-09)

ADDITIONS

- log: timeOrNil() (#157)
- Adding in a .Nil() method to the logged error so you can log and return nil in the same oneliner

## v0.18.0 (Released 2021-03-29)

IMPROVEMENTS

- database: set STRICT_ALL_TABLES on mysql connections

BUILD

- Bump gogo/protobuf to fix CVE
- fix(deps): update module github.com/prometheus/client_golang to v1.10.0

## v0.17.0 (Released 2021-02-18)

ADDITIONS

- Adding in configurable sql connections

## v0.16.0 (Released 2021-02-10)

BREAKING CHANGES

- Adding timezone/location parameter to time Now func

IMPROVEMENTS

- docs: update "Getting Help" section

BUILD

- chore(deps): update module golang-migrate/migrate/v4 to v4.14.1
- chore(deps): update module google/uuid to v1.2.0
- chore(deps): update module mattn/go-sqlite3 to v1.14.6
- chore(deps): update module ory/dockertest/v3 to v3.6.3
- chore(deps): update module prometheus/client_golang to v1.9.0

## v0.15.0 (Released 2020-11-16)

IMPROVEMENTS

- logging: Restricts the logging values from `interface{}` to specific types that are easily converted into a log line.

## v0.14.2 (Released 2020-11-11)

IMPROVEMENTS

- database: run and share single MySQL docker container with all tests

## v0.14.1 (Released 2020-11-06)

FIXES:

- database: fix parallel tests data race in migrations


## v0.14.0 (Released 2020-10-21)

**BREAKING CHANGES**
- database: removed `InMemorySqliteConfig` object

ADDITIONS
- log: add timestamp field to all logs

## v0.13.1

FIXES:

- database: CreateTestSqliteDB didn't use path of SQLite DB for migrations

## v0.13.0 (Released 2020-10-16)

**BREAKING CHANGES**

- log: `Logger.LogError` and `Logger.LogErrorf` no longer return an `error`, they will return `LoggedError` which can be called with `Err()` to return an `error`
- database: changed signature of `New` and `NewAndMigrate` functions by reordering arguments and changing return types
- database: renamed Sqlite to SQlite and MySql to MySQL in database config

ADDITIONS

- database: load sql files for migrations from `/migrations` directory

FIXES:

- database: fix leaked DB connection created by migrator

## v0.12.0 (Released 2020-10-14)

ADDITIONS
- log: package for generating structured logs
- config: package for loading the app configuration
- stime: package for fetching system time and mocking time in tests
- api: schema for base error model


## v0.11.1 (Released 2019-09-09)

ADDITIONS

- http: Add GetSkipAndCount(r *http.Request) to get 'skip' and 'count' vars from request

## v0.11.0 (Released 2020-01-16)

ADDITIONS

- admin: add a handler to print the version on 'GET /version'

IMPROVEMENTS

- http/bind: rename ofac as watchman

BUILD

- Update module prometheus/client_golang to v1.3.0
- Update Copyright headers for 2020
- chore(deps): update module hashicorp/golang-lru to v0.5.4

## v0.10.0 (Released 2019-08-13)

BREAKING CHANGES

We've renamed `http.GetRequestID` and `http.GetUserID` from `http.Get*Id` to match Go's preference for `ID` suffixes.

ADDITIONS

- idempotent: add [`Header(*http.Request) string`](https://godoc.org/github.com/moov-io/base/idempotent#Header) and `HeaderKey`
- http/bind: add Wire HTTP service/port binding
- http/bind: add customers port
- http/bind: rename gl to accounts
- time: expose ISO 8601 format

BUG FIXES

- http: respond with '429 PreconditionFailed' if X-Idempotency-Key has been seen before

IMPROVEMENTS

- idempotent: bump up max header length
- admin: bind on a random port and return it in BindAddr on `:0`
- build: enable windows in TravisCI

## v0.9.0 (Released 2019-03-04)

ADDITIONS

- admin: Added `AddLivenessCheck` and `AddReadinessCheck` for HTTP health checks

## v0.8.0 (Released 2019-02-01)

ADDITIONS

- Added `Has` and `Match` functions to support type-based error handling

## v0.7.0 (Released 2019-01-31)

ADDITIONS

- Add `ID() string` to return a random identifier.

## v0.6.0 (Released 2019-01-25)

ADDITIONS

- admin: [`Server.AddHandler`](https://godoc.org/github.com/moov-io/base/admin#Server.AddHandler) for extendable commands
- http/bind: Add [Fed](https://github.com/moov-io/fed) service

## v0.5.1 (Released 2019-01-17)

BUG FIXES

- http: fix panic in ResponseWriter.WriteHeader

## v0.5.0 (Released 2019-01-17)

BUG FIXES

- http: don't panic if nil idempotent.Recorder is passed to ResponseWriter

ADDITIONS

- http/bind: Add [OFAC](https://github.com/moov-io/ofac) and [GL](https://github.com/moov-io/gl) services
- k8s: Add [`Inside()`](https://godoc.org/github.com/moov-io/base/k8s#Inside) for cluster awareness.
- docker: Add [`Enabled()`](https://godoc.org/github.com/moov-io/base/docker#Enabled) for compatability checks.

## v0.4.0 (Released 2019-01-11)

BREAKING CHANGES

- time: default times to UTC rather than Eastern.

## v0.3.1 (Released 2019-01-09)

- error: Add `ParseError` and `ErrorList` types.
- time: Prevent negative times in `NewTime(t time.Time)`

## v0.3.0 (Released 2019-01-07)

ADDITIONS

- Add ParseError and ErrorList. (See: [moov-io/base #23](https://github.com/moov-io/base/issues/23))

## v0.2.1 (Released 2019-01-03)

BUG FIXES

- http: Add OPTIONS to Access-Control-Allow-Methods

## v0.2.0 (Released 2018-12-18)

ADDITIONS

- Add `base.Time` as an embedded `time.Time` with banktime methods. (AddBankingDay, IsWeekend)

## v0.1.0 (Released 2018-12-17)

- Initial release
