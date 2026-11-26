## moov-io/base/admin

Package admin implements an `http.Server` which can be used for operations and monitoring tools. It's designed to be shipped (and ran) inside an existing Go service.

Here's an example of adding `admin.Server` to serve Prometheus metrics, health checks and `/version`:

```go
package main

import (
	"fmt"

	"github.com/moov-io/base/admin"
	"github.com/moov-io/base/log"
)

func main() {
	logger := log.NewDefaultLogger()

	adminServer, err := admin.New(admin.Opts{
		Addr: ":9090",
	})
	if err != nil {
		logger.Fatal().LogError(err)
		return
	}
	defer adminServer.Shutdown()

	adminServer.AddLivenessCheck("db", func() error { return nil })
	adminServer.AddReadinessCheck("kafka", func() error {
		return fmt.Errorf("not connected")
	})
	adminServer.AddVersionHandler("v1.2.3")

	logger.Info().Logf("admin listening on %s", adminServer.BindAddr())
	if err := adminServer.Listen(); err != nil {
		logger.Fatal().LogErrorf("problem starting admin http: %v", err)
	}
}
```

`Listen` blocks until you call `Shutdown` or the server returns an error. After the example starts, `GET /live` returns `200` with `{"db":"good"}`, `GET /ready` returns `400` with `{"kafka":"not connected"}`, and `GET /version` returns `v1.2.3`. A complete walkthrough is in [example_test.go](example_test.go).

### Endpoints

An Admin server has some default endpoints that are useful for operational support and monitoring.

#### Liveness Probe

This endpoint inspects a set of liveness functions and returns `200 OK` if all functions return without errors. If errors are found then a `400 Bad Request` response with a JSON object is returned describing the errors.

```
GET /live
```

Liveness probes can be registered with the following callback:

```
func (s *Server) AddLivenessCheck(name string, f func() error)
```

#### Readiness Probe

This endpoint inspects a set of readiness functions and returns `200 OK` if all functions return without errors. If errors are found then a `400 Bad Request` response with a JSON object is returned describing the errors.

```
GET /ready
```

Readiness probes can be registered with the following callback:

```
func (s *Server) AddReadinessCheck(name string, f func() error)
```

#### Version

`AddVersionHandler` registers `GET /version` and writes the version string you pass in.

```
func (s *Server) AddVersionHandler(version string)

GET /version
v1.2.3
```

### Extension points

`AddHandler` registers an extra `http.HandlerFunc` on the admin server:

```
func (s *Server) AddHandler(path string, hf http.HandlerFunc)
```

`Subrouter` returns a gorilla/mux subrouter for a path prefix. Middleware on the subrouter does not wrap the parent `/live` and `/ready` routes:

```
sub := adminServer.Subrouter("/prefix")
sub.Use(someMiddleware)
sub.HandleFunc("/resource", ResourceHandler)
```

A request for `/prefix/resource` goes through `someMiddleware`. Requests for `/live` and `/ready` do not.

`Opts.Timeout` sets the read, write and idle timeouts when you call `New`. You can change them later with `SetReadTimeout`, `SetWriteTimeout` and `SetIdleTimeout`.

### Profiling

`GET /debug/pprof/` and the standard Go pprof handlers are registered by default so existing callers keep current behavior.

To omit them (recommended in production), set `Pprof.Enabled` to false:

```
adminServer, err := admin.New(admin.Opts{
	Addr: ":9090",
	Pprof: &admin.Pprof{
		Enabled: false,
	},
})
```

When `Pprof.Secret` is set, `/debug/pprof` requires `Authorization: Bearer <secret>` or `X-Pprof-Token: <secret>`. An empty secret keeps unauthenticated access.

`Pprof.Block` and `Pprof.Mutex` enable runtime block/mutex sampling (`SetBlockProfileRate` / `SetMutexProfileFraction`) and their handlers. They default to false when `Pprof` is set. A nil `Pprof` keeps the historical `PPROF_BLOCK` / `PPROF_MUTEX` behavior.

```
adminServer, err := admin.New(admin.Opts{
	Addr: ":9090",
	Pprof: &admin.Pprof{
		Enabled: true,
		Secret:  secret,
		Block:   true,
		Mutex:   true,
	},
})
```

Per-profile `PPROF_*` environment variables still disable individual handlers when pprof is enabled (for example `PPROF_ALLOCS=no`).

### Metrics

This endpoint returns prometheus metrics registered to the [prometheus/client_golang](https://github.com/prometheus/client_golang) singleton metrics registry. Their `promauto` package can be used to add Counters, Gauges, Histograms, etc. The default Go metrics provided by `prometheus/client_golang` are included.

```
GET /metrics

...
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 0
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
```
