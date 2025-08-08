package httphelper

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// HTTPDo performs an HTTP request with OpenTelemetry tracing.
// It injects the current trace context into the request headers and starts a new span for the request.
// The function takes a context, an HTTP client, an HTTP request, and a span name.
// It returns the HTTP response and any error encountered.
// It is recommended to use this function in conjunction with OpenTelemetry for distributed tracing.
// Example usage:
//
//	ctx := context.Background()
//	client := &http.Client{}
//	req, err := http.NewRequest("GET", "https://example.com", nil)
//	if err != nil {
//	    log.Fatalf("Failed to create request: %v", err)
//	}
//	resp, err := httphelper.HTTPDo(ctx, client, req, "ExampleRequest")
//	if err != nil {
//	    log.Fatalf("HTTP request failed: %v", err)
//	}
//	defer resp.Body.Close()
//	if resp.StatusCode != http.StatusOK {
//	    log.Fatalf("HTTP request failed with status: %s", resp.Status)
//	}
//
// This function is useful for making HTTP requests while maintaining trace context across service boundaries.
// It is particularly useful in microservices architectures where requests may span multiple services.
func HTTPDo(
	ctx context.Context,
	client *http.Client,
	req *http.Request,
	spanName string,
) (*http.Response, error) {
	propagator := otel.GetTextMapPropagator()
	tracer := otel.Tracer("httphelper")

	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	// Inject current trace context into outgoing request headers
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Use passed context for request
	req = req.WithContext(ctx)
	return client.Do(req)
}
