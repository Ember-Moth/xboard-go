package middleware

import (
	"net/http"
	"strings"
	"time"

	"dashgo/internal/service"

	"github.com/gin-gonic/gin"
)

// SecurityMiddleware provides security logging and monitoring
type SecurityMiddleware struct {
	securityService *service.SecurityService
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(securityService *service.SecurityService) *SecurityMiddleware {
	return &SecurityMiddleware{
		securityService: securityService,
	}
}

// LogSecurityEvent logs a security event
func (m *SecurityMiddleware) LogSecurityEvent(eventType, severity string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client information
		sourceIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		
		// Prepare event details
		details := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"referer":    c.GetHeader("Referer"),
			"timestamp":  time.Now().Unix(),
		}

		// Log the security event
		err := m.securityService.LogSecurityEvent(eventType, severity, sourceIP, userAgent, details)
		if err != nil {
			// Log error but don't fail the request
			c.Header("X-Security-Log-Error", "true")
		}

		c.Next()
	}
}

// AuthFailureHandler handles authentication failures with progressive delays
func (m *SecurityMiddleware) AuthFailureHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware should be called after authentication fails
		if c.GetBool("auth_failed") {
			sourceIP := c.ClientIP()
			userAgent := c.GetHeader("User-Agent")
			
			// Get email from request (if available)
			email := c.GetString("attempted_email")
			if email == "" {
				// Try to extract from request body or form
				if c.Request.Method == "POST" {
					email = c.PostForm("email")
					if email == "" {
						email = c.PostForm("username")
					}
				}
			}

			// Check if auth should be blocked
			if m.securityService.ShouldBlockAuth(sourceIP, email) {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Too many authentication failures. Account temporarily blocked.",
				})
				c.Abort()
				return
			}

			// Record the failure and get delay
			delay := m.securityService.RecordAuthFailure(sourceIP, email, userAgent)
			
			// Apply progressive delay
			if delay > 0 {
				time.Sleep(delay)
			}

			// Set response headers to indicate rate limiting
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", time.Now().Add(delay).Format(time.RFC3339))
		}

		c.Next()
	}
}

// SuspiciousActivityDetector detects and logs suspicious activities
func (m *SecurityMiddleware) SuspiciousActivityDetector() gin.HandlerFunc {
	return func(c *gin.Context) {
		sourceIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		path := c.Request.URL.Path
		method := c.Request.Method

		// Detect potential SQL injection attempts
		if m.containsSQLInjectionPatterns(c.Request.URL.RawQuery) {
			details := map[string]interface{}{
				"method":     method,
				"path":       path,
				"query":      c.Request.URL.RawQuery,
				"detection":  "sql_injection_attempt",
			}
			m.securityService.LogSecurityEvent("sql_injection", "high", sourceIP, userAgent, details)
		}

		// Detect potential XSS attempts
		if m.containsXSSPatterns(c.Request.URL.RawQuery) {
			details := map[string]interface{}{
				"method":     method,
				"path":       path,
				"query":      c.Request.URL.RawQuery,
				"detection":  "xss_attempt",
			}
			m.securityService.LogSecurityEvent("xss_attempt", "medium", sourceIP, userAgent, details)
		}

		// Detect potential directory traversal attempts
		if m.containsDirectoryTraversalPatterns(path) {
			details := map[string]interface{}{
				"method":     method,
				"path":       path,
				"detection":  "directory_traversal_attempt",
			}
			m.securityService.LogSecurityEvent("directory_traversal", "high", sourceIP, userAgent, details)
		}

		// Detect suspicious user agents
		if m.isSuspiciousUserAgent(userAgent) {
			details := map[string]interface{}{
				"method":     method,
				"path":       path,
				"user_agent": userAgent,
				"detection":  "suspicious_user_agent",
			}
			m.securityService.LogSecurityEvent("suspicious_activity", "low", sourceIP, userAgent, details)
		}

		c.Next()
	}
}

// containsSQLInjectionPatterns checks for common SQL injection patterns
func (m *SecurityMiddleware) containsSQLInjectionPatterns(query string) bool {
	query = strings.ToLower(query)
	patterns := []string{
		"union select",
		"' or '1'='1",
		"' or 1=1",
		"'; drop table",
		"'; delete from",
		"' union all select",
		"' and 1=1",
		"' and '1'='1",
		"admin'--",
		"admin'/*",
	}

	for _, pattern := range patterns {
		if strings.Contains(query, pattern) {
			return true
		}
	}
	return false
}

// containsXSSPatterns checks for common XSS patterns
func (m *SecurityMiddleware) containsXSSPatterns(query string) bool {
	query = strings.ToLower(query)
	patterns := []string{
		"<script",
		"javascript:",
		"onload=",
		"onerror=",
		"onclick=",
		"onmouseover=",
		"<iframe",
		"<object",
		"<embed",
		"eval(",
	}

	for _, pattern := range patterns {
		if strings.Contains(query, pattern) {
			return true
		}
	}
	return false
}

// containsDirectoryTraversalPatterns checks for directory traversal patterns
func (m *SecurityMiddleware) containsDirectoryTraversalPatterns(path string) bool {
	patterns := []string{
		"../",
		"..\\",
		"..%2f",
		"..%5c",
		"%2e%2e%2f",
		"%2e%2e%5c",
	}

	pathLower := strings.ToLower(path)
	for _, pattern := range patterns {
		if strings.Contains(pathLower, pattern) {
			return true
		}
	}
	return false
}

// isSuspiciousUserAgent checks for suspicious user agents
func (m *SecurityMiddleware) isSuspiciousUserAgent(userAgent string) bool {
	if userAgent == "" {
		return true // Empty user agent is suspicious
	}

	userAgentLower := strings.ToLower(userAgent)
	suspiciousPatterns := []string{
		"sqlmap",
		"nikto",
		"nmap",
		"masscan",
		"burp",
		"zap",
		"w3af",
		"havij",
		"pangolin",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(userAgentLower, pattern) {
			return true
		}
	}
	return false
}

// RateLimitByIP provides basic rate limiting by IP address
func (m *SecurityMiddleware) RateLimitByIP(maxRequests int, window time.Duration) gin.HandlerFunc {
	// This is a simple in-memory rate limiter
	// In production, you might want to use Redis or another distributed cache
	requestCounts := make(map[string][]time.Time)
	
	return func(c *gin.Context) {
		sourceIP := c.ClientIP()
		now := time.Now()

		// Clean old entries
		if requests, exists := requestCounts[sourceIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < window {
					validRequests = append(validRequests, reqTime)
				}
			}
			requestCounts[sourceIP] = validRequests
		}

		// Check rate limit
		if len(requestCounts[sourceIP]) >= maxRequests {
			// Log rate limit violation
			details := map[string]interface{}{
				"method":       c.Request.Method,
				"path":         c.Request.URL.Path,
				"requests":     len(requestCounts[sourceIP]),
				"max_requests": maxRequests,
				"window":       window.String(),
			}
			m.securityService.LogSecurityEvent("rate_limit_exceeded", "medium", sourceIP, c.GetHeader("User-Agent"), details)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		// Add current request
		requestCounts[sourceIP] = append(requestCounts[sourceIP], now)
		c.Next()
	}
}