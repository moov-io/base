package telemetry_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/moov-io/base/telemetry"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
)

func TestStartSpan__NoPanic(t *testing.T) {
	ctx, span := telemetry.StartSpan(nil, "no-panics") //nolint:staticcheck
	require.NotNil(t, ctx)
	require.NotNil(t, span)
}

func TestSpan_SetAttributes(t *testing.T) {
	var conf telemetry.Config
	shutdown, err := telemetry.SetupTelemetry(context.Background(), conf, "v0.0.0")
	require.NoError(t, err)
	t.Cleanup(func() { shutdown() })

	ctx, span := telemetry.StartSpan(context.Background(), "set-attributes")
	defer span.End()

	// First Set
	span.SetAttributes(attribute.String("kafka.topic", "test.cmd.v1"))

	// Second Set
	span.SetAttributes(attribute.String("event.type", "my-favorite-event"))

	// Verify the attributes which are set
	ss := telemetry.SpanFromContext(ctx)
	require.Equal(t, "*trace.recordingSpan", fmt.Sprintf("%T", ss))
	ro, ok := ss.(trace.ReadOnlySpan)
	require.True(t, ok)

	attrs := ro.Attributes()
	for i := range attrs {
		switch attrs[i].Key {
		case "kafka.topic", "event.type":
			// do nothing
		default:
			t.Errorf("attribute[%d]=%#v\n", i, attrs[i])
		}
	}
}
