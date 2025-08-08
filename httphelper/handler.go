// Package httphelper provides utilities for handling HTTP requests with OpenTelemetry tracing.
package httphelper

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// HTTPHandler wraps an HTTP handler function with OpenTelemetry tracing.
// It extracts the trace context from the HTTP request headers and starts a new span.
// The handler function receives a context with the trace span and the HTTP response writer and request.
// The span name can be customized with the `spanName` parameter.
// Example usage:
//
//	http.Handle("/my-endpoint", httphelper.HTTPHandler(myHandler, "MyEndpointSpan"))
//
// // The handler function should be defined as:
//
//	func myHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
//		// Your handler logic here
//		w.Write([]byte("Hello, World!"))
//	}
//
// It is recommended to use this function in conjunction with OpenTelemetry for distributed tracing.
func HTTPHandler(
	handler func(ctx context.Context, w http.ResponseWriter, r *http.Request),
	spanName string,
) http.HandlerFunc {
	propagator := otel.GetTextMapPropagator()
	tracer := otel.Tracer("httphelper")

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, spanName)
		defer span.End()
		handler(ctx, w, r)
	}
}
