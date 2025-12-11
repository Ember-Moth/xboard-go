package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateDownloadURL(t *testing.T) {
	sv := NewSecurityValidator()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid HTTPS URL",
			url:     "https://download.example.com/agent",
			wantErr: false,
		},
		{
			name:    "HTTP URL should fail",
			url:     "http://download.example.com/agent",
			wantErr: true,
		},
		{
			name:    "empty URL should fail",
			url:     "",
			wantErr: true,
		},
		{
			name:    "FTP URL should fail",
			url:     "ftp://download.example.com/agent",
			wantErr: true,
		},
		{
			name:    "file URL should fail",
			url:     "file:///tmp/agent",
			wantErr: true,
		},
		{
			name:    "malformed URL should fail",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "HTTPS without host should fail",
			url:     "https://",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sv.ValidateDownloadURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDownloadURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDownloadURLWithWhitelist(t *testing.T) {
	whitelist := []string{"download.example.com", "cdn.example.com"}
	sv := NewSecurityValidatorWithWhitelist(whitelist)

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "whitelisted host should pass",
			url:     "https://download.example.com/agent",
			wantErr: false,
		},
		{
			name:    "another whitelisted host should pass",
			url:     "https://cdn.example.com/agent",
			wantErr: false,
		},
		{
			name:    "non-whitelisted host should fail",
			url:     "https://malicious.com/agent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sv.ValidateDownloadURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDownloadURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	sv := NewSecurityValidator()

	tests := []struct {
		name       string
		path       string
		wantErr    bool
		skipOnWin  bool
	}{
		{
			name:    "valid relative path",
			path:    "xboard-agent",
			wantErr: false,
		},
		{
			name:      "valid absolute path",
			path:      "/opt/xboard/agent",
			wantErr:   false,
			skipOnWin: true, // Windows uses different path format
		},
		{
			name:    "path traversal with .. should fail",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:      "path to /etc should fail",
			path:      "/etc/shadow",
			wantErr:   true,
			skipOnWin: true, // Windows doesn't have /etc
		},
		{
			name:      "path to /sys should fail",
			path:      "/sys/kernel/config",
			wantErr:   true,
			skipOnWin: true, // Windows doesn't have /sys
		},
		{
			name:      "path to /proc should fail",
			path:      "/proc/self/mem",
			wantErr:   true,
			skipOnWin: true, // Windows doesn't have /proc
		},
		{
			name:    "empty path should fail",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOnWin && filepath.Separator == '\\' {
				t.Skip("Skipping Unix-specific test on Windows")
			}
			err := sv.ValidateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFilePermissions(t *testing.T) {
	sv := NewSecurityValidator()

	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setupFile   func() string
		wantErr     bool
		errContains string
		skipOnWin   bool
	}{
		{
			name: "executable file should pass",
			setupFile: func() string {
				var path string
				if filepath.Separator == '\\' {
					// Windows: create .exe file
					path = filepath.Join(tmpDir, "executable.exe")
				} else {
					// Unix: create file with executable permissions
					path = filepath.Join(tmpDir, "executable")
				}
				f, _ := os.Create(path)
				f.Close()
				if filepath.Separator != '\\' {
					os.Chmod(path, 0755) // rwxr-xr-x (Unix only)
				}
				return path
			},
			wantErr: false,
		},
		{
			name: "non-executable file should fail",
			setupFile: func() string {
				var path string
				if filepath.Separator == '\\' {
					// Windows: create file with non-.exe extension
					path = filepath.Join(tmpDir, "non-executable.txt")
				} else {
					// Unix: create file without executable permissions
					path = filepath.Join(tmpDir, "non-executable")
				}
				f, _ := os.Create(path)
				f.Close()
				if filepath.Separator != '\\' {
					os.Chmod(path, 0644) // rw-r--r-- (Unix only)
				}
				return path
			},
			wantErr:     true,
			errContains: "executable",
		},
		{
			name: "world-writable file should fail",
			setupFile: func() string {
				path := filepath.Join(tmpDir, "world-writable")
				f, _ := os.Create(path)
				f.Close()
				os.Chmod(path, 0777) // rwxrwxrwx
				return path
			},
			wantErr:     true,
			errContains: "world-writable",
			skipOnWin:   true, // Windows doesn't have world-writable concept
		},
		{
			name: "non-existent file should fail",
			setupFile: func() string {
				return filepath.Join(tmpDir, "non-existent")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOnWin && filepath.Separator == '\\' {
				t.Skip("Skipping Unix-specific test on Windows")
			}
			filePath := tt.setupFile()
			err := sv.ValidateFilePermissions(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateFilePermissions() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	sv := NewSecurityValidator()

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   "abcdef1234567890",
			wantErr: false,
		},
		{
			name:    "long valid token",
			token:   "abcdef1234567890abcdef1234567890",
			wantErr: false,
		},
		{
			name:    "empty token should fail",
			token:   "",
			wantErr: true,
		},
		{
			name:    "short token should fail",
			token:   "short",
			wantErr: true,
		},
		{
			name:    "token with newline should fail",
			token:   "abcdef1234567890\n",
			wantErr: true,
		},
		{
			name:    "token with null byte should fail",
			token:   "abcdef1234567890\x00",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sv.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUpdateInfo(t *testing.T) {
	sv := NewSecurityValidator()

	validUpdateInfo := &UpdateInfo{
		LatestVersion: "v1.2.0",
		DownloadURL:   "https://download.example.com/agent",
		SHA256:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		FileSize:      1024 * 1024, // 1MB
		Strategy:      "auto",
		ReleaseNotes:  "Test release",
	}

	tests := []struct {
		name       string
		updateInfo *UpdateInfo
		token      string
		wantErr    bool
	}{
		{
			name:       "valid update info",
			updateInfo: validUpdateInfo,
			token:      "valid-token-1234",
			wantErr:    false,
		},
		{
			name:       "nil update info should fail",
			updateInfo: nil,
			token:      "valid-token-1234",
			wantErr:    true,
		},
		{
			name: "invalid token should fail",
			updateInfo: validUpdateInfo,
			token:      "short",
			wantErr:    true,
		},
		{
			name: "HTTP URL should fail",
			updateInfo: &UpdateInfo{
				LatestVersion: "v1.2.0",
				DownloadURL:   "http://download.example.com/agent",
				SHA256:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				FileSize:      1024 * 1024,
				Strategy:      "auto",
			},
			token:   "valid-token-1234",
			wantErr: true,
		},
		{
			name: "invalid SHA256 length should fail",
			updateInfo: &UpdateInfo{
				LatestVersion: "v1.2.0",
				DownloadURL:   "https://download.example.com/agent",
				SHA256:        "short-hash",
				FileSize:      1024 * 1024,
				Strategy:      "auto",
			},
			token:   "valid-token-1234",
			wantErr: true,
		},
		{
			name: "zero file size should fail",
			updateInfo: &UpdateInfo{
				LatestVersion: "v1.2.0",
				DownloadURL:   "https://download.example.com/agent",
				SHA256:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				FileSize:      0,
				Strategy:      "auto",
			},
			token:   "valid-token-1234",
			wantErr: true,
		},
		{
			name: "negative file size should fail",
			updateInfo: &UpdateInfo{
				LatestVersion: "v1.2.0",
				DownloadURL:   "https://download.example.com/agent",
				SHA256:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				FileSize:      -1,
				Strategy:      "auto",
			},
			token:   "valid-token-1234",
			wantErr: true,
		},
		{
			name: "file too large should fail",
			updateInfo: &UpdateInfo{
				LatestVersion: "v1.2.0",
				DownloadURL:   "https://download.example.com/agent",
				SHA256:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				FileSize:      600 * 1024 * 1024, // 600MB
				Strategy:      "auto",
			},
			token:   "valid-token-1234",
			wantErr: true,
		},
		{
			name: "invalid strategy should fail",
			updateInfo: &UpdateInfo{
				LatestVersion: "v1.2.0",
				DownloadURL:   "https://download.example.com/agent",
				SHA256:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				FileSize:      1024 * 1024,
				Strategy:      "invalid",
			},
			token:   "valid-token-1234",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sv.ValidateUpdateInfo(tt.updateInfo, tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdateInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBeforeDownload(t *testing.T) {
	sv := NewSecurityValidator()

	tests := []struct {
		name        string
		downloadURL string
		destPath    string
		token       string
		wantErr     bool
		skipOnWin   bool
	}{
		{
			name:        "valid inputs",
			downloadURL: "https://download.example.com/agent",
			destPath:    "/opt/xboard/agent",
			token:       "valid-token-1234",
			wantErr:     false,
			skipOnWin:   true, // Windows uses different path format
		},
		{
			name:        "valid inputs Windows",
			downloadURL: "https://download.example.com/agent",
			destPath:    "C:\\opt\\xboard\\agent.exe",
			token:       "valid-token-1234",
			wantErr:     false,
		},
		{
			name:        "invalid token",
			downloadURL: "https://download.example.com/agent",
			destPath:    "/opt/xboard/agent",
			token:       "short",
			wantErr:     true,
		},
		{
			name:        "invalid URL",
			downloadURL: "http://download.example.com/agent",
			destPath:    "/opt/xboard/agent",
			token:       "valid-token-1234",
			wantErr:     true,
		},
		{
			name:        "invalid path",
			downloadURL: "https://download.example.com/agent",
			destPath:    "/etc/passwd",
			token:       "valid-token-1234",
			wantErr:     true,
			skipOnWin:   true, // Windows doesn't have /etc
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOnWin && filepath.Separator == '\\' {
				t.Skip("Skipping Unix-specific test on Windows")
			}
			err := sv.ValidateBeforeDownload(tt.downloadURL, tt.destPath, tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBeforeDownload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAfterDownload(t *testing.T) {
	sv := NewSecurityValidator()
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setupFile func() string
		wantErr   bool
	}{
		{
			name: "valid executable file",
			setupFile: func() string {
				var path string
				if filepath.Separator == '\\' {
					// Windows: create .exe file
					path = filepath.Join(tmpDir, "valid-agent.exe")
				} else {
					// Unix: create file with executable permissions
					path = filepath.Join(tmpDir, "valid-agent")
				}
				f, _ := os.Create(path)
				f.Close()
				if filepath.Separator != '\\' {
					os.Chmod(path, 0755)
				}
				return path
			},
			wantErr: false,
		},
		{
			name: "non-executable file should fail",
			setupFile: func() string {
				var path string
				if filepath.Separator == '\\' {
					// Windows: create file with non-.exe extension
					path = filepath.Join(tmpDir, "non-exec-agent.txt")
				} else {
					// Unix: create file without executable permissions
					path = filepath.Join(tmpDir, "non-exec-agent")
				}
				f, _ := os.Create(path)
				f.Close()
				if filepath.Separator != '\\' {
					os.Chmod(path, 0644)
				}
				return path
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile()
			err := sv.ValidateAfterDownload(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAfterDownload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


