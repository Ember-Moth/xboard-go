// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// validateFilePermissionsImpl is the Windows-specific implementation
func validateFilePermissionsImpl(filePath string, fileInfo os.FileInfo) error {
	// On Windows, we check if the file has .exe extension or no extension
	// Files without extensions are allowed (for Unix compatibility)
	// Files with other extensions are not considered executable
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// If file has an extension, it must be .exe
	if ext != "" && ext != ".exe" {
		return fmt.Errorf("file does not have .exe extension: %s", filePath)
	}

	// Check if file is readable
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("file is not readable: %w", err)
	}
	file.Close()

	return nil
}
