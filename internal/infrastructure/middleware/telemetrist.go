package middleware

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/metrics"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/pkg/apperror"
	"voyago/core-api/internal/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type Telemetrist struct {
	LogProvider     logger.Logger
	TracerProvider  tracer.Tracer
	MetricsProvider metrics.Metrics
}

func NewTelemetrist(
	log logger.Logger,
	trc tracer.Tracer,
	metrics metrics.Metrics,
) *Telemetrist {
	return &Telemetrist{
		LogProvider:     log,
		TracerProvider:  trc,
		MetricsProvider: metrics,
	}
}

// TraceMiddleware initiates the request span.
// It must run first so other middlewares can attach data to this span.
func (m *Telemetrist) HandleTrace() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := m.TracerProvider.StartSpan(c.UserContext(), fmt.Sprintf("HTTP %s %s", c.Method(), c.Path()))
		defer span.Finish()

		tID, _, _ := m.TracerProvider.ExtractTraceInfo(ctx)
		c.Locals("trace_id", tID)
		c.Set("X-Trace-Id", tID)

		c.SetUserContext(ctx)
		err := c.Next()

		var routePath string
		if r := c.Route(); r != nil {
			routePath = r.Path
		}

		if routePath == "" {
			routePath = c.Path()
		}

		span.SetOperationName(fmt.Sprintf("HTTP %s %s", c.Method(), routePath))

		span.SetTag("http.method", c.Method())
		span.SetTag("http.path", c.Path())
		span.SetTag("http.route", routePath)

		statusCode := c.Response().StatusCode()
		if appErr, ok := err.(*apperror.AppError); ok {
			statusCode = appErr.GetHttpStatus()
		}
		span.SetTag("http.status_code", statusCode)

		if err != nil || statusCode >= 400 {
			span.SetTag("error", true)
			if err != nil {
				span.SetTag("error.message", err.Error())
			}
		}
		return err
	}
}

// MetricsMiddleware records latency and throughput.
func (m *Telemetrist) HandleMetrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		method := string(c.Request().Header.Method())
		path := string(c.Path())

		duration := time.Since(start).Seconds()

		var routePath string
		if r := c.Route(); r != nil {
			routePath = r.Path
		}

		if routePath == "" {
			routePath = path
		}

		statusCode := c.Response().StatusCode()
		if appErr, ok := err.(*apperror.AppError); ok {
			statusCode = appErr.GetHttpStatus()
		}

		m.MetricsProvider.RecordHTTP(method, path, routePath, statusCode, duration)

		return err
	}
}

// LogMiddleware provides the final audit trail of the request.
func (m *Telemetrist) HandleLog() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		ctx := c.UserContext()

		err := c.Next()

		var routePath string
		if r := c.Route(); r != nil {
			routePath = r.Path
		}

		if routePath == "" {
			routePath = c.Path()
		}

		latency := float64(time.Since(start).Nanoseconds()) / 1e6

		if err != nil {
			if err := c.App().ErrorHandler(c, err); err != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		statusCode := c.Response().StatusCode()
		if err != nil && statusCode == fiber.StatusOK {
			statusCode = fiber.StatusInternalServerError
		}

		reqContentType := string(c.Request().Header.ContentType())
		resContentType := string(c.Response().Header.ContentType())

		logEntry := m.LogProvider.WithContext(ctx).WithFields(map[string]any{
			"component": "telemetry.middleware",

			"transport":  "http",
			"method":     c.Method(),
			"path":       c.Path(),
			"route":      routePath,
			"status":     statusCode,
			"latency_ms": latency,
			"ip":         c.IP(),
			"trace_id":   c.Locals("trace_id"),

			"request": map[string]any{
				"headers": utils.MaskHttpHeaders(c.GetReqHeaders()),
				"query":   utils.MaskSensitive(c.Queries()),
				"params":  utils.MaskSensitive(c.AllParams()),
				"body":    m.parseBody(c.Body(), reqContentType),
			},

			"response": map[string]any{
				"body": m.parseBody(c.Response().Body(), resContentType),
			},
		})

		if err != nil || statusCode >= 500 {
			logEntry.WithField("error", err.Error()).Error("http request completed with error")
		} else if statusCode >= 400 {
			logEntry.Warn("http request completed with client error")
		} else {
			logEntry.Info("http request completed")
		}

		return nil
	}
}

func (m *Telemetrist) getFiberErrStatusCode(err error) int {
	statusCode := 500
	if fiberErr, ok := err.(*fiber.Error); ok {
		statusCode = fiberErr.Code
	} else {
		statusCode = fiber.StatusInternalServerError
	}
	return statusCode
}

// ParseBody processes raw bytes from request or response, enforces size limits,
// and applies sensitivity masking if the content type is JSON.
func (m *Telemetrist) parseBody(body []byte, contentType string) any {
	if len(body) == 0 {
		return nil
	}

	// Only attempt to parse and mask if it's JSON
	if !strings.Contains(strings.ToLower(contentType), "application/json") {
		return "[non-json or binary content]"
	}

	// Enforce limit to prevent log bloat and high memory usage during unmarshaling
	const limit = 2 * 1024 // 2KB
	if len(body) > limit {
		return fmt.Sprintf("[body too large: %d bytes]", len(body))
	}

	var obj any
	if err := json.Unmarshal(body, &obj); err != nil {
		return "[parse error: invalid json]"
	}

	return utils.MaskSensitive(obj)
}
