package telemetry

import (
	"context"
	"io"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	ServiceName            string
	ServiceNamespace       *string
	Stdout                 bool
	OpenTelemetryCollector *OtelConfig
	Honeycomb              *HoneycombConfig

	// Allows for testing of the output of telemetry without affecting use with config files
	testWriter io.Writer
}

// Allows for testing where the output of the traces are sent to a io.Writer instance.
func TestConfig(w io.Writer) Config {
	return Config{
		ServiceName: "test-service",
		testWriter:  w,
	}
}

type ShutdownFunc func() error

var NoopShutdown ShutdownFunc = func() error {
	return nil
}

func SetupTelemetry(ctx context.Context, config Config, version string) (ShutdownFunc, error) {
	var (
		err error
		exp tracesdk.SpanExporter
	)

	if config.testWriter != nil {
		exp, err = newJsonExporter(config.testWriter)
		if err != nil {
			return NoopShutdown, err
		}

	} else if isOtelEnvironmentSet() {
		exp, err = newOtelExporterFromEnvironment(ctx)
		if err != nil {
			return NoopShutdown, err
		}

	} else if isHoneycombEnvironmentSet() {
		exp, err = newHoneycombExporterFromEnvironment(ctx)
		if err != nil {
			return NoopShutdown, err
		}

	} else if config.Stdout {
		exp, err = newStdoutExporter()
		if err != nil {
			return NoopShutdown, err
		}

	} else if config.OpenTelemetryCollector != nil {
		exp, err = newOpenTelementyCollectorExporter(ctx, *config.OpenTelemetryCollector)
		if err != nil {
			return NoopShutdown, err
		}

	} else if config.Honeycomb != nil {
		exp, err = newHoneycombExporterFromConfig(ctx, *config.Honeycomb)
		if err != nil {
			return NoopShutdown, err
		}
	}

	// Make sure something is set for the exporter
	if exp == nil {
		exp, err = newDiscardExporter()
		if err != nil {
			return NoopShutdown, err
		}
	}

	tp := newTraceProvider(exp, config, version)

	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func() error {
		ctx := context.Background()
		tp.ForceFlush(ctx)
		return tp.Shutdown(ctx)
	}, nil
}

func newTraceProvider(exp tracesdk.SpanExporter, config Config, version string) TracerProvider {
	if config.ServiceName == "" {
		config.ServiceName = os.Getenv("MOOV_SERVICE_NAME")
	}

	if config.ServiceNamespace == nil || *config.ServiceNamespace == "" {
		ns := os.Getenv("MOOV_SERVICE_NAMESPACE")
		config.ServiceNamespace = &ns
	}

	// Wrap it so we can filter out useless traces from consuming
	exp = NewFilteredExporter(exp)

	batcher := tracesdk.WithBatcher(exp,
		tracesdk.WithMaxQueueSize(3*tracesdk.DefaultMaxQueueSize),
		tracesdk.WithMaxExportBatchSize(3*tracesdk.DefaultMaxExportBatchSize),
		tracesdk.WithBatchTimeout(5*time.Second),
	)

	// If we're using the testWriter we want to make sure its not buffering anything in the background
	if config.testWriter != nil {
		batcher = tracesdk.WithSyncer(exp)
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(config.ServiceName),
		semconv.ServiceVersionKey.String(version),
		semconv.ServiceNamespaceKey.String(*config.ServiceNamespace),
	)

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.AlwaysSample())),
		batcher,
		tracesdk.WithResource(resource),
	)

	return &tracerProvider{
		TracerProvider: tp,
	}
}

type TracerProvider interface {
	trace.TracerProvider

	ForceFlush(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type tracerProvider struct {
	*tracesdk.TracerProvider
}
