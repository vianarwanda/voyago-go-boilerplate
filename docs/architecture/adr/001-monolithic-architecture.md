# ADR-001: Adoption of Modular Monolith Architecture

**Status:** Accepted  
**Date:** 2026-02-07  
**Decision Makers:** Engineering Team  
**Tags:** #architecture #monolith #modular-design

---

## Context

Voyago Core API required an architectural pattern that balances simplicity with scalability. We evaluated:

1. **Microservices Architecture** - Distributed system with independent service deployment
2. **Modular Monolith** - Single deployable unit with strong module boundaries
3. **Traditional Monolith** - Monolithic codebase without domain separation

### Key Factors

- **Product Stage:** Early-stage with evolving requirements
- **Team Structure:** Small to medium-sized engineering team
- **Domain Complexity:** Multiple bounded contexts (Booking, Merchant, Payments, etc.)
- **Tech Stack:** Go (Golang) with Fiber framework, PostgreSQL
- **Traffic:** Moderate traffic expected initially (< 10K RPS)

---

## Decision

**We adopted a Modular Monolith Architecture** with:

- **Single Deployable Unit:** All modules run in one process
- **Strong Module Boundaries:** Each domain module is fully isolated
- **Per-Module Database Schemas:** Independent schema and migrations
- **Clean Architecture per Module:** Entity → Repository → UseCase → Delivery
- **Shared Infrastructure:** HTTP server, database connection, logging, telemetry

---

## Consequences

### Benefits ✅

- **Simplified Operations:** Single binary, no orchestration complexity
- **Fast Development:** No network boundaries, easier debugging and refactoring
- **ACID Transactions:** Strong consistency across related entities
- **Lower Cost:** Single server for low-to-medium traffic
- **Future-Proof:** Module boundaries enable service extraction later

### Trade-offs ⚠️

- **Scaling Constraints:** Must replicate entire app (not individual modules)
  - *Mitigation:* Read replicas, caching, extract hot modules when needed
  
- **Technology Lock-in:** All modules share Go stack
  - *Mitigation:* Module boundaries allow future polyglot migration
  
- **Build Coupling:** Changes require full app redeployment
  - *Mitigation:* Fast CI/CD, feature flags for gradual rollouts
  
- **Single Point of Failure:** Critical bug affects entire system
  - *Mitigation:* Comprehensive testing, circuit breakers

---

## Alternative Considered

### 1. Microservices Architecture

**Rejected Reasons:**
- **Premature Complexity:** Introduces distributed system challenges (network latency, service discovery, distributed tracing) before we have the traffic to justify it
- **Operational Overhead:** Requires sophisticated DevOps (Kubernetes, service mesh, centralized logging)
- **Team Size:** Small team would struggle to manage multiple repositories, deployments, and services
- **Data Consistency:** Distributed transactions and eventual consistency are harder to reason about

**When to Reconsider:**
- Traffic exceeds vertical scaling limits (e.g., > 10K RPS)
- Different modules have drastically different scaling needs
- Team grows to 20+ engineers with dedicated service ownership
- Regulatory/compliance requires physical isolation

### 2. Traditional Monolith (No Module Boundaries)

**Rejected Reasons:**
- **Poor Maintainability:** Spaghetti code risk as the codebase grows
- **Tight Coupling:** Difficult to refactor without breaking unrelated features
- **No Migration Path:** Cannot extract services later without massive rewrite

**Why Modular Monolith is Better:**
- Provides the simplicity of a monolith with the flexibility to evolve
- Enforces separation of concerns through module boundaries
- Allows gradual migration to microservices if needed

---

## Implementation Guidelines

### Module Structure Standard

All modules must follow this structure:

```
internal/modules/{MODULE_NAME}/
├── README.md                   # API documentation (mandatory)
├── delivery/http/              # HTTP handlers and routes
├── entity/                     # Domain entities with validation
├── repository/                 # CQRS repositories (command/query)
│   ├── contract.go
│   ├── command/
│   └── query/
├── usecase/                    # Business logic with DTOs
│   ├── contract.go
│   └── {action}_{entity}.go
└── module.go                   # Dependency injection
```

### Module Communication Options

Modules can communicate through various patterns based on requirements:

1. **Direct UseCase Calls** (In-process, synchronous)
   - Fast and simple for low-coupling scenarios
   - Use dependency injection for testability

2. **gRPC/REST API** (Internal HTTP calls)
   - Prepares modules for future service extraction
   - Useful when network boundary might be needed later

3. **Async Patterns**
   - Fire-and-forget goroutines for background tasks
   - Event Outbox Pattern for reliable async messaging
   - Message queue integration (optional)

4. **Shared Kernel** (Use sparingly)
   - Common utilities in `internal/pkg/`
   - Infrastructure components in `internal/infrastructure/`
   - **Rule:** Do NOT add business logic to shared kernel

5. **Database Isolation** (Mandatory)
   - Each module has its own schema/tables
   - Migrations are per-module (`migrations/{MODULE_NAME}/`)
   - No foreign keys across module boundaries

---

## Migration Path (When to Extract Microservices)

### Extraction Triggers

Extract a module to a microservice when:

1. **Performance Bottleneck:** Module requires independent scaling (e.g., high read traffic)
2. **Team Ownership:** Dedicated team needs full autonomy over deployment
3. **Technology Requirements:** Module benefits from a different tech stack
4. **Regulatory Isolation:** Compliance requires physical separation

### Extraction Process

1. **Evaluate Module Boundaries:** Ensure module has minimal dependencies
2. **Introduce API Gateway:** Add HTTP API between modules
3. **Implement Async Communication:** Use message queue for events
4. **Separate Database:** Deploy module database independently
5. **Deploy Service:** Extract to independent deployment
6. **Monitor & Iterate:** Validate performance and reliability improvements

---

## References

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Modular Monolith: A Primer](https://www.kamilgrzybek.com/blog/posts/modular-monolith-primer)
- [Voyago Core API README](../../README.md)

---

## Decision Review

**Next Review Date:** 2026-08-10 (6 months)  
**Review Criteria:**
- System traffic and performance metrics
- Team size and velocity
- Module coupling metrics
- Deployment frequency and complexity
