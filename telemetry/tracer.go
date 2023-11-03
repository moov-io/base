package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	InstrumentationName = "moov.io"
	AttributeTag        = "otel"
	MaxArrayAttributes  = 10
)

func GetTracer(opts ...trace.TracerOption) trace.Tracer {
	return otel.GetTracerProvider().Tracer(InstrumentationName, opts...)
}

func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if ctx == nil {
		ctx = context.Background()
	}
	return GetTracer().Start(ctx, spanName, opts...)
}

func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func AddEvent(ctx context.Context, name string, options ...trace.EventOption) {
	SpanFromContext(ctx).AddEvent(name, options...)
}

// RecordError will record err as an exception span event for this span. It will also
// return the err passed in.
func RecordError(ctx context.Context, err error, options ...trace.EventOption) error {
	options = append(options, trace.WithStackTrace(true))
	SpanFromContext(ctx).RecordError(err, options...)
	return err
}

// SetAttributes sets kv as attributes of the Span. If a key from kv
// already exists for an attribute of the Span it will be overwritten with
// the value contained in kv.
func SetAttributes(ctx context.Context, kv ...attribute.KeyValue) {
	SpanFromContext(ctx).SetAttributes(kv...)
}
