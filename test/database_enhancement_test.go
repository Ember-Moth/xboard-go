package test

import (
	"fmt"
	"testing"
	"time"
)

// withRetry executes a database operation with exponential backoff retry mechanism
func withRetry(operation func() error, maxRetries int) error {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := operation(); err != nil {
			lastErr = err
			
			// Check if it's a database lock error that we should retry
			if isDatabaseLockError(err) && attempt < maxRetries {
				// Exponential backoff: 100ms, 200ms, 400ms, 800ms, 1600ms
				backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
				time.Sleep(backoff)
				continue
			}
			
			return err
		}
		
		return nil
	}
	
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

// isDatabaseLockError checks if the error is related to database locking
func isDatabaseLockError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	lockErrors := []string{
		"database is locked",
		"database table is locked",
		"SQLITE_BUSY",
		"SQLITE_LOCKED",
		"database schema has changed",
	}
	
	for _, lockErr := range lockErrors {
		if contains(errStr, lockErr) {
			return true
		}
	}
	
	return false
}

func TestRetryMechanism(t *testing.T) {
	attempts := 0
	maxRetries := 3

	// Test successful operation
	err := withRetry(func() error {
		attempts++
		return nil
	}, maxRetries)

	if err != nil {
		t.Errorf("Expected no error for successful operation, got: %v", err)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt for successful operation, got: %d", attempts)
	}

	// Reset counter
	attempts = 0

	// Test operation that fails then succeeds
	err = withRetry(func() error {
		attempts++
		if attempts < 3 {
			return &mockDatabaseError{message: "database is locked"}
		}
		return nil
	}, maxRetries)

	if err != nil {
		t.Errorf("Expected no error after retries, got: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}
}

// mockDatabaseError simulates a database lock error
type mockDatabaseError struct {
	message string
}

func (e *mockDatabaseError) Error() string {
	return e.message
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr || 
		      containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}