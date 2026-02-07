// Package database provides the core persistence infrastructure.
// It acts as a lightweight wrapper around GORM to manage connection lifecycles
// and context-aware sessions across the application.
package database

import (
	"context"
	"errors"
	"strings"
	"voyago/core-api/internal/infrastructure/config"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/pkg/apperror"
	baserepo "voyago/core-api/internal/pkg/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// Database defines the primary contract for database interactions.
// It extends repository.TransactionManager, allowing the implementation to orchestrate
// ACID-compliant transactions directly from the infrastructure layer.
type Database interface {
	// TransactionManager provides the Atomic(ctx, fn) method to wrap multiple
	// operations within a single database transaction.
	//
	// Use this in the UseCase layer to ensure that either all operations succeed
	// or none of them are committed.
	baserepo.TransactionManager

	// WithContext returns a shallow copy of the database connection
	// assigned to the provided context. It is "Transaction-Aware": if the context
	// contains an active transaction session (via Atomic), it returns that session.
	//
	// Use this for tracing, logging, and ensuring timeouts/cancellations are respected.
	WithContext(ctx context.Context) *gorm.DB

	// GetDB returns the direct GORM database instance.
	// Use this for global operations or when context scoping is not required.
	GetDB() *gorm.DB

	// Close gracefully shuts down the database connection pool.
	// This should be called during application shutdown to prevent memory leaks.
	Close() error
}

// NewDatabase is a factory function that initializes a new Database implementation.
// It configures the connection pool, logger bridge, and telemetry based on the provided config.
//
// Parameters:
//   - cfg: Database connection and pooling settings.
//   - log: Application logger to be used as a GORM log sink.
//   - trc: Tracer for injecting OpenTelemetry hooks into database queries.
func NewDatabase(cfg *config.DatabaseConfig, log logger.Logger, trc tracer.Tracer) Database {
	return NewGormDatabase(cfg, log, trc)
}

// --------- Error Mapping ---------

// MapDBError converts raw database errors into structured AppErrors.
// Note: gorm.ErrRecordNotFound is EXCLUDED from this mapper to allow
// repositories to handle "not found" cases individually (e.g., returning nil, nil).
func MapDBError(err error) error {
	if err == nil {
		return nil
	}

	// 1. Skip GORM Record Not Found
	// We return the original error so the Repo can decide what to do.
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 2. Handle System/Context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return apperror.NewTransient(apperror.CodeDbTimeout, "database operation timed out", err)
	}

	// 3. Driver specific mappers (Postgres)
	if pgErr := mapPgError(err); pgErr != nil {
		return pgErr
	}

	// 4. Default fallback for connection issues (string matching for dial errors)
	msg := err.Error()
	if strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "can't assign requested address") ||
		strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "broken pipe") {
		return apperror.NewTransient(apperror.CodeDbConnectionFailed, "database connection failed", err)
	}

	// 5. Ultimate Fallback
	return apperror.NewInternal(apperror.CodeInternalError, "unexpected database error", err)
}

// mapPgError handles Postgres specific errors using pgconn driver codes.
func mapPgError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}

	switch pgErr.Code {
	// --- Transient Errors (Retryable) ---

	// Connection issues
	case "08000", "08003", "08006", "57P01":
		return apperror.NewTransient(apperror.CodeDbConnectionFailed, "database connection failed", pgErr)

	// Deadlocks
	case "40P01":
		return apperror.NewTransient(apperror.CodeDbDeadlock, "database deadlock detected, please retry", pgErr)

	// Lock Timeouts
	case "55P03":
		return apperror.NewTransient(apperror.CodeDbTimeout, "database lock timeout", pgErr)

	// --- Permanent Errors (Client Side / Data Issue) ---

	// Unique Violation (e.g., duplicate email/code)
	case "23505": // Unique Violation
		return apperror.NewPersistance(apperror.CodeDbConflict, "duplicate data", pgErr).
			WithDetail("constraint", pgErr.ConstraintName).
			WithDetail("detail", pgErr.Detail)

	// Other Constraint Violations (Foreign Key, Not Null, etc.)
	case "23503", "23502", "23000":
		return apperror.NewPersistance(apperror.CodeDbConstraint, "database constraint violation: "+pgErr.Message, pgErr)

	// --- Internal Errors (Developer / Config Issue) ---

	// Syntax errors, Undefined column, Undefined table
	case "42703", "42601", "42P01":
		return apperror.NewInternal(apperror.CodeInternalError, "database schema or syntax error", pgErr)
	}

	return nil
}
