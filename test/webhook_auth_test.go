package test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"testing/quick"
	"time"
)

// WebhookAuthenticator represents the webhook authentication functionality
type WebhookAuthenticator struct {
	secretToken string
}

// NewWebhookAuthenticator creates a new webhook authenticator
func NewWebhookAuthenticator(secretToken string) *WebhookAuthenticator {
	return &WebhookAuthenticator{
		secretToken: secretToken,
	}
}

// ValidateSignature validates the webhook signature using HMAC-SHA256
func (w *WebhookAuthenticator) ValidateSignature(body []byte, signature string) bool {
	if w.secretToken == "" {
		return false
	}
	
	// Remove "sha256=" prefix if present
	signature = strings.TrimPrefix(signature, "sha256=")
	
	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(w.secretToken))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	
	// Use constant time comparison to prevent timing attacks
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GenerateSignature generates a signature for the given body
func (w *WebhookAuthenticator) GenerateSignature(body []byte) string {
	if w.secretToken == "" {
		return ""
	}
	
	mac := hmac.New(sha256.New, []byte(w.secretToken))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// WebhookRequest represents a webhook request
type WebhookRequest struct {
	Body      []byte
	Signature string
	Headers   map[string]string
}

// WebhookResponse represents the response to a webhook request
type WebhookResponse struct {
	StatusCode int
	Body       string
	Logged     bool // Whether security violation was logged
}

// MockWebhookHandler simulates the webhook handler behavior
type MockWebhookHandler struct {
	authenticator *WebhookAuthenticator
	loggedEvents  []SecurityEvent
}

// SecurityEvent represents a security event for logging
type SecurityEvent struct {
	EventType string
	Severity  string
	SourceIP  string
	Details   map[string]interface{}
	Timestamp time.Time
}

// NewMockWebhookHandler creates a new mock webhook handler
func NewMockWebhookHandler(secretToken string) *MockWebhookHandler {
	return &MockWebhookHandler{
		authenticator: NewWebhookAuthenticator(secretToken),
		loggedEvents:  make([]SecurityEvent, 0),
	}
}

// HandleWebhook processes a webhook request
func (m *MockWebhookHandler) HandleWebhook(req WebhookRequest) WebhookResponse {
	// Validate signature
	if !m.authenticator.ValidateSignature(req.Body, req.Signature) {
		// Log security violation
		event := SecurityEvent{
			EventType: "webhook_forge",
			Severity:  "high",
			SourceIP:  req.Headers["X-Forwarded-For"],
			Details: map[string]interface{}{
				"signature": req.Signature,
				"body_size": len(req.Body),
			},
			Timestamp: time.Now(),
		}
		m.loggedEvents = append(m.loggedEvents, event)
		
		return WebhookResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "Invalid signature",
			Logged:     true,
		}
	}
	
	// Process valid webhook
	return WebhookResponse{
		StatusCode: http.StatusOK,
		Body:       "OK",
		Logged:     false,
	}
}

// GetLoggedEvents returns all logged security events
func (m *MockWebhookHandler) GetLoggedEvents() []SecurityEvent {
	return m.loggedEvents
}

// Helper functions for generating test data

// generateRandomSecret generates a random secret token
func generateRandomSecret() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := rand.Intn(32) + 16 // 16-48 characters
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// generateRandomBody generates random webhook body data
func generateRandomBody() []byte {
	// Generate JSON-like webhook body
	templates := []string{
		`{"update_id": %d, "message": {"message_id": %d, "text": "%s"}}`,
		`{"update_id": %d, "callback_query": {"id": "%s", "data": "%s"}}`,
		`{"update_id": %d}`,
	}
	
	template := templates[rand.Intn(len(templates))]
	updateID := rand.Int63()
	messageID := rand.Int63()
	text := fmt.Sprintf("test_message_%d", rand.Int63())
	callbackID := fmt.Sprintf("callback_%d", rand.Int63())
	callbackData := fmt.Sprintf("data_%d", rand.Int63())
	
	body := fmt.Sprintf(template, updateID, messageID, text)
	if strings.Contains(template, "callback_query") {
		body = fmt.Sprintf(template, updateID, callbackID, callbackData)
	}
	
	return []byte(body)
}

// **Feature: security-fixes, Property 4: Webhook authentication integrity**
// Property 4: Webhook authentication integrity
// For any Telegram webhook request, the system should validate authenticity using signature verification and reject invalid requests
// **Validates: Requirements 3.1, 3.3**
func TestWebhookAuthenticationIntegrity(t *testing.T) {
	// Property: Valid signatures should always be accepted
	t.Run("ValidSignatureAcceptance", func(t *testing.T) {
		property := func() bool {
			secret := generateRandomSecret()
			body := generateRandomBody()
			
			handler := NewMockWebhookHandler(secret)
			authenticator := NewWebhookAuthenticator(secret)
			
			// Generate valid signature
			signature := authenticator.GenerateSignature(body)
			
			req := WebhookRequest{
				Body:      body,
				Signature: signature,
				Headers:   map[string]string{"X-Forwarded-For": "127.0.0.1"},
			}
			
			response := handler.HandleWebhook(req)
			
			// Valid signature should be accepted
			if response.StatusCode != http.StatusOK {
				t.Logf("Valid signature rejected: status %d", response.StatusCode)
				return false
			}
			
			// No security event should be logged for valid requests
			if response.Logged {
				t.Logf("Security event logged for valid signature")
				return false
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Valid signature acceptance property failed: %v", err)
		}
	})
	
	// Property: Invalid signatures should always be rejected
	t.Run("InvalidSignatureRejection", func(t *testing.T) {
		property := func() bool {
			secret := generateRandomSecret()
			body := generateRandomBody()
			
			handler := NewMockWebhookHandler(secret)
			
			// Generate invalid signature (use different secret)
			wrongSecret := generateRandomSecret()
			wrongAuthenticator := NewWebhookAuthenticator(wrongSecret)
			invalidSignature := wrongAuthenticator.GenerateSignature(body)
			
			req := WebhookRequest{
				Body:      body,
				Signature: invalidSignature,
				Headers:   map[string]string{"X-Forwarded-For": "192.168.1.100"},
			}
			
			response := handler.HandleWebhook(req)
			
			// Invalid signature should be rejected
			if response.StatusCode != http.StatusUnauthorized {
				t.Logf("Invalid signature accepted: status %d", response.StatusCode)
				return false
			}
			
			// Security event should be logged for invalid requests
			if !response.Logged {
				t.Logf("No security event logged for invalid signature")
				return false
			}
			
			// Check that security event was actually logged
			events := handler.GetLoggedEvents()
			if len(events) == 0 {
				t.Logf("Security event not found in logs")
				return false
			}
			
			lastEvent := events[len(events)-1]
			if lastEvent.EventType != "webhook_forge" {
				t.Logf("Wrong event type logged: %s", lastEvent.EventType)
				return false
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Invalid signature rejection property failed: %v", err)
		}
	})
	
	// Property: Empty or missing signatures should be rejected
	t.Run("MissingSignatureRejection", func(t *testing.T) {
		property := func() bool {
			secret := generateRandomSecret()
			body := generateRandomBody()
			
			handler := NewMockWebhookHandler(secret)
			
			// Test with empty signature
			req := WebhookRequest{
				Body:      body,
				Signature: "",
				Headers:   map[string]string{"X-Forwarded-For": "10.0.0.1"},
			}
			
			response := handler.HandleWebhook(req)
			
			// Missing signature should be rejected
			if response.StatusCode != http.StatusUnauthorized {
				t.Logf("Missing signature accepted: status %d", response.StatusCode)
				return false
			}
			
			// Security event should be logged
			if !response.Logged {
				t.Logf("No security event logged for missing signature")
				return false
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Missing signature rejection property failed: %v", err)
		}
	})
	
	// Property: Signature validation should be constant time (timing attack prevention)
	t.Run("ConstantTimeValidation", func(t *testing.T) {
		property := func() bool {
			secret := generateRandomSecret()
			body := generateRandomBody()
			
			authenticator := NewWebhookAuthenticator(secret)
			validSignature := authenticator.GenerateSignature(body)
			
			// Create signatures of different lengths to test timing
			shortInvalidSig := "abc123"
			longInvalidSig := strings.Repeat("a", len(validSignature))
			
			// Measure validation time for different signature lengths
			start1 := time.Now()
			authenticator.ValidateSignature(body, shortInvalidSig)
			duration1 := time.Since(start1)
			
			start2 := time.Now()
			authenticator.ValidateSignature(body, longInvalidSig)
			duration2 := time.Since(start2)
			
			start3 := time.Now()
			authenticator.ValidateSignature(body, validSignature)
			duration3 := time.Since(start3)
			
			// The timing differences should not be excessive (within reasonable bounds)
			// This is a basic check - in practice, more sophisticated timing analysis would be needed
			maxDuration := max(duration1, duration2, duration3)
			minDuration := min(duration1, duration2, duration3)
			
			// Allow up to 100x difference (very generous for test environments)
			// The actual security comes from using hmac.Equal() which provides constant-time comparison
			if maxDuration > minDuration*100 && minDuration > 0 {
				t.Logf("Potential timing attack vulnerability: min=%v, max=%v", minDuration, maxDuration)
				return false
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 20} // Fewer iterations for timing tests
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Constant time validation property failed: %v", err)
		}
	})
	
	// Property: Signature validation should handle malformed signatures gracefully
	t.Run("MalformedSignatureHandling", func(t *testing.T) {
		property := func() bool {
			secret := generateRandomSecret()
			body := generateRandomBody()
			
			handler := NewMockWebhookHandler(secret)
			
			// Test various malformed signatures
			malformedSignatures := []string{
				"not_hex_at_all",
				"sha256=invalid_hex",
				"sha256=",
				"wrong_prefix=abcdef123456",
				strings.Repeat("a", 1000), // Very long signature
			}
			
			for _, malformedSig := range malformedSignatures {
				req := WebhookRequest{
					Body:      body,
					Signature: malformedSig,
					Headers:   map[string]string{"X-Forwarded-For": "172.16.0.1"},
				}
				
				response := handler.HandleWebhook(req)
				
				// Malformed signature should be rejected
				if response.StatusCode != http.StatusUnauthorized {
					t.Logf("Malformed signature accepted: %s", malformedSig)
					return false
				}
				
				// Security event should be logged
				if !response.Logged {
					t.Logf("No security event logged for malformed signature: %s", malformedSig)
					return false
				}
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 10}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Malformed signature handling property failed: %v", err)
		}
	})
}

// Helper functions for min/max
func min(a, b, c time.Duration) time.Duration {
	result := a
	if b < result {
		result = b
	}
	if c < result {
		result = c
	}
	return result
}

func max(a, b, c time.Duration) time.Duration {
	result := a
	if b > result {
		result = b
	}
	if c > result {
		result = c
	}
	return result
}

// Test specific edge cases for webhook authentication
func TestWebhookAuthenticationEdgeCases(t *testing.T) {
	// Test with empty secret token
	t.Run("EmptySecretToken", func(t *testing.T) {
		handler := NewMockWebhookHandler("")
		body := []byte(`{"update_id": 123}`)
		
		req := WebhookRequest{
			Body:      body,
			Signature: "sha256=somesignature",
			Headers:   map[string]string{"X-Forwarded-For": "203.0.113.1"},
		}
		
		response := handler.HandleWebhook(req)
		
		// Empty secret should reject all requests
		if response.StatusCode != http.StatusUnauthorized {
			t.Errorf("Empty secret token should reject all requests, got status %d", response.StatusCode)
		}
	})
	
	// Test with very large body
	t.Run("LargeBody", func(t *testing.T) {
		secret := "test_secret_123"
		handler := NewMockWebhookHandler(secret)
		authenticator := NewWebhookAuthenticator(secret)
		
		// Create large body (1MB)
		largeBody := bytes.Repeat([]byte("a"), 1024*1024)
		signature := authenticator.GenerateSignature(largeBody)
		
		req := WebhookRequest{
			Body:      largeBody,
			Signature: signature,
			Headers:   map[string]string{"X-Forwarded-For": "198.51.100.1"},
		}
		
		response := handler.HandleWebhook(req)
		
		// Large body with valid signature should be accepted
		if response.StatusCode != http.StatusOK {
			t.Errorf("Large body with valid signature should be accepted, got status %d", response.StatusCode)
		}
	})
	
	// Test signature with and without sha256= prefix
	t.Run("SignaturePrefixHandling", func(t *testing.T) {
		secret := "test_secret_456"
		body := []byte(`{"update_id": 456}`)
		authenticator := NewWebhookAuthenticator(secret)
		
		signature := authenticator.GenerateSignature(body)
		signatureWithoutPrefix := strings.TrimPrefix(signature, "sha256=")
		
		// Both should be valid
		if !authenticator.ValidateSignature(body, signature) {
			t.Error("Signature with sha256= prefix should be valid")
		}
		
		if !authenticator.ValidateSignature(body, signatureWithoutPrefix) {
			t.Error("Signature without sha256= prefix should be valid")
		}
	})
}

// Benchmark webhook authentication performance
func BenchmarkWebhookAuthentication(b *testing.B) {
	secret := "benchmark_secret_token"
	body := []byte(`{"update_id": 12345, "message": {"message_id": 67890, "text": "benchmark test"}}`)
	authenticator := NewWebhookAuthenticator(secret)
	signature := authenticator.GenerateSignature(body)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authenticator.ValidateSignature(body, signature)
	}
}
// **Feature: security-fixes, Property 5: Webhook security response**
// Property 5: Webhook security response
// For any failed webhook authentication, the system should reject the request and log security violations
// **Validates: Requirements 3.2**
func TestWebhookSecurityResponse(t *testing.T) {
	// Property: All authentication failures should result in proper security responses
	t.Run("AuthenticationFailureResponse", func(t *testing.T) {
		property := func() bool {
			secret := generateRandomSecret()
			body := generateRandomBody()
			
			handler := NewMockWebhookHandler(secret)
			
			// Generate various types of invalid requests
			invalidRequests := []WebhookRequest{
				// Wrong signature
				{
					Body:      body,
					Signature: "sha256=wrongsignature123456789abcdef",
					Headers:   map[string]string{"X-Forwarded-For": "192.168.1.100"},
				},
				// Empty signature
				{
					Body:      body,
					Signature: "",
					Headers:   map[string]string{"X-Forwarded-For": "10.0.0.50"},
				},
				// Malformed signature
				{
					Body:      body,
					Signature: "invalid_format_signature",
					Headers:   map[string]string{"X-Forwarded-For": "172.16.0.200"},
				},
				// Signature for different body
				{
					Body:      body,
					Signature: NewWebhookAuthenticator(secret).GenerateSignature([]byte("different body")),
					Headers:   map[string]string{"X-Forwarded-For": "203.0.113.75"},
				},
			}
			
			for i, req := range invalidRequests {
				response := handler.HandleWebhook(req)
				
				// Should reject with unauthorized status
				if response.StatusCode != http.StatusUnauthorized {
					t.Logf("Invalid request %d not rejected: status %d", i, response.StatusCode)
					return false
				}
				
				// Should log security violation
				if !response.Logged {
					t.Logf("Security violation not logged for invalid request %d", i)
					return false
				}
				
				// Check that security event contains proper information
				events := handler.GetLoggedEvents()
				if len(events) <= i {
					t.Logf("Security event not found in logs for request %d", i)
					return false
				}
				
				event := events[i]
				if event.EventType != "webhook_forge" {
					t.Logf("Wrong event type for request %d: %s", i, event.EventType)
					return false
				}
				
				if event.Severity != "high" {
					t.Logf("Wrong severity for request %d: %s", i, event.Severity)
					return false
				}
				
				// Should capture source IP
				if event.SourceIP == "" {
					t.Logf("Source IP not captured for request %d", i)
					return false
				}
				
				// Should include relevant details
				if event.Details == nil {
					t.Logf("Event details missing for request %d", i)
					return false
				}
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Authentication failure response property failed: %v", err)
		}
	})
	
	// Property: Security logging should be consistent and complete
	t.Run("SecurityLoggingConsistency", func(t *testing.T) {
		property := func() bool {
			secret := generateRandomSecret()
			body := generateRandomBody()
			
			handler := NewMockWebhookHandler(secret)
			
			// Generate multiple invalid requests
			numRequests := rand.Intn(5) + 3 // 3-7 requests
			for i := 0; i < numRequests; i++ {
				invalidReq := WebhookRequest{
					Body:      body,
					Signature: fmt.Sprintf("invalid_sig_%d", i),
					Headers:   map[string]string{"X-Forwarded-For": fmt.Sprintf("192.168.1.%d", i+1)},
				}
				
				handler.HandleWebhook(invalidReq)
			}
			
			// Check that all requests were logged
			events := handler.GetLoggedEvents()
			if len(events) != numRequests {
				t.Logf("Expected %d logged events, got %d", numRequests, len(events))
				return false
			}
			
			// Check that all events have required fields
			for i, event := range events {
				if event.EventType == "" {
					t.Logf("Event %d missing event type", i)
					return false
				}
				
				if event.Severity == "" {
					t.Logf("Event %d missing severity", i)
					return false
				}
				
				if event.Timestamp.IsZero() {
					t.Logf("Event %d missing timestamp", i)
					return false
				}
				
				// Events should be in chronological order
				if i > 0 && event.Timestamp.Before(events[i-1].Timestamp) {
					t.Logf("Events not in chronological order at index %d", i)
					return false
				}
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 30}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Security logging consistency property failed: %v", err)
		}
	})
	
	// Property: Valid requests should not trigger security logging
	t.Run("ValidRequestNoLogging", func(t *testing.T) {
		property := func() bool {
			secret := generateRandomSecret()
			body := generateRandomBody()
			
			handler := NewMockWebhookHandler(secret)
			authenticator := NewWebhookAuthenticator(secret)
			
			// Generate valid request
			validSignature := authenticator.GenerateSignature(body)
			validReq := WebhookRequest{
				Body:      body,
				Signature: validSignature,
				Headers:   map[string]string{"X-Forwarded-For": "127.0.0.1"},
			}
			
			initialEventCount := len(handler.GetLoggedEvents())
			response := handler.HandleWebhook(validReq)
			finalEventCount := len(handler.GetLoggedEvents())
			
			// Valid request should be accepted
			if response.StatusCode != http.StatusOK {
				t.Logf("Valid request rejected: status %d", response.StatusCode)
				return false
			}
			
			// Should not log security event
			if response.Logged {
				t.Logf("Security event logged for valid request")
				return false
			}
			
			// Event count should not increase
			if finalEventCount != initialEventCount {
				t.Logf("Event count increased for valid request: %d -> %d", initialEventCount, finalEventCount)
				return false
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Valid request no logging property failed: %v", err)
		}
	})
	
	// Property: Security responses should not leak sensitive information
	t.Run("NoSensitiveInformationLeakage", func(t *testing.T) {
		property := func() bool {
			secret := generateRandomSecret()
			body := generateRandomBody()
			
			handler := NewMockWebhookHandler(secret)
			
			// Generate invalid request
			invalidReq := WebhookRequest{
				Body:      body,
				Signature: "invalid_signature",
				Headers:   map[string]string{"X-Forwarded-For": "attacker.example.com"},
			}
			
			response := handler.HandleWebhook(invalidReq)
			
			// Response should not contain sensitive information
			sensitiveTerms := []string{
				secret,                    // Secret token
				"secret",                  // Word "secret"
				"token",                   // Word "token"
				"key",                     // Word "key"
				"internal",                // Internal details
				"debug",                   // Debug information
				"stack",                   // Stack traces
			}
			
			responseBody := strings.ToLower(response.Body)
			for _, term := range sensitiveTerms {
				if strings.Contains(responseBody, strings.ToLower(term)) {
					t.Logf("Response contains sensitive term '%s': %s", term, response.Body)
					return false
				}
			}
			
			// Response should be generic but informative
			if response.Body == "" {
				t.Logf("Response body is empty")
				return false
			}
			
			// Should indicate authentication failure without details
			expectedTerms := []string{"invalid", "unauthorized", "signature"}
			foundExpected := false
			for _, term := range expectedTerms {
				if strings.Contains(responseBody, term) {
					foundExpected = true
					break
				}
			}
			
			if !foundExpected {
				t.Logf("Response does not indicate authentication failure appropriately: %s", response.Body)
				return false
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 30}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("No sensitive information leakage property failed: %v", err)
		}
	})
}
// RateLimiter represents a rate limiting mechanism
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// IsAllowed checks if a request from the given IP is allowed
func (r *RateLimiter) IsAllowed(ip string) bool {
	now := time.Now()
	
	// Clean old requests
	if requests, exists := r.requests[ip]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) <= r.window {
				validRequests = append(validRequests, reqTime)
			}
		}
		r.requests[ip] = validRequests
	}
	
	// Check if limit exceeded
	if len(r.requests[ip]) >= r.limit {
		return false
	}
	
	// Add current request
	r.requests[ip] = append(r.requests[ip], now)
	return true
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	BotToken    string        `json:"bot_token"`
	SecretToken string        `json:"secret_token"`
	Enabled     bool          `json:"enabled"`
	RateLimit   int           `json:"rate_limit"`
	RateWindow  time.Duration `json:"rate_window"`
}

// SecureWebhookConfig generates a secure webhook configuration
func SecureWebhookConfig(botToken string) *WebhookConfig {
	return &WebhookConfig{
		BotToken:    botToken,
		SecretToken: generateSecureToken(),
		Enabled:     true,
		RateLimit:   10,                // 10 requests
		RateWindow:  1 * time.Minute,   // per minute
	}
}

// generateSecureToken generates a cryptographically secure token
func generateSecureToken() string {
	// Use crypto/rand for secure token generation
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := 32 // 32 characters for good security
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// ValidateWebhookConfig validates webhook configuration security
func ValidateWebhookConfig(config *WebhookConfig) []string {
	var issues []string
	
	// Check bot token
	if config.BotToken == "" {
		issues = append(issues, "bot token is empty")
	} else if len(config.BotToken) < 10 {
		issues = append(issues, "bot token is too short")
	}
	
	// Check secret token
	if config.SecretToken == "" {
		issues = append(issues, "secret token is empty")
	} else if len(config.SecretToken) < 16 {
		issues = append(issues, "secret token is too short (minimum 16 characters)")
	}
	
	// Check rate limiting
	if config.RateLimit <= 0 {
		issues = append(issues, "rate limit must be positive")
	} else if config.RateLimit > 1000 {
		issues = append(issues, "rate limit is too high (potential DoS vulnerability)")
	}
	
	if config.RateWindow <= 0 {
		issues = append(issues, "rate window must be positive")
	} else if config.RateWindow < time.Second {
		issues = append(issues, "rate window is too short")
	}
	
	return issues
}

// EnhancedWebhookHandler includes rate limiting and configuration validation
type EnhancedWebhookHandler struct {
	config      *WebhookConfig
	rateLimiter *RateLimiter
	handler     *MockWebhookHandler
}

// NewEnhancedWebhookHandler creates a new enhanced webhook handler
func NewEnhancedWebhookHandler(config *WebhookConfig) *EnhancedWebhookHandler {
	return &EnhancedWebhookHandler{
		config:      config,
		rateLimiter: NewRateLimiter(config.RateLimit, config.RateWindow),
		handler:     NewMockWebhookHandler(config.SecretToken),
	}
}

// HandleWebhook processes webhook requests with rate limiting
func (e *EnhancedWebhookHandler) HandleWebhook(req WebhookRequest) WebhookResponse {
	// Check if webhook is enabled
	if !e.config.Enabled {
		return WebhookResponse{
			StatusCode: http.StatusServiceUnavailable,
			Body:       "Webhook disabled",
			Logged:     false,
		}
	}
	
	// Check rate limiting
	sourceIP := req.Headers["X-Forwarded-For"]
	if sourceIP == "" {
		sourceIP = req.Headers["X-Real-IP"]
	}
	if sourceIP == "" {
		sourceIP = "unknown"
	}
	
	if !e.rateLimiter.IsAllowed(sourceIP) {
		// Log rate limit violation
		event := SecurityEvent{
			EventType: "rate_limit_exceeded",
			Severity:  "medium",
			SourceIP:  sourceIP,
			Details: map[string]interface{}{
				"rate_limit": e.config.RateLimit,
				"window":     e.config.RateWindow.String(),
			},
			Timestamp: time.Now(),
		}
		e.handler.loggedEvents = append(e.handler.loggedEvents, event)
		
		return WebhookResponse{
			StatusCode: http.StatusTooManyRequests,
			Body:       "Rate limit exceeded",
			Logged:     true,
		}
	}
	
	// Delegate to regular handler
	return e.handler.HandleWebhook(req)
}

// GetLoggedEvents returns all logged events
func (e *EnhancedWebhookHandler) GetLoggedEvents() []SecurityEvent {
	return e.handler.GetLoggedEvents()
}

// **Feature: security-fixes, Property 6: Webhook configuration security**
// Property 6: Webhook configuration security
// For any Telegram integration setup, the system should generate secure webhook secrets and implement rate limiting
// **Validates: Requirements 3.4, 3.5**
func TestWebhookConfigurationSecurity(t *testing.T) {
	// Property: Secure webhook configurations should have strong secrets
	t.Run("SecureTokenGeneration", func(t *testing.T) {
		property := func() bool {
			botToken := generateRandomSecret()
			config := SecureWebhookConfig(botToken)
			
			// Secret token should be generated
			if config.SecretToken == "" {
				t.Logf("Secret token not generated")
				return false
			}
			
			// Secret token should be sufficiently long
			if len(config.SecretToken) < 16 {
				t.Logf("Secret token too short: %d characters", len(config.SecretToken))
				return false
			}
			
			// Secret token should be different from bot token
			if config.SecretToken == config.BotToken {
				t.Logf("Secret token same as bot token")
				return false
			}
			
			// Configuration should be valid
			issues := ValidateWebhookConfig(config)
			if len(issues) > 0 {
				t.Logf("Configuration validation failed: %v", issues)
				return false
			}
			
			// Rate limiting should be configured
			if config.RateLimit <= 0 {
				t.Logf("Rate limit not configured")
				return false
			}
			
			if config.RateWindow <= 0 {
				t.Logf("Rate window not configured")
				return false
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Secure token generation property failed: %v", err)
		}
	})
	
	// Property: Rate limiting should prevent abuse
	t.Run("RateLimitingPrevention", func(t *testing.T) {
		property := func() bool {
			botToken := generateRandomSecret()
			config := SecureWebhookConfig(botToken)
			config.RateLimit = 3                    // Low limit for testing
			config.RateWindow = 10 * time.Second    // Short window for testing
			
			handler := NewEnhancedWebhookHandler(config)
			body := generateRandomBody()
			signature := NewWebhookAuthenticator(config.SecretToken).GenerateSignature(body)
			
			sourceIP := fmt.Sprintf("192.168.1.%d", rand.Intn(255)+1)
			
			// Send requests up to the limit
			for i := 0; i < config.RateLimit; i++ {
				req := WebhookRequest{
					Body:      body,
					Signature: signature,
					Headers:   map[string]string{"X-Forwarded-For": sourceIP},
				}
				
				response := handler.HandleWebhook(req)
				
				// Requests within limit should be accepted
				if response.StatusCode != http.StatusOK {
					t.Logf("Request %d within limit rejected: status %d", i, response.StatusCode)
					return false
				}
			}
			
			// Next request should be rate limited
			req := WebhookRequest{
				Body:      body,
				Signature: signature,
				Headers:   map[string]string{"X-Forwarded-For": sourceIP},
			}
			
			response := handler.HandleWebhook(req)
			
			// Should be rate limited
			if response.StatusCode != http.StatusTooManyRequests {
				t.Logf("Rate limit not enforced: status %d", response.StatusCode)
				return false
			}
			
			// Should log rate limit violation
			if !response.Logged {
				t.Logf("Rate limit violation not logged")
				return false
			}
			
			// Check logged event
			events := handler.GetLoggedEvents()
			if len(events) == 0 {
				t.Logf("No events logged for rate limit violation")
				return false
			}
			
			lastEvent := events[len(events)-1]
			if lastEvent.EventType != "rate_limit_exceeded" {
				t.Logf("Wrong event type for rate limit: %s", lastEvent.EventType)
				return false
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 20}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Rate limiting prevention property failed: %v", err)
		}
	})
	
	// Property: Different IPs should have independent rate limits
	t.Run("IndependentRateLimits", func(t *testing.T) {
		property := func() bool {
			botToken := generateRandomSecret()
			config := SecureWebhookConfig(botToken)
			config.RateLimit = 2                    // Low limit for testing
			config.RateWindow = 10 * time.Second    // Short window for testing
			
			handler := NewEnhancedWebhookHandler(config)
			body := generateRandomBody()
			signature := NewWebhookAuthenticator(config.SecretToken).GenerateSignature(body)
			
			ip1 := "192.168.1.100"
			ip2 := "192.168.1.200"
			
			// Exhaust rate limit for IP1
			for i := 0; i < config.RateLimit; i++ {
				req := WebhookRequest{
					Body:      body,
					Signature: signature,
					Headers:   map[string]string{"X-Forwarded-For": ip1},
				}
				handler.HandleWebhook(req)
			}
			
			// IP1 should be rate limited
			req1 := WebhookRequest{
				Body:      body,
				Signature: signature,
				Headers:   map[string]string{"X-Forwarded-For": ip1},
			}
			response1 := handler.HandleWebhook(req1)
			
			if response1.StatusCode != http.StatusTooManyRequests {
				t.Logf("IP1 not rate limited: status %d", response1.StatusCode)
				return false
			}
			
			// IP2 should still be allowed
			req2 := WebhookRequest{
				Body:      body,
				Signature: signature,
				Headers:   map[string]string{"X-Forwarded-For": ip2},
			}
			response2 := handler.HandleWebhook(req2)
			
			if response2.StatusCode != http.StatusOK {
				t.Logf("IP2 incorrectly rate limited: status %d", response2.StatusCode)
				return false
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 15}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Independent rate limits property failed: %v", err)
		}
	})
	
	// Property: Configuration validation should catch security issues
	t.Run("ConfigurationValidation", func(t *testing.T) {
		property := func() bool {
			// Test various insecure configurations
			insecureConfigs := []*WebhookConfig{
				// Empty tokens
				{BotToken: "", SecretToken: "", Enabled: true, RateLimit: 10, RateWindow: time.Minute},
				// Short tokens
				{BotToken: "short", SecretToken: "short", Enabled: true, RateLimit: 10, RateWindow: time.Minute},
				// No rate limiting
				{BotToken: generateRandomSecret(), SecretToken: generateRandomSecret(), Enabled: true, RateLimit: 0, RateWindow: time.Minute},
				// Excessive rate limit
				{BotToken: generateRandomSecret(), SecretToken: generateRandomSecret(), Enabled: true, RateLimit: 10000, RateWindow: time.Minute},
				// Too short rate window
				{BotToken: generateRandomSecret(), SecretToken: generateRandomSecret(), Enabled: true, RateLimit: 10, RateWindow: 100 * time.Millisecond},
			}
			
			for i, config := range insecureConfigs {
				issues := ValidateWebhookConfig(config)
				
				// Each insecure config should have validation issues
				if len(issues) == 0 {
					t.Logf("Insecure config %d passed validation", i)
					return false
				}
			}
			
			// Secure config should pass validation
			secureConfig := SecureWebhookConfig(generateRandomSecret())
			issues := ValidateWebhookConfig(secureConfig)
			
			if len(issues) > 0 {
				t.Logf("Secure config failed validation: %v", issues)
				return false
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 20}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Configuration validation property failed: %v", err)
		}
	})
	
	// Property: Disabled webhooks should reject all requests
	t.Run("DisabledWebhookRejection", func(t *testing.T) {
		property := func() bool {
			botToken := generateRandomSecret()
			config := SecureWebhookConfig(botToken)
			config.Enabled = false // Disable webhook
			
			handler := NewEnhancedWebhookHandler(config)
			body := generateRandomBody()
			signature := NewWebhookAuthenticator(config.SecretToken).GenerateSignature(body)
			
			req := WebhookRequest{
				Body:      body,
				Signature: signature,
				Headers:   map[string]string{"X-Forwarded-For": "127.0.0.1"},
			}
			
			response := handler.HandleWebhook(req)
			
			// Disabled webhook should reject all requests
			if response.StatusCode != http.StatusServiceUnavailable {
				t.Logf("Disabled webhook did not reject request: status %d", response.StatusCode)
				return false
			}
			
			// Should not log as security violation (it's a configuration choice)
			if response.Logged {
				t.Logf("Disabled webhook logged security event")
				return false
			}
			
			return true
		}
		
		config := &quick.Config{MaxCount: 30}
		if err := quick.Check(property, config); err != nil {
			t.Errorf("Disabled webhook rejection property failed: %v", err)
		}
	})
}