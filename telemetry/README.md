# Telemetry

This package acts to wire up the various components of OpenTelemetry transparently to Moov developers.
As such, basic understanding of how OpenTelemetry project works is critical to understanding of this library.
The documentation can be found here: https://opentelemetry.io/docs/instrumentation/go/

------------------
### OpenTelemetry Purpose
To allow users, especially of distributed systems, to track execution through the stack via a single Trace that is composed of a series of Spans.
The Trace will persist from ingress in the system until processing is complete. Traces are separated by spans, that are typically started and stopped at microservice boundaries but can also be started/stopped anywhere the client chooses. Spans started within other spans have a parent child relationship to the span that they were derived from.

OpenTelemetry not only lets you trace execution, allowing you to delineate functional blocks of the overarching execution via spans, but also allows you to include meta-data such as attributes and baggage (both covered below) to ensure the metrics being generated contain adequate information to properly query from.

------------------
### OpenTelemetry Components

**Attributes**: Key/value pairs associated with a span, can be added at any time during execution. Attaching attributes to span is one of the most common ways we as Moov developers will add tracing information.

**Context**: While technically not an OpenTelemetry construct, it is very important to understand that this is how spans are propagated (they are embedded inside the context, and passed across microservice boundaries via baggage).

**Baggage**: A bucket for metadata that is passed along the entire Trace. Library warns that any data you pass in this you should expect to be visible and to be careful. Can somewhat be thought of Attributes associated with a Trace instead of a Span. Baggage is the means that transfers span information allowing it to continue through all the microservices. Baggage is added to all internal API requests and to the producing of events.

**Exporter**: Sends spans (that have been sent to it via a batch processor) to the system to be recorded. Two most common variants are the stdoutexporter and the gRPC exporter. Only used by this code base's boilerplate.

**Instrumentation Libraries**: Pre-made libraries that provide quality of life instrumentation, these typically cover auxiliary libraries such as kafka or net/http.

**Propagators**: Establishes the technique to pass Spans across microservice boundaries. Example below allows the Baggage to be included in the context that is marshalled/unmarshalled across boundaries. Only used by this code base's boilerplate.

**Resource**: Represents metadata about the physical device the code is executing on. Only used by this code base's boilerplate. Example below:

**Sampler**: Determines what percent of spans are recorded. You can always send, never send, or any ratio between. Only used by this code base's boilerplate.

**Span**: The main unit of work in OpenTelemetry, and the API Moov developers need to be most familiar with, spans delineates a functional block of execution. Creating a span from a tracer accepts a ctx, if that context already contains a span the new span becomes the child of that span. Contains an API that allows adding attributes, span status, and error details.

Use `telemetry.StartSpan` to start a span:

```go
ctx, span := telemetry.StartSpan(ctx, "SpanName")
defer span.End()
```

**SpanProcessor**: Receives span information and is the pipeline that provides it to the exporter. It is configured within the TraceProvider, multiple SpanProcessors can be configured. BatchSpanProcessor which sends telemetry in batches and should be used for production applications (use NewBatchSpanProcessor). SimpleSpanProcessor is a non-production debugging processor (use NewSimpleSpanProcessor)

**Tracer**: Created from its factory object TracerProvider and provided a name. By OpenTelemetry standards this name should match the package name, but we do not follow this explicitly. Tracers are factories for Spans, and the main purpose of the Tracer is to associate the spans it spawns with the correct package name.

**TracerProvider**: The "glue" of OpenTelemetry, you define a factory and provide it with Sampler, Batcher, A factory of Tracers, defined only one time in an application. Accessed globally via the following:
otel.SetTracerProvider(tp)
otel.GetTracerProvider()

------------------
### Setup

Call `SetupTelemetry` once when the process starts. It picks an exporter in this order:

1. An in-process test writer (`telemetry.TestConfig(w)`)
2. `OTEL_EXPORTER_OTLP_ENDPOINT` (OTLP gRPC collector)
3. `HONEYCOMB_API_KEY` (Honeycomb at `api.honeycomb.io:443`)
4. `Config.Stdout`
5. `Config.OpenTelemetryCollector` (`OtelConfig{Host, TLS}`)
6. `Config.Honeycomb` (`HoneycombConfig{URL, Team}`)
7. A discard exporter, so traces are dropped

If `Config.ServiceName` or `Config.ServiceNamespace` is empty, `SetupTelemetry` uses `MOOV_SERVICE_NAME` and `MOOV_SERVICE_NAMESPACE`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/moov-io/base/telemetry"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	ctx := context.Background()
	shutdown, err := telemetry.SetupTelemetry(ctx, telemetry.Config{
		ServiceName: "example-service",
	}, "v1.0.0")
	if err != nil {
		panic(err)
	}
	defer shutdown()

	ctx, span := telemetry.StartSpan(ctx, "example-operation")
	defer span.End()

	telemetry.SetAttributes(ctx, attribute.String("account_id", "12345"))
	if err := doWork(ctx); err != nil {
		_ = telemetry.RecordError(ctx, err)
	}
}

func doWork(ctx context.Context) error {
	_, span := telemetry.StartSpan(ctx, "do-work")
	defer span.End()
	fmt.Println("working")
	return nil
}
```

`StartSpan` creates a child span when `ctx` already holds a span, or a root span otherwise. End every span. `RecordError` records the error on the current span and returns it. `SpanFromContext` and `AddEvent` are also available.

A compile-checked walkthrough is in [example_test.go](example_test.go).

The lower-level exporter and `TracerProvider` setup that `SetupTelemetry` performs is in [config.go](config.go), [collector.go](collector.go) and [honey.go](honey.go). You do not need to call `otel.SetTracerProvider` yourself after `SetupTelemetry`.
