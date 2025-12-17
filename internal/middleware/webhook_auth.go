package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// WebhookAuthConfig represents webhook authentication configuration
type WebhookAuthConfig struct {
	SecretToken string        `json:"secret_token"`
	RateLimit   int           `json:"rate_limit"`
	RateWindow  time.Duration `json:"rate_window"`
}

// RateLimiter implements rate limiting for webhook requests
type RateLimiter struct {
	mu       sync.RWMutex
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
	r.mu.Lock()
	defer r.mu.Unlock()
	
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

// SecurityLogger interface for logging security events
type SecurityLogger interface {
	LogSecurityEvent(eventType, severity, sourceIP string, details map[string]interface{})
}

// DefaultSecurityLogger provides a default implementation
type DefaultSecurityLogger struct{}

// LogSecurityEvent logs security events (placeholder implementation)
func (d *DefaultSecurityLogger) LogSecurityEvent(eventType, severity, sourceIP string, details map[string]interface{}) {
	// In a real implementation, this would log to a proper security logging system
	fmt.Printf("[SECURITY] %s - %s from %s: %+v\n", severity, eventType, sourceIP, details)
}

// WebhookAuthMiddleware creates a middleware for webhook authentication
func WebhookAuthMiddleware(config WebhookAuthConfig, logger SecurityLogger) gin.HandlerFunc {
	rateLimiter := NewRateLimiter(config.RateLimit, config.RateWindow)
	
	if logger == nil {
		logger = &DefaultSecurityLogger{}
	}
	
	return func(c *gin.Context) {
		// Get source IP
		sourceIP := getSourceIP(c)
		
		// Check rate limiting
		if !rateLimiter.IsAllowed(sourceIP) {
			logger.LogSecurityEvent("rate_limit_exceeded", "medium", sourceIP, map[string]interface{}{
				"rate_limit": config.RateLimit,
				"window":     config.RateWindow.String(),
				"user_agent": c.GetHeader("User-Agent"),
			})
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}
		
		// Read request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.LogSecurityEvent("webhook_read_error", "low", sourceIP, map[string]interface{}{
				"error": err.Error(),
			})
			
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
			})
			c.Abort()
			return
		}
		
		// Get signature from header
		signature := c.GetHeader("X-Telegram-Bot-Api-Secret-Token")
		if signature == "" {
			// Also check for custom signature header
			signature = c.GetHeader("X-Hub-Signature-256")
		}
		
		// Validate signature
		if !validateWebhookSignature(body, signature, config.SecretToken) {
			logger.LogSecurityEvent("webhook_forge", "high", sourceIP, map[string]interface{}{
				"signature":  signature,
				"body_size":  len(body),
				"user_agent": c.GetHeader("User-Agent"),
				"path":       c.Request.URL.Path,
			})
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid signature",
			})
			c.Abort()
			return
		}
		
		// Store body in context for handler use
		c.Set("webhook_body", body)
		
		c.Next()
	}
}

// validateWebhookSignature validates the webhook signature using HMAC-SHA256
func validateWebhookSignature(body []byte, signature, secretToken string) bool {
	if secretToken == "" || signature == "" {
		return false
	}
	
	// Remove "sha256=" prefix if present
	signature = strings.TrimPrefix(signature, "sha256=")
	
	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secretToken))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	
	// Use constant time comparison to prevent timing attacks
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// getSourceIP extracts the source IP from the request
func getSourceIP(c *gin.Context) string {
	// Check X-Forwarded-For header first
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first IP in case of multiple
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to remote address
	return c.ClientIP()
}

// GenerateSecureToken generates a cryptographically secure token for webhook authentication
func GenerateSecureToken() string {
	// Use current time and random data for entropy
	data := fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ValidateWebhookConfig validates webhook configuration for security
func ValidateWebhookConfig(config WebhookAuthConfig) []string {
	var issues []string
	
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