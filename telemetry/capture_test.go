package telemetry

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

func TestCaptureError_NilError(_ *testing.T) {
	// Reset config to ensure clean state
	TelemetryConfig = config{}

	ctx := context.Background()

	// Should not panic with nil error
	CaptureError(ctx, nil, "test message")

	// Test passes if no panic occurs
}

func TestCaptureError_WithError_NoTelemetryEnabled(_ *testing.T) {
	// Reset config to disable all telemetry
	TelemetryConfig = config{}

	ctx := context.Background()
	testErr := errors.New("test error")

	// Should not panic even without telemetry enabled
	CaptureError(ctx, testErr, "test error occurred")

	// Test passes if no panic occurs
}

func TestCaptureError_WithSentryEnabled(_ *testing.T) {
	// Configure with Sentry enabled (but not actually initialized to avoid network calls)
	TelemetryConfig = config{
		SentryEnabled: true,
	}

	ctx := context.Background()
	testErr := errors.New("test sentry error")

	// Should execute Sentry-related code paths without panicking
	CaptureError(ctx, testErr, "sentry error occurred")

	// Test passes if no panic occurs
}

func TestCaptureError_WithTraceEnabled(_ *testing.T) {
	// Set up OpenTelemetry tracer for testing
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	// Configure with tracing enabled
	TelemetryConfig = config{
		TraceEnabled: true,
	}

	// Create a context with an active span
	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	testErr := errors.New("test trace error")

	// Should execute OpenTelemetry-related code paths
	CaptureError(ctx, testErr, "trace error occurred")

	// Verify the span status was set to error
	// Note: In a real test, you'd want to export spans to verify they were recorded correctly
}

func TestCaptureError_WithBothSentryAndTraceEnabled(_ *testing.T) {
	// Set up OpenTelemetry tracer
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	// Configure with both Sentry and tracing enabled
	TelemetryConfig = config{
		SentryEnabled: true,
		TraceEnabled:  true,
	}

	// Create a context with an active span
	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span-both")
	defer span.End()

	testErr := errors.New("test error for both systems")

	// Should execute both Sentry and OpenTelemetry code paths
	CaptureError(ctx, testErr, "error for both systems")

	// Test passes if no panic occurs
}

func TestCaptureError_DifferentErrorTypes(t *testing.T) {
	// Configure with tracing for better testing
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	TelemetryConfig = config{
		TraceEnabled: true,
	}

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "error-types-test")
	defer span.End()

	testCases := []struct {
		name    string
		err     error
		message string
	}{
		{
			name:    "simple error",
			err:     errors.New("simple error"),
			message: "A simple error occurred",
		},
		{
			name:    "wrapped error",
			err:     errors.New("wrapped: original error"),
			message: "A wrapped error occurred",
		},
		{
			name:    "empty message",
			err:     errors.New("error with empty message"),
			message: "",
		},
		{
			name:    "long message",
			err:     errors.New("error"),
			message: "This is a very long error message that contains a lot of details about what went wrong in the system and should be handled properly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			// Should handle different error types and messages without issues
			CaptureError(ctx, tc.err, tc.message)
			// Test passes if no panic occurs
		})
	}
}

func TestCaptureError_ContextVariations(t *testing.T) {
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	TelemetryConfig = config{
		TraceEnabled: true,
	}

	testErr := errors.New("context test error")

	type contextKey string
	testCases := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "background context",
			ctx:  context.Background(),
		},
		{
			name: "context with value",
			ctx:  context.WithValue(context.Background(), contextKey("key"), "value"),
		},
		{
			name: "cancelled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			// Should handle different context types
			CaptureError(tc.ctx, testErr, "context variation test")
			// Test passes if no panic occurs
		})
	}
}

func TestCaptureError_ConcurrentCalls(_ *testing.T) {
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	TelemetryConfig = config{
		TraceEnabled:  true,
		SentryEnabled: true,
	}

	// Test concurrent calls to CaptureError
	const numGoroutines = 10
	const numCallsPerGoroutine = 5

	results := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(_ int) {
			tracer := otel.Tracer("concurrent-test")

			for j := 0; j < numCallsPerGoroutine; j++ {
				ctx, span := tracer.Start(context.Background(), "concurrent-error")
				testErr := errors.New("concurrent error")

				// Should not cause race conditions
				CaptureError(ctx, testErr, "concurrent error test")

				span.End()
			}

			results <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-results
	}

	// Test passes if no panics occurred
}

func TestCaptureError_MessageFormatting(t *testing.T) {
	TelemetryConfig = config{} // No telemetry enabled for simplicity

	ctx := context.Background()
	testErr := errors.New("formatting test error")

	// Test various message formats
	messages := []string{
		"Simple message",
		"Message with %s formatting", // This shouldn't be treated as format string
		"Message with unicode: 🚨 ❌ ⚠️",
		"Message\nwith\nnewlines",
		"Message with \"quotes\" and 'apostrophes'",
		"Message with special chars: !@#$%^&*()_+{}[]|\\:;\"'<>?,./~`",
	}

	for _, msg := range messages {
		t.Run("msg_"+msg[:minInt(20, len(msg))], func(_ *testing.T) {
			// Should handle various message formats without issues
			CaptureError(ctx, testErr, msg)
			// Test passes if no panic occurs
		})
	}
}

func TestCaptureError_ErrorMessageExtraction(t *testing.T) {
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	TelemetryConfig = config{
		TraceEnabled: true,
	}

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "error-extraction-test")
	defer span.End()

	// Test different error message lengths and formats
	errors := []error{
		errors.New(""),      // empty error message
		errors.New("short"), // short message
		errors.New("medium length error message with some details"), // medium message
		errors.New("very long error message that exceeds typical length limits and contains extensive details about the failure condition that occurred in the system during processing"), // long message
	}

	for i, err := range errors {
		t.Run("error_"+string(rune(i+'0')), func(_ *testing.T) {
			CaptureError(ctx, err, "test message")
			// Test passes if no panic occurs
		})
	}
}

// Helper function for minInt (avoiding conflict with builtin min)
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
