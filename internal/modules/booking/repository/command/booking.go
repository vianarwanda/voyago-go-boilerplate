package command

import (
	database "voyago/core-api/internal/infrastructure/db"
	"voyago/core-api/internal/modules/booking/entity"
	"voyago/core-api/internal/modules/booking/repository"
	baserepo "voyago/core-api/internal/pkg/repository"
)

// BookingRepository handles persistence logic for the Booking module.
//
// It embeds repository.BaseRepository to inherit standard CRUD operations
// (Create, Update, Delete) for the entity.Booking model.
//
// Technical Note:
// We use a Pointer Embedding (*BaseRepository) to ensure that all methods
// with pointer receivers in the BaseRepository are correctly promoted
// and satisfy the BookingCommandRepository interface.
type BookingRepository struct {
	*baserepo.BaseRepository[entity.Booking]
}

// Compile-time check to ensure BookingRepository implements the required interface.
// This prevents runtime panics or dependency injection failures if the interface changes.
var _ repository.BookingCommandRepository = (*BookingRepository)(nil)

func NewBookingRepository(db database.Database) repository.BookingCommandRepository {
	return &BookingRepository{
		BaseRepository: &baserepo.BaseRepository[entity.Booking]{
			DB:          db,
			ErrorMapper: database.MapDBError,
		},
	}
}
