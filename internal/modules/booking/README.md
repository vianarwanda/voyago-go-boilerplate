# Booking Module

> **Domain**: Booking Management
> 
> **Responsibility**: Handles booking creation, management, and related business logic for the Voyago platform.

---

## Overview

The Booking module manages the complete lifecycle of bookings, including:
- Creating new bookings with multiple line items (details)
- Validating booking data and business rules
- Managing booking status and payment tracking
- Ensuring data consistency between total amounts and detail subtotals

**Key Features:**
- Multi-item bookings (supports multiple products per booking)
- Unique booking code generation and validation
- Amount consistency validation
- Status tracking (PENDING, CONFIRMED, CANCELLED, COMPLETED)
- Payment status integration

---

## API Endpoints

### Base Path
All booking endpoints are relative to the domain base URL:
```
{BASE_URL}/bookings
```

---

### Create Booking

Creates a new booking with one or more product details.

**Endpoint:**
```
POST {BASE_URL}/bookings
```

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "code": "BKG-2024-001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "total_amount": 150.00,
  "details": [
    {
      "product_id": "660e8400-e29b-41d4-a716-446655440001",
      "product_name": "Premium Package",
      "qty": 2,
      "price_per_unit": 50.00,
      "sub_total": 100.00
    },
    {
      "product_id": "660e8400-e29b-41d4-a716-446655440002",
      "product_name": "Add-on Service",
      "qty": 1,
      "price_per_unit": 50.00,
      "sub_total": 50.00
    }
  ]
}
```

**Request Schema:**

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `code` | string | ✅ Yes | min=3, max=50 | Unique booking code |
| `user_id` | string | ✅ Yes | uuid | UUID of the user creating the booking |
| `total_amount` | number | ✅ Yes | gte=0 | Total booking amount (must match sum of detail subtotals) |
| `details` | array | ✅ Yes | min=1 | Array of booking detail items |
| `details[].product_id` | string | ✅ Yes | uuid_rfc4122 | UUID of the product |
| `details[].product_name` | string | ❌ No | max=100 | Optional product name for display |
| `details[].qty` | integer | ✅ Yes | gt=0 | Quantity (must be positive) |
| `details[].price_per_unit` | number | ✅ Yes | gt=0 | Price per unit (must be positive) |
| `details[].sub_total` | number | ✅ Yes | gt=0 | Subtotal for this line item (qty × price_per_unit) |

**Success Response (201 Created):**
```json
{
  "success": true,
  "message": "Booking created successfully",
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440003",
    "code": "BKG-2024-001",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "total_amount": 150.00,
    "details": [
      {
        "product_id": "660e8400-e29b-41d4-a716-446655440001",
        "product_name": "Premium Package",
        "qty": 2,
        "price_per_unit": 50.00,
        "sub_total": 100.00
      },
      {
        "product_id": "660e8400-e29b-41d4-a716-446655440002",
        "product_name": "Add-on Service",
        "qty": 1,
        "price_per_unit": 50.00,
        "sub_total": 50.00
      }
    ]
  }
}
```

**Error Responses:**

See [Error Codes](#error-codes) section below for complete list.

**Example - Validation Error (400 Bad Request):**
```json
{
  "success": false,
  "message": "Invalid request",
  "error": {
    "code": "INVALID_REQUEST",
    "details": [
      {
        "field": "code",
        "message": "Booking code must be at least 3 characters"
      }
    ]
  }
}
```

**Example - Business Logic Error (400 Bad Request):**
```json
{
  "success": false,
  "message": "total amount does not match with details subtotal",
  "error": {
    "code": "BOOKING_AMOUNT_INCONSISTENT"
  }
}
```

**Example - Duplicate Code (409 Conflict):**
```json
{
  "success": false,
  "message": "booking code already exists",
  "error": {
    "code": "BOOKING_CODE_ALREADY_EXISTS",
    "details": {
      "booking_code": "BKG-2024-001"
    }
  }
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/bookings \
  -H "Content-Type: application/json" \
  -d '{
    "code": "BKG-2024-001",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "total_amount": 150.00,
    "details": [
      {
        "product_id": "660e8400-e29b-41d4-a716-446655440001",
        "product_name": "Premium Package",
        "qty": 2,
        "price_per_unit": 50.00,
        "sub_total": 100.00
      }
    ]
  }'
```

---

## Error Codes

All booking-specific errors use the `BOOKING_*` prefix for easy identification.

### Entity Errors

| Code | Message | HTTP Status | Description |
|------|---------|-------------|-------------|
| `BOOKING_NOT_FOUND` | booking record not found | 404 | The requested booking ID does not exist in the database |
| `BOOKING_CODE_ALREADY_EXISTS` | booking code already exists | 409 | Attempted to create a booking with a code that already exists (duplicate) |

### Validation Errors

| Code | Message | HTTP Status | Description |
|------|---------|-------------|-------------|
| `BOOKING_DETAILS_REQUIRED` | booking must have at least one detail | 400 | The `details` array is empty or missing |
| `BOOKING_AMOUNT_INCONSISTENT` | total amount does not match with details subtotal | 400 | The `total_amount` field does not equal the sum of all detail `sub_total` values |

### Infrastructure Errors

The module also uses standard infrastructure error codes:

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_REQUEST` | 400 | Request validation failed (missing required fields, invalid format, etc.) |
| `MALFORMED_REQUEST` | 400 | Invalid JSON format or data type mismatch |
| `DB_CONFLICT` | 409 | Database constraint violation (unique, foreign key) |
| `DB_CONSTRAINT` | 400 | Database constraint violation (check, not null) |
| `INTERNAL_ERROR` | 500 | Unexpected system error |

---

## Database Schema

### Bookings Table

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | uuid | PRIMARY KEY | Auto-generated booking ID |
| `booking_code` | varchar(50) | NOT NULL, UNIQUE | User-provided or generated booking code |
| `user_id` | uuid | NOT NULL | Reference to user who created the booking |
| `total_amount` | decimal(15,2) | NOT NULL, DEFAULT 0 | Total booking amount |
| `status` | varchar(20) | NOT NULL, DEFAULT 'PENDING' | Current booking status |
| `payment_status` | varchar(20) | NOT NULL, DEFAULT 'UNPAID' | Payment status |
| `created_at` | bigint | NOT NULL | Unix timestamp (milliseconds) |
| `updated_at` | bigint | NULL | Unix timestamp (milliseconds) |
| `deleted_at` | bigint | NULL | Soft delete timestamp |

**Status Values:**
- `PENDING` - Initial state after creation
- `CONFIRMED` - Booking confirmed by user or system
- `CANCELLED` - Booking cancelled
- `COMPLETED` - Booking fulfilled/completed

### Booking Details Table

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | uuid | PRIMARY KEY | Auto-generated detail ID |
| `booking_id` | uuid | NOT NULL, FOREIGN KEY | Reference to parent booking |
| `product_id` | uuid | NOT NULL | Reference to product |
| `product_name` | varchar(100) | NULL | Snapshot of product name at booking time |
| `qty` | integer | NOT NULL | Quantity ordered |
| `price_per_unit` | decimal(15,2) | NOT NULL | Price per unit at booking time |
| `sub_total` | decimal(15,2) | NOT NULL | Calculated subtotal (qty × price_per_unit) |
| `created_at` | bigint | NOT NULL | Unix timestamp (milliseconds) |
| `updated_at` | bigint | NULL | Unix timestamp (milliseconds) |

---

## Business Rules

### 1. Booking Code Uniqueness
- Booking codes must be unique across all bookings
- Attempting to create a booking with an existing code returns `BOOKING_CODE_ALREADY_EXISTS` (409)

### 2. Amount Consistency
- The `total_amount` must exactly match the sum of all detail `sub_total` values
- Validation occurs at the entity level before persistence
- Mismatch returns `BOOKING_AMOUNT_INCONSISTENT` (400)

### 3. Required Details
- Every booking must have at least one detail item
- Empty `details` array returns `BOOKING_DETAILS_REQUIRED` (400)

### 4. Detail Subtotal Calculation
- Each detail's `sub_total` must equal `qty × price_per_unit` (with tolerance of 0.01 for floating-point precision)
- This validation prevents data inconsistency

### 5. Positive Values
- All monetary values (`total_amount`, `price_per_unit`, `sub_total`) must be positive (> 0)
- Quantity (`qty`) must be a positive integer (> 0)

---

## Testing

### Test Coverage

| Layer | Type | Coverage | Location |
|-------|------|----------|----------|
| Entity | Unit | 100% | `test/unit/booking/entity/` |
| UseCase | Unit | 93.1% | `test/unit/booking/usecase/` |
| Handler | Unit | ~90% | `test/unit/booking/handler/` |
| Repository | Integration | 4 tests | `test/integration/booking/` |
| Full Stack | E2E | 7 tests | `test/e2e/booking/` |

**Total: 44 tests** (33 unit + 4 integration + 7 E2E)

### Running Tests

```bash
# Unit tests
go test -v ./test/unit/booking/...

# Integration tests (requires PostgreSQL)
go test -v -tags=integration ./test/integration/booking/...

# E2E tests
go test -v -tags=e2e ./test/e2e/booking/...

# All tests
go test -v -tags=integration,e2e ./test/...
```

See [`test/README.md`](../../../test/README.md) for detailed testing documentation.

---

## Development Notes

### Adding New Endpoints

When adding new endpoints to this module:

1. **Define DTOs** in `usecase/contract.go`
2. **Create UseCase** with interface and implementation
3. **Add Handler** method in `delivery/http/handler.go`
4. **Register Route** in `delivery/http/route.go`
5. **Update this README** with API documentation
6. **Add Tests** (unit, integration, E2E)

### Error Handling Pattern

Follow the established error handling pattern:

```go
// In entity - domain errors
if len(b.Details) == 0 {
    return ErrBookingDetailRequired
}

// In repository - map DB errors
if err != nil {
    return database.MapDBError(err)
}

// In usecase - bubble up errors
if err := booking.Validate(); err != nil {
    log.Warn("booking validation failed", "error", err)
    return nil, err
}

// In handler - let global error handler manage it
if err := h.Uc.CreateBookingUseCase.Execute(ctx, request); err != nil {
    return err  // Will be caught by error handler middleware
}
```

### Observability

This module follows the **Zero-Log Handover** principle:
- **Handlers**: Minimal logging (anchor log only)
- **UseCases**: Business event logging (Warn for business errors, Error for system errors)
- **Repository**: No logging (uses GORM metrics via OpenTelemetry)

All requests are automatically traced with distributed tracing IDs.

---

## Dependencies

**Internal:**
- `internal/pkg/apperror` - Standardized error handling
- `internal/pkg/response` - HTTP response formatting
- `internal/infrastructure/database` - Database connection and utilities
- `internal/infrastructure/validator` - Request validation
- `internal/infrastructure/logger` - Structured logging

**External:**
- `github.com/gofiber/fiber/v2` - HTTP framework
- `gorm.io/gorm` - ORM for database operations

---

## Migration

Database migrations are located in:
```
migrations/booking/
```

Apply migrations:
```bash
migrate -path ./migrations/booking -database "postgres://..." up
```

