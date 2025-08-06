# TMSLabs/tooling

## Initializing Telemetry

```go
func main() {
	shutdown, err := telemetry.Init()
	if err != nil {
		// Using slog after telemetry.Init sets up slog
		// If slog isn't available, fallback to standard log
		println("telemetry init failed:", err.Error())
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
