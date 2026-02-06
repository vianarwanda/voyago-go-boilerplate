package database

import (
	"context"

	"gorm.io/gorm"
)

// ErrorMapper is a function type for mapping database errors to application errors.
type ErrorMapper func(error) error

// GormBaseRepository provides a generic implementation of common persistence operations (CRUD)
// using GORM as the underlying ORM.
//
// This is an infrastructure-specific implementation. Domain repositories should
// expose their own interfaces in contract.go and use this as an embedded helper.
//
// Type T represents the domain entity or database model (e.g., Booking, User).
type GormBaseRepository[T any] struct {
	// DB is the database provider. It should be injected during the initialization
	// of the domain-specific repository.
	DB Database

	// ErrorMapper is the function that maps database errors to application errors.
	ErrorMapper ErrorMapper
}

// getDB is an internal helper to resolve the current database session.
// It delegates the resolution to the Database interface, ensuring that the returned *gorm.DB
// is correctly scoped with the current context and any active transaction.
func (r *GormBaseRepository[T]) getDB(ctx context.Context) *gorm.DB {
	return r.DB.WithContext(ctx)
}

// mapErr passes errors through the ErrorMapper if configured.
func (r *GormBaseRepository[T]) mapErr(err error) error {
	if err == nil || r.ErrorMapper == nil {
		return err
	}
	return r.ErrorMapper(err)
}

// Create inserts a new record of type T into the database.
//
// Context Awareness:
// It uses the provided context to handle timeouts and cancellations. If the context
// contains a transaction session (e.g., from an Atomic block), GORM will
// automatically execute this operation within that transaction.
//
// Error Handling:
// All database errors are passed through the ErrorMapper to ensure the returned
// error is a structured apperror.AppError (e.g., mapping a unique violation
// to apperror.CodeDbConflict).
//
// Parameters:
//   - ctx: The execution context.
//   - entity: A pointer to the model instance to be persisted.
func (r *GormBaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.mapErr(r.getDB(ctx).Create(entity).Error)
}

// Update performs a full update of the entity using GORM's Save method.
// WARNING: Save updates all columns. If you pass a partial struct, fields not set
// will be overwritten with zero values in the database.
//
// For partial updates, implement a custom method using .Updates() in the domain repository.
func (r *GormBaseRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.mapErr(r.getDB(ctx).Save(entity).Error)
}

// Delete removes the record of type T from the database.
// Performs a Soft Delete if the model T includes gorm.DeletedAt; otherwise, it performs a Hard Delete.
func (r *GormBaseRepository[T]) Delete(ctx context.Context, entity *T) error {
	return r.mapErr(r.getDB(ctx).Delete(entity).Error)
}
