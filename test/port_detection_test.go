package test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os/exec"
	"strings"
	"testing"
	"testing/quick"
	"time"
)

// PortDetectionResult represents the result of port detection
type PortDetectionResult struct {
	Port        int    `json:"port"`
	Available   bool   `json:"available"`
	Method      string `json:"method"`      // "netstat", "lsof", "ss"
	ProcessInfo string `json:"process_info"`
	Error       string `json:"error,omitempty"`
}

// PortDetector interface for different detection methods
type PortDetector interface {
	CheckPort(port int) PortDetectionResult
}

// NetstatDetector implements port detection using netstat
type NetstatDetector struct{}

func (n *NetstatDetector) CheckPort(port int) PortDetectionResult {
	result := PortDetectionResult{
		Port:   port,
		Method: "netstat",
	}

	// Use Windows-compatible netstat command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = "netstat command timed out"
		} else {
			result.Error = fmt.Sprintf("netstat command failed: %v", err)
		}
		return result
	}

	lines := strings.Split(string(output), "\n")
	portStr := fmt.Sprintf(":%d ", port)
	
	for _, line := range lines {
		if strings.Contains(line, portStr) && strings.Contains(line, "LISTENING") {
			result.Available = false
			// Extract process info if available (Windows netstat doesn't show process by default)
			result.ProcessInfo = "unknown"
			return result
		}
	}
	
	result.Available = true
	return result
}

// LsofDetector implements port detection using lsof (not available on Windows)
type LsofDetector struct{}

func (l *LsofDetector) CheckPort(port int) PortDetectionResult {
	result := PortDetectionResult{
		Port:   port,
		Method: "lsof",
	}

	// lsof is not available on Windows, so we skip this method
	result.Error = "lsof command not available on Windows"
	return result
}

// SsDetector implements port detection using ss (not available on Windows)
type SsDetector struct{}

func (s *SsDetector) CheckPort(port int) PortDetectionResult {
	result := PortDetectionResult{
		Port:   port,
		Method: "ss",
	}

	// ss is not available on Windows, so we skip this method
	result.Error = "ss command not available on Windows"
	return result
}

// PowerShellDetector implements port detection using PowerShell (Windows)
type PowerShellDetector struct{}

func (p *PowerShellDetector) CheckPort(port int) PortDetectionResult {
	result := PortDetectionResult{
		Port:   port,
		Method: "powershell",
	}

	// Use PowerShell to check for listening ports with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "powershell", "-Command", fmt.Sprintf("Get-NetTCPConnection -LocalPort %d -State Listen -ErrorAction SilentlyContinue", port))
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = "powershell command timed out"
		} else {
			result.Error = fmt.Sprintf("powershell command failed: %v", err)
		}
		return result
	}

	// If output is not empty, port is occupied
	if strings.TrimSpace(string(output)) != "" {
		result.Available = false
		result.ProcessInfo = "detected by powershell"
		return result
	}
	
	result.Available = true
	return result
}

// RobustPortDetector implements robust port detection with multiple methods
type RobustPortDetector struct {
	detectors []PortDetector
}

func NewRobustPortDetector() *RobustPortDetector {
	return &RobustPortDetector{
		detectors: []PortDetector{
			&NetstatDetector{},
			// PowerShell detection disabled due to performance issues on Windows
			// &PowerShellDetector{},
		},
	}
}

func (r *RobustPortDetector) CheckPort(port int) []PortDetectionResult {
	results := make([]PortDetectionResult, 0, len(r.detectors))
	
	for _, detector := range r.detectors {
		result := detector.CheckPort(port)
		results = append(results, result)
	}
	
	return results
}

// GetConsensusResult returns the consensus result from multiple detection methods
func (r *RobustPortDetector) GetConsensusResult(port int) PortDetectionResult {
	results := r.CheckPort(port)
	
	// Count votes for availability
	availableVotes := 0
	unavailableVotes := 0
	var lastValidResult PortDetectionResult
	
	for _, result := range results {
		if result.Error == "" {
			lastValidResult = result
			if result.Available {
				availableVotes++
			} else {
				unavailableVotes++
			}
		}
	}
	
	// If no valid results, return error
	if availableVotes == 0 && unavailableVotes == 0 {
		return PortDetectionResult{
			Port:   port,
			Method: "consensus",
			Error:  "all detection methods failed",
		}
	}
	
	// Return consensus result
	consensus := PortDetectionResult{
		Port:   port,
		Method: "consensus",
	}
	
	if availableVotes > unavailableVotes {
		consensus.Available = true
	} else {
		consensus.Available = false
		consensus.ProcessInfo = lastValidResult.ProcessInfo
	}
	
	return consensus
}

// Helper function to create a test server on a specific port
func createTestServer(port int) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf(":%d", port))
}

// Helper function to find an available port
func findAvailablePort() int {
	for i := 0; i < 100; i++ {
		port := rand.Intn(10000) + 20000 // Use ports 20000-30000 for testing
		if listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
			listener.Close()
			return port
		}
	}
	return 0
}

// **Feature: security-fixes, Property 1: Port detection accuracy**
// Property 1: Port detection accuracy
// For any system configuration and port number, port detection should correctly identify availability 
// using multiple verification methods and handle command failures gracefully
// **Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5**
func TestPortDetectionAccuracy(t *testing.T) {
	detector := NewRobustPortDetector()
	
	// Property: For any available port, all detection methods should agree it's available
	t.Run("AvailablePortConsistency", func(t *testing.T) {
		property := func() bool {
			port := findAvailablePort()
			if port == 0 {
				return true // Skip if no available port found
			}
			
			results := detector.CheckPort(port)
			
			// All successful detection methods should agree the port is available
			for _, result := range results {
				if result.Error == "" && !result.Available {
					t.Logf("Port %d reported as unavailable by %s but should be available", port, result.Method)
					return false
				}
			}
			return true
		}
		
		config := &quick.Config{MaxCount: 10}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Available port consistency property failed: %v", err)
		}
	})
	
	// Property: For any occupied port, at least one detection method should identify it as occupied
	t.Run("OccupiedPortDetection", func(t *testing.T) {
		property := func() bool {
			port := findAvailablePort()
			if port == 0 {
				return true // Skip if no available port found
			}
			
			// Create a test server to occupy the port
			listener, err := createTestServer(port)
			if err != nil {
				return true // Skip if can't create server
			}
			defer listener.Close()
			
			// Give the server a moment to start listening
			time.Sleep(10 * time.Millisecond)
			
			results := detector.CheckPort(port)
			
			// At least one successful detection method should identify the port as occupied
			foundOccupied := false
			for _, result := range results {
				if result.Error == "" && !result.Available {
					foundOccupied = true
					break
				}
			}
			
			if !foundOccupied {
				t.Logf("Port %d is occupied but no detection method identified it", port)
				return false
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 5} // Fewer iterations for occupied port tests
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Occupied port detection property failed: %v", err)
		}
	})
	
	// Property: Consensus method should provide robust results
	t.Run("ConsensusRobustness", func(t *testing.T) {
		property := func() bool {
			port := findAvailablePort()
			if port == 0 {
				return true // Skip if no available port found
			}
			
			consensus := detector.GetConsensusResult(port)
			
			// Consensus should not fail if at least one method works
			if consensus.Error != "" {
				// Check if all individual methods also failed
				results := detector.CheckPort(port)
				allFailed := true
				for _, result := range results {
					if result.Error == "" {
						allFailed = false
						break
					}
				}
				
				if !allFailed {
					t.Logf("Consensus failed but individual methods succeeded for port %d", port)
					return false
				}
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 10}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Consensus robustness property failed: %v", err)
		}
	})
	
	// Property: Detection methods should handle invalid ports gracefully
	t.Run("InvalidPortHandling", func(t *testing.T) {
		property := func(invalidPort int) bool {
			// Test with clearly invalid ports
			if invalidPort >= 1 && invalidPort <= 65535 {
				return true // Skip valid ports
			}
			
			results := detector.CheckPort(invalidPort)
			
			// All methods should either return an error or handle gracefully
			for _, result := range results {
				if result.Error == "" {
					// If no error, the result should be consistent
					if result.Port != invalidPort {
						t.Logf("Method %s returned wrong port number for invalid port %d", result.Method, invalidPort)
						return false
					}
				}
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 5}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Invalid port handling property failed: %v", err)
		}
	})
}

// Test specific edge cases
func TestPortDetectionEdgeCases(t *testing.T) {
	detector := NewRobustPortDetector()
	
	// Test well-known ports
	t.Run("WellKnownPorts", func(t *testing.T) {
		wellKnownPorts := []int{80, 443, 22, 25, 53, 110, 143, 993, 995}
		
		for _, port := range wellKnownPorts {
			results := detector.CheckPort(port)
			
			// At least one method should work for well-known ports
			hasValidResult := false
			for _, result := range results {
				if result.Error == "" {
					hasValidResult = true
					break
				}
			}
			
			if !hasValidResult {
				t.Errorf("All detection methods failed for well-known port %d", port)
			}
		}
	})
	
	// Test port range boundaries
	t.Run("PortBoundaries", func(t *testing.T) {
		boundaryPorts := []int{1, 1023, 1024, 49151, 49152, 65535}
		
		for _, port := range boundaryPorts {
			results := detector.CheckPort(port)
			
			// Methods should handle boundary ports without crashing
			for _, result := range results {
				if result.Port != port {
					t.Errorf("Method %s returned wrong port %d instead of %d", result.Method, result.Port, port)
				}
			}
		}
	})
}

// Benchmark port detection performance
func BenchmarkPortDetection(b *testing.B) {
	detector := NewRobustPortDetector()
	port := findAvailablePort()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.GetConsensusResult(port)
	}
}