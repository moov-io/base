package telemetry_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/moov-io/base/telemetry"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func Test_Setup_Honey(t *testing.T) {
	shutdown, err := telemetry.SetupTelemetry(context.Background(), telemetry.Config{
		ServiceName: "test",
		Honeycomb: &telemetry.HoneycombConfig{
			URL:  "api.honeycomb.io:443",
			Team: "HoneycombAPIKey",
		},
	}, "v0.0.1")

	require.NoError(t, err)

	err = shutdown()
	require.NoError(t, err)
}

func Test_Setup_Otel(t *testing.T) {
	shutdown, err := telemetry.SetupTelemetry(context.Background(), telemetry.Config{
		ServiceName: "test",
		OpenTelemetryCollector: &telemetry.OtelConfig{
			Host: "collector",
		},
	}, "v0.0.1")

	require.NoError(t, err)

	err = shutdown()
	require.NoError(t, err)
}

func Test_Setup_Stdout(t *testing.T) {
	shutdown, err := telemetry.SetupTelemetry(context.Background(), telemetry.Config{
		ServiceName: "test",
		Stdout:      true,
	}, "v0.0.1")

	require.NoError(t, err)

	_, spn := telemetry.StartSpan(context.Background(), "test")
	spn.AddEvent("added an event!")
	spn.End()

	err = shutdown()
	require.NoError(t, err)
}

func Test_Keeping_Consumers(t *testing.T) {
	buf := setupTelemetry(t)

	_, spn := telemetry.StartSpan(context.Background(), "consuming", trace.WithSpanKind(trace.SpanKindConsumer))
	time.Sleep(5 * time.Millisecond)
	spn.End()

	require.Contains(t, buf.String(), `"Name":"consuming"`)
}

func Test_Dropping_Empty_Consumers(t *testing.T) {
	buf := setupTelemetry(t)

	_, spn := telemetry.StartSpan(context.Background(), "consuming", trace.WithSpanKind(trace.SpanKindConsumer))
	// instantaneously returns
	spn.End()

	require.NotContains(t, buf.String(), "consuming")
}

func setupTelemetry(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	config := telemetry.TestConfig(buf)

	shutdown, err := telemetry.SetupTelemetry(context.Background(), config, "v0.0.1")
	t.Cleanup(func() {
		shutdown()
	})

	require.NoError(t, err)

	return buf
}
