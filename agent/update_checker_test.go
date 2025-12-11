package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateChecker_CheckUpdate(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path
		if r.URL.Path != "/api/v1/agent/version" {
			t.Errorf("Expected path /api/v1/agent/version, got %s", r.URL.Path)
		}

		// Verify Authorization header
		if r.Header.Get("Authorization") != "test-token-1234567890" {
			t.Errorf("Expected Authorization header 'test-token-1234567890', got '%s'", r.Header.Get("Authorization"))
		}

		// Return mock update info
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"latest_version": "v1.2.0",
				"download_url":   "https://example.com/xboard-agent",
				"sha256":         "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				"file_size":      6090936,
				"strategy":       "auto",
				"release_notes":  "Test release",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create version manager and update checker
	vm := NewVersionManager("v1.0.0")
	uc := NewUpdateChecker(server.URL, "test-token-1234567890", vm)

	// Execute check update
	updateInfo, err := uc.CheckUpdate("v1.0.0")
	if err != nil {
		t.Fatalf("CheckUpdate failed: %v", err)
	}

	// Verify returned update info
	if updateInfo.LatestVersion != "v1.2.0" {
		t.Errorf("Expected version v1.2.0, got %s", updateInfo.LatestVersion)
	}

	if updateInfo.Strategy != "auto" {
		t.Errorf("Expected strategy auto, got %s", updateInfo.Strategy)
	}
}

func TestUpdateChecker_ShouldUpdate(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		shouldUpdate   bool
	}{
		{
			name:           "need update - current version older",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.1.0",
			shouldUpdate:   true,
		},
		{
			name:           "no update needed - same version",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.0.0",
			shouldUpdate:   false,
		},
		{
			name:           "no update needed - current version newer",
			currentVersion: "v1.1.0",
			latestVersion:  "v1.0.0",
			shouldUpdate:   false,
		},
		{
			name:           "need update - minor version update",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.1.0",
			shouldUpdate:   true,
		},
		{
			name:           "need update - patch version update",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.0.1",
			shouldUpdate:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVersionManager(tt.currentVersion)
			uc := NewUpdateChecker("http://example.com", "test-token-1234567890", vm)

			updateInfo := &UpdateInfo{
				LatestVersion: tt.latestVersion,
			}

			shouldUpdate, err := uc.ShouldUpdate(updateInfo)
			if err != nil {
				t.Fatalf("ShouldUpdate failed: %v", err)
			}

			if shouldUpdate != tt.shouldUpdate {
				t.Errorf("Expected shouldUpdate=%v, got %v", tt.shouldUpdate, shouldUpdate)
			}
		})
	}
}

func TestUpdateChecker_ShouldUpdate_NilUpdateInfo(t *testing.T) {
	vm := NewVersionManager("v1.0.0")
	uc := NewUpdateChecker("http://example.com", "test-token-1234567890", vm)

	_, err := uc.ShouldUpdate(nil)
	if err == nil {
		t.Error("Expected error for nil updateInfo, got nil")
	}
}

func TestUpdateChecker_CheckUpdate_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	vm := NewVersionManager("v1.0.0")
	uc := NewUpdateChecker(server.URL, "test-token-1234567890", vm)

	_, err := uc.CheckUpdate("v1.0.0")
	if err == nil {
		t.Error("Expected error for invalid response, got nil")
	}
}
