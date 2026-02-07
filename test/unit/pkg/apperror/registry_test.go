// Package apperror_test contains unit tests for the application's
// standardized error handling system.
//
// registry_test.go specifically verifies the Modular Error Registry mechanism,
// which allows domain-driven error codes to be mapped to specific HTTP status codes
// without polluting the domain layer with infrastructure concerns.
package apperror_test

import (
	"testing"
	"voyago/core-api/internal/pkg/apperror"
)

func TestRegisterStatus(t *testing.T) {
	code := "CUSTOM_TEST_ERROR"
	expectedStatus := 418 // I'm a teapot

	apperror.RegisterStatus(code, expectedStatus)

	err := &apperror.AppError{
		Code: code,
		Kind: apperror.KindPersistance,
	}

	if status := err.GetHttpStatus(); status != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, status)
	}
}

func TestGetHttpStatusFallback(t *testing.T) {
	err := &apperror.AppError{
		Code: "UNREGISTERED_ERROR",
		Kind: apperror.KindTransient,
	}

	expectedStatus := 503 // Fallback for KindTransient
	if status := err.GetHttpStatus(); status != expectedStatus {
		t.Errorf("expected fallback status %d, got %d", expectedStatus, status)
	}
}
