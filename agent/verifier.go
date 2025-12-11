package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// FileVerifier handles file verification operations
type FileVerifier struct{}

// NewFileVerifier creates a new FileVerifier instance
func NewFileVerifier() *FileVerifier {
	return &FileVerifier{}
}

// VerifySize verifies that the file size matches the expected size
func (fv *FileVerifier) VerifySize(filePath string, expectedSize int64) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	actualSize := fileInfo.Size()
	if actualSize != expectedSize {
		return fmt.Errorf("file size mismatch: expected %d bytes, got %d bytes", expectedSize, actualSize)
	}

	return nil
}

// VerifySHA256 verifies that the file's SHA256 hash matches the expected hash
func (fv *FileVerifier) VerifySHA256(filePath, expectedHash string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("SHA256 hash mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

// VerifyExecutable verifies that the file has executable permissions
func (fv *FileVerifier) VerifyExecutable(filePath string) error {
	return verifyExecutable(filePath)
}

// VerifyAll performs all verification checks on the file
func (fv *FileVerifier) VerifyAll(filePath string, expectedSize int64, expectedHash string) error {
	// Verify file size
	if err := fv.VerifySize(filePath, expectedSize); err != nil {
		return fmt.Errorf("size verification failed: %w", err)
	}

	// Verify SHA256 hash
	if err := fv.VerifySHA256(filePath, expectedHash); err != nil {
		return fmt.Errorf("hash verification failed: %w", err)
	}

	// Verify executable permissions
	if err := fv.VerifyExecutable(filePath); err != nil {
		return fmt.Errorf("executable verification failed: %w", err)
	}

	return nil
}
