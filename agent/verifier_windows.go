//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
)

// verifyExecutable checks if the file is a regular file on Windows
// Windows doesn't use Unix-style permission bits, so we just verify it's a regular file
func verifyExecutable(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	mode := fileInfo.Mode()
	if !mode.IsRegular() {
		return fmt.Errorf("file is not a regular file: mode is %v", mode)
	}

	return nil
}
