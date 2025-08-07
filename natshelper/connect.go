// Package natshelper provides a helper for connecting to NATS.
package natshelper

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

var (
	// NatsConn is the global NATS connection.
	NatsConn *nats.Conn
)

// Connect initializes the NATS connection.
func Connect(natsURL string) error {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Set the global NATS connection
	NatsConn = nc
	return nil
}
