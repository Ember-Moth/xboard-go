package test

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestSimpleNetstat(t *testing.T) {
	cmd := exec.Command("netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("netstat failed: %v", err)
	}
	
	fmt.Printf("Netstat output length: %d\n", len(output))
	
	// Check if we can find any LISTENING ports
	lines := string(output)
	if len(lines) > 0 {
		t.Logf("Netstat working correctly")
	}
}

func TestSimplePowerShell(t *testing.T) {
	cmd := exec.Command("powershell", "-Command", "Get-NetTCPConnection -State Listen | Select-Object -First 5")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("PowerShell failed: %v", err)
	}
	
	fmt.Printf("PowerShell output: %s\n", string(output))
}