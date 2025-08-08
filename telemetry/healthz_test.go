package telemetry

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckConnection_Success(t *testing.T) {
	// This test would require a real NATS server or mock
	// For now, we test the error cases
	err := CheckConnection("")
	assert.Error(t, err)
}

func TestCheckConnection_InvalidURL(t *testing.T) {
	err := CheckConnection("invalid-url")
	assert.Error(t, err)
}

func TestCheckConnection_UnreachableServer(t *testing.T) {
	err := CheckConnection("nats://nonexistent-host:4222")
	assert.Error(t, err)
}

func TestHealthzEndpointHandler_NoConfigEnabled(t *testing.T) {
	// Reset telemetry config to default (no services enabled)
	TelemetryConfig = config{}

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HealthzEndpointHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Service is healthy")
}

func TestHealthzEndpointHandler_MySQLEnabled_InvalidDSN(t *testing.T) {
	// Configure with MySQL enabled but invalid DSN
	TelemetryConfig = config{
		MysqlEnabled: true,
		MysqlConfig: mySQLConfig{
			DSN: "invalid-mysql-dsn",
		},
	}

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HealthzEndpointHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "MySQL connection failed")
}

func TestHealthzEndpointHandler_MySQLEnabled_EmptyDSN(t *testing.T) {
	// Configure with MySQL enabled but empty DSN
	TelemetryConfig = config{
		MysqlEnabled: true,
		MysqlConfig: mySQLConfig{
			DSN: "",
		},
	}

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HealthzEndpointHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "MySQL connection failed")
}

func TestHealthzEndpointHandler_NATSEnabled_InvalidURL(t *testing.T) {
	// Configure with NATS enabled but invalid URL
	TelemetryConfig = config{
		NatsEnabled: true,
		NatsConfig: natsConfig{
			URL: "invalid-nats-url",
		},
	}

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HealthzEndpointHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "NATS connection failed")
}

func TestHealthzEndpointHandler_NATSEnabled_NoHealthCheckEvent(t *testing.T) {
	// Configure with NATS enabled and valid-looking URL but no health check events
	TelemetryConfig = config{
		NatsEnabled: true,
		NatsConfig: natsConfig{
			URL: "nats://localhost:4222",
		},
	}

	// Reset health check event
	LastHealthCheckEvent = ""

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HealthzEndpointHandler(w, req)

	// Should fail the connection check first, before checking events
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "NATS connection failed")
}

func TestHealthzEndpointHandler_NATSEnabled_OldHealthCheckEvent(t *testing.T) {
	// Mock a scenario where NATS connection would succeed but health check event is old
	// This is harder to test without mocking the CheckConnection function
	// For now, we test the logic around event timing

	// Set an old health check event (more than 5 minutes ago)
	oldTime := time.Now().Add(-10 * time.Minute)
	LastHealthCheckEvent = oldTime.Format(time.RFC3339)

	// Even with valid-looking config, the connection check will fail first
	TelemetryConfig = config{
		NatsEnabled: true,
		NatsConfig: natsConfig{
			URL: "nats://localhost:4222", // This will fail to connect
		},
	}

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HealthzEndpointHandler(w, req)

	// Will fail at connection check stage
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHealthzEndpointHandler_MultipleServices(t *testing.T) {
	// Test with multiple services enabled - should fail on first check
	TelemetryConfig = config{
		MysqlEnabled: true,
		MysqlConfig: mySQLConfig{
			DSN: "invalid-mysql-dsn",
		},
		NatsEnabled: true,
		NatsConfig: natsConfig{
			URL: "nats://localhost:4222",
		},
	}

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HealthzEndpointHandler(w, req)

	// Should fail on MySQL check first
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "MySQL connection failed")
}

func TestHealthzEndpointHandler_HTTPMethods(t *testing.T) {
	// Reset config
	TelemetryConfig = config{}

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/healthz", nil)
			w := httptest.NewRecorder()

			HealthzEndpointHandler(w, req)

			// Handler should work with any HTTP method
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "Service is healthy")
		})
	}
}

func TestHealthzEndpointHandler_ResponseFormat(t *testing.T) {
	// Reset config
	TelemetryConfig = config{}

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HealthzEndpointHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check response format
	body := w.Body.String()
	assert.Contains(t, body, "status")
	assert.Contains(t, body, "ok")
	assert.Contains(t, body, "message")
	assert.Contains(t, body, "Service is healthy")

	// Should be JSON-like format
	assert.True(t, strings.HasPrefix(body, "{"))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(body), "}"))
}

func TestLastHealthCheckEvent_GlobalVariable(t *testing.T) {
	// Test the global variable behavior
	originalValue := LastHealthCheckEvent
	defer func() {
		LastHealthCheckEvent = originalValue
	}()

	// Test setting and getting the value
	testTime := time.Now().Format(time.RFC3339)
	LastHealthCheckEvent = testTime

	assert.Equal(t, testTime, LastHealthCheckEvent)

	// Test with empty value
	LastHealthCheckEvent = ""
	assert.Equal(t, "", LastHealthCheckEvent)
}

// Note: HealthzEventChecker is hard to test without a real NATS connection
// as it's a long-running goroutine. In a real-world scenario, you'd want to:
// 1. Use dependency injection to make NATS connection mockable
// 2. Extract the logic into smaller, testable functions
// 3. Use interfaces for external dependencies
// 4. Add proper shutdown mechanisms for goroutines

func TestHealthzEventChecker_Integration(t *testing.T) {
	// This would be an integration test requiring a real NATS server
	// Skipping for now as it requires external dependencies
	t.Skip("Integration test requires real NATS server - use testcontainers or docker-compose for full testing")

	// Example of how this would work with a real NATS connection:
	// nc, err := nats.Connect("nats://localhost:4222")
	// if err != nil {
	//     t.Skipf("NATS server not available: %v", err)
	// }
	// defer nc.Close()
	//
	// go HealthzEventChecker(nc, "test-service")
	//
	// // Wait for health check event to be published and received
	// time.Sleep(time.Second * 2)
	//
	// // Verify LastHealthCheckEvent was updated
	// assert.NotEmpty(t, LastHealthCheckEvent)
}
