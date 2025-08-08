# TMSLabs Go Tooling

[![Tests](https://github.com/TMSLabs/go-tooling/actions/workflows/test.yml/badge.svg)](https://github.com/TMSLabs/go-tooling/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/TMSLabs/go-tooling)](https://goreportcard.com/report/github.com/TMSLabs/go-tooling)
[![codecov](https://codecov.io/gh/TMSLabs/go-tooling/branch/main/graph/badge.svg)](https://codecov.io/gh/TMSLabs/go-tooling)
[![Go Reference](https://pkg.go.dev/badge/github.com/TMSLabs/go-tooling.svg)](https://pkg.go.dev/github.com/TMSLabs/go-tooling)

A comprehensive Go utility library providing helper packages for building applications with telemetry, observability, and infrastructure integrations. This library simplifies the setup and integration of common tools like OpenTelemetry tracing, Sentry error tracking, structured logging, database connections, NATS messaging, and Kubernetes utilities.

## 🌟 Features

- **🔍 Telemetry Integration**: Unified initialization for OpenTelemetry, Sentry, and structured logging
- **🌐 HTTP Utilities**: OpenTelemetry-enabled HTTP handlers and request utilities
- **🗄️ Database Support**: MySQL connection management with health checks
- **📡 Message Queues**: NATS messaging utilities for pub/sub patterns
- **☸️ Kubernetes Integration**: Environment detection and namespace utilities
- **🛡️ Production Ready**: Comprehensive error handling, testing, and observability
- **⚡ Zero Dependencies**: Clean interfaces without forcing specific implementations

## 📋 Table of Contents

- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Package Documentation](#-package-documentation)
  - [Telemetry](#telemetry-package)
  - [HTTP Helper](#httphelper-package)
  - [MySQL Helper](#mysqlhelper-package)
  - [NATS Helper](#natshelper-package)
  - [Kubernetes Helper](#k8shelper-package)
- [Configuration](#-configuration)
- [Examples](#-examples)
- [Testing](#-testing)
- [Contributing](#-contributing)
- [Troubleshooting](#-troubleshooting)
- [License](#-license)

## 🚀 Installation

### Prerequisites

- **Go**: Version 1.22 or higher
- **Optional External Services** (for full functionality):
  - MySQL database server
  - NATS messaging server
  - Sentry account (for error tracking)
  - OpenTelemetry collector (for distributed tracing)

### Install the Library

```bash
go get github.com/TMSLabs/go-tooling
```

### Import Packages

```go
import (
    "github.com/TMSLabs/go-tooling/telemetry"
    "github.com/TMSLabs/go-tooling/httphelper"
    "github.com/TMSLabs/go-tooling/mysqlhelper"
    "github.com/TMSLabs/go-tooling/natshelper"
    "github.com/TMSLabs/go-tooling/k8shelper"
)
```

## ⚡ Quick Start

Here's a minimal example to get started with telemetry and HTTP handling:

```go
package main

import (
    "context"
    "log/slog"
    "net/http"
    "os"

    "github.com/TMSLabs/go-tooling/telemetry"
    "github.com/TMSLabs/go-tooling/httphelper"
    "github.com/TMSLabs/go-tooling/k8shelper"
)

func main() {
    // Initialize telemetry with structured logging
    shutdown, err := telemetry.Init(
        "my-service",
        k8shelper.GetEnvironment(),
        telemetry.WithSlog(),
    )
    if err != nil {
        slog.Error("telemetry init failed", "error", err)
        os.Exit(1)
    }
    defer shutdown()

    // Set up HTTP handler with tracing
    http.Handle("/hello", httphelper.HTTPHandler(helloHandler, "hello-endpoint"))
    
    slog.Info("Server starting", "addr", ":8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        slog.Error("server failed", "error", err)
    }
}

func helloHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    slog.InfoContext(ctx, "handling hello request")
    w.Write([]byte("Hello, World!"))
}
```

## 📚 Package Documentation

### Telemetry Package

The telemetry package provides unified initialization for observability tools including OpenTelemetry tracing, Sentry error tracking, and structured logging.

#### Features

- **Structured Logging**: Built-in slog configuration
- **Distributed Tracing**: OpenTelemetry integration
- **Error Tracking**: Sentry integration with automatic error capture
- **Health Checks**: NATS-based health monitoring
- **Configuration**: Flexible option-based configuration

#### Basic Usage

```go
// Minimal setup with just logging
shutdown, err := telemetry.Init("my-service", "production",
    telemetry.WithSlog(),
)
defer shutdown()

// Full setup with all integrations
shutdown, err := telemetry.Init("my-service", "production",
    telemetry.WithSlog(telemetry.SlogLogLevel(slog.LevelDebug)),
    telemetry.WithSentry(telemetry.SentryDSN(os.Getenv("SENTRY_DSN"))),
    telemetry.WithTrace(telemetry.TraceExporterURL(os.Getenv("OTEL_EXPORTER_ENDPOINT"))),
    telemetry.WithNATS(telemetry.NATSURL(os.Getenv("NATS_SERVERS"))),
    telemetry.WithMySQL(telemetry.MySQLDSN(os.Getenv("MYSQL_DSN"))),
)
defer shutdown()
```

#### Configuration Options

| Option | Purpose | Environment Variable |
|--------|---------|---------------------|
| `WithSlog()` | Enable structured logging | - |
| `WithSentry()` | Enable error tracking | `SENTRY_DSN` |
| `WithTrace()` | Enable distributed tracing | `OTEL_EXPORTER_ENDPOINT` |
| `WithNATS()` | Enable NATS health checks | `NATS_SERVERS` |
| `WithMySQL()` | Enable MySQL health checks | `MYSQL_DSN` |

#### Error Capture

```go
import "github.com/TMSLabs/go-tooling/telemetry"

// Capture errors with context
telemetry.CaptureError(err, "Database connection failed")

// Capture with additional context
telemetry.CaptureErrorWithContext(ctx, err, "Processing user request", map[string]interface{}{
    "user_id": userID,
    "operation": "create_user",
})
```

#### Health Checks

The telemetry package provides HTTP health check endpoints:

```go
http.Handle("/healthz", telemetry.HealthzEndpointHandler())
```

### HTTPHelper Package

Provides HTTP utilities with automatic OpenTelemetry tracing integration.

#### Features

- **Request Tracing**: Automatic span creation and context propagation
- **Sentry Integration**: Automatic breadcrumb creation for HTTP requests
- **Context Propagation**: Seamless trace context handling

#### HTTP Handler Wrapper

```go
import "github.com/TMSLabs/go-tooling/httphelper"

// Wrap your handler function
http.Handle("/api/users", httphelper.HTTPHandler(handleUsers, "users-endpoint"))

func handleUsers(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    // Your handler logic with automatic tracing
    span := trace.SpanFromContext(ctx)
    span.AddEvent("processing user request")
    
    // Handle the request...
    slog.InfoContext(ctx, "user request processed")
}
```

#### Making HTTP Requests

```go
import "github.com/TMSLabs/go-tooling/httphelper"

// Create a traced HTTP request
client := &http.Client{}
req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)

resp, err := httphelper.HTTPDo(ctx, client, req)
if err != nil {
    slog.ErrorContext(ctx, "request failed", "error", err)
    return
}
defer resp.Body.Close()
```

### MySQLHelper Package

Provides utilities for MySQL database connections with health checking.

#### Features

- **Connection Management**: Simple database connection setup
- **Health Checks**: Built-in connection health verification
- **Error Handling**: Comprehensive error reporting

#### Basic Usage

```go
import "github.com/TMSLabs/go-tooling/mysqlhelper"

// Connect to database
dsn := "user:password@tcp(localhost:3306)/dbname?parseTime=true"
db, err := mysqlhelper.Connect(dsn)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Check connection health
if err := mysqlhelper.CheckConnection(dsn); err != nil {
    log.Printf("Database health check failed: %v", err)
}
```

#### Integration with Telemetry

```go
// Initialize telemetry with MySQL health checks
shutdown, err := telemetry.Init("my-service", "production",
    telemetry.WithSlog(),
    telemetry.WithMySQL(telemetry.MySQLDSN(dsn)),
)
defer shutdown()
```

### NATSHelper Package

Provides utilities for NATS messaging integration.

#### Features

- **Connection Management**: Global NATS connection handling
- **Pub/Sub Support**: Publishing and subscribing utilities
- **Error Handling**: Comprehensive connection error management

#### Basic Usage

```go
import "github.com/TMSLabs/go-tooling/natshelper"

// Connect to NATS
err := natshelper.Connect("nats://localhost:4222")
if err != nil {
    log.Fatal(err)
}

// Use the global connection
natshelper.NatsConn.Publish("subject", []byte("message"))

// Subscribe to messages
sub, err := natshelper.NatsConn.SubscribeSync("subject")
if err != nil {
    log.Fatal(err)
}

msg, err := sub.NextMsg(time.Second * 10)
if err != nil {
    log.Printf("No message received: %v", err)
}
```

### K8sHelper Package

Provides Kubernetes utilities for environment detection and namespace management.

#### Features

- **Environment Detection**: Automatic environment determination from namespace
- **Namespace Reading**: Access to current Kubernetes namespace

#### Basic Usage

```go
import "github.com/TMSLabs/go-tooling/k8shelper"

// Get current environment based on Kubernetes namespace
env := k8shelper.GetEnvironment()
fmt.Printf("Current environment: %s\n", env)

// Possible return values: "development", "testing", "staging", "production"
```

#### Environment Mapping

| Namespace Contains | Detected Environment |
|-------------------|---------------------|
| `prod` | `production` |
| `test` | `testing` |
| `staging` | `staging` |
| _other_ | `development` |

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `SENTRY_DSN` | Sentry project DSN for error tracking | No | `https://key@sentry.io/project` |
| `OTEL_EXPORTER_ENDPOINT` | OpenTelemetry collector endpoint | No | `localhost:4317` |
| `NATS_SERVERS` | NATS server URLs (comma-separated) | No | `nats://localhost:4222` |
| `MYSQL_DSN` | MySQL connection string | No | `user:pass@tcp(host:port)/db` |

### Configuration Example

Create a `.env` file in your project root:

```env
# Telemetry Configuration
SENTRY_DSN=https://your-sentry-dsn@sentry.io/project-id
OTEL_EXPORTER_ENDPOINT=localhost:4317
NATS_SERVERS=nats://localhost:4222
MYSQL_DSN=user:password@tcp(localhost:3306)/myapp?parseTime=true

# Application Configuration  
APP_ENV=development
APP_PORT=8080
```

Load environment variables in your application:

```go
import "github.com/joho/godotenv"

func main() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }
    
    // Initialize telemetry with environment variables
    shutdown, err := telemetry.Init("my-service", os.Getenv("APP_ENV"),
        telemetry.WithSlog(),
        telemetry.WithSentry(telemetry.SentryDSN(os.Getenv("SENTRY_DSN"))),
        telemetry.WithTrace(telemetry.TraceExporterURL(os.Getenv("OTEL_EXPORTER_ENDPOINT"))),
        telemetry.WithNATS(telemetry.NATSURL(os.Getenv("NATS_SERVERS"))),
        telemetry.WithMySQL(telemetry.MySQLDSN(os.Getenv("MYSQL_DSN"))),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown()
}
```

## 💡 Examples

### Complete Web Service Example

```go
package main

import (
    "context"
    "encoding/json"
    "log/slog"
    "net/http"
    "os"
    "time"

    "github.com/TMSLabs/go-tooling/telemetry"
    "github.com/TMSLabs/go-tooling/httphelper"
    "github.com/TMSLabs/go-tooling/mysqlhelper"
    "github.com/TMSLabs/go-tooling/k8shelper"
    "go.opentelemetry.io/otel/trace"
)

func main() {
    // Initialize telemetry
    shutdown, err := telemetry.Init(
        "user-service",
        k8shelper.GetEnvironment(),
        telemetry.WithSlog(),
        telemetry.WithSentry(telemetry.SentryDSN(os.Getenv("SENTRY_DSN"))),
        telemetry.WithTrace(telemetry.TraceExporterURL(os.Getenv("OTEL_EXPORTER_ENDPOINT"))),
        telemetry.WithMySQL(telemetry.MySQLDSN(os.Getenv("MYSQL_DSN"))),
    )
    if err != nil {
        slog.Error("telemetry initialization failed", "error", err)
        os.Exit(1)
    }
    defer shutdown()

    // Initialize database
    db, err := mysqlhelper.Connect(os.Getenv("MYSQL_DSN"))
    if err != nil {
        slog.Error("database connection failed", "error", err)
        os.Exit(1)
    }
    defer db.Close()

    // Setup routes
    http.Handle("/users", httphelper.HTTPHandler(handleUsers, "users-endpoint"))
    http.Handle("/healthz", telemetry.HealthzEndpointHandler())

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    slog.Info("server starting", "port", port, "environment", k8shelper.GetEnvironment())
    
    server := &http.Server{
        Addr:         ":" + port,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    }

    if err := server.ListenAndServe(); err != nil {
        slog.Error("server failed", "error", err)
    }
}

func handleUsers(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    span := trace.SpanFromContext(ctx)
    span.AddEvent("processing user request")

    switch r.Method {
    case http.MethodGet:
        getUsers(ctx, w, r)
    case http.MethodPost:
        createUser(ctx, w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func getUsers(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    slog.InfoContext(ctx, "fetching users")
    
    users := []map[string]interface{}{
        {"id": 1, "name": "John Doe", "email": "john@example.com"},
        {"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func createUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    slog.InfoContext(ctx, "creating user")
    
    var user map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        telemetry.CaptureErrorWithContext(ctx, err, "Invalid JSON payload", nil)
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Simulate user creation
    user["id"] = 123
    user["created_at"] = time.Now()

    w.Header().Set("Content-Type", "application/json")
    w.WriteStatus(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
    
    slog.InfoContext(ctx, "user created successfully", "user_id", user["id"])
}
```

### NATS Publisher/Subscriber Example

```go
package main

import (
    "log"
    "log/slog"
    "os"
    "time"

    "github.com/TMSLabs/go-tooling/telemetry"
    "github.com/TMSLabs/go-tooling/natshelper"
)

func main() {
    // Initialize telemetry
    shutdown, err := telemetry.Init("nats-service", "development",
        telemetry.WithSlog(),
        telemetry.WithNATS(telemetry.NATSURL(os.Getenv("NATS_SERVERS"))),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown()

    // Connect to NATS
    if err := natshelper.Connect(os.Getenv("NATS_SERVERS")); err != nil {
        slog.Error("NATS connection failed", "error", err)
        return
    }

    // Subscribe to messages
    sub, err := natshelper.NatsConn.Subscribe("user.created", func(msg *nats.Msg) {
        slog.Info("received message", "subject", msg.Subject, "data", string(msg.Data))
    })
    if err != nil {
        log.Fatal(err)
    }
    defer sub.Unsubscribe()

    // Publish messages
    for i := 0; i < 10; i++ {
        message := fmt.Sprintf(`{"user_id": %d, "timestamp": "%s"}`, i+1, time.Now().Format(time.RFC3339))
        
        if err := natshelper.NatsConn.Publish("user.created", []byte(message)); err != nil {
            slog.Error("failed to publish message", "error", err)
        } else {
            slog.Info("published message", "message", message)
        }
        
        time.Sleep(time.Second)
    }

    // Wait for messages to be processed
    time.Sleep(5 * time.Second)
}
```

## 🧪 Testing

This repository includes comprehensive unit and integration tests for all core modules. The test suite covers success scenarios, error cases, and boundary conditions without requiring real external services.

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./telemetry
go test -v ./mysqlhelper
go test -v ./httphelper

# Run integration tests specifically
go test -v -run Integration

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Test Coverage

The test suite includes:

- **telemetry**: Tests for initialization, health checks, error capture, and configuration options
- **mysqlhelper**: Tests for database connection and health check functionality
- **httphelper**: Tests for HTTP request tracing and handler wrapping
- **natshelper**: Tests for NATS connection management
- **Integration tests**: Tests demonstrating integration between telemetry and external services

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

## 🤝 Contributing

We welcome contributions to the TMSLabs Go Tooling project! Here's how you can help:

### Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-tooling.git
   cd go-tooling
   ```
3. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

### Development Setup

1. **Install Go 1.22 or higher**
2. **Install dependencies**:
   ```bash
   go mod download
   ```
3. **Install development tools**:
   ```bash
   # Install golangci-lint for linting
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
   
   # Install gosec for security scanning
   go install github.com/securecode/gosec/v2/cmd/gosec@latest
   ```

### Code Standards

- **Go Style**: Follow standard Go conventions and use `gofmt`
- **Linting**: Code must pass `golangci-lint run`
- **Security**: Code must pass `gosec ./...`
- **Testing**: All new code must include tests with good coverage
- **Documentation**: Public APIs must have godoc comments

### Making Changes

1. **Write your code** following the project conventions
2. **Add tests** for any new functionality
3. **Run tests** to ensure everything works:
   ```bash
   go test ./...
   go test -race ./...
   ```
4. **Run linting**:
   ```bash
   golangci-lint run
   gosec ./...
   ```
5. **Update documentation** if needed

### Submitting Changes

1. **Commit your changes** with clear commit messages:
   ```bash
   git add .
   git commit -m "feat: add new telemetry option for custom exporters"
   ```
2. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```
3. **Create a Pull Request** on GitHub

### Commit Message Format

We use conventional commits for clear git history:

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Test additions or modifications
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

### Pull Request Guidelines

- **Clear Description**: Explain what changes you've made and why
- **Link Issues**: Reference any related GitHub issues
- **Test Coverage**: Ensure tests pass and maintain good coverage
- **Small PRs**: Keep pull requests focused and reasonably sized
- **Documentation**: Update README or godoc as needed

### Reporting Issues

When reporting bugs or requesting features:

1. **Search existing issues** to avoid duplicates
2. **Use issue templates** when available
3. **Provide details**:
   - Go version
   - Operating system
   - Steps to reproduce
   - Expected vs actual behavior
   - Error messages or logs

### Code Review Process

1. **Automated checks** must pass (tests, linting, security)
2. **Peer review** by at least one maintainer
3. **Discussion** if needed for design decisions
4. **Approval** and merge by maintainers

### Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Code Comments**: Ask questions in pull requests

Thank you for contributing to TMSLabs Go Tooling!

## 🔧 Troubleshooting

### Common Issues

#### Import Path Issues

**Problem**: Module import errors or "package not found" errors.

**Solution**: 
```bash
# Ensure you're using the correct import path
go mod tidy
go mod download

# Verify Go version (requires 1.22+)
go version
```

#### Telemetry Initialization Fails

**Problem**: `telemetry.Init()` returns errors about missing configuration.

**Solution**:
```go
// Check that required environment variables are set
if os.Getenv("SENTRY_DSN") == "" {
    log.Println("Warning: SENTRY_DSN not set, Sentry disabled")
}

// Use minimal configuration for development
shutdown, err := telemetry.Init("my-service", "development",
    telemetry.WithSlog(), // Only enable logging
)
```

#### Database Connection Issues

**Problem**: MySQL connection fails with "connection refused" or timeout errors.

**Solution**:
```go
// Test connection before using
dsn := os.Getenv("MYSQL_DSN")
if err := mysqlhelper.CheckConnection(dsn); err != nil {
    log.Printf("Database not available: %v", err)
    // Handle graceful degradation
}

// Use proper DSN format
// Example: "user:password@tcp(localhost:3306)/database?parseTime=true"
```

#### NATS Connection Problems

**Problem**: NATS connection fails or messages aren't being delivered.

**Solution**:
```go
// Check NATS server availability
natsURL := os.Getenv("NATS_SERVERS")
if natsURL == "" {
    natsURL = "nats://localhost:4222" // Default
}

err := natshelper.Connect(natsURL)
if err != nil {
    log.Printf("NATS not available: %v", err)
    // Continue without NATS functionality
}
```

#### OpenTelemetry Tracing Issues

**Problem**: Traces not appearing in observability platforms.

**Solution**:
```go
// Verify exporter URL format
exporterURL := os.Getenv("OTEL_EXPORTER_ENDPOINT")
// Should be: "localhost:4317" (not "http://localhost:4317")

// Check if collector is running
// docker run -p 4317:4317 otel/opentelemetry-collector

// Verify spans are being created
span := trace.SpanFromContext(ctx)
span.AddEvent("debug checkpoint")
```

#### Sentry Integration Issues

**Problem**: Errors not appearing in Sentry dashboard.

**Solution**:
```go
// Verify DSN format
// Should be: "https://key@sentry.io/project-id"

// Test error capture
telemetry.CaptureError(errors.New("test error"), "Testing Sentry integration")

// Check environment configuration
shutdown, err := telemetry.Init("my-service", "development", // Use correct environment
    telemetry.WithSentry(telemetry.SentryDSN(os.Getenv("SENTRY_DSN"))),
)
```

### Performance Issues

#### High Memory Usage

**Problem**: Application using excessive memory.

**Solution**:
```go
// Reduce log level in production
telemetry.WithSlog(telemetry.SlogLogLevel(slog.LevelWarn))

// Implement proper shutdown
defer shutdown() // Always call shutdown function

// Monitor span creation
// Avoid creating too many spans in hot code paths
```

#### Slow HTTP Requests

**Problem**: HTTP requests taking longer than expected.

**Solution**:
```go
// Add timeouts to HTTP clients
client := &http.Client{
    Timeout: 30 * time.Second,
}

// Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

resp, err := httphelper.HTTPDo(ctx, client, req)
```

### Development Issues

#### Tests Failing in CI

**Problem**: Tests pass locally but fail in CI/CD.

**Solution**:
```bash
# Run tests with race detection locally
go test -race ./...

# Run tests with verbose output
go test -v ./...

# Check for environment-specific issues
# Tests should not depend on external services
```

#### Linting Errors

**Problem**: `golangci-lint` reports errors.

**Solution**:
```bash
# Run linter locally
golangci-lint run

# Fix common issues
gofmt -w .
go mod tidy

# Check specific linter configuration in .golangci.yml
```

### Environment-Specific Issues

#### Kubernetes Deployment

**Problem**: Application not working correctly in Kubernetes.

**Solution**:
```yaml
# Ensure environment variables are set in deployment
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: myapp
        env:
        - name: SENTRY_DSN
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: sentry-dsn
        - name: MYSQL_DSN
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: mysql-dsn
```

#### Docker Issues

**Problem**: Application not working in Docker container.

**Solution**:
```dockerfile
# Use correct Go version in Dockerfile
FROM golang:1.24-alpine

# Set working directory
WORKDIR /app

# Copy and build
COPY . .
RUN go mod download
RUN go build -o myapp

# Use multi-stage build for smaller images
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/myapp /myapp
CMD ["/myapp"]
```

### Getting Help

If you're still experiencing issues:

1. **Check the logs**: Enable debug logging with `telemetry.WithSlog(telemetry.SlogLogLevel(slog.LevelDebug))`
2. **Review examples**: Look at the complete examples in this README
3. **Check GitHub Issues**: Search for similar problems
4. **Create an Issue**: If you find a bug, please report it with:
   - Go version (`go version`)
   - Operating system
   - Complete error messages
   - Minimal reproduction case

## 📚 Additional Resources

### External Documentation

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Sentry Go SDK Documentation](https://docs.sentry.io/platforms/go/)
- [Go slog Package](https://pkg.go.dev/log/slog)
- [NATS Go Client](https://docs.nats.io/using-nats/developer/connecting/go)
- [MySQL Go Driver](https://github.com/go-sql-driver/mysql)

### Related Projects

- [OpenTelemetry](https://opentelemetry.io/) - Observability framework
- [Sentry](https://sentry.io/) - Error tracking platform  
- [NATS](https://nats.io/) - Message-oriented middleware
- [Jaeger](https://www.jaegertracing.io/) - Distributed tracing platform
- [Prometheus](https://prometheus.io/) - Monitoring and alerting

### Community

- [Go Community](https://golang.org/community/)
- [OpenTelemetry Community](https://opentelemetry.io/community/)
- [NATS Community](https://natsio.slack.com/)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Made with ❤️ by TMSLabs**

For more information, visit our [GitHub organization](https://github.com/TMSLabs) or check out our other projects.
