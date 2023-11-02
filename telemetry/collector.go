package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	// Add in gzip
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
)

type OtelConfig struct {
	Host string
	TLS  bool
}

func newOpenTelementyCollectorExporter(ctx context.Context, config OtelConfig) (*otlptrace.Exporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.Host),
	}

	if config.TLS {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	} else {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	client := otlptracegrpc.NewClient(opts...)
	return otlptrace.New(ctx, client)
}
