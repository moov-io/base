package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/codes"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var _ tracesdk.SpanExporter = &filteredExporter{}

func NewFilteredExporter(inner tracesdk.SpanExporter) tracesdk.SpanExporter {
	return &filteredExporter{inner: inner}
}

type filteredExporter struct {
	inner tracesdk.SpanExporter
}

func (fe *filteredExporter) Shutdown(ctx context.Context) error {
	return fe.inner.Shutdown(ctx)
}

func (fe *filteredExporter) ExportSpans(ctx context.Context, spans []tracesdk.ReadOnlySpan) error {
	in := []tracesdk.ReadOnlySpan{}

	for _, span := range spans {
		if fe.AlwaysInclude(span) {
			in = append(in, span)
			continue
		}

		if HasSpanDrop(span) {
			continue
		}

		if IsEmptyConsume(span) {
			continue
		}

		in = append(in, span)
	}

	return fe.inner.ExportSpans(ctx, in)
}

func (fe *filteredExporter) AlwaysInclude(s tracesdk.ReadOnlySpan) bool {
	return len(s.Links()) > 0 ||
		len(s.Events()) > 0 ||
		s.ChildSpanCount() > 0 ||
		s.Status().Code == codes.Error
}

// Allows for services to just flag a span to be dropped.
func HasSpanDrop(s tracesdk.ReadOnlySpan) bool {
	for _, attr := range s.Attributes() {
		if attr.Key == DropSpanKey && attr.Value.AsBool() {
			return true
		}
	}

	return false
}

// Detects if its an event that was consumed but ignored.
// These can cause a lot of cluttering in the traces and we want to filter them out.
func IsEmptyConsume(s tracesdk.ReadOnlySpan) bool {
	if s.SpanKind() == trace.SpanKindConsumer {

		// If it took less than a millisecond and has no child spans, the event was most likely ignored...
		if s.EndTime().Sub(s.StartTime()) < time.Millisecond {
			return true
		}
	}

	return false
}
