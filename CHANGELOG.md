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
