package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartLinkedRootSpan starts a new root span where the parent and child spans share links to each other. This
// is particularly useful in batch processing applications where separate spans are wanted for each subprocess
// in the batch, but without cluttering the parent span.
func StartLinkedRootSpan(ctx context.Context, name string, options ...trace.SpanStartOption) *LinkedSpan {

	// new root for the children
	childOpts := append([]trace.SpanStartOption{
		trace.WithNewRoot(),
		trace.WithLinks(trace.LinkFromContext(ctx, attribute.String("link.name", "parent"))), // link to parent from child
	}, options...)

	childCtx, childSpan := StartSpan(ctx, name, childOpts...)

	// start a new span on the parent and link to the child span from the parent one.
	parentOpts := append([]trace.SpanStartOption{
		trace.WithLinks(trace.LinkFromContext(childCtx, attribute.String("link.name", "child"))), // link to parent from child
	}, options...)

	parentCtx, parentSpan := StartSpan(ctx, name, parentOpts...)

	return &LinkedSpan{
		childCtx:   childCtx,
		childSpan:  childSpan,
		parentCtx:  parentCtx,
		parentSpan: parentSpan,
	}
}

type LinkedSpan struct {
	childCtx  context.Context
	childSpan trace.Span

	parentCtx  context.Context
	parentSpan trace.Span
}

func (l *LinkedSpan) End(options ...trace.SpanEndOption) {
	l.childSpan.End(options...)
	l.parentSpan.End(options...)
}

func (l *LinkedSpan) AddEvent(name string, options ...trace.EventOption) {
	l.childSpan.AddEvent(name, options...)
	l.parentSpan.AddEvent(name, options...)
}

func (l *LinkedSpan) RecordError(err error, options ...trace.EventOption) {
	l.childSpan.RecordError(err, options...)
	l.parentSpan.RecordError(err, options...)
}

func (l *LinkedSpan) SetStatus(code codes.Code, description string) {
	l.childSpan.SetStatus(code, description)
	l.parentSpan.SetStatus(code, description)
}

func (l *LinkedSpan) SetAttributes(kv ...attribute.KeyValue) {
	l.childSpan.SetAttributes(kv...)
	l.parentSpan.SetAttributes(kv...)
}

func (l *LinkedSpan) SetName(name string) {
	l.childSpan.SetName(name)
	l.parentSpan.SetName(name)
}

func (l *LinkedSpan) ChildSpan() trace.Span {
	return l.childSpan
}

func (l *LinkedSpan) ChildContext() context.Context {
	return l.childCtx
}

func (l *LinkedSpan) ParentSpan() trace.Span {
	return l.parentSpan
}

func (l *LinkedSpan) ParentContext() context.Context {
	return l.parentCtx
}
