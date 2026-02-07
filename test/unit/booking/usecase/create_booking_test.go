package usecase_test

import (
	"context"
	"errors"
	"testing"

	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"
	"voyago/core-api/internal/modules/booking/entity"
	"voyago/core-api/internal/modules/booking/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ============================================================================
// MOCKS
// ============================================================================

// MockLogger is a mock implementation of logger.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) WithContext(ctx context.Context) logger.Logger {
	args := m.Called(ctx)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) WithField(key string, value any) logger.Logger {
	args := m.Called(key, value)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) WithFields(fields map[string]any) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) Debug(message string) {
	m.Called(message)
}

func (m *MockLogger) Info(message string) {
	m.Called(message)
}

func (m *MockLogger) Warn(message string) {
	m.Called(message)
}

func (m *MockLogger) Error(message string) {
	m.Called(message)
}

// MockSpan is a mock implementation of tracer.Span
type MockSpan struct {
	mock.Mock
}

func (m *MockSpan) SetOperationName(name string) {
	m.Called(name)
}

func (m *MockSpan) Finish() {
	m.Called()
}

func (m *MockSpan) SetTag(key string, value any) {
	m.Called(key, value)
}

// MockTracer is a mock implementation of tracer.Tracer
type MockTracer struct {
	mock.Mock
}

func (m *MockTracer) StartSpan(ctx context.Context, name string) (tracer.Span, context.Context) {
	args := m.Called(ctx, name)
	return args.Get(0).(tracer.Span), args.Get(1).(context.Context)
}

func (m *MockTracer) UseGorm(db *gorm.DB) {
	m.Called(db)
}

func (m *MockTracer) ExtractTraceInfo(ctx context.Context) (traceID, spanID string, ok bool) {
	args := m.Called(ctx)
	return args.String(0), args.String(1), args.Bool(2)
}

func (m *MockTracer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockTransactionManager is a mock implementation of baserepo.TransactionManager
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) Atomic(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)

	// Execute the function if we're testing success scenarios
	if args.Error(0) == nil {
		return fn(ctx)
	}

	return args.Error(0)
}

// MockBookingCommandRepository is a mock implementation of repository.BookingCommandRepository
type MockBookingCommandRepository struct {
	mock.Mock
}

func (m *MockBookingCommandRepository) Create(ctx context.Context, booking *entity.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *MockBookingCommandRepository) Update(ctx context.Context, booking *entity.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *MockBookingCommandRepository) Delete(ctx context.Context, booking *entity.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

// MockBookingQueryRepository is a mock implementation of repository.BookingQueryRepository
type MockBookingQueryRepository struct {
	mock.Mock
}

func (m *MockBookingQueryRepository) ExistsByBookingCode(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

func (m *MockBookingQueryRepository) FindByID(ctx context.Context, id string) (*entity.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Booking), args.Error(1)
}

func (m *MockBookingQueryRepository) FindByCode(ctx context.Context, code string) (*entity.Booking, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Booking), args.Error(1)
}

// ============================================================================
// TEST HELPERS
// ============================================================================

func setupTest(t *testing.T) (
	*MockLogger,
	*MockTracer,
	*MockSpan,
	*MockTransactionManager,
	*MockBookingCommandRepository,
	*MockBookingQueryRepository,
	usecase.CreateBookingUseCase,
) {
	mockLog := new(MockLogger)
	mockTracer := new(MockTracer)
	mockSpan := new(MockSpan)
	mockTxManager := new(MockTransactionManager)
	mockBookingCmd := new(MockBookingCommandRepository)
	mockBookingQry := new(MockBookingQueryRepository)

	// Setup common mock expectations for logger
	mockLog.On("WithField", "action", "usecase:booking.create").Return(mockLog)
	mockLog.On("WithContext", mock.Anything).Return(mockLog)
	mockLog.On("WithField", "method", "Exec").Return(mockLog)
	mockLog.On("WithFields", mock.Anything).Return(mockLog)
	mockLog.On("Info", mock.Anything).Return()
	mockLog.On("Warn", mock.Anything).Return()
	mockLog.On("Error", mock.Anything).Return()

	// Setup common mock expectations for tracer
	mockTracer.On("StartSpan", mock.Anything, "usecase:booking.create").Return(mockSpan, context.Background())
	mockSpan.On("Finish").Return()
	// RecordSpanError calls SetTag multiple times: error (bool), error.message, error.code, error.kind
	// Use Maybe() to allow 0 or more calls
	mockSpan.On("SetTag", mock.Anything, mock.Anything).Return().Maybe()

	uc := usecase.NewCreateBookingUseCase(
		mockLog,
		mockTracer,
		mockTxManager,
		usecase.CreateBookingRepositories{
			BookingCmd: mockBookingCmd,
			BookingQry: mockBookingQry,
		},
	)

	return mockLog, mockTracer, mockSpan, mockTxManager, mockBookingCmd, mockBookingQry, uc
}

func createValidRequest() *usecase.CreateBookingRequest {
	productName := "Test Product"
	return &usecase.CreateBookingRequest{
		BookingCode: "BOOK001",
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
}

// ============================================================================
// TEST CASES
// ============================================================================

func TestCreateBookingUseCase_Execute_Success(t *testing.T) {
	// Arrange
	_, _, mockSpan, mockTxManager, mockBookingCmd, mockBookingQry, uc := setupTest(t)
	req := createValidRequest()

	mockBookingQry.On("ExistsByBookingCode", mock.Anything, req.BookingCode).Return(false, nil)
	mockTxManager.On("Atomic", mock.Anything, mock.Anything).Return(nil)
	mockBookingCmd.On("Create", mock.Anything, mock.Anything).Return(nil)

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.BookingCode, resp.BookingCode)
	assert.Equal(t, req.UserID, resp.UserID)
	assert.Equal(t, req.TotalAmount, resp.TotalAmount)
	assert.Len(t, resp.Details, 1)
	assert.NotEmpty(t, resp.BookingID)

	mockBookingQry.AssertExpectations(t)
	mockBookingCmd.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestCreateBookingUseCase_Execute_ValidationError_EmptyDetails(t *testing.T) {
	// Arrange
	_, _, mockSpan, _, _, _, uc := setupTest(t)
	req := createValidRequest()
	req.Details = []usecase.CreateBookingDetailRequest{} // Empty details

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, entity.ErrBookingDetailsRequired, err)

	mockSpan.AssertExpectations(t)
}

func TestCreateBookingUseCase_Execute_ValidationError_AmountInconsistent(t *testing.T) {
	// Arrange
	_, _, mockSpan, _, _, _, uc := setupTest(t)
	req := createValidRequest()
	req.TotalAmount = 200.0 // Inconsistent with details subtotal (100.0)

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, entity.ErrBookingAmountInconsistent, err)

	mockSpan.AssertExpectations(t)
}

func TestCreateBookingUseCase_Execute_ValidationError_SubTotalInconsistent(t *testing.T) {
	// Arrange
	_, _, mockSpan, _, _, _, uc := setupTest(t)
	req := createValidRequest()
	req.Details[0].SubTotal = 90.0 // Inconsistent with price * qty (100.0)
	req.TotalAmount = 90.0

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Exactly(t, err.Error(), "detail subtotal does not match with expected subtotal")

	mockSpan.AssertExpectations(t)
}

func TestCreateBookingUseCase_Execute_BookingCodeAlreadyExists(t *testing.T) {
	// Arrange
	_, _, mockSpan, _, _, mockBookingQry, uc := setupTest(t)
	req := createValidRequest()

	mockBookingQry.On("ExistsByBookingCode", mock.Anything, req.BookingCode).Return(true, nil)

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, entity.ErrBookingCodeAlreadyExists, err)

	mockBookingQry.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestCreateBookingUseCase_Execute_ExistsByBookingCodeError(t *testing.T) {
	// Arrange
	_, _, mockSpan, _, _, mockBookingQry, uc := setupTest(t)
	req := createValidRequest()

	expectedErr := errors.New("database connection error")
	mockBookingQry.On("ExistsByBookingCode", mock.Anything, req.BookingCode).Return(false, expectedErr)

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)

	mockBookingQry.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestCreateBookingUseCase_Execute_CreateError(t *testing.T) {
	// Arrange
	_, _, mockSpan, mockTxManager, mockBookingCmd, mockBookingQry, uc := setupTest(t)
	req := createValidRequest()

	expectedErr := errors.New("database insert error")
	mockBookingQry.On("ExistsByBookingCode", mock.Anything, req.BookingCode).Return(false, nil)
	mockBookingCmd.On("Create", mock.Anything, mock.Anything).Return(expectedErr)
	mockTxManager.On("Atomic", mock.Anything, mock.Anything).Return(nil)

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)

	mockBookingQry.AssertExpectations(t)
	mockBookingCmd.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestCreateBookingUseCase_Execute_TransactionError(t *testing.T) {
	// Arrange
	_, _, mockSpan, mockTxManager, _, mockBookingQry, uc := setupTest(t)
	req := createValidRequest()

	expectedErr := errors.New("transaction error")
	mockBookingQry.On("ExistsByBookingCode", mock.Anything, req.BookingCode).Return(false, nil)
	// Mock Atomic to return error without executing the function
	mockTxManager.On("Atomic", mock.Anything, mock.Anything).Return(expectedErr)

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)

	mockBookingQry.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestCreateBookingUseCase_Execute_MultipleDetails(t *testing.T) {
	// Arrange
	_, _, mockSpan, mockTxManager, mockBookingCmd, mockBookingQry, uc := setupTest(t)

	productName1 := "Product 1"
	productName2 := "Product 2"
	req := &usecase.CreateBookingRequest{
		BookingCode: "BOOK002",
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		TotalAmount: 250.0,
		Details: []usecase.CreateBookingDetailRequest{
			{
				ProductID:    "650e8400-e29b-41d4-a716-446655440001",
				ProductName:  &productName1,
				Qty:          2,
				PricePerUnit: 50.0,
				SubTotal:     100.0,
			},
			{
				ProductID:    "650e8400-e29b-41d4-a716-446655440002",
				ProductName:  &productName2,
				Qty:          3,
				PricePerUnit: 50.0,
				SubTotal:     150.0,
			},
		},
	}

	mockBookingQry.On("ExistsByBookingCode", mock.Anything, req.BookingCode).Return(false, nil)
	mockTxManager.On("Atomic", mock.Anything, mock.Anything).Return(nil)
	mockBookingCmd.On("Create", mock.Anything, mock.Anything).Return(nil)

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.BookingCode, resp.BookingCode)
	assert.Equal(t, 250.0, resp.TotalAmount)
	assert.Len(t, resp.Details, 2)

	mockBookingQry.AssertExpectations(t)
	mockBookingCmd.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}
