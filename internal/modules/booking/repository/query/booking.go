package query

import (
	"context"
	"errors"
	database "voyago/core-api/internal/infrastructure/db"
	"voyago/core-api/internal/modules/booking/entity"
	"voyago/core-api/internal/modules/booking/repository"

	"gorm.io/gorm"
)

type BookingRepository struct {
	DB database.Database
}

// Compile-time check to ensure BookingRepository implements the required interface.
// This prevents runtime panics or dependency injection failures if the interface changes.
var _ repository.BookingQueryRepository = (*BookingRepository)(nil)

func NewBookingRepository(db database.Database) repository.BookingQueryRepository {
	return &BookingRepository{
		DB: db,
	}
}

func (r *BookingRepository) ExistsByBookingCode(ctx context.Context, code string) (bool, error) {
	if code == "" {
		return false, nil
	}
	var count int64
	if err := r.DB.WithContext(ctx).
		Model(&entity.Booking{}).
		Where("booking_code = ?", code).
		Limit(1).
		Count(&count).
		Error; err != nil {
		return false, database.MapDBError(err)
	}
	return count > 0, nil
}

func (r *BookingRepository) FindByCode(ctx context.Context, code string) (*entity.Booking, error) {
	if code == "" {
		return nil, nil
	}
	var booking entity.Booking
	err := r.DB.WithContext(ctx).
		Model(&entity.Booking{}).
		Select(
			"id",
			"booking_code",
			"user_id",
			"total_amount",
			"status",
			"payment_status",
			"created_at",
			"updated_at",
		).
		Where("booking_code = ?", code).
		First(&booking).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, database.MapDBError(err)
	}

	return &booking, nil
}
func (r *BookingRepository) FindByID(ctx context.Context, id string) (*entity.Booking, error) {
	if id == "" {
		return nil, nil
	}
	var booking entity.Booking
	err := r.DB.WithContext(ctx).
		Model(&entity.Booking{}).
		Select(
			"id",
			"booking_code",
			"user_id",
			"total_amount",
			"status",
			"payment_status",
			"created_at",
			"updated_at",
		).
		Where("id = ?", id).
		Preload("BookingDetails", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "booking_id", "product_id", "product_name", "qty", "price_per_unit", "sub_total")
		}).
		First(&booking).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, database.MapDBError(err)
	}

	return &booking, nil
}
