# TMSLabs/tooling

## Testing

This repository includes comprehensive unit and integration tests for all core modules. The test suite covers success scenarios, error cases, and boundary conditions without requiring real external services.

### Running Tests

To run all tests:
```bash
go test ./...
```

To run tests with verbose output:
```bash
go test -v ./...
```

To run tests for a specific package:
```bash
go test -v ./telemetry
go test -v ./mysqlhelper
go test -v ./httphelper
```

To run integration tests specifically:
```bash
go test -v -run Integration
```

To run tests with coverage:
```bash
go test -cover ./...
```

### Test Coverage

The test suite includes:

- **telemetry**: Tests for initialization, health checks, error capture, and configuration options
- **mysqlhelper**: Tests for database connection and health check functionality
- **httphelper**: Tests for HTTP request tracing and handler wrapping
- **Integration tests**: Tests demonstrating integration between telemetry and MySQL/NATS connectivity

### Test Dependencies

Tests use mocks and fakes to avoid dependencies on real external services:
- No actual MySQL server required (connection failures are tested)
- No actual NATS server required (connection failures are tested)  
- No actual Sentry or OpenTelemetry endpoints required
- HTTP tests use `httptest` for isolated testing

### Running Tests in CI/CD

The tests are designed to run reliably in CI/CD environments without external dependencies:

```bash
# Basic test run
go test ./...

# With race detection
go test -race ./...

# With coverage reporting
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Initializing Telemetry

```go
func main() {

	_ = godotenv.Load()

	shutdown, err := telemetry.Init(
		"nats-event-logger",
		k8shelper.GetEnvironment(),
		telemetry.WithSentry(
			telemetry.SentryDSN(
				os.Getenv("SENTRY_DSN"),
			),
		),
		telemetry.WithSlog(),
		telemetry.WithNATS(telemetry.NATSURL(os.Getenv("NATS_SERVERS"))),
		telemetry.WithMySQL(telemetry.MySQLDSN(os.Getenv("MYSQL_DSN"))),
		telemetry.WithTrace(telemetry.TraceExporterURL(os.Getenv("OTEL_EXPORTER_ENDPOINT"))),
	)
	if err != nil {
		slog.Error("telemetry init failed", "error", err)
		sentry.CaptureException(err)
		os.Exit(1)
	}
	defer shutdown()

	// Your application logic here
}
```

## Http handler

```go
http.Handle("/hello", handlers.OtelHandler(http.HandlerFunc(HelloHandler)))
// slog is already set as default logger by telemetry.Init
slog.Info("Server starting", "addr", ":7777")
if err := http.ListenAndServe(":7777", nil); err != nil {
    slog.Error("http server error", "err", err)
    sentry.CaptureException(err)
}

// HelloHandler is a simple HTTP handler demonstrating tracing, logging, and Sentry usage.
func HelloHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	span := trace.SpanFromContext(ctx)
	span.AddEvent("handling hello route")

	if req.Method != http.MethodGet {
		err := http.ErrNotSupported
		slog.Error("invalid request method", "method", req.Method)
		span.SetStatus(codes.Error, "Method Not Allowed")
		span.RecordError(err)
		sentry.CaptureException(err)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	_, _ = io.WriteString(w, "Hello, world!\n")
	slog.Info("hello handler served")
}
```
