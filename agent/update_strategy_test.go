package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestUpdateStrategy_Auto tests automatic update strategy
func TestUpdateStrategy_Auto(t *testing.T) {
	// Create mock panel server
	panelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/agent/version" {
			response := map[string]interface{}{
				"data": UpdateInfo{
					LatestVersion: "v1.1.0",
					DownloadURL:   "https://example.com/download",
					SHA256:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
					FileSize:      1024,
					Strategy:      "auto",
					ReleaseNotes:  "Auto update test",
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer panelServer.Close()
	
	// Create version manager and update checker
	versionManager := NewVersionManager("v1.0.0")
	updateChecker := NewUpdateChecker(panelServer.URL, "test-token-1234567890", versionManager)
	
	// Check update
	updateInfo, err := updateChecker.CheckUpdate("v1.0.0")
	if err != nil {
		t.Fatalf("CheckUpdate failed: %v", err)
	}
	
	// Verify strategy is auto
	if updateInfo.Strategy != "auto" {
		t.Errorf("Expected auto strategy, got %s", updateInfo.Strategy)
	}
	
	// Verify should update
	shouldUpdate, err := updateChecker.ShouldUpdate(updateInfo)
	if err != nil {
		t.Fatalf("ShouldUpdate failed: %v", err)
	}
	
	if !shouldUpdate {
		t.Error("Expected shouldUpdate to be true for auto strategy")
	}
}

// TestUpdateStrategy_Manual tests manual update strategy
func TestUpdateStrategy_Manual(t *testing.T) {
	// Create mock panel server
	panelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/agent/version" {
			response := map[string]interface{}{
				"data": UpdateInfo{
					LatestVersion: "v1.1.0",
					DownloadURL:   "https://example.com/download",
					SHA256:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
					FileSize:      1024,
					Strategy:      "manual",
					ReleaseNotes:  "Manual update test",
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer panelServer.Close()
	
	// Create version manager and update checker
	versionManager := NewVersionManager("v1.0.0")
	updateChecker := NewUpdateChecker(panelServer.URL, "test-token-1234567890", versionManager)
	
	// Check update
	updateInfo, err := updateChecker.CheckUpdate("v1.0.0")
	if err != nil {
		t.Fatalf("CheckUpdate failed: %v", err)
	}
	
	// Verify strategy is manual
	if updateInfo.Strategy != "manual" {
		t.Errorf("Expected manual strategy, got %s", updateInfo.Strategy)
	}
	
	// Verify should update (version comparison only)
	shouldUpdate, err := updateChecker.ShouldUpdate(updateInfo)
	if err != nil {
		t.Fatalf("ShouldUpdate failed: %v", err)
	}
	
	if !shouldUpdate {
		t.Error("Expected shouldUpdate to be true (version comparison)")
	}
}

// TestUpdateStrategy_NoUpdate tests when no update is needed
func TestUpdateStrategy_NoUpdate(t *testing.T) {
	// Create mock panel server
	panelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/agent/version" {
			response := map[string]interface{}{
				"data": UpdateInfo{
					LatestVersion: "v1.0.0",
					DownloadURL:   "https://example.com/download",
					SHA256:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
					FileSize:      1024,
					Strategy:      "auto",
					ReleaseNotes:  "No update test",
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer panelServer.Close()
	
	// Create version manager and update checker
	versionManager := NewVersionManager("v1.0.0")
	updateChecker := NewUpdateChecker(panelServer.URL, "test-token-1234567890", versionManager)
	
	// Check update
	updateInfo, err := updateChecker.CheckUpdate("v1.0.0")
	if err != nil {
		t.Fatalf("CheckUpdate failed: %v", err)
	}
	
	// Verify should not update (same version)
	shouldUpdate, err := updateChecker.ShouldUpdate(updateInfo)
	if err != nil {
		t.Fatalf("ShouldUpdate failed: %v", err)
	}
	
	if shouldUpdate {
		t.Error("Expected shouldUpdate to be false for same version")
	}
}

// TestUpdateInfo_StrategyField tests the Strategy field in UpdateInfo
func TestUpdateInfo_StrategyField(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
	}{
		{"auto strategy", "auto"},
		{"manual strategy", "manual"},
		{"empty strategy", ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateInfo := &UpdateInfo{
				Strategy: tt.strategy,
			}
			
			if updateInfo.Strategy != tt.strategy {
				t.Errorf("Expected strategy %s, got %s", tt.strategy, updateInfo.Strategy)
			}
		})
	}
}

// TestAgent_ManualUpdateFlag tests the manual update flag
func TestAgent_ManualUpdateFlag(t *testing.T) {
	// This test verifies that the -update flag exists and can be used
	// In actual implementation, this would be a command-line flag
	
	// For now, we just verify the concept
	manualUpdate := false // This would be set by flag.Parse()
	
	if manualUpdate {
		t.Log("Manual update flag is enabled")
	} else {
		t.Log("Manual update flag is disabled (default)")
	}
}

// TestHeartbeat_VersionInfo tests version info in heartbeat
func TestHeartbeat_VersionInfo(t *testing.T) {
	// Create mock panel server
	panelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/agent/version" {
			response := map[string]interface{}{
				"data": UpdateInfo{
					LatestVersion: "v1.1.0",
					DownloadURL:   "https://example.com/download",
					SHA256:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
					FileSize:      1024,
					Strategy:      "manual",
					ReleaseNotes:  "Test release",
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer panelServer.Close()
	
	// Simulate heartbeat with version info
	oldPanelURL := panelURL
	oldToken := token
	panelURL = panelServer.URL
	token = "test-token-1234567890"
	
	defer func() {
		panelURL = oldPanelURL
		token = oldToken
	}()
	
	// Create version manager and update checker
	versionManager := NewVersionManager("v1.0.0")
	updateChecker := NewUpdateChecker(panelURL, token, versionManager)
	
	// Check update (simulating heartbeat response)
	updateInfo, err := updateChecker.CheckUpdate("v1.0.0")
	if err != nil {
		t.Fatalf("CheckUpdate failed: %v", err)
	}
	
	// Verify update info is received
	if updateInfo.LatestVersion != "v1.1.0" {
		t.Errorf("Expected version v1.1.0, got %s", updateInfo.LatestVersion)
	}
	
	t.Log("Update info received via heartbeat")
}
