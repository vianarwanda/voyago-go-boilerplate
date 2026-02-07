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
  "message": "Validation failed",
  "error_code": "INVALID_REQUEST",
  "errors": [
    {
      "code": "required",
      "field": "code",
      "message": "Booking code is required",
      "param": ""
    }
  ],
  "trace_id": "bace8705956301997fceea98ef5deb91"
}
```

**Example - Business Logic Error (400 Bad Request):**
```json
{
  "success": false,
  "message": "invalid subtotal for product 019c3162-f0e3-71d7-8aae-7a96c11a79bc",
  "error_code": "BOOKING_AMOUNT_INCONSISTENT",
  "trace_id": "bace8705956301997fceea98ef5deb91"
}
```

**Example - Duplicate Code (409 Conflict):**
```json
{
    "success": false,
    "message": "booking code already exists",
    "error_code": "BOOKING_CODE_ALREADY_EXISTS",
    "trace_id": "bace8705956301997fceea98ef5deb91"
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

| Code                          | Message                     | HTTP Status | Description                                                              |
|-------------------------------|-----------------------------|-------------|--------------------------------------------------------------------------|
| `BOOKING_NOT_FOUND`           | booking record not found    | 404         | The requested booking ID does not exist in the database                  |
| `BOOKING_CODE_ALREADY_EXISTS` | booking code already exists | 409         | Attempted to create a booking with a code that already exists (duplicate)|

### Validation Errors

| Code                          | Message                                           | HTTP Status | Description                                                                      |
|-------------------------------|---------------------------------------------------|-------------|----------------------------------------------------------------------------------|
| `BOOKING_DETAILS_REQUIRED`             | booking must have at least one detail                 | 400         | The `details` array is empty or missing                                          |
| `BOOKING_AMOUNT_INCONSISTENT`          | total amount does not match with details subtotal     | 400         | The `total_amount` field does not equal the sum of all detail `sub_total` values |
| `BOOKING_DETAIL_SUBTOTAL_INCONSISTENT` | detail subtotal does not match with expected subtotal | 400         | A detail item's subtotal does not match its quantity multiplied by price         |

### Infrastructure Errors
> Common infrastructure errors (e.g., `INVALID_REQUEST`, `INTERNAL_ERROR`) are documented in the [Root README](../../../../README.md#infrastructure-error-codes).

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
