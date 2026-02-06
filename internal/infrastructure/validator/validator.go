package validator

// ValidationError represents a single field validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Validator defines the contract for request data validation.
type Validator interface {
	// Validate performs a structural validation on the provided input.
	// It returns nil if the input is valid, or an error if validation fails.
	Validate(i any) error

	// ToCustomError converts a standard validator error into a slice of ValidationError structs.
	// Useful for internal logic where you need to process errors as objects.
	ToCustomError(err error) []ValidationError

	// ToMap converts validation errors into a map where keys are field names.
	// Primarily used for simpler, legacy-style error responses.
	ToMap(err error) map[string]any

	// ToDetails converts validation errors into a slice of key-value maps.
	// Designed for API responses to provide "field" and "message" keys for Front-End consumption.
	ToDetails(err error) []map[string]any
}
