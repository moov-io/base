package telemetry

import (
	"io"
	"os"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

func newDiscardExporter() (trace.SpanExporter, error) {
	return newJsonExporter(io.Discard)
}

func newStdoutExporter() (trace.SpanExporter, error) {
	return newJsonExporter(os.Stdout)
}

// newExporter returns a console exporter.
func newJsonExporter(w io.Writer) (trace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
	)
}
