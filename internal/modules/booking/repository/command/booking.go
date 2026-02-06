/*
|------------------------------------------------------------------------------------
| REPOSITORY ARCHITECTURAL STANDARDS & PERSISTENCE MANIFESTO
|------------------------------------------------------------------------------------
|
| The Repository layer is responsible for low-level data persistence. It acts as
| a bridge between the Domain Entities and the Physical Database.
|
| [1. ERROR MAPPING & TRANSLATION]
| - Repositories MUST NOT return raw database errors (e.g., gorm.ErrRecordNotFound).
| - All errors must be passed through an ErrorMapper to be translated into
|   standardized apperror.AppError (e.g., ErrCodeNotFound).
|
| [2. AUTOMATIC OBSERVABILITY]
| - Persistence operations are automatically traced via GORM Callbacks/Middleware.
| - REPEAT LOGGING PROHIBITION: Do not log errors here if the Database Bridge
|   (GORM Logger) already emits structured logs. This maintains "Log Hygiene".
|
| [3. ATOMICITY COMPLIANCE]
| - Commands MUST respect the 'ctx' (context) to ensure they participate in
|   active transactions managed by the TransactionManager (Runner).
|
| [4. GENERIC CONSTRAINTS]
| - Use BaseRepository embedding for standard CRUD to reduce boilerplate, but
|   override methods if specific business logic or optimization is required.
|
|------------------------------------------------------------------------------------
*/
package command

import (
	database "voyago/core-api/internal/infrastructure/db"
	"voyago/core-api/internal/modules/booking/entity"
	"voyago/core-api/internal/modules/booking/repository"
	baserepo "voyago/core-api/internal/pkg/repository"
)

// bookingRepository provides the concrete implementation of BookingCommandRepository.
// By embedding BaseRepository, it gains robust CRUD capabilities while maintaining
// strict type safety for the entity.Booking model.
type bookingRepository struct {
	// We use Pointer Embedding to inherit method sets and ensure the repository
	// behaves as a reference type across the application.
	*baserepo.BaseRepository[entity.Booking]
}

// [INTERFACE COMPLIANCE CHECK]
// This static check ensures that if the BookingCommandRepository interface changes,
// this file will fail to compile, preventing runtime dependency injection errors.
var _ repository.BookingCommandRepository = (*bookingRepository)(nil)

// NewBookingRepository initializes the repository with a Database connection
// and a centralized ErrorMapper.
//
// Technical Note: The ErrorMapper is crucial for translating SQL-specific
// errors into Domain-friendly AppErrors before they reach the UseCase.
func NewBookingRepository(db database.Database) repository.BookingCommandRepository {
	return &bookingRepository{
		BaseRepository: &baserepo.BaseRepository[entity.Booking]{
			DB:          db,
			ErrorMapper: database.MapDBError,
		},
	}
}
