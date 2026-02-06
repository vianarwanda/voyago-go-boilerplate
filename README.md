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

Every module **must** include `contract.go` files in both `repository` and `usecase` directories. These files serve as the "Public API" of each layer.

#### Repository Contracts (CQRS Pattern)
```go
// repository/contract.go
type BookingCommandRepository interface {
    Create(ctx context.Context, booking *entity.Booking) error
}

type BookingQueryRepository interface {
    FindByID(ctx context.Context, id string) (*entity.Booking, error)
}
```

#### UseCase Contracts
DTOs (Request/Response) and the UseCase interface itself must be defined here.
```go
// usecase/contract.go
type CreateBookingRequest struct { ... }
type CreateBookingResponse struct { ... }

type CreateBookingUseCase interface {
    Execute(ctx context.Context, req *CreateBookingRequest) (*CreateBookingResponse, error)
}
```

**Rationale:**
- **Decoupling** — Layers depend on abstractions, not implementations.
- **Testability** — Enables seamless mocking for unit tests.
- **Clarity** — All public-facing structures are easily discoverable in one file.

---

### 2. Implementation Naming & Compliance (Mandatory)

To avoid naming collisions and strictly enforce the use of interfaces, implementation structs **must** be private (lowercase), while the factory functions and interfaces remain public.

```go
// usecase/create_booking.go

// 1. Interface Compliance Check
var _ CreateBookingUseCase = (*createBookingUseCase)(nil)

// 2. Private Implementation
type createBookingUseCase struct {
    Repo repository.BookingCommandRepository
}

// 3. Public Factory (returns the interface)
func NewCreateBookingUseCase(...) CreateBookingUseCase {
    return &createBookingUseCase{...}
}
```

---

### 3. DTO & Data Flow Standards (Mandatory)

We strictly separate Domain Entities from external API contracts.

**The Flow:**
1. **Handler** receives an HTTP request and parses it into a **Request DTO**.
2. **UseCase** receives the **Request DTO**, processes it using **Entities**, and persists via **Repository**.
3. **Repository** returns **Entities** to the UseCase.
4. **UseCase** MUST map the **Entities** back into a **Response DTO** before returning to the Handler.

**Rationale:**
- Prevents database schema leaks to the API.
- Allows the internal domain to evolve independently of the external contract.
- Ensures the Handler (Upstream) only deals with cleaned, formatted data.

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
│   ├── contract.go             # ⭐ MANDATORY: Interface & DTO definitions
│   └── {action}_{entity}.go    # Implementation of the business logic
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
| - Interface-First: UseCases MUST be defined as interfaces in contract.go.
| - Traceability, Observability, Validation, Atomicity, Side Effects.
|
| [2. LOGGING OPERATIONAL SCOPE]
| - MINIMAL LOGS: "started" and "completed/failed" only.
| - ERROR BUBBLING: Do not log downstream errors.
|
| [3. STANDARD ERROR HANDLING]
| - RECORD → ENRICH → LOG → BUBBLE → HALT.
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
