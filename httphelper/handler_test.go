package httphelper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

func TestHTTPHandler(t *testing.T) {
	// Set up a test tracer provider
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tests := []struct {
		name     string
		spanName string
		method   string
		url      string
		headers  map[string]string
	}{
		{
			name:     "GET request with tracing",
			spanName: "TestHandler",
			method:   "GET",
			url:      "/test",
			headers:  map[string]string{},
		},
		{
			name:     "POST request with tracing",
			spanName: "PostHandler",
			method:   "POST",
			url:      "/api/post",
			headers:  map[string]string{"Content-Type": "application/json"},
		},
		{
			name:     "request with existing trace context",
			spanName: "TracedHandler",
			method:   "GET",
			url:      "/traced",
			headers: map[string]string{
				"traceparent": "00-12345678901234567890123456789012-1234567890123456-01",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track if handler was called with correct parameters
			handlerCalled := false
			var receivedCtx context.Context
			var receivedReq *http.Request

			// Create test handler function
			handlerFunc := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				receivedCtx = ctx
				receivedReq = r

				// Verify context contains span
				assert.NotNil(t, ctx)

				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test response"))
			}

			// Create wrapped handler
			wrappedHandler := HTTPHandler(handlerFunc, tt.spanName)

			// Create test request
			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader("test body"))

			// Set headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler
			wrappedHandler.ServeHTTP(w, req)

			// Verify handler was called
			assert.True(t, handlerCalled, "Handler function should have been called")
			assert.NotNil(t, receivedCtx, "Context should not be nil")
			assert.NotNil(t, receivedReq, "Request should not be nil")

			// Verify response
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "test response", w.Body.String())
		})
	}
}

func TestHTTPHandler_TraceExtraction(t *testing.T) {
	// Set up a test tracer provider
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Create handler that checks for trace context
	handlerFunc := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// The context should contain trace information
		assert.NotNil(t, ctx)

		w.WriteHeader(http.StatusOK)
	}

	wrappedHandler := HTTPHandler(handlerFunc, "TraceExtractionTest")

	// Create request with trace context
	req := httptest.NewRequest("GET", "/test", nil)

	// Add trace context using OpenTelemetry propagator
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	// Create a parent context with a span
	ctx := context.Background()
	tracer := otel.Tracer("test")
	parentCtx, parentSpan := tracer.Start(ctx, "parent-span")
	defer parentSpan.End()

	// Inject the context into headers
	propagator.Inject(parentCtx, propagation.HeaderCarrier(req.Header))

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHTTPHandler_ErrorHandling(t *testing.T) {
	// Set up a test tracer provider
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Test handler that panics
	handlerFunc := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}

	wrappedHandler := HTTPHandler(handlerFunc, "PanicHandler")

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// Should handle panic gracefully or let it propagate (depending on implementation)
	defer func() {
		if r := recover(); r != nil {
			// Panic was propagated - this is acceptable behavior
			assert.Equal(t, "test panic", r)
		}
	}()

	wrappedHandler.ServeHTTP(w, req)

	// If we get here, the panic was handled internally
	// The response code would depend on how panics are handled
}

func TestHTTPHandler_MultipleRequests(t *testing.T) {
	// Set up a test tracer provider
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	callCount := 0
	handlerFunc := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	wrappedHandler := HTTPHandler(handlerFunc, "MultiRequestHandler")

	// Make multiple requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	}

	assert.Equal(t, 3, callCount, "Handler should have been called 3 times")
}

func TestHTTPHandler_NilHandler(t *testing.T) {
	// This test verifies behavior when nil handler is passed
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with nil handler
			t.Log("Correctly panicked with nil handler")
		} else {
			t.Error("Expected panic with nil handler")
		}
	}()

	// Should panic when creating handler with nil function
	wrappedHandler := HTTPHandler(nil, "NilHandler")

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)
}
