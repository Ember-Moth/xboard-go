// +build !windows

package main

import (
	"fmt"
	"os"
)

// validateFilePermissionsImpl is the Unix-specific implementation
func validateFilePermissionsImpl(filePath string, fileInfo os.FileInfo) error {
	mode := fileInfo.Mode()

	// Check if file has executable permissions for owner
	if mode.Perm()&0100 == 0 {
		return fmt.Errorf("file is not executable: %s", filePath)
	}

	// Check if file is world-writable (security risk)
	if mode.Perm()&0002 != 0 {
		return fmt.Errorf("file is world-writable (security risk): %s", filePath)
	}

	return nil
}
