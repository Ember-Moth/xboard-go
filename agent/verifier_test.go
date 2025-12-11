package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifySize(t *testing.T) {
	verifier := NewFileVerifier()

	// Create a temporary file with known content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bin")
	content := []byte("Hello, World!")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		expectedSize int64
		wantErr      bool
	}{
		{
			name:         "correct size",
			expectedSize: int64(len(content)),
			wantErr:      false,
		},
		{
			name:         "incorrect size - too small",
			expectedSize: int64(len(content)) - 1,
			wantErr:      true,
		},
		{
			name:         "incorrect size - too large",
			expectedSize: int64(len(content)) + 1,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifier.VerifySize(testFile, tt.expectedSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifySize_NonExistentFile(t *testing.T) {
	verifier := NewFileVerifier()
	err := verifier.VerifySize("/nonexistent/file.bin", 100)
	if err == nil {
		t.Error("VerifySize() should return error for non-existent file")
	}
}

func TestVerifySHA256(t *testing.T) {
	verifier := NewFileVerifier()

	// Create a temporary file with known content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bin")
	content := []byte("Hello, World!")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate the correct hash
	hasher := sha256.New()
	hasher.Write(content)
	correctHash := hex.EncodeToString(hasher.Sum(nil))

	tests := []struct {
		name         string
		expectedHash string
		wantErr      bool
	}{
		{
			name:         "correct hash",
			expectedHash: correctHash,
			wantErr:      false,
		},
		{
			name:         "incorrect hash",
			expectedHash: "0000000000000000000000000000000000000000000000000000000000000000",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifier.VerifySHA256(testFile, tt.expectedHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySHA256() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifySHA256_NonExistentFile(t *testing.T) {
	verifier := NewFileVerifier()
	err := verifier.VerifySHA256("/nonexistent/file.bin", "abc123")
	if err == nil {
		t.Error("VerifySHA256() should return error for non-existent file")
	}
}

func TestVerifyExecutable(t *testing.T) {
	verifier := NewFileVerifier()
	tmpDir := t.TempDir()

	// Create a regular file
	testFile := filepath.Join(tmpDir, "test.bin")
	if err := os.WriteFile(testFile, []byte("test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// On all platforms, a regular file should pass
	err := verifier.VerifyExecutable(testFile)
	if err != nil {
		t.Errorf("VerifyExecutable() should pass for regular file, got error: %v", err)
	}
}

func TestVerifyExecutable_NonExistentFile(t *testing.T) {
	verifier := NewFileVerifier()
	err := verifier.VerifyExecutable("/nonexistent/file.bin")
	if err == nil {
		t.Error("VerifyExecutable() should return error for non-existent file")
	}
}

func TestVerifyAll(t *testing.T) {
	verifier := NewFileVerifier()
	tmpDir := t.TempDir()

	// Create a test file with known content and executable permissions
	testFile := filepath.Join(tmpDir, "test.bin")
	content := []byte("Test content for verification")
	if err := os.WriteFile(testFile, content, 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate correct hash
	hasher := sha256.New()
	hasher.Write(content)
	correctHash := hex.EncodeToString(hasher.Sum(nil))

	tests := []struct {
		name         string
		expectedSize int64
		expectedHash string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "all checks pass",
			expectedSize: int64(len(content)),
			expectedHash: correctHash,
			wantErr:      false,
		},
		{
			name:         "size mismatch",
			expectedSize: int64(len(content)) + 1,
			expectedHash: correctHash,
			wantErr:      true,
			errContains:  "size verification failed",
		},
		{
			name:         "hash mismatch",
			expectedSize: int64(len(content)),
			expectedHash: "0000000000000000000000000000000000000000000000000000000000000000",
			wantErr:      true,
			errContains:  "hash verification failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifier.VerifyAll(testFile, tt.expectedSize, tt.expectedHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("VerifyAll() error should contain %q, got %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestVerifyAll_NonExecutableFile(t *testing.T) {
	// This test only makes sense on Unix-like systems
	if os.PathSeparator == '\\' {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	verifier := NewFileVerifier()
	tmpDir := t.TempDir()

	// Create a test file without executable permissions
	testFile := filepath.Join(tmpDir, "test.bin")
	content := []byte("Test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate correct hash
	hasher := sha256.New()
	hasher.Write(content)
	correctHash := hex.EncodeToString(hasher.Sum(nil))

	err := verifier.VerifyAll(testFile, int64(len(content)), correctHash)
	if err == nil {
		t.Error("VerifyAll() should fail for non-executable file")
	}
	if !contains(err.Error(), "executable verification failed") {
		t.Errorf("VerifyAll() error should mention executable verification, got: %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestVerifyExecutable_NonExecutableOnUnix(t *testing.T) {
	// This test only makes sense on Unix-like systems
	if os.PathSeparator == '\\' {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	verifier := NewFileVerifier()
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		mode    os.FileMode
		wantErr bool
	}{
		{
			name:    "executable by owner",
			mode:    0744,
			wantErr: false,
		},
		{
			name:    "executable by group",
			mode:    0754,
			wantErr: false,
		},
		{
			name:    "executable by others",
			mode:    0755,
			wantErr: false,
		},
		{
			name:    "not executable",
			mode:    0644,
			wantErr: true,
		},
		{
			name:    "read-only",
			mode:    0444,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.name+".bin")
			if err := os.WriteFile(testFile, []byte("test"), tt.mode); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			err := verifier.VerifyExecutable(testFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyExecutable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
