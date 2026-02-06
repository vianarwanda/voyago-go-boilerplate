package repository

import (
	"context"
	"voyago/core-api/internal/modules/booking/entity"
)

// -------- Repository Command --------

type BookingCommandRepository interface {
	Create(ctx context.Context, booking *entity.Booking) error
	Update(ctx context.Context, booking *entity.Booking) error
	Delete(ctx context.Context, booking *entity.Booking) error
}

// -------- Repository Query --------

type BookingQueryRepository interface {
	ExistsByBookingCode(ctx context.Context, code string) (bool, error)
	FindByID(ctx context.Context, id string) (*entity.Booking, error)
	FindByCode(ctx context.Context, code string) (*entity.Booking, error)
}
