package main

import (
	"testing"
)

func TestVersionManager_GetCurrentVersion(t *testing.T) {
	vm := NewVersionManager("v1.2.3")
	if vm.GetCurrentVersion() != "v1.2.3" {
		t.Errorf("Expected v1.2.3, got %s", vm.GetCurrentVersion())
	}
}

func TestVersionManager_ParseVersion(t *testing.T) {
	vm := NewVersionManager("v1.0.0")

	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"valid version", "v1.2.3", false},
		{"valid version without v", "1.2.3", false},
		{"valid prerelease", "v1.2.3-beta.1", false},
		{"valid with build metadata", "v1.2.3+20231211", false},
		{"valid short version", "v1.2", false}, // semver library accepts this as v1.2.0
		{"invalid version", "invalid", true},
		{"invalid format", "not.a.version", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vm.ParseVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionManager_CompareVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		remote  string
		want    int
		wantErr bool
	}{
		{"current older", "v1.0.0", "v1.2.0", -1, false},
		{"current newer", "v2.0.0", "v1.0.0", 1, false},
		{"versions equal", "v1.2.3", "v1.2.3", 0, false},
		{"major version difference", "v1.0.0", "v2.0.0", -1, false},
		{"minor version difference", "v1.1.0", "v1.2.0", -1, false},
		{"patch version difference", "v1.2.1", "v1.2.3", -1, false},
		{"prerelease vs release", "v1.2.3-beta.1", "v1.2.3", -1, false},
		{"invalid current version", "invalid", "v1.0.0", 0, true},
		{"invalid remote version", "v1.0.0", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVersionManager(tt.current)
			got, err := vm.CompareVersion(tt.remote)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CompareVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionManager_CompareVersion_Transitivity(t *testing.T) {
	// Test transitivity: if v1 < v2 and v2 < v3, then v1 < v3
	vm1 := NewVersionManager("v1.0.0")
	vm2 := NewVersionManager("v1.5.0")

	cmp12, _ := vm1.CompareVersion("v1.5.0")
	cmp23, _ := vm2.CompareVersion("v2.0.0")
	cmp13, _ := vm1.CompareVersion("v2.0.0")

	if cmp12 != -1 || cmp23 != -1 || cmp13 != -1 {
		t.Errorf("Transitivity failed: v1.0.0 < v1.5.0 < v2.0.0")
	}
}

func TestVersionManager_CompareVersion_Symmetry(t *testing.T) {
	// Test symmetry: if v1 < v2, then v2 > v1
	vm1 := NewVersionManager("v1.0.0")
	vm2 := NewVersionManager("v2.0.0")

	cmp12, _ := vm1.CompareVersion("v2.0.0")
	cmp21, _ := vm2.CompareVersion("v1.0.0")

	if cmp12 != -1 || cmp21 != 1 {
		t.Errorf("Symmetry failed: expected v1.0.0 < v2.0.0 and v2.0.0 > v1.0.0")
	}
}

func TestVersionConstant(t *testing.T) {
	// Verify the Version constant is valid
	vm := NewVersionManager(Version)
	_, err := vm.ParseVersion(Version)
	if err != nil {
		t.Errorf("Version constant %s is not a valid semver: %v", Version, err)
	}
}
