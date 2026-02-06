/*
|------------------------------------------------------------------------------------
| REPOSITORY ARCHITECTURAL STANDARDS & QUERY OPTIMIZATION MANIFESTO
|------------------------------------------------------------------------------------
|
| The Query Repository is dedicated to data retrieval. It follows the R-side of
| CQRS, focusing on performance, filtering, and non-mutating operations.
|
| [1. SELECTIVE RETRIEVAL (NO SELECT *)]
| - Always specify required fields in .Select(). Avoid 'SELECT *' to minimize
|   database I/O and prevent sensitive data leakage.
|
| [2. NULLABLE VS ERROR]
| - If a record is NOT FOUND, return (nil, nil) instead of an error for Query
|   methods (unless the business logic dictates that the absence is an anomaly).
| - Database connection issues or syntax errors MUST still be mapped and returned.
|
| [3. READ-ONLY CONTEXT]
| - Ensure .WithContext(ctx) is called to respect timeouts, cancellations,
|   and tracing propagation.
|
| [4. PRELOAD DISCIPLINE]
| - Only Preload relationships that are strictly necessary for the requested
|   operation to avoid N+1 query problems or heavy payload bloat.
|
|------------------------------------------------------------------------------------
*/
package query

import (
	"context"
	"errors"
	database "voyago/core-api/internal/infrastructure/db"
	"voyago/core-api/internal/modules/booking/entity"
	"voyago/core-api/internal/modules/booking/repository"

	"gorm.io/gorm"
)

// bookingRepository implements the repository.BookingQueryRepository interface.
// It focuses on efficient data fetching and complex filtering logic.
type bookingRepository struct {
	DB database.Database
}

// [INTERFACE COMPLIANCE CHECK]
var _ repository.BookingQueryRepository = (*bookingRepository)(nil)

// NewBookingRepository creates a new instance for reading Booking data.
func NewBookingRepository(db database.Database) repository.BookingQueryRepository {
	return &bookingRepository{
		DB: db,
	}
}

func (r *bookingRepository) ExistsByBookingCode(ctx context.Context, code string) (bool, error) {
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

func (r *bookingRepository) FindByCode(ctx context.Context, code string) (*entity.Booking, error) {
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
func (r *bookingRepository) FindByID(ctx context.Context, id string) (*entity.Booking, error) {
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
