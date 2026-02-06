// Package metrics provides an abstraction layer for application telemetry.
package metrics

import (
	"time"
	"voyago/core-api/internal/infrastructure/config"
)

// Metrics defines the interface for recording application performance data.
// It allows the application to be agnostic of the underlying provider (Datadog, OTel, etc).
type Metrics interface {
	// Incr increments a counter by 1. Use this for tracking event occurrences.
	Incr(name string, tags []string)

	// Distribution records numeric values for statistical analysis (e.g., payload size).
	Distribution(name string, value float64, tags []string)

	// Timing records the duration of an operation.
	Timing(name string, value time.Duration, tags []string)

	// RecordHTTP captures performance data for an incoming HTTP request.
	//
	// Parameters:
	//   - method: The HTTP verb used (e.g., "GET", "POST").
	//   - path: The URL pattern or template (e.g., "/bookings/:id") to avoid high cardinality.
	//   - status: The final HTTP response code (e.g., 200, 404, 500).
	//   - duration: Total execution time in seconds (float64).
	//
	// Implementation should ideally update a Counter for throughput/errors
	// and a Histogram/Summary for latency distribution (P99, P95).
	RecordHTTP(method string, path string, routePath string, statusCode int, duration float64)

	// Close flushes any buffered metrics and closes the connection to the provider.
	Close() error
}

// New creates a new Metrics instance based on the provided TelemetryConfig.
// It returns a NoOp (No-Operation) implementation if telemetry is disabled.
// Supported types: "datadog", "otel".
//
// Parameters:
//   - cfg: The telemetry settings.
//   - env: The deployment environment (e.g., "production", "staging", "development").
//
// Example:
//
//	m, err := metrics.New(&cfg.Telemetry, "production")
func New(cfg *config.TelemetryConfig, env string) (Metrics, error) {
	if !cfg.Enabled {
		return NewNoOpMetrics(), nil
	}

	switch cfg.Type {
	case "datadog":
		return NewDatadogMetrics(
			cfg.MetricsAddress,
			cfg.Namespace,
			[]string{"env:" + env},
		)
	case "otel":
		return NewOTelMetrics(
			cfg.MetricsAddress,
			cfg.Namespace,
			[]string{"env:" + env},
		)
	default:
		return NewNoOpMetrics(), nil
	}
}
