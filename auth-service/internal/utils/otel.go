package utils

import (
	"context"
	"errors"
	"fmt" // ⭐️ Added for resource setup error formatting
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"       // ⭐️ NEW: OTLP Log Exporter
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc" // ⭐️ NEW: OTLP Metric Exporter
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"   // ⭐️ NEW: OTLP Trace Exporter
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	// ⭐️ NEW: Dependencies for gRPC
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SetupOTelSDK bootstraps the OpenTelemetry pipeline and sends data via OTLP.
func SetupOTelSDK(ctx context.Context, serviceName, serviceVersion, otelExporterURL string) (func(context.Context) error, error) { // ⭐️ CHANGED: Added otelExporterURL
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// --- 1. Resource: Define service attributes ---
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			attribute.String("environment", "development"),
		),
	)
	if err != nil {
		handleErr(fmt.Errorf("failed to create resource: %w", err))
		return shutdown, err
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// --- GRPC Connection for OTLP Exporters (SigNoz typically uses 4317 for gRPC) ---
	// NOTE: Using insecure for simplicity, use TLS for production.
	conn, err := grpc.NewClient(otelExporterURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		handleErr(fmt.Errorf("failed to create gRPC connection to collector: %w", err))
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error {
		return conn.Close()
	})

	// Set up trace provider.
	tracerProvider, err := newTracerProvider(ctx, res, conn) // ⭐️ PASS CONN
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx, res, conn) // ⭐️ PASS CONN
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(ctx, res, conn) // ⭐️ PASS CONN
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return shutdown, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// ⭐️ UPDATED signature to accept context and gRPC connection
func newTracerProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (*trace.TracerProvider, error) {
	// ⭐️ OTLP Trace Exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()), // Use AlwaysSample for development
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
	)
	return tracerProvider, nil
}

// ⭐️ UPDATED signature to accept context and gRPC connection
func newMeterProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (*metric.MeterProvider, error) {
	// ⭐️ OTLP Metric Exporter
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
	)
	return meterProvider, nil
}

// ⭐️ UPDATED signature to accept context and gRPC connection
func newLoggerProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (*log.LoggerProvider, error) {
	// ⭐️ OTLP Log Exporter
	logExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}
