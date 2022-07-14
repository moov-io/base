## moov-io/base/admin

Package admin implements an `http.Server` which can be used for operations and monitoring tools. It's designed to be shipped (and ran) inside an existing Go service.

Here's an example of adding `admin.Server` to serve Prometheus metrics:

```Go
import (
    "fmt"
    "os"

    "github.com/moov-io/base/admin"

    "github.com/go-kit/log"
)

var logger log.Logger

// in main.go or cmd/server/main.go

adminServer := admin.NewServer(*adminAddr)
go func() {
	logger.Log("admin", fmt.Sprintf("listening on %s", adminServer.BindAddr()))
	if err := adminServer.Listen(); err != nil {
		err = fmt.Errorf("problem starting admin http: %v", err)
		logger.Log("admin", err)
		// errs <- err // send err to shutdown channel
	}
}()
defer adminServer.Shutdown()
```

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

### Metrics

This endpoint returns prometheus metrics registered to the [prometheus/client_golang](https://github.com/prometheus/client_golang) singleton metrics registry. Their `promauto` package can be used to add Counters, Guages, Histograms, etc. The default Go metrics provided by `prometheus/client_golang` are included.

```
GET /metrics

...
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 0
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP stream_file_processing_errors Counter of stream submitted ACH files that failed processing
# TYPE stream_file_processing_errors counter
stream_file_processing_errors 0
```
