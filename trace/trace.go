// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package trace

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.11.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

const (
	LogFieldTraceID = "trace_id"
)

var (
	traceProviderName string
	serviceName       string
)

func Initialize(traceProvider, service string) {
	traceProviderName = traceProvider
	serviceName = service
}

// SetUpTracer sets up a GRPC reciever for serviceName with url as the endpoint of the collector.
// If a connection is not establised within connectTimeout, it is aborted and returns an error
func SetUpTracer(ctx context.Context, url string, connectTimeout time.Duration) (func(), error) {
	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(url),
		otlptracegrpc.WithDialOption(grpc.WithBlock()))

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to set up exporter: %w", err)
	}

	tp, err := setupTraceproviderWithExporter(serviceName, exporter)
	if err != nil {
		return nil, err
	}

	return func() {
		if err := tp.ForceFlush(ctx); err != nil {
			log.Print(err)
		}
		if err := tp.Shutdown(ctx); err != nil {
			log.Print(err)
		}
	}, nil
}

func setupTraceproviderWithExporter(serviceName string, exporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	resc, err := resource.New(
		context.Background(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
		resource.WithFromEnv(),
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	resc, err = resource.Merge(
		resource.Default(),
		resc,
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(resc))
	otel.SetTracerProvider(tp)
	propagationB3 := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}, propagationB3))
	return tp, nil
}

func NewRootSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(traceProviderName).Start(ctx, name, append(opts, trace.WithNewRoot())...)
	ctx = LoggerAddField(ctx, LogFieldTraceID, TraceIDFromContext(ctx))

	return ctx, span
}

func NewChildSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(traceProviderName).Start(ctx, name, opts...)
}

func NewAutoNamedChildSpan(ctx context.Context, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(traceProviderName).Start(ctx, getCallingFuncName(), opts...)
}

func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func TraceIDFromContext(ctx context.Context) string {
	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	if traceID == "" {
		log.Println("trace_id not found")
	}
	return traceID
}
