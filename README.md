moov-io/base
===
[![GoDoc](https://godoc.org/github.com/moov-io/base?status.svg)](https://pkg.go.dev/github.com/moov-io/base)
[![Build Status](https://github.com/moov-io/base/workflows/Go/badge.svg)](https://github.com/moov-io/base/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/moov-io/base)](https://goreportcard.com/report/github.com/moov-io/base)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/moov-io/base/master/LICENSE)

Package `github.com/moov-io/base` implements core libraries used in multiple Moov projects. Other Go services can import these packages. Some packages (for example [`http`](https://pkg.go.dev/github.com/moov-io/base/http) and [`http/bind`](https://pkg.go.dev/github.com/moov-io/base/http/bind)) are written for Moov services and may not fit every project.

## Packages

| Package | Description |
|---------|-------------|
| [github.com/moov-io/base](https://pkg.go.dev/github.com/moov-io/base) | Random IDs, error lists (`ErrorList`, `Has`, `Match`) and Federal Reserve banking-day helpers (`Time`) |
| [admin](https://pkg.go.dev/github.com/moov-io/base/admin) | HTTP server for liveness, readiness, Prometheus metrics and pprof. See [admin/README.md](admin/README.md). |
| [build](https://pkg.go.dev/github.com/moov-io/base/build) | Logs Go module and build information |
| [config](https://pkg.go.dev/github.com/moov-io/base/config) | Loads YAML defaults and optional `APP_CONFIG` / `APP_CONFIG_SECRETS` overlays. See [config/README.md](config/README.md). |
| [database](https://pkg.go.dev/github.com/moov-io/base/database) | MySQL, PostgreSQL and Spanner connections and migrations |
| [docker](https://pkg.go.dev/github.com/moov-io/base/docker) | Reports whether the `docker` binary is on `PATH` |
| [http](https://pkg.go.dev/github.com/moov-io/base/http) | Problem responses, CORS and request logging for Moov services |
| [http/bind](https://pkg.go.dev/github.com/moov-io/base/http/bind) | Well-known local bind addresses for Moov services |
| [k8s](https://pkg.go.dev/github.com/moov-io/base/k8s) | Detects whether the process runs inside Kubernetes |
| [log](https://pkg.go.dev/github.com/moov-io/base/log) | Structured logging. See [log/README.md](log/README.md). |
| [mask](https://pkg.go.dev/github.com/moov-io/base/mask) | Masks passwords for log output |
| [randx](https://pkg.go.dev/github.com/moov-io/base/randx) | Random integers in a range |
| [ratex](https://pkg.go.dev/github.com/moov-io/base/ratex) | Rate limiting and retry helpers |
| [sql](https://pkg.go.dev/github.com/moov-io/base/sql) | Observed `database/sql` helpers |
| [stime](https://pkg.go.dev/github.com/moov-io/base/stime) | `TimeService` with system and static clocks |
| [strx](https://pkg.go.dev/github.com/moov-io/base/strx) | Small string helpers (`Or`, `Yes`) |
| [telemetry](https://pkg.go.dev/github.com/moov-io/base/telemetry) | OpenTelemetry setup. See [telemetry/README.md](telemetry/README.md). |

Runnable examples are in [example_test.go](example_test.go) and in the `admin`, `config`, `log` and `telemetry` packages.

## Getting Started

This project uses [Go Modules](https://go.dev/blog/using-go-modules) and Go 1.26 or newer. See [Golang's install instructions](https://go.dev/doc/install) for help setting up Go.

To add the module to your service:

```
go get github.com/moov-io/base@latest
go mod tidy
```

`go get github.com/moov-io/base@latest` records the root module. Import a subpackage (for example `github.com/moov-io/base/admin`) and run `go mod tidy` so that the subpackage dependencies are added to `go.sum`.

To clone the repository:

```
git clone https://github.com/moov-io/base.git
```

We recommend a [tagged release](https://github.com/moov-io/base/releases/latest) in production.

### Quick start

```go
package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/moov-io/base"
	"github.com/moov-io/base/log"
)

func main() {
	logger := log.NewDefaultLogger()
	logger.Info().Logf("id=%s", base.ID())

	first := errors.New("first")
	var list base.ErrorList
	list.Add(first)
	fmt.Println(base.Has(list, first))

	eastern, err := time.LoadLocation("America/New_York")
	if err != nil {
		logger.Fatal().LogError(err)
		return
	}
	// Thanksgiving Day 2026 is a Federal Reserve holiday.
	day := base.NewTime(time.Date(2026, time.November, 26, 10, 0, 0, 0, eastern))
	fmt.Println(day.IsHoliday(), day.IsBankingDay())
	fmt.Println(day.AddBankingDay(1).Format("2006-01-02"))
}
```

`Time` follows the [Federal Reserve Bank holiday schedule](https://www.frbservices.org/about/holiday-schedules). For a holiday on Saturday, Reserve Banks are open on the Friday before. For a holiday on Sunday, they are closed on the following Monday. `GetHoliday` can still return that holiday on the Friday before a Saturday holiday, while `IsHoliday` is false and `IsBankingDay` is true.

Config and admin examples are in [config/README.md](config/README.md) and [admin/README.md](admin/README.md).

## Configuration

Environment variables used by this repository:

| Environmental Variable | Package | Description | Default |
|------------------------|---------|-------------|---------|
| `APP_CONFIG` | config | Path to a YAML file that overlays the default config | unset |
| `APP_CONFIG_SECRETS` | config | Path to a YAML file of secret overlays, merged after `APP_CONFIG` | unset |
| `MOOV_LOG_FORMAT` | log | Log output format: `json`, `logfmt`, `nop` or `noop`. `LOG_FORMAT` is used if `MOOV_LOG_FORMAT` is empty. | `logfmt` |
| `LOG_FORMAT` | log | Fallback for `MOOV_LOG_FORMAT` | unset |
| `PPROF_ALLOCS`, `PPROF_BLOCK`, `PPROF_GOROUTINE`, `PPROF_GOROUTINELEAK`, `PPROF_HEAP`, `PPROF_MUTEX`, `PPROF_PROFILE`, `PPROF_SYMBOL`, `PPROF_THREADCREATE`, `PPROF_TRACE` | admin | Set a profile to `no` (or any value other than empty, `yes` or `true`) to disable that pprof handler. `PPROF_BLOCK` and `PPROF_MUTEX` apply only when `Opts.Pprof` is nil. | enabled |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | telemetry | When set, `SetupTelemetry` exports spans to this OTLP collector | unset |
| `HONEYCOMB_API_KEY` | telemetry | When set (and no OTLP endpoint is set), `SetupTelemetry` exports spans to Honeycomb | unset |
| `MOOV_SERVICE_NAME` | telemetry | Service name for traces when `Config.ServiceName` is empty | unset |
| `MOOV_SERVICE_NAMESPACE` | telemetry | Service namespace for traces when `Config.ServiceNamespace` is empty | unset |
| `MYSQL_TIMEOUT` | database | Timeout parameter on the MySQL DSN | `30s` |
| `MYSQL_MAX_CONNECTIONS` | database | `SetMaxOpenConns` for MySQL | `16` |
| `MOOV_CORS_ALLOW_ORIGINS` | http | Comma-separated list of exact `Origin` values allowed for credentialed CORS | localhost origins only |
| `KUBERNETES_SERVICE_ACCOUNT_FILEPATH` | k8s | Extra path checked by `k8s.Inside()` | `/var/run/secrets/kubernetes.io` and the default service-account token path |

## Testing

Run unit tests with:

```
go test -short ./...
```

`-short` skips tests that need MySQL, PostgreSQL or Spanner. `go test ./...` without `-short` fails those tests when the services are not running, and the `sql` package retries for about 22 seconds for each connection.

To run the integration tests:

1. Install Docker, make, git, wget and OpenSSL.
2. Run `make setup`. This writes TLS certificates to `./testcerts` (`database/testdata/gencerts.sh`) and starts MySQL (`mysql:26-oracle` on port 3306), the Spanner emulator (ports 9010 and 9020) and PostgreSQL (`postgres:19.0` on port 5432) from [docker-compose.yml](docker-compose.yml).
3. Run `go test ./...`.
4. Run `make teardown` to stop the containers.

`make check` downloads [lint-project.sh](https://raw.githubusercontent.com/moov-io/infra/master/go/lint-project.sh) and runs the project linters. On Windows, `make check` runs `go test ./... -short`.

## Getting Help

 channel | info
 ------- | -------
Twitter [@moov](https://twitter.com/moov)	| You can follow Moov.io's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories.
[GitHub Issue](https://github.com/moov-io/base/issues) | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.
[moov-io slack](https://slack.moov.io/) | Join our slack channel to have an interactive discussion about the development of the project.

## Supported and Tested Platforms

- 64-bit Linux (Ubuntu, Debian), macOS, and Windows

## Contributing

Yes please! Please review our [Contributing guide](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md) to get started!

This project uses [Go Modules](https://go.dev/blog/using-go-modules) and uses Go 1.26 or higher. See [Golang's install instructions](https://go.dev/doc/install) for help setting up Go. You can download the source code and we offer [tagged and released versions](https://github.com/moov-io/base/releases/latest) as well. We highly recommend you use a tagged release for production.

## License

Apache License 2.0 See [LICENSE](LICENSE) for details.
