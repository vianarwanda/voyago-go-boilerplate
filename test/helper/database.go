package helper

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"voyago/core-api/internal/infrastructure/config"
	database "voyago/core-api/internal/infrastructure/db"
	"voyago/core-api/internal/infrastructure/logger"
	"voyago/core-api/internal/infrastructure/telemetry/tracer"

	"gorm.io/gorm"
)

// TestDatabaseConfig holds test database configuration
type TestDatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// DefaultTestDBConfig returns test database configuration from environment variables
// with fallback to sensible defaults for local development
func DefaultTestDBConfig() *TestDatabaseConfig {
	port := 5432
	if portStr := os.Getenv("TEST_DB_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	return &TestDatabaseConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     port,
		User:     getEnv("TEST_DB_USER", "booking_user"),
		Password: getEnv("TEST_DB_PASSWORD", ""), // MUST be set via env var
		DBName:   getEnv("TEST_DB_NAME", "voyago_test"),
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// SetupTestDB creates a test database connection
// It should be called at the beginning of each integration test
func SetupTestDB(t *testing.T) database.Database {
	t.Helper()

	cfg := DefaultTestDBConfig()

	dbCfg := &config.DatabaseConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Name:     cfg.DBName,
		Pool: struct {
			Idle     int `mapstructure:"idle"`
			Max      int `mapstructure:"max"`
			Lifetime int `mapstructure:"lifetime"`
		}{
			Idle:     5,
			Max:      20,
			Lifetime: 300,
		},
	}

	// Use NoOp logger and tracer for tests
	log := logger.NewNoOpLogger()
	trc := tracer.NewNoOpTracer()

	db := database.NewDatabase(dbCfg, log, trc)

	// Verify connection
	sqlDB, err := db.GetDB().DB()
	if err != nil {
		t.Fatalf("Failed to get underlying DB: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping test database: %v. "+
			"Make sure test database '%s' exists and is accessible at %s:%d",
			err, cfg.DBName, cfg.Host, cfg.Port)
	}

	return db
}

// CleanupTestDB closes the database connection
func CleanupTestDB(t *testing.T, db database.Database) {
	t.Helper()

	if err := db.Close(); err != nil {
		t.Errorf("Failed to close test database: %v", err)
	}
}

// TruncateTable truncates a specific table for cleanup
func TruncateTable(t *testing.T, db *gorm.DB, tableName string) {
	t.Helper()

	if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)).Error; err != nil {
		t.Errorf("Failed to truncate table %s: %v", tableName, err)
	}
}

// TruncateTables truncates multiple tables for cleanup
func TruncateTables(t *testing.T, db *gorm.DB, tableNames ...string) {
	t.Helper()

	for _, tableName := range tableNames {
		TruncateTable(t, db, tableName)
	}
}

// WithTestTransaction wraps a test function in a transaction that will be rolled back
// This ensures test isolation - changes made during the test won't persist
func WithTestTransaction(t *testing.T, db database.Database, testFn func(tx *gorm.DB)) {
	t.Helper()

	err := db.GetDB().Transaction(func(tx *gorm.DB) error {
		// Run the test function with the transaction
		testFn(tx)

		// Always rollback the transaction to maintain test isolation
		return fmt.Errorf("rollback test transaction")
	})

	// We expect the error since we always rollback
	if err != nil && err.Error() != "rollback test transaction" {
		t.Fatalf("Unexpected error in test transaction: %v", err)
	}
}
