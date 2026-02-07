package response

import "github.com/gofiber/fiber/v2"

// ResponseApi defines the standardized JSON structure for all API responses.
// It bridges the gap between the server and client by providing consistent
// metadata, domain data, and observability IDs (TraceID).
type ResponseApi struct {
	// Success indicates if the operation was completed without business or technical errors.
	Success bool `json:"success"`

	// Message is a human-readable summary of the response.
	Message string `json:"message"`

	// Data holds the primary payload of the response (e.g., entity, list of objects).
	Data any `json:"data,omitempty"`

	// Meta holds additional metadata such as pagination (e.g., page, total_rows).
	Meta any `json:"meta,omitempty"`

	// ErrorCode is a unique application-specific string used for programmatic error handling.
	ErrorCode string `json:"error_code,omitempty"`

	// IsRetryable hints to the client whether repeating the same request might eventually succeed.
	IsRetryable bool `json:"is_retryable,omitempty"`

	// Errors contains granular validation details or field-specific error messages.
	Errors any `json:"errors,omitempty"`

	// TraceID is the unique identifier for the request's lifecycle.
	// Clients should provide this ID when reporting issues to support teams.
	TraceID string `json:"trace_id,omitempty"`
}

// responseBuilder handles the construction of API responses.
type responseBuilder struct {
	ctx *fiber.Ctx
}

// NewResponseApi initializes a new response builder.
// It captures the context once to avoid redundant passing in subsequent method calls.
func NewResponseApi(c *fiber.Ctx) *responseBuilder {
	return &responseBuilder{ctx: c}
}

// OK sends a standardized successful response (HTTP 200).
func (b *responseBuilder) OK(response ResponseApi) error {
	response.Success = true
	response.TraceID, _ = b.ctx.Locals("trace_id").(string)
	return b.ctx.Status(fiber.StatusOK).JSON(response)
}

// Created sends a standardized resource creation response (HTTP 201).
// Use this when a resource has been successfully created.
func (b *responseBuilder) Created(response ResponseApi) error {
	response.Success = true
	response.TraceID, _ = b.ctx.Locals("trace_id").(string)
	return b.ctx.Status(fiber.StatusCreated).JSON(response)
}

// Accepted sends a standardized response for asynchronous processing (HTTP 202).
// Use this when a request is valid and queued.
func (b *responseBuilder) Accepted(response ResponseApi) error {
	response.Success = true
	response.TraceID, _ = b.ctx.Locals("trace_id").(string)
	return b.ctx.Status(fiber.StatusAccepted).JSON(response)
}

// NoContent sends a successful response with no body (HTTP 204).
// Use this when an action is successful but there is no data to return.
func (b *responseBuilder) NoContent() error {
	return b.ctx.SendStatus(fiber.StatusNoContent)
}
