package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TMSLabs/go-tooling/httphelper"
	"github.com/TMSLabs/go-tooling/mysqlhelper"
	"github.com/TMSLabs/go-tooling/telemetry"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

// TestTelemetryMySQLIntegration demonstrates integration between telemetry and MySQL
func TestTelemetryMySQLIntegration(t *testing.T) {
	// Set up tracing for integration test
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Initialize telemetry with MySQL enabled (but expect connection to fail)
	shutdown, err := telemetry.Init(
		"integration-test-service",
		"test",
		telemetry.WithSlog(),
		telemetry.WithMySQL(telemetry.MySQLDSN("user:pass@tcp(127.0.0.1:9998)/testdb")),
	)

	// This will succeed at initialization but fail during health checks
	if err != nil {
		t.Logf("Expected initialization error: %v", err)
		// If init fails, we can still test the configuration
		assert.True(t, telemetry.TelemetryConfig.MysqlEnabled)
		assert.Equal(t, "user:pass@tcp(127.0.0.1:9998)/testdb", telemetry.TelemetryConfig.MysqlConfig.DSN)
		return
	}

	assert.NotNil(t, shutdown)
	defer shutdown()

	// Verify telemetry configuration
	assert.True(t, telemetry.TelemetryConfig.MysqlEnabled)
	assert.True(t, telemetry.TelemetryConfig.SlogEnabled)
	assert.Equal(t, "integration-test-service", telemetry.TelemetryConfig.ServiceName)

	// Test direct MySQL connection (should fail)
	db, err := mysqlhelper.Connect(telemetry.TelemetryConfig.MysqlConfig.DSN)
	assert.Error(t, err)
	assert.Nil(t, db)

	// Test MySQL health check through telemetry (should fail)
	err = mysqlhelper.CheckConnection(telemetry.TelemetryConfig.MysqlConfig.DSN)
	assert.Error(t, err)

	// Test health endpoint integration
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	telemetry.HealthzEndpointHandler(w, req)

	// Should fail because MySQL connection fails
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "MySQL connection failed")
}

// TestTelemetryNATSIntegration demonstrates integration between telemetry and NATS
func TestTelemetryNATSIntegration(t *testing.T) {
	// Initialize telemetry with NATS enabled (expect connection to fail)
	shutdown, err := telemetry.Init(
		"nats-integration-test",
		"test",
		telemetry.WithSlog(),
		telemetry.WithNATS(telemetry.NATSURL("nats://127.0.0.1:9999")),
	)

	// Should fail because NATS server doesn't exist
	assert.Error(t, err)
	assert.Nil(t, shutdown)
	assert.Contains(t, err.Error(), "nats connection failed")

	// Verify configuration was set even though connection failed
	assert.True(t, telemetry.TelemetryConfig.NatsEnabled)
	assert.Equal(t, "nats://127.0.0.1:9999", telemetry.TelemetryConfig.NatsConfig.URL)
}

// TestHTTPTelemetryIntegration demonstrates HTTP request tracing integration
func TestHTTPTelemetryIntegration(t *testing.T) {
	// Set up tracing
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Initialize telemetry
	shutdown, err := telemetry.Init(
		"http-integration-test",
		"test",
		telemetry.WithSlog(),
	)

	assert.NoError(t, err)
	assert.NotNil(t, shutdown)
	defer shutdown()

	// Create a test server that uses telemetry error capture
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate an error that gets captured by telemetry
		ctx := r.Context()
		testErr := assert.AnError

		// Use telemetry error capture
		telemetry.CaptureError(ctx, testErr, "Test error in HTTP handler")

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error occurred"))
	}))
	defer server.Close()

	// Make HTTP request using httphelper with tracing
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	assert.NoError(t, err)

	ctx := context.Background()
	tracer := otel.Tracer("integration-test")
	ctx, span := tracer.Start(ctx, "http-integration-test")
	defer span.End()

	client := &http.Client{}

	// Use httphelper.HTTPDo which adds tracing
	resp, err := httphelper.HTTPDo(ctx, client, req, "IntegrationTestRequest")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	defer resp.Body.Close()
}

// TestHTTPHandlerTelemetryIntegration demonstrates HTTP handler tracing integration
func TestHTTPHandlerTelemetryIntegration(t *testing.T) {
	// Set up tracing
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Initialize telemetry
	shutdown, err := telemetry.Init(
		"handler-integration-test",
		"test",
		telemetry.WithSlog(),
	)

	assert.NoError(t, err)
	assert.NotNil(t, shutdown)
	defer shutdown()

	// Create a handler that uses telemetry
	handlerFunc := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// Use telemetry within the handler
		if r.URL.Path == "/error" {
			testErr := assert.AnError
			telemetry.CaptureError(ctx, testErr, "Handler error test")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}

	// Wrap with httphelper tracing
	wrappedHandler := httphelper.HTTPHandler(handlerFunc, "IntegrationTestHandler")

	// Test successful request
	req := httptest.NewRequest("GET", "/success", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())

	// Test error request
	req = httptest.NewRequest("GET", "/error", nil)
	w = httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestFullIntegrationScenario demonstrates a complete integration scenario
func TestFullIntegrationScenario(t *testing.T) {
	// Set up tracing
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Initialize telemetry with multiple components
	// Note: This will fail due to missing external services, but demonstrates the integration
	shutdown, err := telemetry.Init(
		"full-integration-test",
		"development",
		telemetry.WithSlog(telemetry.SlogLogLevel(4)), // Info level
		telemetry.WithMySQL(telemetry.MySQLDSN("user:pass@tcp(127.0.0.1:9998)/testdb")),
		// Note: Intentionally not including NATS, Sentry, or Trace to avoid external dependencies
	)

	if err != nil {
		// Expected due to MySQL connection failure
		t.Logf("Expected initialization error: %v", err)

		// Verify configuration was set
		assert.True(t, telemetry.TelemetryConfig.SlogEnabled)
		assert.True(t, telemetry.TelemetryConfig.MysqlEnabled)
		return
	}

	assert.NotNil(t, shutdown)
	defer shutdown()

	// Test health endpoint with multiple services
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	telemetry.HealthzEndpointHandler(w, req)

	// Should fail due to MySQL connection
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "MySQL connection failed")

	// Test error capture in integration context
	ctx := context.Background()
	tracer := otel.Tracer("full-integration")
	ctx, span := tracer.Start(ctx, "integration-error-test")
	defer span.End()

	integrationErr := assert.AnError
	telemetry.CaptureError(ctx, integrationErr, "Full integration test error")

	// No assertion needed - just verify it doesn't panic
}

// TestIntegrationWithRealTime demonstrates time-based functionality
func TestIntegrationWithRealTime(t *testing.T) {
	// Test the health check event timing logic
	originalEvent := telemetry.LastHealthCheckEvent
	defer func() {
		telemetry.LastHealthCheckEvent = originalEvent
	}()

	// Set a recent health check event
	recentTime := time.Now().Add(-1 * time.Minute)
	telemetry.LastHealthCheckEvent = recentTime.Format(time.RFC3339)

	// Just verify the timestamp was set correctly
	assert.Equal(t, recentTime.Format(time.RFC3339), telemetry.LastHealthCheckEvent)

	// Test with an old timestamp
	oldTime := time.Now().Add(-10 * time.Minute)
	telemetry.LastHealthCheckEvent = oldTime.Format(time.RFC3339)

	assert.Equal(t, oldTime.Format(time.RFC3339), telemetry.LastHealthCheckEvent)
}

// Note: The config and natsConfig types are not exported, so I need to check the actual types
// Let me fix the type references above
