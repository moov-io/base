package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
)

func isOtelEnvironmentSet() bool {
	return os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != ""
}

// Creates a exporter thats completely built by environment flags.
// References:
// - https://opentelemetry.io/docs/specs/otel/protocol/exporter/
// - https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/
func newOtelExporterFromEnvironment(ctx context.Context) (*otlptrace.Exporter, error) {
	client := otlptracegrpc.NewClient()
	return otlptrace.New(ctx, client)
}
