//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"os"
)

// verifyExecutable checks if the file has executable permissions on Unix-like systems
func verifyExecutable(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	mode := fileInfo.Mode()
	// Check if the file has any executable bit set (owner, group, or others)
	if mode&0111 == 0 {
		return fmt.Errorf("file is not executable: permissions are %v", mode)
	}

	return nil
}
