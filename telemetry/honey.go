package telemetry

import (
	"context"
	"os"

	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	// Add in gzip
	_ "google.golang.org/grpc/encoding/gzip"
)

type HoneycombConfig struct {
	URL  string
	Team string
}

func newHoneycombExporterFromConfig(ctx context.Context, config HoneycombConfig) (*otlptrace.Exporter, error) {
	return newHoneycombExporter(ctx, config.URL, config.Team)
}

func isHoneycombEnvironmentSet() bool {
	return os.Getenv("HONEYCOMB_API_KEY") != ""
}

func newHoneycombExporterFromEnvironment(ctx context.Context) (*otlptrace.Exporter, error) {
	return newHoneycombExporter(ctx, "api.honeycomb.io:443", os.Getenv("HONEYCOMB_API_KEY"))
}

func newHoneycombExporter(ctx context.Context, endpoint string, team string) (*otlptrace.Exporter, error) {
	// Configuration to export data to Honeycomb:
	//
	// 1. The Honeycomb endpoint
	// 2. Your API key, set as the x-honeycomb-team header
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithCompressor("gzip"),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithHeaders(map[string]string{
			"x-honeycomb-team": team,
		}),
		otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
	}

	client := otlptracegrpc.NewClient(opts...)
	return otlptrace.New(ctx, client)
}
