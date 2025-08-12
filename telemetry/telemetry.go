// Package telemetry provides initialization for OpenTelemetry, Sentry, and slog logging.
package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryotel "github.com/getsentry/sentry-go/otel"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"

	"log/slog"
)

// TelemetryConfig is the configuration for telemetry.
var TelemetryConfig = config{}

// --- slog helpers ---
type otelHandler struct {
	slog.Handler
}

func newOTelHandler(base slog.Handler) *otelHandler {
	return &otelHandler{Handler: base}
}

func (h *otelHandler) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		r.AddAttrs(
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

// --- end ---

// initTelemetry initializes slog, OpenTelemetry, and Sentry.
// Returns the TracerProvider for shutdown.
func initTelemetry(
	serviceName string,
	environment string,
	opts ...Option,
) (*sdktrace.TracerProvider, error) {

	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.ServiceName = serviceName
	cfg.Environment = environment
	TelemetryConfig = *cfg

	// --- NATS init ---
	if cfg.NatsEnabled {
		if cfg.NatsConfig.URL == "" {
			slog.Error("NATS URL is required but not set")
			return nil, fmt.Errorf("nats URL is required but not set")
		}
		nc, err := nats.Connect(cfg.NatsConfig.URL, nats.Name(serviceName))
		if err != nil {
			slog.Error("NATS connection failed", "err", err)
			return nil, fmt.Errorf("nats connection failed: %w", err)
		}
		slog.Info("NATS initialized", "url", cfg.NatsConfig.URL)

		// Subscribe to health check Environment
		go HealthzEventChecker(nc, serviceName)
	}

	// --- slog init ---
	if cfg.SlogEnabled {
		logLevel := slog.LevelInfo
		if cfg.SlogConfig.logLevel != slog.LevelInfo {
			logLevel = cfg.SlogConfig.logLevel
		}
		baseHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
		otelHandler := newOTelHandler(baseHandler)
		logger := slog.New(otelHandler)
		slog.SetDefault(logger)
		slog.Info("slog initialized", "level", logLevel)
	}

	// --- Sentry init ---
	if cfg.SentryEnabled {
		if cfg.SentryConfig.DSN == "" {
			slog.Error("Sentry DSN is required but not set")
			return nil, fmt.Errorf("sentry DSN is required but not set")
		}

		sentryConfig := sentry.ClientOptions{
			AttachStacktrace: true,
			SendDefaultPII:   true,
			EnableTracing:    true,
			TracesSampler: sentry.TracesSampler(func(ctx sentry.SamplingContext) float64 {
				if ctx.Span != nil && ctx.Span.Status == sentry.SpanStatusInternalError {
					return 1.0 // Send trace for errors
				}
				return 0.0 // Don't send trace for non-error spans
			}),
		}

		sentryConfig.Environment = environment
		if cfg.SentryConfig.Environment != "" {
			sentryConfig.Environment = cfg.SentryConfig.Environment
		}
		if cfg.SentryConfig.Release != "" {
			sentryConfig.Release = cfg.SentryConfig.Release
		}
		if cfg.SentryConfig.DSN != "" {
			sentryConfig.Dsn = cfg.SentryConfig.DSN
		}

		if err := sentry.Init(sentryConfig); err != nil {
			slog.Error("Sentry initialization failed", "err", err)
			return nil, err
		}

		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSpanProcessor(sentryotel.NewSentrySpanProcessor()),
		)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(sentryotel.NewSentryPropagator())

		slog.Info("Sentry initialized")
	}

	// --- OpenTelemetry init ---
	if cfg.TraceEnabled {
		// check if OTEL_EXPORTER_ENDPOINT is set
		if cfg.TraceConfig.ExporterURL == "" {
			slog.Error("OpenTelemetry Exporter URL is required but not set")
			return nil, fmt.Errorf("OpenTelemetry Exporter URL is required but not set")
		}

		ctx := context.Background()
		exporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(cfg.TraceConfig.ExporterURL),
		)
		if err != nil {
			slog.Error("otel exporter init failed", "err", err)
			return nil, err
		}

		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(
				resource.NewWithAttributes(
					semconv.SchemaURL,
					semconv.ServiceName(serviceName),
					semconv.DeploymentEnvironment(cfg.Environment),
					semconv.ServiceVersion(cfg.SentryConfig.Release),
				),
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

	return nil, nil
}

// ShutdownFunc is a function type for cleaning up telemetry resources
type ShutdownFunc func()

// Init initializes all telemetry and returns a shutdown function to defer in main.
func Init(serviceName string, environment string, opts ...Option) (ShutdownFunc, error) {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	tp, err := initTelemetry(serviceName, environment, opts...)
	if err != nil {
		return nil, err
	}
	if cfg.TraceEnabled && cfg.SentryEnabled {
		return func() {
			sentry.Flush(2 * time.Second)
			if err := tp.Shutdown(context.Background()); err != nil {
				slog.Error("Error shutting down tracer provider", "err", err)
			}
		}, nil
	} else if cfg.SentryEnabled {
		return func() {
			sentry.Flush(2 * time.Second)
		}, nil
	}
	return func() {}, nil
}
