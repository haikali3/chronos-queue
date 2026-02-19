package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func InitTracer(ctx context.Context, serviceName string) (*sdktrace.TracerProvider, error) {
	// 1. exporter - where traces go
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	// 2. resource - info about the service
	res := resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(serviceName))

	// 3. tracer provider - ties exporter and resource
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// 4. set as global provider
	otel.SetTracerProvider(tp)

	// 5. return to main.go to shutdown on exit
	return tp, nil
}

func ShutdownTracer(tp *sdktrace.TracerProvider) error {
	return tp.Shutdown(context.Background())
}
