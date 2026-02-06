package baserepo

import (
	"context"

	"gorm.io/gorm"
)

type ErrorMapper func(error) error

// DBProvider defines the contract required by BaseRepository to obtain a database session.
// It is designed to decouple the repository layer from the concrete infrastructure implementation,
// preventing circular dependencies.
//
// Any struct that implements WithContext(context.Context) *gorm.DB (like your gormDatabase)
// will automatically satisfy this interface.
type DBProvider interface {
	// WithContext returns a GORM instance scoped to the provided context.
	// Implementation should ideally check for active transactions within the context.
	WithContext(ctx context.Context) *gorm.DB
}

// BaseRepository provides a generic implementation of common persistence operations (CRUD).
// It ensures that all operations are context-aware and automatically participate in
// active transactions if managed by a TransactionManager/Atomic runner.
//
// Type T represents the domain entity or database model (e.g., Booking, User).
type BaseRepository[T any] struct {
	// DB is the database provider. It should be injected during the initialization
	// of the domain-specific repository.
	DB DBProvider

	// ErrorMapper is the function that maps database errors to application errors.
	ErrorMapper ErrorMapper
}

// getDB is an internal helper to resolve the current database session.
// It delegates the resolution to the DBProvider, ensuring that the returned *gorm.DB
// is correctly scoped with the current context and any active transaction.
func (r *BaseRepository[T]) getDB(ctx context.Context) *gorm.DB {
	return r.DB.WithContext(ctx)
}

// Create inserts a new record of type T into the database.
// If executed within an Atomic block, it will automatically participate in the transaction.
func (r *BaseRepository[T]) mapErr(err error) error {
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
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.mapErr(r.getDB(ctx).Create(entity).Error)
}

// Update performs a full update of the entity using GORM's Save method.
// WARNING: Save updates all columns. If you pass a partial struct, fields not set
// will be overwritten with zero values in the database.
//
// For partial updates, implement a custom method using .Updates() in the domain repository.
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.mapErr(r.getDB(ctx).Save(entity).Error)
}

// Delete removes the record of type T from the database.
// Performs a Soft Delete if the model T includes gorm.DeletedAt; otherwise, it performs a Hard Delete.
func (r *BaseRepository[T]) Delete(ctx context.Context, entity *T) error {
	return r.mapErr(r.getDB(ctx).Delete(entity).Error)
}
