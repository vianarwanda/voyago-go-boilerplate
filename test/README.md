# Test Setup Guide

This guide explains how to set up and run the test suite.

## Test Structure

```
test/
├── unit/           # Fast, isolated tests with mocks
├── integration/    # Medium-speed tests with real database
├── e2e/           # Full-stack HTTP tests
└── helper/        # Shared test utilities
```

## Prerequisites

### For All Tests
- Go 1.25.7
- PostgreSQL 16 (for integration/e2e tests)

### Test Database Setup

1. **Create test database**:
   ```bash
   createdb voyago_test
   ```

2. **Run migrations** on test database:
   ```bash
   # Use your migration tool to create tables in voyago_test
   # Example: migrate -path ./migrations -database "postgresql://..." up
   ```

3. **Configure environment variables**:
   ```bash
   # Copy the example file
   cp .env.test.example .env.test
   
   # Edit .env.test with your actual credentials
   # This file is gitignored for security
   ```

4. **Set environment variables** before running tests:
   ```bash
   # Option 1: Export manually
   export TEST_DB_PASSWORD=your_password_here
   
   # Option 2: Use direnv or similar tool
   # Option 3: Load from .env.test in your test runner
   ```

## Running Tests

### Unit Tests (Default, Fast)
```bash
# Run all unit tests
go test ./test/unit/...

# With coverage
go test -coverprofile=coverage.out ./test/unit/...
go tool cover -html=coverage.out
```

### Integration Tests (Requires DB)
```bash
# Run integration tests
TEST_DB_PASSWORD=your_password go test -tags=integration -v ./test/integration/...

# Or with .env.test loaded
go test -tags=integration -v ./test/integration/...
```

### E2E Tests (Full Stack)
```bash
# Run E2E tests
TEST_DB_PASSWORD=your_password go test -tags=e2e -v ./test/e2e/...
```

### All Tests
```bash
# Run everything
TEST_DB_PASSWORD=your_password go test -tags="integration e2e" -v ./test/...
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `TEST_DB_HOST` | Database host | `localhost` | No |
| `TEST_DB_PORT` | Database port | `5432` | No |
| `TEST_DB_USER` | Database user | `booking_user` | No |
| `TEST_DB_PASSWORD` | Database password | _(empty)_ | **Yes** |
| `TEST_DB_NAME` | Database name | `voyago_test` | No |

## Security Note

⚠️ **NEVER commit `.env.test` to git!**

- `.env.test` is in `.gitignore` to prevent credential leaks
- Use `.env.test.example` as a template
- Each developer should create their own `.env.test` locally

## Test Coverage Goals

- **Unit tests**: 90%+ (current: 93%+)
- **Integration tests**: 70%+ repository layer
- **E2E tests**: 100% critical user flows

## Troubleshooting

### "Failed to ping test database"
- Ensure PostgreSQL is running
- Check `TEST_DB_*` environment variables
- Verify test database exists: `psql -l | grep voyago_test`

### "No such table"
- Run migrations on test database
- Ensure you're using `voyago_test` not production DB

### Build tag tests not running
- Use `-tags` flag: `-tags=integration` or `-tags=e2e`
- Check file has correct build tag: `//go:build integration`
