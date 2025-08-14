package telemetry

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/TMSLabs/go-tooling/mysqlhelper"
	"github.com/nats-io/nats.go"
)

var (
	// LastHealthCheckEvent stores the timestamp of the last health check event.
	LastHealthCheckEvent = ""
)

// HealthzEventChecker subscribes to the health check and publishes health check events periodically.
func HealthzEventChecker(nc *nats.Conn, serviceName string) {
	_, err := nc.Subscribe(serviceName+".healthz", func(_ *nats.Msg) {
		// fmt.Printf("Received health check event\n")
		LastHealthCheckEvent = time.Now().Format(time.RFC3339)
	})
	if err != nil {
		slog.Error("Error subscribing to health check event", "error", err)
		return
	}

	for {
		data := []byte("Health check event")
		err := nc.Publish(serviceName+".healthz", data)
		if err != nil {
			slog.Error("Error publishing health check event", "error", err)
			return
		}
		time.Sleep(60 * time.Second)
	}
}

// HealthzEndpointHandler handles the health check endpoint for the service.
func HealthzEndpointHandler(w http.ResponseWriter, _ *http.Request) {

	if TelemetryConfig.MysqlEnabled {
		if err := mysqlhelper.CheckConnection(TelemetryConfig.MysqlConfig.DSN); err != nil {
			slog.Error("MySQL connection check failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintln(w, "MySQL connection failed:", err)
			return
		}
	}

	if TelemetryConfig.NatsEnabled {
		if err := CheckConnection(TelemetryConfig.NatsConfig.URL); err != nil {
			slog.Error("NATS connection check failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintln(w, "NATS connection failed:", err)
			return
		}

		if LastHealthCheckEvent == "" {
			slog.Warn("No health check event received yet")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintln(w, "No health check event received yet")
			return
		}

		if LastHealthCheckEvent < time.Now().Add(-5*time.Minute).Format(time.RFC3339) {
			slog.Warn(
				"Last health check event is older than 5 minutes",
				"last_event",
				LastHealthCheckEvent,
			)
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintln(w, "Last health check event is older than 5 minutes")
			return
		}

		slog.Debug(
			"Health check event received",
			"last_event",
			LastHealthCheckEvent,
		)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "{\"status\": \"ok\", \"message\": \"Service is healthy\"}")

}

// CheckConnection checks if the NATS server is reachable.
func CheckConnection(dsn string) error {
	nc, err := nats.Connect(dsn)
	if err != nil {
		return err
	}
	defer nc.Close()

	return nil
}
