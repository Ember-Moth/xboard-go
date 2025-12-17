package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ResilienceService handles system resilience and error recovery
type ResilienceService struct {
	db              *gorm.DB
	circuitBreakers map[string]*CircuitBreaker
	mu              sync.RWMutex
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name           string
	maxFailures    int
	timeout        time.Duration
	state          CircuitBreakerState
	failures       int
	lastFailTime   time.Time
	mu             sync.RWMutex
	onStateChange  func(name string, from, to CircuitBreakerState)
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts     int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	Jitter          bool
}

// NetworkOperation represents a network operation that can be retried
type NetworkOperation func() error

// DatabaseOperation represents a database operation that can be retried
type DatabaseOperation func() error

// NewResilienceService creates a new resilience service
func NewResilienceService(db *gorm.DB) *ResilienceService {
	return &ResilienceService{
		db:              db,
		circuitBreakers: make(map[string]*CircuitBreaker),
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:        name,
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       CircuitBreakerClosed,
	}
}

// Execute executes an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(operation func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if circuit breaker should transition from open to half-open
	if cb.state == CircuitBreakerOpen {
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.setState(CircuitBreakerHalfOpen)
		} else {
			return fmt.Errorf("circuit breaker %s is open", cb.name)
		}
	}

	// Execute the operation
	err := operation()
	
	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

// onFailure handles operation failure
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.state == CircuitBreakerHalfOpen {
		cb.setState(CircuitBreakerOpen)
	} else if cb.failures >= cb.maxFailures {
		cb.setState(CircuitBreakerOpen)
	}
}

// onSuccess handles operation success
func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0
	if cb.state == CircuitBreakerHalfOpen {
		cb.setState(CircuitBreakerClosed)
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState CircuitBreakerState) {
	oldState := cb.state
	cb.state = newState
	
	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, oldState, newState)
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetCircuitBreaker gets or creates a circuit breaker
func (rs *ResilienceService) GetCircuitBreaker(name string, maxFailures int, timeout time.Duration) *CircuitBreaker {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if cb, exists := rs.circuitBreakers[name]; exists {
		return cb
	}

	cb := NewCircuitBreaker(name, maxFailures, timeout)
	cb.onStateChange = func(name string, from, to CircuitBreakerState) {
		log.Printf("Circuit breaker %s state changed from %v to %v", name, from, to)
	}
	
	rs.circuitBreakers[name] = cb
	return cb
}

// RetryWithExponentialBackoff retries an operation with exponential backoff
func (rs *ResilienceService) RetryWithExponentialBackoff(operation NetworkOperation, config RetryConfig) error {
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Don't wait after the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}
		
		// Calculate delay with exponential backoff
		delay := rs.calculateBackoffDelay(attempt, config)
		time.Sleep(delay)
	}
	
	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// calculateBackoffDelay calculates the delay for exponential backoff
func (rs *ResilienceService) calculateBackoffDelay(attempt int, config RetryConfig) time.Duration {
	// Calculate exponential backoff: baseDelay * (backoffFactor ^ attempt)
	delay := float64(config.BaseDelay) * math.Pow(config.BackoffFactor, float64(attempt))
	
	// Apply maximum delay limit
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}
	
	duration := time.Duration(delay)
	
	// Add jitter if enabled (±25% random variation)
	if config.Jitter {
		jitter := time.Duration(float64(duration) * 0.25 * (2*rand.Float64() - 1))
		duration += jitter
		
		// Ensure duration is not negative
		if duration < 0 {
			duration = config.BaseDelay
		}
	}
	
	return duration
}

// ExecuteWithCircuitBreaker executes an operation with circuit breaker protection
func (rs *ResilienceService) ExecuteWithCircuitBreaker(name string, operation func() error, maxFailures int, timeout time.Duration) error {
	cb := rs.GetCircuitBreaker(name, maxFailures, timeout)
	return cb.Execute(operation)
}

// ExecuteDatabaseOperationWithResilience executes a database operation with resilience patterns
func (rs *ResilienceService) ExecuteDatabaseOperationWithResilience(operation DatabaseOperation) error {
	config := RetryConfig{
		MaxAttempts:   3,
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      2 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}
	
	return rs.RetryWithExponentialBackoff(func() error {
		return rs.ExecuteWithCircuitBreaker("database", func() error {
			return operation()
		}, 5, 30*time.Second)
	}, config)
}

// ExecuteNetworkOperationWithResilience executes a network operation with resilience patterns
func (rs *ResilienceService) ExecuteNetworkOperationWithResilience(operation NetworkOperation) error {
	config := RetryConfig{
		MaxAttempts:   5,
		BaseDelay:     500 * time.Millisecond,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}
	
	return rs.RetryWithExponentialBackoff(operation, config)
}

// IsNetworkError checks if an error is a network-related error
func (rs *ResilienceService) IsNetworkError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for common network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	
	// Check for HTTP client errors that indicate network issues
	if errors.Is(err, context.DeadlineExceeded) || 
	   errors.Is(err, context.Canceled) {
		return true
	}
	
	// Check error message for common network error patterns
	errMsg := err.Error()
	networkErrorPatterns := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"network is unreachable",
		"no route to host",
		"temporary failure in name resolution",
		"i/o timeout",
	}
	
	for _, pattern := range networkErrorPatterns {
		if contains(errMsg, pattern) {
			return true
		}
	}
	
	return false
}

// IsDatabaseError checks if an error is a database-related error
func (rs *ResilienceService) IsDatabaseError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for GORM errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false // This is not a connection error
	}
	
	// Check error message for database connection issues
	errMsg := err.Error()
	dbErrorPatterns := []string{
		"database is locked",
		"connection refused",
		"connection reset",
		"connection timeout",
		"database/sql: connection refused",
		"driver: bad connection",
		"invalid connection",
	}
	
	for _, pattern := range dbErrorPatterns {
		if contains(errMsg, pattern) {
			return true
		}
	}
	
	return false
}

// GracefulDegradation provides fallback functionality when services are unavailable
func (rs *ResilienceService) GracefulDegradation(primaryOperation, fallbackOperation func() error) error {
	// Try primary operation first
	err := primaryOperation()
	if err == nil {
		return nil
	}
	
	// If primary fails, try fallback
	log.Printf("Primary operation failed: %v, attempting fallback", err)
	fallbackErr := fallbackOperation()
	if fallbackErr == nil {
		log.Printf("Fallback operation succeeded")
		return nil
	}
	
	// Both operations failed
	return fmt.Errorf("both primary and fallback operations failed: primary=%v, fallback=%v", err, fallbackErr)
}

// SanitizeErrorMessage creates a safe error message for user display
func (rs *ResilienceService) SanitizeErrorMessage(err error, userFriendlyMessage string) string {
	if err == nil {
		return ""
	}
	
	// Don't expose internal error details to users
	sensitivePatterns := []string{
		"password",
		"token",
		"secret",
		"key",
		"database",
		"sql",
		"internal",
		"panic",
		"stack trace",
	}
	
	errMsg := err.Error()
	for _, pattern := range sensitivePatterns {
		if contains(errMsg, pattern) {
			return userFriendlyMessage
		}
	}
	
	// Return sanitized version of the error
	return userFriendlyMessage + " (Error code: " + fmt.Sprintf("%x", hash(errMsg)%10000) + ")"
}

// RecoverFromPanic recovers from panics and converts them to errors
func (rs *ResilienceService) RecoverFromPanic() error {
	if r := recover(); r != nil {
		log.Printf("Recovered from panic: %v", r)
		return fmt.Errorf("internal server error occurred")
	}
	return nil
}

// HealthCheck performs a comprehensive health check
func (rs *ResilienceService) HealthCheck() map[string]interface{} {
	result := make(map[string]interface{})
	
	// Check database connectivity
	result["database"] = rs.checkDatabaseHealth()
	
	// Check circuit breaker states
	result["circuit_breakers"] = rs.getCircuitBreakerStates()
	
	// Overall health status
	result["status"] = rs.calculateOverallHealth(result)
	result["timestamp"] = time.Now().Unix()
	
	return result
}

// checkDatabaseHealth checks database connectivity
func (rs *ResilienceService) checkDatabaseHealth() map[string]interface{} {
	result := make(map[string]interface{})
	
	start := time.Now()
	var err error
	
	if rs.db == nil {
		err = fmt.Errorf("database connection not initialized")
	} else {
		err = rs.ExecuteDatabaseOperationWithResilience(func() error {
			sqlDB, err := rs.db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Ping()
		})
	}
	duration := time.Since(start)
	
	result["healthy"] = err == nil
	result["response_time_ms"] = duration.Milliseconds()
	if err != nil {
		result["error"] = rs.SanitizeErrorMessage(err, "Database connectivity issue")
	}
	
	return result
}

// getCircuitBreakerStates returns the current state of all circuit breakers
func (rs *ResilienceService) getCircuitBreakerStates() map[string]interface{} {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	
	states := make(map[string]interface{})
	for name, cb := range rs.circuitBreakers {
		states[name] = map[string]interface{}{
			"state":    cb.GetState(),
			"failures": cb.failures,
		}
	}
	
	return states
}

// calculateOverallHealth calculates the overall system health
func (rs *ResilienceService) calculateOverallHealth(checks map[string]interface{}) string {
	// Check database health
	if dbHealth, ok := checks["database"].(map[string]interface{}); ok {
		if healthy, ok := dbHealth["healthy"].(bool); ok && !healthy {
			return "unhealthy"
		}
	}
	
	// Check circuit breaker states
	if cbStates, ok := checks["circuit_breakers"].(map[string]interface{}); ok {
		for _, state := range cbStates {
			if stateMap, ok := state.(map[string]interface{}); ok {
				if cbState, ok := stateMap["state"].(CircuitBreakerState); ok {
					if cbState == CircuitBreakerOpen {
						return "degraded"
					}
				}
			}
		}
	}
	
	return "healthy"
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
		 (s[:len(substr)] == substr || 
		  s[len(s)-len(substr):] == substr || 
		  findInString(s, substr))))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func hash(s string) uint32 {
	h := uint32(0)
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}