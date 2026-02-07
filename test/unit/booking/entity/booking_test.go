package entity_test

import (
	"testing"

	"voyago/core-api/internal/modules/booking/entity"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// TEST HELPERS
// ============================================================================

func createValidBooking() *entity.Booking {
	productName := "Test Product"
	return &entity.Booking{
		ID:          "booking-id-123",
		BookingCode: "BOOK001",
		UserID:      "user-id-456",
		TotalAmount: 100.0,
		Status:      entity.BookingStatusPending,
		Details: []entity.BookingDetail{
			{
				ID:           "detail-id-789",
				BookingID:    "booking-id-123",
				ProductID:    "product-id-111",
				ProductName:  &productName,
				Qty:          2,
				PricePerUnit: 50.0,
				SubTotal:     100.0,
			},
		},
	}
}

// ============================================================================
// TEST CASES
// ============================================================================

func TestBooking_TableName(t *testing.T) {
	// Arrange
	booking := entity.Booking{}

	// Act
	tableName := booking.TableName()

	// Assert
	assert.Equal(t, "bookings", tableName)
}

func TestBooking_Validate_Success(t *testing.T) {
	// Arrange
	booking := createValidBooking()

	// Act
	err := booking.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestBooking_Validate_EmptyDetails(t *testing.T) {
	// Arrange
	booking := createValidBooking()
	booking.Details = []entity.BookingDetail{} // Empty details

	// Act
	err := booking.Validate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, entity.ErrBookingDetailsRequired, err)
}

func TestBooking_Validate_TotalAmountInconsistent(t *testing.T) {
	// Arrange
	booking := createValidBooking()
	booking.TotalAmount = 200.0 // Should be 100.0

	// Act
	err := booking.Validate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, entity.ErrBookingAmountInconsistent, err)
}

func TestBooking_Validate_DetailSubTotalInconsistent(t *testing.T) {
	// Arrange
	booking := createValidBooking()
	booking.Details[0].SubTotal = 90.0 // Should be 100.0 (50 * 2)
	booking.TotalAmount = 90.0         // Update total to match

	// Act
	err := booking.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid subtotal")
}

func TestBooking_Validate_MultipleDetails_Success(t *testing.T) {
	// Arrange
	productName1 := "Product 1"
	productName2 := "Product 2"
	booking := &entity.Booking{
		ID:          "booking-id-123",
		BookingCode: "BOOK002",
		UserID:      "user-id-456",
		TotalAmount: 250.0,
		Status:      entity.BookingStatusPending,
		Details: []entity.BookingDetail{
			{
				ID:           "detail-id-001",
				BookingID:    "booking-id-123",
				ProductID:    "product-id-111",
				ProductName:  &productName1,
				Qty:          2,
				PricePerUnit: 50.0,
				SubTotal:     100.0,
			},
			{
				ID:           "detail-id-002",
				BookingID:    "booking-id-123",
				ProductID:    "product-id-222",
				ProductName:  &productName2,
				Qty:          3,
				PricePerUnit: 50.0,
				SubTotal:     150.0,
			},
		},
	}

	// Act
	err := booking.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestBooking_Validate_FloatingPointPrecision(t *testing.T) {
	// Test case: verify that epsilon handling works correctly
	// for floating-point precision issues
	productName := "Test Product"
	booking := &entity.Booking{
		ID:          "booking-id-123",
		BookingCode: "BOOK003",
		UserID:      "user-id-456",
		// Use a value that might have floating-point precision issues
		TotalAmount: 59.97,
		Status:      entity.BookingStatusPending,
		Details: []entity.BookingDetail{
			{
				ID:           "detail-id-789",
				BookingID:    "booking-id-123",
				ProductID:    "product-id-111",
				ProductName:  &productName,
				Qty:          3,
				PricePerUnit: 19.99,
				SubTotal:     59.97, // 19.99 * 3 = 59.97
			},
		},
	}

	// Act
	err := booking.Validate()

	// Assert
	assert.NoError(t, err, "Should handle floating-point precision correctly")
}

func TestBooking_Validate_MultipleDetails_OneInvalidSubTotal(t *testing.T) {
	// Arrange
	productName1 := "Product 1"
	productName2 := "Product 2"
	booking := &entity.Booking{
		ID:          "booking-id-123",
		BookingCode: "BOOK004",
		UserID:      "user-id-456",
		TotalAmount: 240.0, // 100 + 140 = 240
		Status:      entity.BookingStatusPending,
		Details: []entity.BookingDetail{
			{
				ID:           "detail-id-001",
				BookingID:    "booking-id-123",
				ProductID:    "product-id-111",
				ProductName:  &productName1,
				Qty:          2,
				PricePerUnit: 50.0,
				SubTotal:     100.0, // Valid
			},
			{
				ID:           "detail-id-002",
				BookingID:    "booking-id-123",
				ProductID:    "product-id-222",
				ProductName:  &productName2,
				Qty:          3,
				PricePerUnit: 50.0,
				SubTotal:     140.0, // Invalid: should be 150.0
			},
		},
	}

	// Act
	err := booking.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid subtotal")
	assert.Contains(t, err.Error(), "product-id-222")
}

func TestBooking_Validate_EdgeCase_ZeroAmount(t *testing.T) {
	// Arrange
	productName := "Free Product"
	booking := &entity.Booking{
		ID:          "booking-id-123",
		BookingCode: "BOOK005",
		UserID:      "user-id-456",
		TotalAmount: 0.0,
		Status:      entity.BookingStatusPending,
		Details: []entity.BookingDetail{
			{
				ID:           "detail-id-789",
				BookingID:    "booking-id-123",
				ProductID:    "product-id-111",
				ProductName:  &productName,
				Qty:          1,
				PricePerUnit: 0.0,
				SubTotal:     0.0,
			},
		},
	}

	// Act
	err := booking.Validate()

	// Assert
	assert.NoError(t, err, "Should allow zero amount bookings")
}

// ============================================================================
// BOOKING DETAIL TESTS
// ============================================================================

func TestBookingDetail_TableName(t *testing.T) {
	// Arrange
	detail := entity.BookingDetail{}

	// Act
	tableName := detail.TableName()

	// Assert
	assert.Equal(t, "booking_details", tableName)
}

func TestBookingDetail_Validate_Success(t *testing.T) {
	// Arrange
	detail := &entity.BookingDetail{}

	// Act
	err := detail.Validate()

	// Assert
	// BookingDetail.Validate() returns nil (no validation rules)
	assert.NoError(t, err)
}
