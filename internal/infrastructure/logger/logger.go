// Package logger provides a unified logging interface for the application,
// supporting multiple drivers and context-aware metadata extraction.
package logger

import (
	"context"
	"voyago/core-api/internal/infrastructure/config"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
)

// Logger defines the standard interface for structured logging across the system.
// It supports chaining for context and field enrichment.
type Logger interface {
	// WithContext extracts metadata from the context (e.g., TraceID, RequestID)
	// and returns a new Logger instance with these fields attached.
	WithContext(ctx context.Context) Logger

	// WithField adds a single key-value pair to the logging context.
	WithField(key string, value any) Logger

	// WithFields adds multiple key-value pairs to the logging context.
	WithFields(fields map[string]any) Logger

	// Debug logs a message at the Debug level. Use this for verbose development info.
	Debug(message string)
	// Info logs a message at the Info level. This is the default for general application flow.
	Info(message string)
	// Warn logs a message at the Warn level. Use for non-critical issues that need attention.
	Warn(message string)
	// Error logs a message at the Error level. Use for critical failures or caught exceptions.
	Error(message string)
}

// New creates and returns a Logger implementation based on the application environment.
//
// Logic:
//   - "production": Returns a Logrus logger (optimized for JSON/structured log aggregation).
//   - "staging": Returns a Logrus logger (optimized for JSON/structured log aggregation).
//   - "development": Returns a Stdout logger (optimized for human readability/tinted output).
//   - default: Returns a NoOp logger (disables all logging).
//
// Example:
//
//	log := logger.New(cfg, trc)
//	log.WithContext(ctx).Info("Application started")
func New(cfg *config.Config, trc tracer.Tracer) Logger {
	switch cfg.App.Env {
	case "production", "staging":
		return NewLogrus(cfg, trc)
	case "development":
		return NewStdoutLogger(cfg, trc)
	default:
		return NewNoOpLogger()
	}
}
