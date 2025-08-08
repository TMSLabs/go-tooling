package telemetry

import (
	"context"
	"log/slog"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// CaptureError captures an error in the context of telemetry.
// It sends the error to Sentry if enabled, and records it in the current OpenTelemetry span.
// If the error is nil, it does nothing.
// This function is useful for logging errors in a consistent way across your application.
// It is recommended to use this function in conjunction with OpenTelemetry for distributed tracing.
// Example usage:
//
//	ctx := context.Background()
//	err := someFunctionThatMightFail()
//
//	if err != nil {
//	    telemetry.CaptureError(ctx, err, "An error occurred in someFunctionThatMightFail")
//	}
func CaptureError(ctx context.Context, err error, message string) {
	if err == nil {
		return
	}

	// Capture the error using Sentry
	if TelemetryConfig.SentryEnabled {
		slog.Error("Sentry error capture", "error", err, "message", message)
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "error",
			Message:  message,
			Data: map[string]any{
				"error":   err.Error(),
				"message": message,
			},
			Level: sentry.LevelError,
		})

		sentry.CaptureException(err)

		sentrySpan := sentry.SpanFromContext(ctx)
		if sentrySpan != nil {
			sentrySpan.Status = sentry.SpanStatusInternalError
		}
	}

	// If OpenTelemetry is enabled, record the error in the current span
	if TelemetryConfig.TraceEnabled {
		slog.Error("OpenTelemetry error capture", "error", err, "message", message)
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(
			attribute.String("error.message", err.Error()),
			attribute.String("error.description", message),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	slog.Error("Error captured", "error", err, "message", message)

}
