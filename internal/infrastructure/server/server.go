// Package server provides the web server infrastructure using the Fiber framework.
package server

import (
	"context"
	"fmt"
	"time"
	"voyago/core-api/internal/infrastructure/config"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/pkg/apperror"
	"voyago/core-api/internal/pkg/response"

	"github.com/gofiber/fiber/v2"
)

// Server represents the HTTP server wrapper for the Fiber application.
// It encapsulates the framework's engine and provides lifecycle management.
type Server struct {
	// App is the underlying Fiber instance.
	// Use this to register routes and middlewares.
	App *fiber.App
	cfg *config.Config
	log logger.Logger
}

// NewServer initializes a new Fiber application with settings from the config.
// It sets up default configurations like AppName and Preforking.
//
// Parameters:
//   - cfg: Application configuration (ports, timeouts, prefork settings).
//   - log: Logger instance for infrastructure-level logging.
func NewServer(
	cfg *config.Config,
	log logger.Logger,
) *Server {
	readTimeout := 10 * time.Second
	if cfg.Http.ReadTimeout != 0 {
		readTimeout = time.Duration(cfg.Http.ReadTimeout) * time.Second
	}

	writeTimeout := 10 * time.Second
	if cfg.Http.WriteTimeout != 0 {
		writeTimeout = time.Duration(cfg.Http.WriteTimeout) * time.Second
	}

	idleTimeout := 30 * time.Second
	if cfg.Http.IdleTimeout != 0 {
		idleTimeout = time.Duration(cfg.Http.IdleTimeout) * time.Second
	}

	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		Prefork:      cfg.Http.Prefork,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		ErrorHandler: errorHdlr,
	})

	return &Server{
		App: app,
		cfg: cfg,
		log: log.WithField("component", "app"),
	}
}

// Start launches the HTTP server on the port defined in the configuration.
// It returns an error if the server fails to bind to the address.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Http.Port)
	s.log.Info(fmt.Sprintf("Server [%s] started and listening on %s", s.cfg.App.Name, addr))
	return s.App.Listen(addr)
}

// Stop gracefully shuts down the server without interrupting active connections.
// It accepts a context for timeout management (e.g., wait 5s before forcing exit).
func (s *Server) Stop(ctx context.Context) error {
	s.log.Warn(fmt.Sprintf("Shutting down server [%s] gracefully...", s.cfg.App.Name))
	return s.App.ShutdownWithContext(ctx)
}

func errorHdlr(c *fiber.Ctx, err error) error {
	// Default response
	code := fiber.ErrInternalServerError.Code
	message := err.Error()
	errCode := fmt.Sprintf("ERR_%d", fiber.ErrInternalServerError.Code)
	var details any
	var isRetryable bool

	// check if it appError
	if e, ok := err.(*apperror.AppError); ok {
		code = e.GetHttpStatus()
		message = e.Message
		errCode = e.Code
		details = e.Details
		isRetryable = e.IsRetryable()
	} else if e, ok := err.(*fiber.Error); ok {
		// Error from Fiber itself (e.g. 404 route not found)
		code = e.Code
		message = e.Message
		errCode = fmt.Sprintf("ERR_%d", e.Code)
	}

	// get trace id from locals
	traceID, _ := c.Locals("trace_id").(string)

	return c.Status(code).JSON(response.ResponseApi{
		Success:     false,
		Message:     message,
		ErrorCode:   errCode,
		Errors:      details,
		TraceID:     traceID,
		IsRetryable: isRetryable,
	})
}
