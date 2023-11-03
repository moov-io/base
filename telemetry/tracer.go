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

// GetTracer returns a unique Tracer scoped to be used by instrumentation code
// to trace computational workflows.
func GetTracer(opts ...trace.TracerOption) trace.Tracer {
	return otel.GetTracerProvider().Tracer(InstrumentationName, opts...)
}

// StartSpan will create a Span and a context containing the newly created Span.
//
// If the context.Context provided contains a Span then the new span will be a child span,
// otherwise the new span will be a root span.
//
// OTEL recommends creating all attributes via `WithAttributes()` SpanOption when the span is created.
//
// Created spans MUST be ended with `.End()` and is the responsibility of callers.
func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if ctx == nil {
		ctx = context.Background()
	}
	return GetTracer().Start(ctx, spanName, opts...)
}

// SpanFromContext returns the current Span from ctx.
//
// If no Span is currently set in ctx an implementation of a Span that performs no operations is returned.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddEvent adds an event the Span in `ctx` with the provided name and options.
func AddEvent(ctx context.Context, name string, options ...trace.EventOption) {
	SpanFromContext(ctx).AddEvent(name, options...)
}

// RecordError will record err as an exception span event for this span. It will also return the err passed in.
func RecordError(ctx context.Context, err error, options ...trace.EventOption) error {
	options = append(options, trace.WithStackTrace(true))
	SpanFromContext(ctx).RecordError(err, options...)
	return err
}

// SetAttributes sets kv as attributes of the Span. If a key from kv already exists for an
// attribute of the Span it will be overwritten with the value contained in kv.
func SetAttributes(ctx context.Context, kv ...attribute.KeyValue) {
	SpanFromContext(ctx).SetAttributes(kv...)
}
