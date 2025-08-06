// Package telemetry provides initialization for OpenTelemetry, Sentry, and slog logging.
package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/TMSLabs/go-tooling/k8s"
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"

	"log/slog"
)

// OtelHandler re-exports otelhttp.NewHandler for convenience
var OtelHandler = otelhttp.NewHandler

// initTelemetry initializes slog, OpenTelemetry, and Sentry.
// Returns the TracerProvider for shutdown.
func initTelemetry(serviceName string) (*sdktrace.TracerProvider, error) {

	// load .env vars if available
	if err := godotenv.Load(); err != nil {
		// .env is optional, continue without it
		slog.Warn("Failed to load .env file, continuing without it", "err", err)
	}

	// --- slog init ---

	logLevel := slog.LevelInfo
	if lvl, ok := os.LookupEnv("LOG_LEVEL"); ok {
		switch lvl {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "WARN":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		}
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)
	slog.Info("slog initialized", "level", logLevel)

	// --- Sentry init ---

	// check if Sentry DSN is set
	if os.Getenv("SENTRY_DSN") == "" {
		slog.Warn("SENTRY_DSN not set, Sentry will not be initialized")
		return nil, fmt.Errorf("SENTRY_DSN environment variable is required")
	}
	// check if Sentry environment is set
	sentryEnv := os.Getenv("SENTRY_ENVIRONMENT")
	if sentryEnv == "" {
		sentryEnv = k8s.GetEnvironment()
	}
	if sentryEnv == "" {
		slog.Warn("SENTRY_ENVIRONMENT not set, using default environment")
		sentryEnv = "development"
	}

	sentryDsn := os.Getenv("SENTRY_DSN")
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDsn,
		Environment:      sentryEnv,
		AttachStacktrace: true,
	}); err != nil {
		slog.Error("Sentry initialization failed", "err", err)
		return nil, err
	}
	slog.Info("Sentry initialized")

	// --- OpenTelemetry init ---

	// check if OTEL_EXPORTER_ENDPOINT is set
	if os.Getenv("OTEL_EXPORTER_ENDPOINT") == "" {
		slog.Warn("OTEL_EXPORTER_ENDPOINT not set, OpenTelemetry will not be initialized")
		return nil, fmt.Errorf("OTEL_EXPORTER_ENDPOINT environment variable is required")
	}

	ctx := context.Background()
	otelEndpoint := os.Getenv("OTEL_EXPORTER_ENDPOINT")
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otelEndpoint),
	)
	if err != nil {
		slog.Error("otel exporter init failed", "err", err)
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(serviceName)),
		),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
	slog.Info("OpenTelemetry initialized")
	return tp, nil
}

// ShutdownFunc is a function type for cleaning up telemetry resources
type ShutdownFunc func()

// Init initializes all telemetry and returns a shutdown function to defer in main.
func Init(serviceName string) (ShutdownFunc, error) {
	tp, err := initTelemetry(serviceName)
	if err != nil {
		return nil, err
	}
	return func() {
		sentry.Flush(2 * time.Second)
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("Error shutting down tracer provider", "err", err)
		}
	}, nil
}
