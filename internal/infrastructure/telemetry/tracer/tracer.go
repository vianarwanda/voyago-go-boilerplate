// Package tracer provides an abstraction for distributed tracing,
// supporting multiple providers like Datadog and OpenTelemetry (OTel).
package tracer

import (
	"context"
	"voyago/core-api/internal/infrastructure/config"

	"gorm.io/gorm"
)

// Tracer defines the interface for managing distributed tracing life-cycles.
// It allows starting spans, extracting trace context, and integrating with database ORMs.
type Tracer interface {
	// StartSpan initiates a new span and returns a child context.
	// Always call Finish() on the returned Span to avoid memory leaks.
	StartSpan(ctx context.Context, name string) (Span, context.Context)

	// UseGorm injects tracing instrumentation into a GORM database instance.
	UseGorm(db *gorm.DB)

	// ExtractTraceInfo retrieves the current TraceID and SpanID from the context.
	// Useful for logging or debugging across service boundaries.
	ExtractTraceInfo(ctx context.Context) (traceID, spanID string, ok bool)

	// Close flushes any remaining spans to the collector and releases resources.
	Close() error
}

// Span represents a single unit of work within a trace.
type Span interface {
	// SetOperationName changes the name of the span after it has been started.
	SetOperationName(name string)

	// Finish marks the end of the span and prepares it for reporting.
	Finish()

	// SetTag attaches metadata to the span for better filtering in dashboards.
	SetTag(key string, value any)
}

// New initializes a new Tracer based on the TelemetryConfig provided.
// It automatically returns a NoOpTracer if telemetry is disabled in the config.
// Supported types: "datadog", "otel".
//
// Parameters:
//   - cfg: The telemetry settings.
//   - env: The deployment environment (e.g., "production", "staging", "development").
//
// Example:
//
//	tr, _ := tracer.New(&cfg.Telemetry, "production")
func New(cfg *config.TelemetryConfig, env string) (Tracer, error) {
	if !cfg.Enabled {
		return NewNoOpTracer(), nil
	}

	switch cfg.Type {
	case "datadog":
		return NewDatadogTracer(
			cfg.Namespace,
			env,
			cfg.TracerAddress,
			cfg.SampleRate,
		), nil
	case "otel":
		return NewOTelTracer(
			cfg.Namespace,
			env,
			cfg.TracerAddress,
			cfg.SampleRate,
		)
	default:
		return NewNoOpTracer(), nil
	}
}
