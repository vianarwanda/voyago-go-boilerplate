package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"voyago/core-api/internal/infrastructure/config"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/validator"
	deliveryhttp "voyago/core-api/internal/modules/booking/delivery/http"
	"voyago/core-api/internal/modules/booking/usecase"
	"voyago/core-api/internal/pkg/apperror"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCreateBookingUseCase is a mock implementation of usecase.CreateBookingUseCase
type MockCreateBookingUseCase struct {
	mock.Mock
}

func (m *MockCreateBookingUseCase) Execute(ctx context.Context, req *usecase.CreateBookingRequest) (*usecase.CreateBookingResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.CreateBookingResponse), args.Error(1)
}

// setupTestHandler creates a test handler with mocked dependencies
func setupTestHandler(t *testing.T) (*deliveryhttp.Handler, *MockCreateBookingUseCase, *fiber.App) {
	t.Helper()

	// Create mocks
	mockUseCase := new(MockCreateBookingUseCase)

	// Create real dependencies
	cfg := &config.Config{
		App: config.AppConfig{
			Name: "test",
			Env:  "test",
		},
	}
	log := logger.NewNoOpLogger()
	val := validator.NewPlaygroundValidator()

	// Create handler
	handler := deliveryhttp.NewHandler(
		cfg,
		log,
		val,
		deliveryhttp.HandlerUseCases{
			CreateBookingUseCase: mockUseCase,
		},
	)

	// Create Fiber app and register routes
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := err.Error()
			errCode := "ERR_500"
			var details any

			if e, ok := err.(*apperror.AppError); ok {
				code = e.GetHttpStatus()
				message = e.Message
				errCode = e.Code
				details = e.Details
			}

			return c.Status(code).JSON(map[string]any{
				"status":     "error",
				"message":    message,
				"error_code": errCode,
				"details":    details,
			})
		},
	})

	// Register route
	app.Post("/bookings/", handler.CreateBooking)

	return handler, mockUseCase, app
}

// makeRequest helper to create and send HTTP request
func makeRequest(t *testing.T, app *fiber.App, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewReader(jsonData)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	recorder.Code = resp.StatusCode
	bodyBytes, _ := io.ReadAll(resp.Body)
	recorder.Body = bytes.NewBuffer(bodyBytes)

	return recorder
}

// TestHandler_CreateBooking_Success tests successful booking creation
func TestHandler_CreateBooking_Success(t *testing.T) {
	// Setup
	_, mockUseCase, app := setupTestHandler(t)

	productName := "Test Product"
	requestBody := map[string]any{
		"code":         "TEST001",
		"user_id":      "550e8400-e29b-41d4-a716-446655440000",
		"total_amount": 100.0,
		"details": []map[string]any{
			{
				"product_id":     "650e8400-e29b-41d4-a716-446655440000",
				"product_name":   productName,
				"qty":            2,
				"price_per_unit": 50.0,
				"sub_total":      100.0,
			},
		},
	}

	expectedResponse := &usecase.CreateBookingResponse{
		BookingID:   "123e4567-e89b-12d3-a456-426614174000",
		BookingCode: "TEST001",
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		TotalAmount: 100.0,
		Details: []usecase.CreateBookingDetailResponse{
			{
				ProductID:    "650e8400-e29b-41d4-a716-446655440000",
				ProductName:  &productName,
				Qty:          2,
				PricePerUnit: 50.0,
				SubTotal:     100.0,
			},
		},
	}

	mockUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *usecase.CreateBookingRequest) bool {
		return req.BookingCode == "TEST001"
	})).Return(expectedResponse, nil)

	// Execute
	resp := makeRequest(t, app, "POST", "/bookings/", requestBody)

	// Assert
	assert.Equal(t, fiber.StatusCreated, resp.Code)

	var response map[string]any
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Booking created successfully", response["message"])
	assert.NotNil(t, response["data"])

	mockUseCase.AssertExpectations(t)
}

// TestHandler_CreateBooking_ValidationErrors tests various validation failures
func TestHandler_CreateBooking_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name           string
		requestBody    map[string]any
		expectedStatus int
		expectedField  string
		expectedCode   string
	}{
		{
			name: "Empty booking code (required)",
			requestBody: map[string]any{
				"code":         "",
				"user_id":      "550e8400-e29b-41d4-a716-446655440000",
				"total_amount": 100.0,
				"details": []map[string]any{
					{
						"product_id":     "650e8400-e29b-41d4-a716-446655440000",
						"qty":            2,
						"price_per_unit": 50.0,
						"sub_total":      100.0,
					},
				},
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedField:  "code",
			expectedCode:   "required",
		},
		{
			name: "Booking code too short (min=3)",
			requestBody: map[string]any{
				"code":         "AB",
				"user_id":      "550e8400-e29b-41d4-a716-446655440000",
				"total_amount": 100.0,
				"details": []map[string]any{
					{
						"product_id":     "650e8400-e29b-41d4-a716-446655440000",
						"qty":            2,
						"price_per_unit": 50.0,
						"sub_total":      100.0,
					},
				},
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedField:  "code",
			expectedCode:   "min",
		},
		{
			name: "Invalid user_id UUID format",
			requestBody: map[string]any{
				"code":         "TEST001",
				"user_id":      "not-a-valid-uuid",
				"total_amount": 100.0,
				"details": []map[string]any{
					{
						"product_id":     "650e8400-e29b-41d4-a716-446655440000",
						"qty":            2,
						"price_per_unit": 50.0,
						"sub_total":      100.0,
					},
				},
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedField:  "user_id",
			expectedCode:   "uuid",
		},
		{
			name: "Negative total_amount (gte=0)",
			requestBody: map[string]any{
				"code":         "TEST001",
				"user_id":      "550e8400-e29b-41d4-a716-446655440000",
				"total_amount": -100.0,
				"details": []map[string]any{
					{
						"product_id":     "650e8400-e29b-41d4-a716-446655440000",
						"qty":            2,
						"price_per_unit": 50.0,
						"sub_total":      100.0,
					},
				},
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedField:  "total_amount",
			expectedCode:   "gte",
		},
		{
			name: "Empty details array (min=1)",
			requestBody: map[string]any{
				"code":         "TEST001",
				"user_id":      "550e8400-e29b-41d4-a716-446655440000",
				"total_amount": 0.0,
				"details":      []map[string]any{},
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedField:  "details",
			expectedCode:   "min",
		},
		{
			name: "Invalid product_id in details (uuid_rfc4122)",
			requestBody: map[string]any{
				"code":         "TEST001",
				"user_id":      "550e8400-e29b-41d4-a716-446655440000",
				"total_amount": 100.0,
				"details": []map[string]any{
					{
						"product_id":     "invalid-uuid",
						"qty":            2,
						"price_per_unit": 50.0,
						"sub_total":      100.0,
					},
				},
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedField:  "product_id",
			expectedCode:   "uuid",
		},
		{
			name: "Negative quantity (gt=0)",
			requestBody: map[string]any{
				"code":         "TEST001",
				"user_id":      "550e8400-e29b-41d4-a716-446655440000",
				"total_amount": 100.0,
				"details": []map[string]any{
					{
						"product_id":     "650e8400-e29b-41d4-a716-446655440000",
						"qty":            -1,
						"price_per_unit": 50.0,
						"sub_total":      100.0,
					},
				},
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedField:  "qty",
			expectedCode:   "gt",
		},
		{
			name: "Negative price_per_unit (gt=0)",
			requestBody: map[string]any{
				"code":         "TEST001",
				"user_id":      "550e8400-e29b-41d4-a716-446655440000",
				"total_amount": 100.0,
				"details": []map[string]any{
					{
						"product_id":     "650e8400-e29b-41d4-a716-446655440000",
						"qty":            2,
						"price_per_unit": -50.0,
						"sub_total":      100.0,
					},
				},
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedField:  "price_per_unit",
			expectedCode:   "gt",
		},
		{
			name: "Negative sub_total (gt=0)",
			requestBody: map[string]any{
				"code":         "TEST001",
				"user_id":      "550e8400-e29b-41d4-a716-446655440000",
				"total_amount": 100.0,
				"details": []map[string]any{
					{
						"product_id":     "650e8400-e29b-41d4-a716-446655440000",
						"qty":            2,
						"price_per_unit": 50.0,
						"sub_total":      -100.0,
					},
				},
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedField:  "sub_total",
			expectedCode:   "gt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			_, _, app := setupTestHandler(t)

			// Execute
			resp := makeRequest(t, app, "POST", "/bookings/", tc.requestBody)

			// Assert
			assert.Equal(t, tc.expectedStatus, resp.Code)

			var response map[string]any
			err := json.Unmarshal(resp.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "error", response["status"])

			// Check details array exists
			details, ok := response["details"].([]any)
			require.True(t, ok, "Details should be an array")
			require.NotEmpty(t, details, "Details should not be empty")

			// Find error for expected field
			found := false
			for _, detail := range details {
				detailMap := detail.(map[string]any)
				if detailMap["field"] == tc.expectedField {
					assert.Equal(t, tc.expectedCode, detailMap["code"])
					found = true
					break
				}
			}
			assert.True(t, found, "Expected validation error for field %s not found", tc.expectedField)
		})
	}
}

// TestHandler_CreateBooking_MalformedJSON tests malformed JSON request
func TestHandler_CreateBooking_MalformedJSON(t *testing.T) {
	// Setup
	_, _, app := setupTestHandler(t)

	// Execute with invalid JSON
	req := httptest.NewRequest("POST", "/bookings/", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]any
	err = json.Unmarshal(bodyBytes, &response)
	require.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["error_code"], "REQ")
}

// TestHandler_CreateBooking_UseCaseError tests use case errors
func TestHandler_CreateBooking_UseCaseError(t *testing.T) {
	// Setup
	_, mockUseCase, app := setupTestHandler(t)

	requestBody := map[string]any{
		"code":         "TEST001",
		"user_id":      "550e8400-e29b-41d4-a716-446655440000",
		"total_amount": 100.0,
		"details": []map[string]any{
			{
				"product_id":     "650e8400-e29b-41d4-a716-446655440000",
				"qty":            2,
				"price_per_unit": 50.0,
				"sub_total":      100.0,
			},
		},
	}

	// Mock use case to return error
	mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(
		nil,
		apperror.NewInternal("TEST_ERROR", "Test error message", errors.New("underlying error")),
	)

	// Execute
	resp := makeRequest(t, app, "POST", "/bookings/", requestBody)

	// Assert
	assert.Equal(t, fiber.StatusInternalServerError, resp.Code)

	var response map[string]any
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Test error message", response["message"])
	assert.Equal(t, "TEST_ERROR", response["error_code"])

	mockUseCase.AssertExpectations(t)
}

// TestHandler_CreateBooking_ProductNameOptional tests product_name is optional
func TestHandler_CreateBooking_ProductNameOptional(t *testing.T) {
	// Setup
	_, mockUseCase, app := setupTestHandler(t)

	requestBody := map[string]any{
		"code":         "TEST001",
		"user_id":      "550e8400-e29b-41d4-a716-446655440000",
		"total_amount": 100.0,
		"details": []map[string]any{
			{
				"product_id": "650e8400-e29b-41d4-a716-446655440000",
				// product_name omitted
				"qty":            2,
				"price_per_unit": 50.0,
				"sub_total":      100.0,
			},
		},
	}

	expectedResponse := &usecase.CreateBookingResponse{
		BookingID:   "123e4567-e89b-12d3-a456-426614174000",
		BookingCode: "TEST001",
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		TotalAmount: 100.0,
		Details: []usecase.CreateBookingDetailResponse{
			{
				ProductID:    "650e8400-e29b-41d4-a716-446655440000",
				ProductName:  nil, // nil is valid
				Qty:          2,
				PricePerUnit: 50.0,
				SubTotal:     100.0,
			},
		},
	}

	mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(expectedResponse, nil)

	// Execute
	resp := makeRequest(t, app, "POST", "/bookings/", requestBody)

	// Assert - should succeed since product_name is optional (omitempty)
	assert.Equal(t, fiber.StatusCreated, resp.Code)

	mockUseCase.AssertExpectations(t)
}
