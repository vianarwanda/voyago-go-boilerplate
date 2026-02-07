package apperror

import (
	"fmt"
	"strings"
)

// Kind defines the category of the error, determining how the system
// should react (e.g., retrying the operation or returning a 4xx/500).
type Kind string

const (
	// KindPersistance represents errors that will fail again if retried
	// without changing the input (e.g., Validation, Resource Conflicts).
	KindPersistance Kind = "PERSISTANCE"

	// KindTransient represents temporary failures that might succeed
	// upon retry (e.g., Network Timeouts, Database Deadlocks).
	KindTransient Kind = "TRANSIENT"

	// KindInternal represents unexpected system failures or bugs
	// (e.g., Nil Pointers, Database Syntax Errors).
	KindInternal Kind = "INTERNAL"
)

// AppError is the standardized error structure for the entire application.
// It wraps raw errors with business codes and metadata for consistent API responses.
type AppError struct {
	// Code is a machine-readable string (e.g., "DB_CONFLICT").
	Code string
	// Message is a human-readable explanation.
	Message string
	// Kind determines the retryability and HTTP mapping.
	Kind Kind
	// Details holds additional context for debugging or frontend hints.
	Details any
	// Err is the original underlying error (useful for stack traces).
	Err error
}

// Error implements the standard error interface.
// It prioritizes the Message, falling back to the Code if the message is empty.
func (e *AppError) Error() string {
	return fmt.Sprintf("%s", e.Message)
}

// Unwrap allows AppError to work with the standard errors.Is and errors.As functions.
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetail adds a key-value pair to the error's details map.
// If the current Details is not a map[string]any, it will be initialized as one.
func (e *AppError) WithDetail(key string, value any) *AppError {
	currentDetails, ok := e.Details.(map[string]any)
	if !ok || currentDetails == nil {
		currentDetails = make(map[string]any)
	}

	currentDetails[key] = value
	e.Details = currentDetails
	return e
}

// WithError wraps an existing error into the AppError context.
// It allows for chaining and provides a way to retain the original
// underlying error for logging or debugging purposes.
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// AddValidationError appends a structured validation error to the Details field.
// It treats Details as a slice of field-message pairs. If Details is not already
// a slice of maps, it initializes a new one.
func (e *AppError) AddValidationError(field, message string) *AppError {
	list, ok := e.Details.([]map[string]string)
	if !ok {
		list = []map[string]string{}
	}

	list = append(list, map[string]string{
		"field":   field,
		"message": message,
	})

	e.Details = list
	return e
}

// AddValidationErrors sets the validation details.
// It overwrites existing details to prevent duplicate error entries
// if validation is triggered multiple times in the same execution flow.
func (e *AppError) AddValidationErrors(errors []map[string]any) *AppError {
	// Directly assign instead of appending to avoid duplication
	e.Details = errors
	return e
}

// IsRetryable is a helper method to check if the error is a Transient failure.
func (e *AppError) IsRetryable() bool {
	return e.Kind == KindTransient
}

// ToMap converts the AppError to a map for logging purposes.
func (e *AppError) ToMap() map[string]any {
	return map[string]any{
		"code":         e.Code,
		"kind":         string(e.Kind),
		"is_retryable": e.IsRetryable(),
		"details":      e.Details,
		"raw_error":    e.Err,
	}
}

var (
	// statusRegistry is a thread-safe (assuming init-time registration) map
	// for module-specific error code to HTTP status mapping.
	statusRegistry = make(map[string]int)
)

// RegisterStatus allows modular registration of error codes to HTTP status codes.
// This should typically be called in a module's init() or during bootstrap.
func RegisterStatus(code string, status int) {
	statusRegistry[code] = status
}

func init() {
	// Initialize with default infrastructure/common mappings
	statusRegistry[CodeDbConnectionFailed] = 500
	statusRegistry[CodeDbTimeout] = 500
	statusRegistry[CodeDbDeadlock] = 500
	statusRegistry[CodeDbConstraint] = 500
	statusRegistry[CodeDbConflict] = 409
	statusRegistry[CodeInternalError] = 500

	statusRegistry[CodeMalformedRequest] = 400
	statusRegistry[CodeInvalidRequest] = 400
	statusRegistry[CodeValidation] = 400
	statusRegistry[CodeUnauthorized] = 401
	statusRegistry[CodeForbidden] = 403
	statusRegistry[CodeNotFound] = 404
	statusRegistry[CodeMethodNotAllowed] = 405
	statusRegistry[CodeNotAcceptable] = 406
	statusRegistry[CodeRequestTimeout] = 408
	statusRegistry[CodeConflict] = 409
	statusRegistry[CodeGone] = 410
	statusRegistry[CodeLengthRequired] = 411
	statusRegistry[CodePreconditionFailed] = 412
	statusRegistry[CodePayloadTooLarge] = 413
	statusRegistry[CodeURITooLong] = 414
	statusRegistry[CodeUnsupportedMediaType] = 415
	statusRegistry[CodeRangeNotSatisfiable] = 416
	statusRegistry[CodeExpectationFailed] = 417
	statusRegistry[CodeTeapot] = 418
	statusRegistry[CodeMisdirectedRequest] = 421
	statusRegistry[CodeUnprocessableEntity] = 422
	statusRegistry[CodeLocked] = 423
	statusRegistry[CodeFailedDependency] = 424
	statusRegistry[CodeTooEarly] = 425
	statusRegistry[CodeUpgradeRequired] = 426
	statusRegistry[CodePreconditionRequired] = 428
	statusRegistry[CodeTooManyRequests] = 429
	statusRegistry[CodeRequestHeaderFieldsTooLarge] = 431
	statusRegistry[CodeUnavailableForLegalReasons] = 451
	statusRegistry[CodeNetworkAuthenticationRequired] = 511
}

// GetHttpStatus resolves the appropriate HTTP status code for the error.
// It first attempts to match the 'Code' against the statusRegistry.
// If no match is found, it falls back to a status based on the 'Kind':
// - KindPersistance -> 400 (Bad Request)
// - KindTransient -> 503 (Service Unavailable)
// - KindInternal  -> 500 (Internal Server Error)
func (e *AppError) GetHttpStatus() int {
	// 1. Check direct code mapping in registry
	if status, exists := statusRegistry[e.Code]; exists {
		return status
	}

	// 2. Case-insensitive check (legacy support if needed)
	if status, exists := statusRegistry[strings.ToUpper(e.Code)]; exists {
		return status
	}

	// 3. Fallback to Kind
	switch e.Kind {
	case KindPersistance:
		return 400
	case KindTransient:
		return 503
	case KindInternal:
		return 500
	default:
		return 500
	}
}
