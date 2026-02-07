//go:build integration
// +build integration

package booking_test

import (
	"context"
	"testing"

	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/modules/booking/entity"
	"voyago/core-api/internal/modules/booking/repository/command"
	"voyago/core-api/internal/modules/booking/repository/query"
	"voyago/core-api/internal/modules/booking/usecase"
	"voyago/core-api/test/helper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateBooking_Integration tests the full flow with real database
func TestCreateBooking_Integration(t *testing.T) {
	// Setup
	db := helper.SetupTestDB(t)
	defer helper.CleanupTestDB(t, db)

	// Clean tables before test
	helper.TruncateTables(t, db.GetDB(), "booking_details", "bookings")

	// Initialize real repositories
	bookingCmd := command.NewBookingRepository(db)
	bookingQry := query.NewBookingRepository(db)

	// Initialize real usecase with real dependencies
	log := logger.NewNoOpLogger()
	trc := tracer.NewNoOpTracer()

	uc := usecase.NewCreateBookingUseCase(
		log,
		trc,
		db, // TransactionManager
		usecase.CreateBookingRepositories{
			BookingCmd: bookingCmd,
			BookingQry: bookingQry,
		},
	)

	// Test data
	productName := "Integration Test Product"
	req := &usecase.CreateBookingRequest{
		BookingCode: "INTEG001",
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		TotalAmount: 150.0,
		Details: []usecase.CreateBookingDetailRequest{
			{
				ProductID:    "650e8400-e29b-41d4-a716-446655440000",
				ProductName:  &productName,
				Qty:          3,
				PricePerUnit: 50.0,
				SubTotal:     150.0,
			},
		},
	}

	// Execute
	ctx := context.Background()
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, req.BookingCode, resp.BookingCode)
	assert.Equal(t, req.UserID, resp.UserID)
	assert.Equal(t, req.TotalAmount, resp.TotalAmount)
	assert.NotEmpty(t, resp.BookingID)

	// Verify data persisted in real database
	found, err := bookingQry.FindByCode(ctx, req.BookingCode)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, resp.BookingID, found.ID)
	assert.Equal(t, req.BookingCode, found.BookingCode)
	assert.Len(t, found.Details, 1)
	assert.Equal(t, entity.BookingStatusPending, found.Status)
}

// TestCreateBooking_Integration_DuplicateCode tests duplicate code detection
func TestCreateBooking_Integration_DuplicateCode(t *testing.T) {
	// Setup
	db := helper.SetupTestDB(t)
	defer helper.CleanupTestDB(t, db)

	// Clean tables
	helper.TruncateTables(t, db.GetDB(), "booking_details", "bookings")

	// Initialize repositories and usecase
	bookingCmd := command.NewBookingRepository(db)
	bookingQry := query.NewBookingRepository(db)
	log := logger.NewNoOpLogger()
	trc := tracer.NewNoOpTracer()

	uc := usecase.NewCreateBookingUseCase(
		log,
		trc,
		db,
		usecase.CreateBookingRepositories{
			BookingCmd: bookingCmd,
			BookingQry: bookingQry,
		},
	)

	// Create first booking
	productName := "Product 1"
	req1 := &usecase.CreateBookingRequest{
		BookingCode: "DUP001",
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		TotalAmount: 100.0,
		Details: []usecase.CreateBookingDetailRequest{
			{
				ProductID:    "650e8400-e29b-41d4-a716-446655440000",
				ProductName:  &productName,
				Qty:          2,
				PricePerUnit: 50.0,
				SubTotal:     100.0,
			},
		},
	}

	ctx := context.Background()
	_, err := uc.Execute(ctx, req1)
	require.NoError(t, err)

	// Try to create duplicate
	req2 := &usecase.CreateBookingRequest{
		BookingCode: "DUP001", // Same code
		UserID:      "660e8400-e29b-41d4-a716-446655440000",
		TotalAmount: 200.0,
		Details: []usecase.CreateBookingDetailRequest{
			{
				ProductID:    "750e8400-e29b-41d4-a716-446655440000",
				ProductName:  &productName,
				Qty:          4,
				PricePerUnit: 50.0,
				SubTotal:     200.0,
			},
		},
	}

	_, err = uc.Execute(ctx, req2)

	// Assert: should fail with duplicate error
	require.Error(t, err)
	assert.Equal(t, entity.ErrBookingCodeAlreadyExists, err)
}

// TestCreateBooking_Integration_TransactionRollback tests transaction rollback
func TestCreateBooking_Integration_TransactionRollback(t *testing.T) {
	// Setup
	db := helper.SetupTestDB(t)
	defer helper.CleanupTestDB(t, db)

	// Clean tables
	helper.TruncateTables(t, db.GetDB(), "booking_details", "bookings")

	// Create a booking fixture with invalid data that will fail validation
	fixture := helper.NewBookingFixture().
		WithBookingCode("ROLLBACK001").
		WithDetails([]helper.BookingDetailFixture{}) // Empty details - will fail validation

	bookingEntity := fixture.ToEntity()
	bookingEntity.TotalAmount = 100.0 // Wrong amount

	// Initialize repositories
	bookingCmd := command.NewBookingRepository(db)
	bookingQry := query.NewBookingRepository(db)

	ctx := context.Background()

	// Attempt to create directly via repository (bypassing usecase validation)
	err := db.Atomic(ctx, func(txCtx context.Context) error {
		return bookingCmd.Create(txCtx, bookingEntity)
	})

	// Should succeed at repository level (no details is DB-valid)
	// But when we try with usecase, it should fail validation

	log := logger.NewNoOpLogger()
	trc := tracer.NewNoOpTracer()

	uc := usecase.NewCreateBookingUseCase(
		log,
		trc,
		db,
		usecase.CreateBookingRepositories{
			BookingCmd: bookingCmd,
			BookingQry: bookingQry,
		},
	)

	req := &usecase.CreateBookingRequest{
		BookingCode: "ROLLBACK002",
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		TotalAmount: 100.0,
		Details:     []usecase.CreateBookingDetailRequest{}, // Empty - will fail
	}

	_, err = uc.Execute(ctx, req)

	// Assert: validation should fail
	require.Error(t, err)
	assert.Equal(t, entity.ErrBookingDetailRequired, err)

	// Verify nothing was persisted
	found, err := bookingQry.FindByCode(ctx, "ROLLBACK002")
	assert.NoError(t, err)
	assert.Nil(t, found, "Booking should not exist after failed validation")
}

// TestCreateBooking_Integration_MultipleDetails tests booking with multiple details
func TestCreateBooking_Integration_MultipleDetails(t *testing.T) {
	// Setup
	db := helper.SetupTestDB(t)
	defer helper.CleanupTestDB(t, db)

	// Clean tables
	helper.TruncateTables(t, db.GetDB(), "booking_details", "bookings")

	// Initialize components
	bookingCmd := command.NewBookingRepository(db)
	bookingQry := query.NewBookingRepository(db)
	log := logger.NewNoOpLogger()
	trc := tracer.NewNoOpTracer()

	uc := usecase.NewCreateBookingUseCase(
		log,
		trc,
		db,
		usecase.CreateBookingRepositories{
			BookingCmd: bookingCmd,
			BookingQry: bookingQry,
		},
	)

	// Create request with multiple details
	product1 := "Product 1"
	product2 := "Product 2"
	product3 := "Product 3"

	req := &usecase.CreateBookingRequest{
		BookingCode: "MULTI001",
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		TotalAmount: 350.0,
		Details: []usecase.CreateBookingDetailRequest{
			{
				ProductID:    "prod-id-001",
				ProductName:  &product1,
				Qty:          2,
				PricePerUnit: 50.0,
				SubTotal:     100.0,
			},
			{
				ProductID:    "prod-id-002",
				ProductName:  &product2,
				Qty:          3,
				PricePerUnit: 50.0,
				SubTotal:     150.0,
			},
			{
				ProductID:    "prod-id-003",
				ProductName:  &product3,
				Qty:          2,
				PricePerUnit: 50.0,
				SubTotal:     100.0,
			},
		},
	}

	// Execute
	ctx := context.Background()
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Details, 3)

	// Verify in database
	found, err := bookingQry.FindByCode(ctx, "MULTI001")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Len(t, found.Details, 3)
	assert.Equal(t, 350.0, found.TotalAmount)
}
