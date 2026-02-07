package helper

import (
	"voyago/core-api/internal/modules/booking/entity"
)

// BookingFixture provides reusable test data builders for booking entities
type BookingFixture struct {
	ID          string
	BookingCode string
	UserID      string
	TotalAmount float64
	Status      entity.BookingStatus
	Details     []BookingDetailFixture
}

type BookingDetailFixture struct {
	ID           string
	ProductID    string
	ProductName  *string
	Qty          int32
	PricePerUnit float64
	SubTotal     float64
}

// NewBookingFixture creates a valid booking fixture with sensible defaults
func NewBookingFixture() *BookingFixture {
	productName := "Test Product"
	return &BookingFixture{
		ID:          "11111111-1111-1111-1111-111111111111",
		BookingCode: "TEST001",
		UserID:      "22222222-2222-2222-2222-222222222222",
		TotalAmount: 100.0,
		Status:      entity.BookingStatusPending,
		Details: []BookingDetailFixture{
			{
				ID:           "33333333-3333-3333-3333-333333333333",
				ProductID:    "44444444-4444-4444-4444-444444444444",
				ProductName:  &productName,
				Qty:          2,
				PricePerUnit: 50.0,
				SubTotal:     100.0,
			},
		},
	}
}

// WithID sets custom booking ID
func (f *BookingFixture) WithID(id string) *BookingFixture {
	f.ID = id
	return f
}

// WithBookingCode sets custom booking code
func (f *BookingFixture) WithBookingCode(code string) *BookingFixture {
	f.BookingCode = code
	return f
}

// WithUserID sets custom user ID
func (f *BookingFixture) WithUserID(userID string) *BookingFixture {
	f.UserID = userID
	return f
}

// WithStatus sets booking status
func (f *BookingFixture) WithStatus(status entity.BookingStatus) *BookingFixture {
	f.Status = status
	return f
}

// WithDetails sets custom booking details
func (f *BookingFixture) WithDetails(details []BookingDetailFixture) *BookingFixture {
	f.Details = details
	// Recalculate total amount
	total := 0.0
	for _, d := range details {
		total += d.SubTotal
	}
	f.TotalAmount = total
	return f
}

// ToEntity converts fixture to entity.Booking
func (f *BookingFixture) ToEntity() *entity.Booking {
	details := make([]entity.BookingDetail, len(f.Details))
	for i, d := range f.Details {
		details[i] = entity.BookingDetail{
			ID:           d.ID,
			BookingID:    f.ID,
			ProductID:    d.ProductID,
			ProductName:  d.ProductName,
			Qty:          d.Qty,
			PricePerUnit: d.PricePerUnit,
			SubTotal:     d.SubTotal,
		}
	}

	return &entity.Booking{
		ID:          f.ID,
		BookingCode: f.BookingCode,
		UserID:      f.UserID,
		TotalAmount: f.TotalAmount,
		Status:      f.Status,
		Details:     details,
	}
}

// NewBookingDetailFixture creates a valid booking detail fixture
func NewBookingDetailFixture(productID string, qty int32, price float64) BookingDetailFixture {
	productName := "Test Product"
	return BookingDetailFixture{
		ID:           "detail-id-" + productID,
		ProductID:    productID,
		ProductName:  &productName,
		Qty:          qty,
		PricePerUnit: price,
		SubTotal:     price * float64(qty),
	}
}
