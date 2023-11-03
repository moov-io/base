# Telemetry

This package acts to wire up the various components of OpenTelemetry transparently to Moov developers.
As such, basic understanding of how OpenTelemetry project works is critical to understanding of this library.
The documentation can be found here: https://opentelemetry.io/docs/instrumentation/go/

------------------
### OpenTelemetry Purpose
To allow users, especially of distributed systems, to track execution through the stack via a single Trace that is composed of a series of Spans.
The Trace will persist from ingress in the system until processing is complete. Traces are seperated by spans, tht are typically started and stopped at microservice boundaries but can also be started/stopped anywhere the client chooses. Spans started within other spans have a praent child relationship to the span than they were derived from.

OpenTelemetry not only lets you trace execution, allowing you to delineate functional locks of the overarching execution via spans, but also allows you to include meta-data such as attibutes and baggage (both covered below) to ensure the metrics being generated contain adequate information to properly query from.

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
```
oneotel.Tracer(packageName).Start(ctx, “SpanName”)
```

**SpanProcessor**: Receives span information and is the pipeline that provides it to the exporter. It is configured within the TraceProvider, multiple SpanProcessors can be configured. BatchSpanProcessor which sends telemetry in batches and should be used for production applications (use NewBatchSpanProcessor). SimpleSpanProcessor is a non-production debugging processor (use NewSimpleSpanProcessor)

**Tracer**: Created from its factory object TracerProvider and provided a name. By OpenTelemetry standards this name should match the package name, but we do not follow this explicitly. Tracers are factories for Spans, and the main purpose of the Tracer is to associate the spans it spawns with the correct package name.

**TracerProvider**: The "glue" of OpenTelemetry, you define a factory and provide it with Sampler, Batcher, A factory of Tracers, defined only one time in an application. Accessed globally via the following:
otel.SetTracerProvider(tp)
otel.GetTracerProvider()

------------------
### OpenTelemetry "Putting it all Together"

To configure a system to generate and export spans the following steps must be done:

1. Define your exporter, where do you want your data to go? In this package, honey.go, stdout.go and collector.go all store methods to create Exporters. The beginning blocks of logic in SetupTelemetry determines which one should be used based on configuration values. Honeycomb exporter shown as example:
```go
opts := []otlptracegrpc.Option{
    otlptracegrpc.WithCompressor("gzip"),
    otlptracegrpc.WithEndpoint(config.URL),
    otlptracegrpc.WithHeaders(map[string]string{
        "x-honeycomb-team": config.Team,
    }),
    otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
}

client := otlptracegrpc.NewClient(opts...)
return otlptrace.New(ctx, client)
```
2. Define your TracerProvider, typically this will require you to define the other constructs TracerProvider uses at the same time:
```go
resource := resource.NewWithAttributes(
    semconv.SchemaURL,
    semconv.ServiceNameKey.String(config.ServiceName),
    semconv.ServiceVersionKey.String(version),
)

return trace.NewTracerProvider(
    trace.WithSampler(trace.AlwaysSample()),
    trace.WithBatcher(exp,
        trace.WithMaxQueueSize(3*trace.DefaultMaxQueueSize),
        trace.WithMaxExportBatchSize(3*trace.DefaultMaxExportBatchSize),
        trace.WithBatchTimeout(5*time.Second),
    ),
    trace.WithResource(resource),
)
```
3. Globally set the defined TracerProvider:
```go
otel.SetTracerProvider(tp)
```
4. Allow Trace information to be propagated across microservice boundaries:
```go
otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{},
    propagation.Baggage{},
))
```
5. Acquire a Tracer from the TracerProvider, and start a Span wrapping your desired block of execution*
[IMPORTANT NOTE]: Our libraries (usually) automatically start spans for you:
   1. Kakfa Consumers:  [start a span for processing each consumed event](https://github.com/moovfinancial/events/blob/82ed357686f9bc920568274828a36dffc1f37fb4/go/consumer/processor_event.go#L137).
   2. Kafka Producers:  [start a span for producing each event](https://github.com/moovfinancial/events/blob/ef809bc0d63f3ddec07f39da47edad3e39145dab/go/producer/producer.go#L170)
   3. HTTPS Endpoints: [start a Span for each request](https://github.com/moovfinancial/go-zero-trust/blob/v2/pkg/middleware/middleware.go#L50) and are protected by the go-zero-trust-middleware.
6. Add attributes, state code, or error messages at any location during a span, acquiring the span from the context if needed:
```go
span := telemetry.SpanFromContext(r.Context())
attributes := []attribute.KeyValue{
    attribute.String("account_id", accountID),
    attribute.String("mode", claims.CallingAccountMode.String()),
}
if claims.CallingAccountID != nil {
    attributes = append(attributes, attribute.String("mode", *claims.CallingAccountID))
}
span.SetAttributes(attributes...)
```
------------------
#### Example of how HTTPS endpoints gain tracing via Zero-Trust library
```go
    if env.ZeroTrustMiddleware == nil {
        gatewayMiddleware, err := middleware.NewServerFromConfig(env.Logger, env.TimeService, env.Config.Gateway)
        if err != nil {
        return nil, env.Logger.Fatal().LogErrorf("failed to startup Gateway middleware: %w", err).Err()
        }
        env.ZeroTrustMiddleware = gatewayMiddleware.Handler
    }
```
followed by attaching it to your router of choice:
```go
    env.PublicRouter.Use(env.ZeroTrustMiddleware)
```
