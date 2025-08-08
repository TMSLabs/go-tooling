# TMSLabs/tooling

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
