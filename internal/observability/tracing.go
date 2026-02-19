package observability

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer(ctx context.Context, serviceName string) (*sdktrace.TracerProvider, error) {
	// 1. exporter - where traces go

	// 2. resource - info about the service
	// 3. tracer provider - ties exporter and resource
	// 4. set as global provider
	// 5. return to main.go to shutdown on exit

	return nil, nil
}

func ShutdownTracer() {
	// flush any remaining spans
}
