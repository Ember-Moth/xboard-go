package main

import (
	"errors"
	"testing"
)

func TestUpdateError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *UpdateError
		expected string
	}{
		{
			name: "network error with underlying error",
			err: &UpdateError{
				Category:  ErrorCategoryNetwork,
				Message:   "connection timeout",
				Retryable: true,
				Err:       errors.New("dial tcp: timeout"),
			},
			expected: "[network] connection timeout: dial tcp: timeout",
		},
		{
			name: "file error without underlying error",
			err: &UpdateError{
				Category:  ErrorCategoryFile,
				Message:   "disk full",
				Retryable: false,
				Err:       nil,
			},
			expected: "[file] disk full",
		},
		{
			name: "verification error",
			err: &UpdateError{
				Category:  ErrorCategoryVerification,
				Message:   "SHA256 mismatch",
				Retryable: false,
				Err:       errors.New("expected abc, got def"),
			},
			expected: "[verification] SHA256 mismatch: expected abc, got def",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestUpdateError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	updateErr := &UpdateError{
		Category:  ErrorCategoryNetwork,
		Message:   "test",
		Retryable: true,
		Err:       originalErr,
	}

	unwrapped := updateErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestNewNetworkError(t *testing.T) {
	originalErr := errors.New("connection refused")
	err := NewNetworkError("failed to connect", originalErr)

	if err.Category != ErrorCategoryNetwork {
		t.Errorf("Category = %v, want %v", err.Category, ErrorCategoryNetwork)
	}
	if err.Message != "failed to connect" {
		t.Errorf("Message = %v, want %v", err.Message, "failed to connect")
	}
	if !err.Retryable {
		t.Error("Retryable should be true for network errors")
	}
	if err.Err != originalErr {
		t.Errorf("Err = %v, want %v", err.Err, originalErr)
	}
}

func TestNewFileError(t *testing.T) {
	originalErr := errors.New("permission denied")
	err := NewFileError("failed to write file", originalErr)

	if err.Category != ErrorCategoryFile {
		t.Errorf("Category = %v, want %v", err.Category, ErrorCategoryFile)
	}
	if err.Retryable {
		t.Error("Retryable should be false for file errors")
	}
}

func TestNewVerificationError(t *testing.T) {
	originalErr := errors.New("hash mismatch")
	err := NewVerificationError("verification failed", originalErr)

	if err.Category != ErrorCategoryVerification {
		t.Errorf("Category = %v, want %v", err.Category, ErrorCategoryVerification)
	}
	if err.Retryable {
		t.Error("Retryable should be false for verification errors")
	}
}

func TestNewUpdateError(t *testing.T) {
	originalErr := errors.New("rollback failed")
	err := NewUpdateError("update failed", originalErr)

	if err.Category != ErrorCategoryUpdate {
		t.Errorf("Category = %v, want %v", err.Category, ErrorCategoryUpdate)
	}
	if err.Retryable {
		t.Error("Retryable should be false for update errors")
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectedRetry   bool
	}{
		{
			name:          "nil error",
			err:           nil,
			expectedRetry: false,
		},
		{
			name:          "network error (retryable)",
			err:           NewNetworkError("connection timeout", errors.New("timeout")),
			expectedRetry: true,
		},
		{
			name:          "file error (not retryable)",
			err:           NewFileError("disk full", errors.New("no space")),
			expectedRetry: false,
		},
		{
			name:          "verification error (not retryable)",
			err:           NewVerificationError("hash mismatch", errors.New("mismatch")),
			expectedRetry: false,
		},
		{
			name:          "update error (not retryable)",
			err:           NewUpdateError("rollback failed", errors.New("failed")),
			expectedRetry: false,
		},
		{
			name:          "unknown error type",
			err:           errors.New("unknown error"),
			expectedRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleError(tt.err)
			if result != tt.expectedRetry {
				t.Errorf("HandleError() = %v, want %v", result, tt.expectedRetry)
			}
		})
	}
}

func TestErrorCategories(t *testing.T) {
	// Test that all error categories are defined correctly
	categories := []UpdateErrorCategory{
		ErrorCategoryNetwork,
		ErrorCategoryFile,
		ErrorCategoryVerification,
		ErrorCategoryUpdate,
		ErrorCategoryUnknown,
	}

	expectedValues := []string{
		"network",
		"file",
		"verification",
		"update",
		"unknown",
	}

	for i, cat := range categories {
		if string(cat) != expectedValues[i] {
			t.Errorf("Category %d = %v, want %v", i, cat, expectedValues[i])
		}
	}
}
