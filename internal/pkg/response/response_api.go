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

// NewResponseApi initializes a new response object and automatically extracts
// the TraceID from the context (populated by telemetries middleware).
func NewResponseApi(c *fiber.Ctx) *ResponseApi {
	// Extract TraceID from locals.
	// This ensures every response, whether success or error, carries its technical identity.
	traceID, _ := c.Locals("trace_id").(string)
	return &ResponseApi{
		TraceID: traceID,
	}
}

// OK sends a standardized successful response (HTTP 200).
// It populates the common fields and ensures the 'Success' flag is set to true.
func (r *ResponseApi) OK(c *fiber.Ctx, response ResponseApi) error {
	r.Success = true
	r.Message = response.Message
	r.Data = response.Data
	r.Meta = response.Meta
	return c.Status(fiber.StatusOK).JSON(r)
}

// Created sends a standardized resource creation response (HTTP 201).
// Use this when a resource has been successfully created (e.g., POST /bookings).
//
// Why: It tells the client that the resource is new and specific headers (like Location) might be relevant.
func (r *ResponseApi) Created(c *fiber.Ctx, response ResponseApi) error {
	r.Success = true
	r.Message = response.Message
	r.Data = response.Data
	r.Meta = response.Meta
	return c.Status(fiber.StatusCreated).JSON(r)
}

// Accepted sends a standardized response for asynchronous processing (HTTP 202).
// Use this when a request is valid and queued but processing is not yet complete (e.g., Generate PDF).
//
// Why: It prevents client timeouts on long-running tasks and indicates that the request is "in progress".
func (r *ResponseApi) Accepted(c *fiber.Ctx, response ResponseApi) error {
	r.Success = true
	r.Message = response.Message
	r.Data = response.Data
	r.Meta = response.Meta
	return c.Status(fiber.StatusAccepted).JSON(r)
}

// NoContent sends a successful response with no body (HTTP 204).
// Use this when an action is successful but there is no data to return (e.g., DELETE /bookings/1).
//
// Why: It saves bandwidth and clearly signals "done, nothing to see here".
func (r *ResponseApi) NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
