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
