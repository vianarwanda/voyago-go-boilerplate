package entity

import (
	"fmt"
	"voyago/core-api/internal/pkg/apperror"
)

// [ENTITY STANDARD: DOMAIN SPECIFIC ERROR]
const (
	CodeBookingNotFound           = "booking.not_found"
	CodeBookingCodeAlreadyExists  = "booking.booking_code.already_exists"
	CodeBookingAmountInconsistent = "booking.amount.inconsistent"
	CodeBookingDetailRequired     = "booking.details.required"
)

var (
	ErrBookingNotFound = apperror.NewPermanent(
		CodeBookingNotFound,
		"booking record not found",
	)

	ErrBookingCodeAlreadyExists = apperror.NewPermanent(
		CodeBookingCodeAlreadyExists,
		"booking code already exists",
	)

	ErrBookingAmountInconsistent = apperror.NewPermanent(
		CodeBookingAmountInconsistent,
		"total amount does not match with details subtotal",
	)

	ErrBookingDetailRequired = apperror.NewPermanent(
		CodeBookingDetailRequired,
		"booking must have at least one detail",
	)
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "PENDING"
	BookingStatusConfirmed BookingStatus = "CONFIRMED"
	BookingStatusCancelled BookingStatus = "CANCELLED"
	BookingStatusCompleted BookingStatus = "COMPLETED"
)

type Booking struct {
	ID            string        `gorm:"column:id;type:uuid;primaryKey"`
	BookingCode   string        `gorm:"column:booking_code;type:varchar(50);not null;unique"`
	UserID        string        `gorm:"column:user_id;type:uuid;not null"`
	TotalAmount   float64       `gorm:"column:total_amount;type:decimal(15,2);not null;default:0"`
	Status        BookingStatus `gorm:"column:status;type:varchar(20);not null;default:'PENDING'"`
	PaymentStatus string        `gorm:"column:payment_status;type:varchar(20);not null;default:'UNPAID'"`
	CreatedAt     int64         `gorm:"column:created_at;type:bigint;not null;autoCreateTime:milli"`
	UpdatedAt     *int64        `gorm:"column:updated_at;type:bigint;autoUpdateTime:false"`
	DeletedAt     *int64        `gorm:"column:deleted_at;autoUpdateTime:false"`

	Details []BookingDetail `gorm:"foreignKey:BookingID;references:ID"`
}

func (Booking) TableName() string {
	return "bookings"
}

// [ENTITY STANDARD: DOMAIN VALIDATION]
func (e *Booking) Validate() error {
	// We enforce this at the domain level to prevent "empty" transactions
	// from polluting the database and financial reports.
	if len(e.Details) == 0 {
		return ErrBookingDetailRequired
	}

	// Ensure the header TotalAmount matches the sum of all line item subtotals.
	// This prevents price manipulation and ensures data integrity.
	var calculatedAmount float64
	for _, detail := range e.Details {
		calculatedAmount += detail.SubTotal

		expectedSubTotal := detail.PricePerUnit * float64(detail.Qty)
		if detail.SubTotal != expectedSubTotal {
			return apperror.NewPermanent(
				CodeBookingAmountInconsistent,
				fmt.Sprintf("invalid subtotal for product %s", detail.ProductID),
				fmt.Errorf("expected: %.2f, got: %.2f", expectedSubTotal, detail.SubTotal),
			)
		}
	}

	// We use a small epsilon or direct comparison depending on your precision needs.
	if e.TotalAmount != calculatedAmount {
		return ErrBookingAmountInconsistent
	}

	return nil
}
