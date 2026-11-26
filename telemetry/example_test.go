package telemetry_test

import (
	"context"
	"fmt"
	"io"

	"github.com/moov-io/base/telemetry"
	"go.opentelemetry.io/otel/attribute"
)

func ExampleSetupTelemetry() {
	ctx := context.Background()
	shutdown, err := telemetry.SetupTelemetry(ctx, telemetry.TestConfig(io.Discard), "v1.0.0")
	if err != nil {
		panic(err)
	}
	defer shutdown()

	ctx, span := telemetry.StartSpan(ctx, "example-operation")
	defer span.End()

	telemetry.SetAttributes(ctx, attribute.String("account_id", "12345"))
	fmt.Println(span.IsRecording())
	// Output:
	// true
}
