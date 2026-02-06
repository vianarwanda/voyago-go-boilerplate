package utils

import (
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/pkg/apperror"
)

// RecordSpanError is a global helper to enrich a trace span with error metadata.
// It automatically detects if the error is an apperror.AppError and extracts
// machine-readable tags (Code, Kind) and debugging details.
//
// Parameters:
//   - span: The active tracer span. If nil, this function does nothing.
//   - err: The error to be recorded. If nil, this function does nothing.
func RecordSpanError(span tracer.Span, err error) {
	if err == nil || span == nil {
		return
	}

	// Standard error tags
	span.SetTag("error", true)
	span.SetTag("error.message", err.Error())

	// Enhanced metadata for AppError
	if appErr, ok := err.(*apperror.AppError); ok {
		span.SetTag("error.code", appErr.Code)
		span.SetTag("error.kind", string(appErr.Kind))
	}
}
