package middleware

import (
	"voyago/core-api/internal/infrastructure/ctxkey"

	"github.com/gofiber/fiber/v2"
)

// RequestID middleware manages the correlation identifier for each incoming request.
// It synchronizes the ID across the request headers and the Go context to ensure
// consistent traceability from the entry point down to the persistence layer.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Attempt to extract the Request ID from incoming headers.
		reqId := c.Get(fiber.HeaderXRequestID)
		// In scenarios where the header is absent, we default to "unknown".
		// This explicitly indicates in the logs that the request did not originate
		// with a predefined correlation ID from the upstream gateway or client.
		if reqId == "" {
			reqId = "unknown"
		}

		// Reflect the Request ID back to the client in the response headers.
		// This allows clients to provide a reference ID when reporting issues to support teams.
		c.Set(fiber.HeaderXRequestID, reqId)

		// Context Propagation:
		// We store the Request ID into the Fiber UserContext using a dedicated context key.
		// This allows the logger and other downstream components to extract the ID
		// and maintain a unified audit trail for the entire execution lifecycle.
		ctx := ctxkey.SetRequestID(c.UserContext(), reqId)
		c.SetUserContext(ctx)

		return c.Next()
	}
}
