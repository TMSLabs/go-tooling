package httphelper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

func setupTestTracer() *trace.TracerProvider {
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)

	// Set up propagator for header injection
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tp
}

func TestHTTPDo(t *testing.T) {
	// Set up a test tracer provider
	tp := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tests := []struct {
		name         string
		method       string
		url          string
		spanName     string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful GET request",
			method:       "GET",
			url:          "/test",
			spanName:     "TestRequest",
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "successful POST request",
			method:       "POST",
			url:          "/test",
			spanName:     "PostRequest",
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name:         "server error response",
			method:       "GET",
			url:          "/error",
			spanName:     "ErrorRequest",
			serverStatus: http.StatusInternalServerError,
			wantErr:      false, // HTTP errors are not Go errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				// Note: Headers might be set depending on tracer configuration
				// We just verify the server receives the request correctly

				w.WriteHeader(tt.serverStatus)
				_, _ = w.Write([]byte("test response"))
			}))
			defer server.Close()

			// Create the request
			req, err := http.NewRequest(tt.method, server.URL+tt.url, strings.NewReader("test body"))
			require.NoError(t, err)

			// Create context with an active span so trace headers get injected
			ctx := context.Background()
			tracer := otel.Tracer("test")
			ctx, span := tracer.Start(ctx, "test-parent-span")
			defer span.End()

			client := &http.Client{}

			// Execute HTTPDo
			resp, err := HTTPDo(ctx, client, req, tt.spanName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.serverStatus, resp.StatusCode)
				resp.Body.Close()
			}
		})
	}
}

func TestHTTPDo_TraceHeaderInjection(t *testing.T) {
	// Set up a test tracer provider
	tp := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	// Track if trace headers were injected
	traceHeaderFound := false

	// Create a test server that checks for trace headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if trace headers are present
		if r.Header.Get("Traceparent") != "" {
			traceHeaderFound = true
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Create the request
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	require.NoError(t, err)

	// Create context with an active span
	ctx := context.Background()
	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(ctx, "parent-span")
	defer span.End()

	client := &http.Client{}

	// Execute HTTPDo
	resp, err := HTTPDo(ctx, client, req, "TraceHeaderTest")
	require.NoError(t, err)
	assert.NotNil(t, resp)
	defer resp.Body.Close()

	// Verify trace header was injected when there's an active span
	assert.True(t, traceHeaderFound, "Expected trace header to be injected when there's an active span")
}

func TestHTTPDo_NetworkError(t *testing.T) {
	// Set up a test tracer provider
	tp := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	// Test network error case using an invalid port instead of DNS lookup
	req, err := http.NewRequest("GET", "http://127.0.0.1:9999", nil)
	require.NoError(t, err)

	ctx := context.Background()
	client := &http.Client{}

	resp, err := HTTPDo(ctx, client, req, "NetworkErrorTest")

	// Network errors should return Go errors
	require.Error(t, err)
	assert.Nil(t, resp)
	// resp is nil so no need to close, but add check for linter
	if resp != nil {
		resp.Body.Close()
	}
}

func TestHTTPDo_ContextCancellation(t *testing.T) {
	// Set up a test tracer provider
	tp := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// This handler won't be reached due to context cancellation
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create the request
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &http.Client{}

	resp, err := HTTPDo(ctx, client, req, "CancelledRequest")

	// Should return context cancellation error
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "context canceled")
	// resp is nil so no need to close, but add check for linter
	if resp != nil {
		resp.Body.Close()
	}
}

func TestHTTPDo_NilParameters(t *testing.T) {
	// Set up a test tracer provider
	tp := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx := context.Background()

	// Test with nil client - should panic or fail
	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with nil client
			t.Log("Correctly panicked with nil client")
		}
	}()

	resp, err := HTTPDo(ctx, nil, req, "NilClientTest")
	// If we get here without panic, it should be an error
	if err == nil {
		t.Error("Expected error or panic with nil client")
	}
	// No response to close since client is nil, but add check for linter
	if resp != nil {
		resp.Body.Close()
	}
}
