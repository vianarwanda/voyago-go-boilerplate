# Voyago Core API

> **Modular Monolith Architecture** — Production-ready Clean Architecture with Domain Isolation.

## Overview

Voyago Core API is the backend service for the Voyago platform. This project implements a **Modular Monolith** architecture where each domain (module) maintains full isolation while running as a single binary.

---

## Project Structure

```
voyago/core-api/
├── cmd/
│   └── api/                    # Application entry point
├── config/
│   ├── config.yaml             # Global configuration (server, telemetry)
│   └── {MODULE_NAME}/          # Per-module configuration (database, logging)
├── migrations/
│   └── {MODULE_NAME}/          # SQL migrations per module
├── internal/
│   ├── app/                    # Application bootstrap
│   ├── infrastructure/         # Shared infrastructure (db, logger, telemetry, validator)
│   ├── modules/                # ⭐ DOMAIN MODULES (development team focus)
│   │   └── {MODULE_NAME}/
│   │       ├── delivery/       # HTTP handlers and routes
│   │       ├── entity/         # Domain entities and validation
│   │       ├── repository/     # Data access layer (CQRS: command & query)
│   │       ├── usecase/        # Business logic and DTOs
│   │       └── module.go       # Dependency injection
│   └── pkg/                    # Shared packages (apperror, response, utils)
└── logs/                       # Per-module log files
```

---

## Team Responsibilities

Development teams should **focus exclusively** on the following directories based on their assigned domain:

| Focus Area | Path | Description |
|------------|------|-------------|
| **Domain Logic** | `./internal/modules/{MODULE_NAME}/` | All business logic implementation |
| **Database Migrations** | `./migrations/{MODULE_NAME}/` | SQL up/down migration scripts |
| **Module Configuration** | `./config/{MODULE_NAME}/` | Database and logging configuration |

### Restricted Areas

The following directories are maintained by the core team and should not be modified:

- `./internal/infrastructure/` — Core infrastructure components
- `./internal/pkg/` — Shared utility packages
- `./internal/app/` — Application bootstrap logic

---

## Architectural Standards

### 1. Interface Definitions in `contract.go` (Mandatory)

Every module **must** include a `repository/contract.go` file that defines interfaces following the **CQRS** (Command Query Responsibility Segregation) pattern:

```go
// repository/contract.go
package repository

// -------- Command Repository --------
type BookingCommandRepository interface {
    Create(ctx context.Context, booking *entity.Booking) error
    Update(ctx context.Context, booking *entity.Booking) error
    Delete(ctx context.Context, booking *entity.Booking) error
}

// -------- Query Repository --------
type BookingQueryRepository interface {
    FindByID(ctx context.Context, id string) (*entity.Booking, error)
    ExistsByBookingCode(ctx context.Context, code string) (bool, error)
}
```

**Rationale:**
- **Dependency Injection** — UseCases depend on interfaces, not implementations
- **Testability** — Enables straightforward mocking for unit tests
- **Compile-time Safety** — Interface compliance verification via `var _ Interface = (*Impl)(nil)`

---

### 2. Data Transfer Objects (DTOs) in UseCase Files (Mandatory)

Each usecase **must** define its request DTO within the same file:

```go
// usecase/create_booking.go
package usecase

// -------- DTO --------
type CreateBookingRequest struct {
    BookingCode string  `json:"code" validate:"required,min=3,max=50"`
    UserID      string  `json:"user_id" validate:"required,uuid"`
    TotalAmount float64 `json:"total_amount" validate:"gte=0"`
    Details     []CreateBookingDetailRequest `json:"details" validate:"required,min=1,dive"`
}

type CreateBookingDetailRequest struct {
    ProductID    string  `json:"product_id" validate:"required,uuid_rfc4122"`
    Qty          int32   `json:"qty" validate:"required,gt=0"`
    PricePerUnit float64 `json:"price_per_unit" validate:"required,gt=0"`
}
```

**Validation Tags Reference:**
| Tag | Description |
|-----|-------------|
| `required` | Field is mandatory |
| `uuid` | Must be a valid UUID format |
| `min=N,max=N` | String length constraints |
| `gte=N`, `gt=N` | Numeric value constraints |
| `dive` | Validates nested array elements |

---

### 3. Entity with Domain Validation (Mandatory)

Entities **must** include:
- Domain-specific error codes
- A `Validate()` method for business rule enforcement
- GORM tags for database mapping

```go
// entity/booking.go
package entity

// Domain-specific error codes
const (
    CodeBookingNotFound          = "booking.not_found"
    CodeBookingCodeAlreadyExists = "booking.booking_code.already_exists"
)

var ErrBookingNotFound = apperror.NewPermanent(
    CodeBookingNotFound,
    "booking record not found",
)

type Booking struct {
    ID          string `gorm:"column:id;type:uuid;primaryKey"`
    BookingCode string `gorm:"column:booking_code;type:varchar(50);not null;unique"`
    // ... additional fields
}

func (e *Booking) Validate() error {
    // Domain validation logic
    if len(e.Details) == 0 {
        return ErrBookingDetailRequired
    }
    return nil
}
```

> [!TIP]
> If no domain validation is required for an entity, simply implement the method with `return nil`.

---

## Module Structure Template

When creating a new module, adhere to the following structure:

```
internal/modules/{MODULE_NAME}/
├── delivery/
│   └── http/
│       ├── handler.go          # HTTP request handlers
│       └── route.go            # Route definitions
├── entity/
│   └── {entity}.go             # Domain entities with Validate() method
├── repository/
│   ├── contract.go             # ⭐ MANDATORY: Interface definitions (CQRS)
│   ├── command/
│   │   └── {entity}.go         # Write operations implementation
│   └── query/
│       └── {entity}.go         # Read operations implementation
├── usecase/
│   └── {action}_{entity}.go    # Business logic with DTOs
└── module.go                   # Dependency injection and module registration
```

---

## Documentation Standards

All handler and usecase files **must** include documentation headers that outline architectural standards and observability guidelines.

### Handler Documentation Template:
```go
/*
|------------------------------------------------------------------------------------
| HTTP HANDLER ARCHITECTURAL STANDARDS & OBSERVABILITY MANIFESTO
|------------------------------------------------------------------------------------
|
| [1. THE SINGLE LOG RULE]
| - Every handler execution MUST emit exactly ONE "Anchor Log"
|
| [2. ZERO POST-ENTRY LOGGING]
| - Once the request is delegated to UseCase, Handler MUST NOT emit logs
|
| [3. LEAN ORCHESTRATION]
| - Validation → Parsing → Error Bubbling (to Global Error Handler)
|
| [4. RESPONSE NORMALIZATION]
| - Always use the standardized 'response' package
|------------------------------------------------------------------------------------
*/
```

### UseCase Documentation Template:
```go
/*
|------------------------------------------------------------------------------------
| USECASE ARCHITECTURAL STANDARDS & OBSERVABILITY MANIFESTO
|------------------------------------------------------------------------------------
|
| [1. COMPLIANCE STANDARDS]
| - Traceability, Observability, Validation, Atomicity, Side Effects
|
| [2. LOGGING OPERATIONAL SCOPE]
| - MINIMAL LOGS: "started" and "completed/failed" only
| - ERROR BUBBLING: Do not log downstream errors
|
| [3. STANDARD ERROR HANDLING]
| - RECORD → ENRICH → LOG → BUBBLE → HALT
|------------------------------------------------------------------------------------
*/
```

---

## Getting Started

### Prerequisites
- Go 1.25.7
- PostgreSQL 16
- golang-migrate (for database migrations)

### Running the Application

```bash
# Install dependencies
go mod download

# Run database migrations (per module)
migrate -path ./migrations/booking -database "postgres://..." up

# Start the API server
go run ./cmd/api/main.go
```

### Configuration

1. **Global configuration**: `./config/config.yaml`
2. **Module configuration**: `./config/{MODULE_NAME}/config.yaml`

**Setup:** Copy the example configuration files before running:
```bash
cp config/booking/config.example.yaml config/booking/config.yaml
cp config/merchant/config.example.yaml config/merchant/config.yaml
```

Environment variables use the `${VAR_NAME:default_value}` syntax:
```yaml
database:
  host: ${DB_HOST:localhost}
  port: ${DB_PORT:5432}
  user: ${DB_USER:postgres}
  password: ${DB_PASSWORD:postgres}
```

> [!NOTE]
> `config.yaml` files are git-ignored. Only `config.example.yaml` templates are committed.

---

## Reference Implementation

The **`booking`** module serves as the complete reference implementation. Use it as a template for new modules:

| File | Purpose |
|------|---------|
| `booking/module.go` | Dependency injection pattern |
| `booking/delivery/http/handler.go` | Handler with observability standards |
| `booking/usecase/create_booking.go` | UseCase with DTOs and error handling |
| `booking/repository/contract.go` | CQRS interface definitions |
| `booking/entity/booking.go` | Entity with domain validation |

---

## Contributing

1. Create a branch from `main`: `git checkout -b feature/{MODULE_NAME}/{feature-name}`
2. Follow the architectural standards outlined above
3. Ensure all interface compliance checks pass
4. Submit a Pull Request for review

---

## Contact

For architectural questions or clarifications, please contact the core maintainers.
