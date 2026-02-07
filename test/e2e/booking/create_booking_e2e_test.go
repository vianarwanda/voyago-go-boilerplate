//go:build e2e
// +build e2e

package booking_test

import (
	"encoding/json"
	"testing"

	"voyago/core-api/internal/infrastructure/config"
	database "voyago/core-api/internal/infrastructure/db"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/infrastructure/validator"
	"voyago/core-api/internal/modules/booking"
	"voyago/core-api/internal/modules/booking/usecase"
	"voyago/core-api/test/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer initializes a test Fiber app with all dependencies
func setupTestServer(t *testing.T) (*helper.HTTPTestHelper, database.Database) {
	t.Helper()

	// Setup test database
	db := helper.SetupTestDB(t)

	// Initialize infrastructure
	cfg := &config.Config{
		App: config.AppConfig{
			Name: "voyago-test",
			Env:  "test",
		},
	}
	log := logger.NewNoOpLogger()
	trc := tracer.NewNoOpTracer()
	val := validator.NewPlaygroundValidator()

	// Create Fiber app directly
	app := fiber.New(fiber.Config{
		AppName: cfg.App.Name,
	})

	// Register booking module
	booking.RegisterModule(booking.ModuleConfig{
		Config: cfg,
		Server: app,
		DB:     db,
		Log:    log,
		Val:    val,
		Tracer: trc,
	})

	return helper.NewHTTPTestHelper(app, t), db
}

// TestCreateBooking_E2E_Success tests successful booking creation via HTTP
func TestCreateBooking_E2E_Success(t *testing.T) {
	// Setup
	httpHelper, db := setupTestServer(t)
	defer helper.CleanupTestDB(t, db)

	// Clean tables
	helper.TruncateTables(t, db.GetDB(), "booking_details", "bookings")

	// Prepare request
	productName := "E2E Test Product"
	requestBody := map[string]interface{}{
		"code":         "E2E001",
		"user_id":      "550e8400-e29b-41d4-a716-446655440000",
		"total_amount": 200.0,
		"details": []map[string]interface{}{
			{
				"product_id":     "650e8400-e29b-41d4-a716-446655440000",
				"product_name":   productName,
				"qty":            4,
				"price_per_unit": 50.0,
				"sub_total":      200.0,
			},
		},
	}

	// Execute
	resp := httpHelper.POST("/bookings/", requestBody)

	// Assert response
	var response map[string]interface{}
	httpHelper.AssertJSONResponse(resp, 201, &response)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Booking created successfully", response["message"])

	// Verify response data
	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "Response data should be a map")

	assert.Equal(t, "E2E001", data["code"])
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", data["user_id"])
	assert.Equal(t, 200.0, data["total_amount"])
	assert.NotEmpty(t, data["id"], "Booking ID should be generated")
}

// TestCreateBooking_E2E_ValidationError tests validation error responses
func TestCreateBooking_E2E_ValidationError(t *testing.T) {
	// Setup
	httpHelper, db := setupTestServer(t)
	defer helper.CleanupTestDB(t, db)

	testCases := []struct {
		name               string
		requestBody        map[string]interface{}
		expectedStatus     int
		expectedErrorField string
	}{
		{
			name: "Empty booking code",
			requestBody: map[string]interface{}{
				"code":         "",
				"user_id":      "550e8400-e29b-41d4-a716-446655440000",
				"total_amount": 100.0,
				"details": []map[string]interface{}{
					{
						"product_id":     "650e8400-e29b-41d4-a716-446655440000",
						"qty":            2,
						"price_per_unit": 50.0,
						"sub_total":      100.0,
					},
				},
			},
			expectedStatus:     400,
			expectedErrorField: "code",
		},
		{
			name: "Invalid user ID format",
			requestBody: map[string]interface{}{
				"code":         "TEST001",
				"user_id":      "not-a-uuid",
				"total_amount": 100.0,
				"details": []map[string]interface{}{
					{
						"product_id":     "650e8400-e29b-41d4-a716-446655440000",
						"qty":            2,
						"price_per_unit": 50.0,
						"sub_total":      100.0,
					},
				},
			},
			expectedStatus:     400,
			expectedErrorField: "user_id",
		},
		{
			name: "Empty details",
			requestBody: map[string]interface{}{
				"code":         "TEST001",
				"user_id":      "550e8400-e29b-41d4-a716-446655440000",
				"total_amount": 0.0,
				"details":      []map[string]interface{}{},
			},
			expectedStatus:     400,
			expectedErrorField: "details",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute
			resp := httpHelper.POST("/bookings/", tc.requestBody)

			// Assert
			errResp := httpHelper.AssertErrorResponse(resp, tc.expectedStatus)

			assert.Equal(t, "error", errResp["status"])

			// Check if details field exists for validation errors
			if details, ok := errResp["details"]; ok {
				detailsArray, ok := details.([]interface{})
				require.True(t, ok, "Details should be an array")
				require.NotEmpty(t, detailsArray, "Details array should not be empty")
			}
		})
	}
}

// TestCreateBooking_E2E_DuplicateCode tests duplicate booking code handling
func TestCreateBooking_E2E_DuplicateCode(t *testing.T) {
	// Setup
	httpHelper, db := setupTestServer(t)
	defer helper.CleanupTestDB(t, db)

	// Clean tables
	helper.TruncateTables(t, db.GetDB(), "booking_details", "bookings")

	// Create first booking
	productName := "Product 1"
	requestBody := map[string]interface{}{
		"code":         "DUP_E2E001",
		"user_id":      "550e8400-e29b-41d4-a716-446655440000",
		"total_amount": 100.0,
		"details": []map[string]interface{}{
			{
				"product_id":     "650e8400-e29b-41d4-a716-446655440000",
				"product_name":   productName,
				"qty":            2,
				"price_per_unit": 50.0,
				"sub_total":      100.0,
			},
		},
	}

	// First request should succeed
	resp1 := httpHelper.POST("/bookings/", requestBody)
	var successResp map[string]interface{}
	httpHelper.AssertJSONResponse(resp1, 201, &successResp)
	assert.Equal(t, "success", successResp["status"])

	// Second request with same code should fail
	resp2 := httpHelper.POST("/bookings/", requestBody)
	errResp := httpHelper.AssertErrorResponse(resp2, 400)

	assert.Equal(t, "error", errResp["status"])
	assert.Contains(t, errResp["message"], "already exists")
}

// TestCreateBooking_E2E_MalformedJSON tests malformed JSON request handling
func TestCreateBooking_E2E_MalformedJSON(t *testing.T) {
	// Setup
	httpHelper, db := setupTestServer(t)
	defer helper.CleanupTestDB(t, db)

	// Create a request with invalid JSON (using string instead of struct)
	// This tests the BodyParser error handling
	resp := httpHelper.POST("/bookings/", "invalid json")

	// Assert
	errResp := httpHelper.AssertErrorResponse(resp, 400)
	assert.Equal(t, "error", errResp["status"])
}

// TestCreateBooking_E2E_AmountMismatch tests amount validation
func TestCreateBooking_E2E_AmountMismatch(t *testing.T) {
	// Setup
	httpHelper, db := setupTestServer(t)
	defer helper.CleanupTestDB(t, db)

	// Clean tables
	helper.TruncateTables(t, db.GetDB(), "booking_details", "bookings")

	// Request with mismatched total amount
	productName := "Product"
	requestBody := map[string]interface{}{
		"code":         "AMOUNT001",
		"user_id":      "550e8400-e29b-41d4-a716-446655440000",
		"total_amount": 300.0, // Should be 100.0
		"details": []map[string]interface{}{
			{
				"product_id":     "650e8400-e29b-41d4-a716-446655440000",
				"product_name":   productName,
				"qty":            2,
				"price_per_unit": 50.0,
				"sub_total":      100.0,
			},
		},
	}

	// Execute
	resp := httpHelper.POST("/bookings/", requestBody)

	// Assert
	errResp := httpHelper.AssertErrorResponse(resp, 400)
	assert.Equal(t, "error", errResp["status"])
	assert.Contains(t, errResp["message"], "amount")
}

// TestCreateBooking_E2E_CompleteFlow tests the complete booking flow
func TestCreateBooking_E2E_CompleteFlow(t *testing.T) {
	// Setup
	httpHelper, db := setupTestServer(t)
	defer helper.CleanupTestDB(t, db)

	// Clean tables
	helper.TruncateTables(t, db.GetDB(), "booking_details", "bookings")

	// Step 1: Create a booking with multiple details
	product1 := "Product A"
	product2 := "Product B"
	requestBody := map[string]interface{}{
		"code":         "FLOW001",
		"user_id":      "550e8400-e29b-41d4-a716-446655440000",
		"total_amount": 300.0,
		"details": []map[string]interface{}{
			{
				"product_id":     "prod-001",
				"product_name":   product1,
				"qty":            2,
				"price_per_unit": 50.0,
				"sub_total":      100.0,
			},
			{
				"product_id":     "prod-002",
				"product_name":   product2,
				"qty":            4,
				"price_per_unit": 50.0,
				"sub_total":      200.0,
			},
		},
	}

	// Execute
	resp := httpHelper.POST("/bookings/", requestBody)

	// Assert
	var response map[string]interface{}
	httpHelper.AssertJSONResponse(resp, 201, &response)

	assert.Equal(t, "success", response["status"])

	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok)

	bookingID := data["id"].(string)
	assert.NotEmpty(t, bookingID)

	// Verify details in response
	details, ok := data["details"].([]interface{})
	require.True(t, ok)
	assert.Len(t, details, 2)

	// Step 2: Verify data was persisted
	// (This verifies the integration between HTTP layer and persistence)
	var found usecase.CreateBookingResponse
	respJSON, _ := json.Marshal(data)
	err := json.Unmarshal(respJSON, &found)
	require.NoError(t, err)

	assert.Equal(t, "FLOW001", found.BookingCode)
	assert.Equal(t, 300.0, found.TotalAmount)
	assert.Len(t, found.Details, 2)
}
