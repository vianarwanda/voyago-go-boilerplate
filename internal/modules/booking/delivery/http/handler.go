/*
|------------------------------------------------------------------------------------
| HTTP HANDLER ARCHITECTURAL STANDARDS & OBSERVABILITY MANIFESTO
|------------------------------------------------------------------------------------
|
| The Handler layer serves as the system's "Front Gate". It is responsible for
| request orchestration, DTO enforcement, and response normalization.
|
| [1. THE SINGLE LOG RULE]
| - Every handler execution MUST emit exactly ONE "Anchor Log" (request received).
| - This log must be enriched with 'business_key' (if available) to bridge the
|   gap between business domains and technical traces.
|
| [2. ZERO POST-ENTRY LOGGING]
| - Once the request is handed over to the UseCase, the Handler MUST NOT emit
|   any further logs (success or failure).
| - Observability for the rest of the execution is handled by the UseCase
|   and Repository layers via TraceID correlation.
|
| [3. LEAN ORCHESTRATION]
| - Validation: Enforce payload integrity using DTO tags before execution.
| - Parsing: Handle malformed requests and immediately return AppError.
| - Bubbling: All errors returned by the UseCase are bubbled up directly to
|   the Global Error Handler to maintain log hygiene.
|
| [4. RESPONSE NORMALIZATION]
| - Always use the standardized 'response' package to ensure consistent
|   API contracts across all modules.
|
|------------------------------------------------------------------------------------
*/
package http

import (
	"voyago/core-api/internal/infrastructure/config"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/validator"
	"voyago/core-api/internal/modules/booking/usecase"
	"voyago/core-api/internal/pkg/apperror"
	"voyago/core-api/internal/pkg/response"

	"github.com/gofiber/fiber/v2"
)

const (
	// handlerName follows the "Layer:Component.Action" pattern.
	// This constant is used as the Span Name in tracing and 'action' field in logs,
	// enabling precise filtering across the entire observability stack.
	handlerName = "http:handler.booking"
)

type HandlerUseCases struct {
	CreateBookingUseCase usecase.CreateBookingUseCase
}

type Handler struct {
	Cfg *config.Config
	Log logger.Logger
	Val validator.Validator
	Uc  HandlerUseCases
}

func NewHandler(cfg *config.Config, log logger.Logger, validator validator.Validator, useCases HandlerUseCases) *Handler {
	return &Handler{
		Cfg: cfg,
		Log: log,
		Val: validator,
		Uc:  useCases,
	}
}

func (h *Handler) CreateBooking(c *fiber.Ctx) error {
	// We use c.UserContext() which has been enriched by the Telemetrist middlewares.
	// There's no need to start a new span here unless we have complex logic
	// within the handler itself. The Telemetrist middlewares span will act as the parent
	// for all subsequent UseCase and Repository spans.
	ctx := c.UserContext()

	// 1. INITIALIZE CONTEXTUAL LOGGER
	// .WithContext(ctx) is vital: it extracts the TraceID from the span created
	// by the Telemetrist middleware, linking this log to the entire trace.
	log := h.Log.WithContext(ctx).WithField("method", "CreateBooking")

	// 2. PARSE REQUEST BODY
	request := new(usecase.CreateBookingRequest)
	if err := c.BodyParser(request); err != nil {
		// [LOG HYGIENE]: We don't log here. The error is bubbled to the Global Error Handler,
		// which will emit a single error log with full context and TraceID.
		return apperror.ErrCodeMalformedRequest.WithError(err)
	}

	// 3. VALIDATE REQUEST DTO
	// Standardizing validation at the entry point ensures UseCase only receives clean data.
	if err := h.Val.Validate(request); err != nil {
		// [LOG HYGIENE]: Bubble up directly. Redundant logging at this stage would
		// only clutter the aggregator since the failure is already captured in the response.
		return apperror.ErrCodeInvalidRequest.WithError(err).AddValidationErrors(h.Val.ToDetails(err))
	}

	// 4. THE ANCHOR LOG & BUSINESS CORRELATION
	// The 'businessKey' serves as a human-readable bridge (e.g., Booking Code, User ID).
	// While TraceID links technical spans, Business Keys link technical logs to
	// real-world customer support tickets.
	//
	// By logging this ONCE at the entry point, we create an 'Anchor Log'.
	// Ops teams can search by 'booking_code' to find the TraceID, then use
	// that TraceID to uncover every subsequent event in the entire request lifecycle.
	businessKey := map[string]any{
		"booking_code": request.BookingCode,
	}

	// [LOGGING OPERATIONAL SCOPE: ENTRY]
	// To prevent "Field Pollution" in our log aggregator,
	// we ONLY enrich this entry log with the business_key.
	// Subsequent logs (from UseCase/Repo) will be linked via TraceID automatically.
	if len(businessKey) > 0 {
		log.WithFields(map[string]any{
			"business_key": businessKey,
		}).Info("request received")
	} else {
		log.Info("request received")
	}

	// --- HANDOVER TO DOMAIN LAYER (THE ZERO-LOG HANDOVER) ---
	// From this point forward, the handler delegates all operational observability
	// to the UseCase. This boundary ensures that domain-specific errors and
	// performance metrics are captured at the source of truth.
	//
	// We strictly avoid logging before or after this call to maintain log hygiene.
	createBooking, err := h.Uc.CreateBookingUseCase.Execute(ctx, request)
	if err != nil {
		// [ERROR BUBBLING STRATEGY]
		// If an error occurs, it has already been:
		// 1. Traced (RecordSpanError) in UseCase or Repository.
		// 2. Logged with full business context (is_retryable, app_code, etc.).
		//
		// We bubble the error to Fiber's Global Error Handler to ensure a
		// standardized JSON response and to prevent redundant log entries
		// that would otherwise clutter our observability stack.
		return err
	}

	// [SUCCESS RESPONSE NORMALIZATION]
	// No success log is emitted here. The successful execution trace and the
	// 'usecase completed' log are sufficient for audit and monitoring.
	// We normalize the output using a consistent API response contract.
	return response.NewHttp(c).Created(response.Http{
		Message: "Booking created successfully",
		Data:    createBooking, // Use the processed entity from UseCase
	})
}
