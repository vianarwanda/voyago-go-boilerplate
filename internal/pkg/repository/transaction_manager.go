package baserepo

import "context"

type TransactionManager interface {
	// Atomic executes the provided function within a database transaction.
	// STANDARD: Every "Command" (Insert/Update/Delete) that involves
	// multiple tables or requires data consistency MUST use this method
	// If the function returns an error, the transaction is automatically rolled back.
	// Otherwise, it is committed.
	Atomic(ctx context.Context, fn func(ctx context.Context) error) error
}
