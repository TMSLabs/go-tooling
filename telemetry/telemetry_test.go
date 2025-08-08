package telemetry

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_MinimalConfiguration(t *testing.T) {
	// Test basic initialization with no optional components
	shutdown, err := Init("test-service", "development")

	require.NoError(t, err)
	assert.NotNil(t, shutdown)

	// Call shutdown function - should not panic
	shutdown()
}

func TestInit_WithSlog(t *testing.T) {
	shutdown, err := Init(
		"test-service",
		"test",
		WithSlog(SlogLogLevel(slog.LevelDebug)),
	)

	require.NoError(t, err)
	assert.NotNil(t, shutdown)

	// Verify slog is configured
	assert.True(t, TelemetryConfig.SlogEnabled)
	assert.Equal(t, slog.LevelDebug, TelemetryConfig.SlogConfig.logLevel)

	shutdown()
}

func TestInit_WithSentry_MissingDSN(t *testing.T) {
	shutdown, err := Init(
		"test-service",
		"test",
		WithSentry(), // No DSN provided
	)

	// Should fail because DSN is required
	require.Error(t, err)
	assert.Nil(t, shutdown)
	assert.Contains(t, err.Error(), "sentry DSN is required")
}

func TestInit_WithSentry_InvalidDSN(t *testing.T) {
	shutdown, err := Init(
		"test-service",
		"test",
		WithSentry(SentryDSN("invalid-dsn")),
	)

	// Should fail because DSN is invalid
	require.Error(t, err)
	assert.Nil(t, shutdown)
}

func TestInit_WithSentry_ValidDSN(t *testing.T) {
	// Use a test DSN format (not a real endpoint)
	testDSN := "https://test@o123456.ingest.us.sentry.io/123456"

	shutdown, err := Init(
		"test-service",
		"test",
		WithSentry(
			SentryDSN(testDSN),
			SentryEnvironment("test"),
			SentryRelease("v1.0.0"),
		),
	)

	// This might fail or succeed depending on network, but we can check config was set
	if err == nil {
		assert.NotNil(t, shutdown)
		shutdown()
	}

	// Verify configuration was set
	assert.True(t, TelemetryConfig.SentryEnabled)
	assert.Equal(t, testDSN, TelemetryConfig.SentryConfig.DSN)
	assert.Equal(t, "test", TelemetryConfig.SentryConfig.Environment)
	assert.Equal(t, "v1.0.0", TelemetryConfig.SentryConfig.Release)
}

func TestInit_WithTrace_MissingExporterURL(t *testing.T) {
	shutdown, err := Init(
		"test-service",
		"test",
		WithTrace(), // No exporter URL provided
	)

	// Should fail because exporter URL is required
	require.Error(t, err)
	assert.Nil(t, shutdown)
	assert.Contains(t, err.Error(), "OpenTelemetry Exporter URL is required")
}

func TestInit_WithTrace_InvalidExporterURL(t *testing.T) {
	shutdown, err := Init(
		"test-service",
		"test",
		WithTrace(TraceExporterURL("http://127.0.0.1:9999")),
	)

	// The behavior depends on the OpenTelemetry implementation
	// Some invalid URLs might still work during initialization but fail during export
	// Let's just verify that the configuration was set correctly
	if err == nil {
		assert.NotNil(t, shutdown)
		shutdown()
		// Configuration should still be set even if connection works
		assert.True(t, TelemetryConfig.TraceEnabled)
		assert.Equal(t, "http://127.0.0.1:9999", TelemetryConfig.TraceConfig.ExporterURL)
	} else {
		assert.Nil(t, shutdown)
	}
}

func TestInit_WithMySQL_MissingDSN(t *testing.T) {
	shutdown, err := Init(
		"test-service",
		"test",
		WithMySQL(), // No DSN provided - this should work, just not enable health checks
	)

	require.NoError(t, err)
	assert.NotNil(t, shutdown)
	assert.True(t, TelemetryConfig.MysqlEnabled)
	assert.Empty(t, TelemetryConfig.MysqlConfig.DSN)

	shutdown()
}

func TestInit_WithNATS_MissingURL(t *testing.T) {
	shutdown, err := Init(
		"test-service",
		"test",
		WithNATS(), // No URL provided
	)

	// Should fail because NATS URL is required
	require.Error(t, err)
	assert.Nil(t, shutdown)
	assert.Contains(t, err.Error(), "nats URL is required")
}

func TestInit_WithNATS_InvalidURL(t *testing.T) {
	shutdown, err := Init(
		"test-service",
		"test",
		WithNATS(NATSURL("nats://127.0.0.1:9999")),
	)

	// Should fail because NATS server is unreachable
	require.Error(t, err)
	assert.Nil(t, shutdown)
	assert.Contains(t, err.Error(), "nats connection failed")
}

func TestInit_ComplexConfiguration(t *testing.T) {
	// Test configuration with multiple components (that should fail gracefully)
	testDSN := "https://test@o123456.ingest.us.sentry.io/123456"

	shutdown, err := Init(
		"complex-test-service",
		"production",
		WithSlog(SlogLogLevel(slog.LevelInfo)),
		WithSentry(
			SentryDSN(testDSN),
			SentryEnvironment("production"),
			SentryRelease("v2.0.0"),
		),
		WithMySQL(MySQLDSN("user:pass@tcp(127.0.0.1:9998)/testdb")),
		WithNATS(NATSURL("nats://127.0.0.1:9999")),
		WithTrace(TraceExporterURL("localhost:4317")),
	)

	// This will likely fail due to missing services, but we can check configuration
	if err != nil {
		// Expected - services aren't running
		t.Logf("Expected error due to missing services: %v", err)
	} else {
		assert.NotNil(t, shutdown)
		shutdown()
	}

	// Verify all configurations were set
	assert.True(t, TelemetryConfig.SlogEnabled)
	assert.True(t, TelemetryConfig.SentryEnabled)
	assert.True(t, TelemetryConfig.MysqlEnabled)
	assert.True(t, TelemetryConfig.NatsEnabled)
	assert.True(t, TelemetryConfig.TraceEnabled)
	assert.Equal(t, "complex-test-service", TelemetryConfig.ServiceName)
}

func TestShutdownFunc(t *testing.T) {
	// Create a minimal shutdown function
	shutdown, err := Init("shutdown-test", "test", WithSlog())
	require.NoError(t, err)
	assert.NotNil(t, shutdown)

	// Calling shutdown multiple times should not panic
	shutdown()
	shutdown()
	shutdown()
}

func TestConfig_OptionFunctions(t *testing.T) {
	tests := []struct {
		name    string
		option  Option
		checkFn func(*config)
	}{
		{
			name:   "WithSlog sets slog config",
			option: WithSlog(SlogLogLevel(slog.LevelWarn)),
			checkFn: func(cfg *config) {
				assert.True(t, cfg.SlogEnabled)
				assert.Equal(t, slog.LevelWarn, cfg.SlogConfig.logLevel)
			},
		},
		{
			name: "WithSentry sets sentry config",
			option: WithSentry(
				SentryDSN("test-dsn"),
				SentryEnvironment("test-env"),
				SentryRelease("test-release"),
			),
			checkFn: func(cfg *config) {
				assert.True(t, cfg.SentryEnabled)
				assert.Equal(t, "test-dsn", cfg.SentryConfig.DSN)
				assert.Equal(t, "test-env", cfg.SentryConfig.Environment)
				assert.Equal(t, "test-release", cfg.SentryConfig.Release)
			},
		},
		{
			name:   "WithTrace sets trace config",
			option: WithTrace(TraceExporterURL("test-url")),
			checkFn: func(cfg *config) {
				assert.True(t, cfg.TraceEnabled)
				assert.Equal(t, "test-url", cfg.TraceConfig.ExporterURL)
			},
		},
		{
			name:   "WithMySQL sets mysql config",
			option: WithMySQL(MySQLDSN("test-mysql-dsn")),
			checkFn: func(cfg *config) {
				assert.True(t, cfg.MysqlEnabled)
				assert.Equal(t, "test-mysql-dsn", cfg.MysqlConfig.DSN)
			},
		},
		{
			name:   "WithNATS sets nats config",
			option: WithNATS(NATSURL("test-nats-url")),
			checkFn: func(cfg *config) {
				assert.True(t, cfg.NatsEnabled)
				assert.Equal(t, "test-nats-url", cfg.NatsConfig.URL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			cfg := &config{}
			tt.option(cfg)
			tt.checkFn(cfg)
		})
	}
}

// Test environment variable integration (if any)
func TestInit_EnvironmentIntegration(t *testing.T) {
	// Save original env vars
	originalSentry := os.Getenv("SENTRY_DSN")
	originalNats := os.Getenv("NATS_SERVERS")
	originalMySQL := os.Getenv("MYSQL_DSN")
	originalOtel := os.Getenv("OTEL_EXPORTER_ENDPOINT")

	defer func() {
		// Restore original env vars
		_ = os.Setenv("SENTRY_DSN", originalSentry)
		_ = os.Setenv("NATS_SERVERS", originalNats)
		_ = os.Setenv("MYSQL_DSN", originalMySQL)
		_ = os.Setenv("OTEL_EXPORTER_ENDPOINT", originalOtel)
	}()

	// This test demonstrates how the library would be used with environment variables
	// but doesn't actually use them in the current implementation

	shutdown, err := Init("env-test", "test")
	require.NoError(t, err)
	assert.NotNil(t, shutdown)
	shutdown()
}

func TestInit_ConcurrentCalls(t *testing.T) {
	// Test that concurrent calls to Init don't cause race conditions
	results := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(_ int) {
			shutdown, err := Init("concurrent-test", "test", WithSlog())
			if shutdown != nil {
				// Add small delay to test shutdown timing
				time.Sleep(time.Millisecond * 10)
				shutdown()
			}
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < 10; i++ {
		err := <-results
		// Most should succeed (basic slog init is simple)
		if err != nil {
			t.Logf("Concurrent init %d failed: %v", i, err)
		}
	}
}
