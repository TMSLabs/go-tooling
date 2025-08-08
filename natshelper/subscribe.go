package natshelper

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Subscribe subscribes to a NATS subject and processes messages with the provided handler.
// It extracts trace context from NATS headers if present and starts a new span for message processing.
// The handler function receives a context and the NATS message.
// It returns the subscription and any error encountered.
// It is recommended to use this function in conjunction with OpenTelemetry for distributed tracing.
// Example usage:
//
//	nc, err := nats.Connect("nats://localhost:4222")
//	if err != nil {
//	    log.Fatalf("Failed to connect to NATS: %v", err)
//	}
//	sub, err := natshelper.Subscribe(nc, "my.subject", func(ctx context.Context, msg *nats.Msg) {
//	    // Process the message
//	    fmt.Printf("Received message: %s\n", string(msg.Data))
//	})
//	if err != nil {
//	    log.Fatalf("Failed to subscribe: %v", err)
//	}
//	defer sub.Unsubscribe()
//	// Keep the connection alive to receive messages
//	select {}
//	}
func Subscribe(
	nc *nats.Conn,
	subj string,
	handler func(ctx context.Context, msg *nats.Msg),
) (*nats.Subscription, error) {
	tracer := otel.Tracer("natshelper")

	return nc.Subscribe(subj, func(msg *nats.Msg) {
		ctx := context.Background()
		// Extract trace context from NATS headers if present
		if msg.Header != nil {
			ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(msg.Header))
		}
		// Start a new span for message processing
		ctx, span := tracer.Start(ctx, fmt.Sprintf("nats.receive.%s", subj))
		defer span.End()
		handler(ctx, msg)
	})
}
