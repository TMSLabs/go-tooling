package natshelper

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Publish publishes a message to a NATS subject with OpenTelemetry tracing.
// It injects the trace context into the message headers.
// The function starts a new span for the publish operation and returns any error encountered.
// It is recommended to use this function in conjunction with OpenTelemetry for distributed tracing.
// Example usage:
//
//	nc, err := nats.Connect("nats://localhost:4222")
//	if err != nil {
//	    log.Fatalf("Failed to connect to NATS: %v", err)
//	}
//	err = natshelper.Publish(context.Background(), nc, "my.subject", []byte("Hello, World!"))
//	if err != nil {
//	    log.Fatalf("Failed to publish message: %v", err)
//	}
func Publish(ctx context.Context, nc *nats.Conn, subj string, data []byte) error {
	tracer := otel.Tracer("natshelper")

	// Start a span for publish
	ctx, span := tracer.Start(ctx, "nats.publish."+subj)
	defer span.End()

	msg := &nats.Msg{
		Subject: subj,
		Data:    data,
		Header:  nats.Header{},
	}
	// Inject trace context into headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(msg.Header))

	return nc.PublishMsg(msg)
}

// PublishMsg publishes a NATS message with OpenTelemetry tracing.
// It injects the trace context into the message headers.
// The function starts a new span for the publish operation and returns any error encountered.
// It is recommended to use this function in conjunction with OpenTelemetry for distributed tracing.
// Example usage:
//
//	nc, err := nats.Connect("nats://localhost:4222")
//	if err != nil {
//	    log.Fatalf("Failed to connect to NATS: %v", err)
//	}
//	msg := &nats.Msg{
//	    Subject: "my.subject",
//	    Data:    []byte("Hello, World!"),
//	    Header:  nats.Header{},
//	}
//	err = natshelper.PublishMsg(context.Background(), nc, msg)
//	if err != nil {
//	    log.Fatalf("Failed to publish message: %v", err)
//	}
func PublishMsg(ctx context.Context, nc *nats.Conn, msg *nats.Msg) error {
	tracer := otel.Tracer("natshelper")

	// Start a span for publish
	ctx, span := tracer.Start(ctx, "nats.publish."+msg.Subject)
	defer span.End()

	// Inject trace context into headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(msg.Header))

	return nc.PublishMsg(msg)
}
