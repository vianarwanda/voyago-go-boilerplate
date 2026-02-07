/*
|------------------------------------------------------------------------------------
| USECASE ARCHITECTURAL STANDARDS & OBSERVABILITY MANIFESTO
|------------------------------------------------------------------------------------
|
| Every UseCase implementation MUST satisfy these high-level pillars to
| maintain system integrity and observability hygiene.
|
| [1. COMPLIANCE STANDARDS]
| - Interface-First: UseCases MUST be defined as interfaces to enable decoupled
|   communication and seamless unit testing (mocking).
| - Traceability: Maintain a continuous trace chain from entry to exit.
| - Observability: Ensure actions are searchable via business keys.
| - Validation: Enforce strict DTO validation before domain processing.
| - Atomicity: Guarantee data consistency via TransactionManager.
| - Side Effects: Trigger external events ONLY after a successful commit.
|
| [2. LOGGING OPERATIONAL SCOPE]
| - MINIMAL LOGS: Each execution logs "started" and either "completed"
|   (if successful) or "failed" (ONLY for internal UseCase logic errors).
| - ERROR BUBBLING: Downstream errors (Repo/Service) are bubbled up
|   without redundant logging to prevent aggregator pollution.
| - BUSINESS KEY: ONLY attach business_key to the "started" log to serve
|   as an 'Anchor Log'. Correlate subsequent logs via TraceID.
| - FIELD POLLUTION: Metadata enrichment only if it contains actual data.
|
| [3. STANDARD ERROR HANDLING]
| Operational steps when an error originates within this UseCase:
| 1. RECORD: Capture error details into the span (utils.RecordSpanError).
| 2. ENRICH: Wrap/Cast raw error into apperror.AppError (Code & Kind).
| 3. LOG:    Emit structured log ONLY if originating from UseCase logic.
| 4. BUBBLE: If the error originates from an underlying Repository/Service that has
|            already logged/traced the error, pass it directly to the caller to
|            maintain log hygiene and avoid redundancy.
| 5. HALT:   Return the standardized AppError immediately.
|
|------------------------------------------------------------------------------------
*/
package usecase

import (
	"context"
	"errors"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/modules/booking/entity"
	"voyago/core-api/internal/modules/booking/repository"
	"voyago/core-api/internal/pkg/apperror"
	baserepo "voyago/core-api/internal/pkg/repository"
	"voyago/core-api/internal/pkg/uid"
	"voyago/core-api/internal/pkg/utils"
)

type CreateBookingRepositories struct {
	BookingCmd repository.BookingCommandRepository
	BookingQry repository.BookingQueryRepository
}

// createBookingUseCase is the private implementation of CreateBookingUseCase.
// Use NewCreateBookingUseCase constructor to instantiate.
type createBookingUseCase struct {
	Log    logger.Logger
	Tracer tracer.Tracer
	Runner baserepo.TransactionManager
	Repo   CreateBookingRepositories
}

const (
	// useCaseName follows the "Layer:Component.Action" pattern.
	// This constant is used as the Span Name in tracing and 'action' field in logs,
	// enabling precise filtering across the entire observability stack.
	useCaseName = "usecase:booking.create"
)

// Compile-time check to ensure BookingRepository implements the required interface.
// This prevents runtime panics or dependency injection failures if the interface changes.
var _ CreateBookingUseCase = (*createBookingUseCase)(nil)

func NewCreateBookingUseCase(log logger.Logger, trc tracer.Tracer, runner baserepo.TransactionManager, repo CreateBookingRepositories) CreateBookingUseCase {
	return &createBookingUseCase{
		// WithField creates a sub-logger that automatically attaches the "action" context.
		Log:    log.WithField("action", useCaseName),
		Tracer: trc,
		Runner: runner,
		Repo:   repo,
	}
}

func (uc *createBookingUseCase) Execute(ctx context.Context, req *CreateBookingRequest) (*CreateBookingResponse, error) {
	// 1. START TRACING
	// StartSpan initializes a new trace span. The returned 'ctx' carries the span
	// information and must be passed downstream to maintain the trace chain.
	span, ctx := uc.Tracer.StartSpan(ctx, useCaseName)

	// Ensures the span is closed and flushed to the collector (e.g., OpenTelemetry)
	// when the function returns.
	defer span.Finish()

	log := uc.Log.WithContext(ctx).WithField("method", "Exec")

	// businessKey serves as a human-readable domain identifier (e.g., Booking ID, Transaction Code).
	// While TraceID links technical spans across services, Business Keys bridge the gap
	// between customer support tickets and system logs, allowing Ops teams to search
	// logs using real-world data provided by the user.
	//
	// Example: a domain-specific key from request request id, transaction id, etc.
	// businessKey := map[string]any{
	// 	"booking_id": req.BookingID,
	// }

	businessKey := map[string]any{
		"booking_code":  req.BookingCode,
		"count_details": len(req.Details),
	}

	// [LOGGING OPERATIONAL SCOPE: STARTED]
	// Operational Scope: We only enrich the "started" log with business_key if it contains data.
	// This serves as the 'Anchor Log' for searchability without polluting subsequent entries.
	if len(businessKey) > 0 {
		// We use .WithFields to attach multiple business identifiers at once,
		// marking the "Entry Point" log for this specific request.
		log.WithFields(map[string]any{
			"business_key": businessKey,
		}).Info("usecase started")
	} else {
		// Fallback log for requests without specific business identifiers.
		log.Info("usecase started")
	}

	// --- BUSINESS LOGIC & ERROR HANDLING ---
	// Follow [STANDARD ERROR HANDLING]:
	//
	// err := uc.Repo.SomeAction(ctx)
	// if err != nil {
	//    utils.RecordSpanError(span, err)
	//    return nil, err // BUBBLE UP: Let Repo handle the logging
	// }
	bookingID := uid.NewUUID()
	totalAmount := 0.0
	var details []entity.BookingDetail
	for _, d := range req.Details {
		detailID := uid.NewUUID()
		totalAmount += d.PricePerUnit * float64(d.Qty)
		details = append(details, entity.BookingDetail{
			ID:           detailID,
			ProductID:    d.ProductID,
			ProductName:  d.ProductName,
			Qty:          d.Qty,
			PricePerUnit: d.PricePerUnit,
			SubTotal:     d.SubTotal,
		})
	}

	e := entity.Booking{
		ID:            bookingID,
		BookingCode:   req.BookingCode,
		UserID:        req.UserID,
		TotalAmount:   req.TotalAmount,
		Status:        entity.BookingStatusPending,
		PaymentStatus: "UNPAID",
		Details:       details,
	}

	// --- PILLAR: DOMAIN VALIDATION ---
	// Execute domain-specific business rules defined within the entity.
	// This ensures the entity is in a valid state before persisting to the database.
	if err := e.Validate(); err != nil {
		// [STANDARD ERROR HANDLING]:
		// NOTE: For clean code, consider encapsulating this block into a private
		// helper (e.g., logAndTraceError) to maintain usecase readability.
		// Since this error originates in the UseCase (Domain Logic), we MUST log it here.

		// 1. RECORD: Capture the domain error in the distributed trace span.
		utils.RecordSpanError(span, err)

		// 2. LOG: Emit a structured log.
		// We cast to AppError (if possible) or use ToMap() to ensure
		// all metadata (Code, Kind) is visible in log aggregators.
		var appErr *apperror.AppError
		logFields := map[string]any{"error": err.Error()}
		if errors.As(err, &appErr) {
			if appErr.Err != nil {
				logFields["internal_detail"] = appErr.Err.Error()
			}
			logFields["retryable"] = appErr.IsRetryable()
		}
		log.WithFields(logFields).Warn("domain logic validation failed")

		// 3. HALT: Return the error immediately.
		// Since e.Validate() returns an AppError, the transport layer
		// will automatically resolve the correct HTTP status code.
		return nil, err
	}

	// --- PILLAR: BUSINESS RULE VALIDATION ---
	// Checking for uniqueness is a business rule that requires external context (DB).
	exists, err := uc.Repo.BookingQry.ExistsByBookingCode(ctx, e.BookingCode)
	if err != nil {
		// [STANDARD ERROR HANDLING]: BUBBLE UP
		// We only record the span error to ensure the trace reflects the failure.
		// Logging is already handled by the Repository/DB bridge.
		utils.RecordSpanError(span, err)
		return nil, err
	}

	if exists {
		// [STANDARD ERROR HANDLING]: Logged because it's a UseCase-level business violation.
		// We add an attribute to the span to mark this specific business failure.
		logAndTraceError(span, log, entity.ErrBookingCodeAlreadyExists, "domain logic validation failed", false)
		return nil, entity.ErrBookingCodeAlreadyExists
	}

	// --- PILLAR: PERSISTENCE (ATOMIC TRANSACTION) ---
	// We MUST wrap all command/write operations within an atomic transaction.
	// This guarantees ACID compliance—ensuring that the Booking header,
	// associated line items, and any state changes are committed as a single unit.
	// If any repository call fails, the entire transaction will roll back to prevent data corruption.
	errRunner := uc.Runner.Atomic(ctx, func(txCtx context.Context) error {
		if err := uc.Repo.BookingCmd.Create(txCtx, &e); err != nil {
			return err
		}
		return nil
	})
	if errRunner != nil {
		// [STANDARD ERROR HANDLING]: BUBBLE UP
		// We only record the span error to ensure the trace reflects the failure.
		// Logging is already handled by the Repository/DB bridge.
		utils.RecordSpanError(span, errRunner)
		return nil, errRunner
	}

	// [LOGGING OPERATIONAL SCOPE: COMPLETED]
	// Clean exit log: relying on TraceID for correlation with the "started" log.
	// No business_key here (already in 'started')
	log.Info("usecase completed")

	// Map Entity to Response DTO
	var detailsResponse []CreateBookingDetailResponse
	for _, d := range e.Details {
		detailsResponse = append(detailsResponse, CreateBookingDetailResponse{
			ProductID:    d.ProductID,
			ProductName:  d.ProductName,
			Qty:          d.Qty,
			PricePerUnit: d.PricePerUnit,
			SubTotal:     d.SubTotal,
		})
	}

	return &CreateBookingResponse{
		BookingID:   e.ID,
		BookingCode: e.BookingCode,
		UserID:      e.UserID,
		TotalAmount: e.TotalAmount,
		Details:     detailsResponse,
	}, nil
}

func logAndTraceError(span tracer.Span, log logger.Logger, err error, msg string, isCritical bool) {
	if err == nil {
		return
	}

	utils.RecordSpanError(span, err)

	var appErr *apperror.AppError
	logFields := map[string]any{"error": err.Error()}
	if errors.As(err, &appErr) {
		if appErr.Err != nil {
			logFields["internal_detail"] = appErr.Err.Error()
		}
		logFields["retryable"] = appErr.IsRetryable()
	}
	l := log.WithFields(logFields)
	if isCritical {
		l.Error(msg)
	} else {
		l.Warn(msg)
	}
}
