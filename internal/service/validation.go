package service

import (
	"fmt"
	"html"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// ValidationService handles input validation and sanitization
type ValidationService struct {
	securityService *SecurityService
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Name        string
	Pattern     *regexp.Regexp
	MinLength   int
	MaxLength   int
	Required    bool
	Sanitize    bool
	AllowedChars string
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid    bool
	Errors   []string
	Sanitized string
}

// NewValidationService creates a new validation service
func NewValidationService(securityService *SecurityService) *ValidationService {
	return &ValidationService{
		securityService: securityService,
	}
}

// ValidateAndSanitize validates and sanitizes input according to rules
func (vs *ValidationService) ValidateAndSanitize(input string, rules ValidationRule) ValidationResult {
	result := ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Sanitized: input,
	}

	// Check if required field is empty
	if rules.Required && strings.TrimSpace(input) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s is required", rules.Name))
		return result
	}

	// Skip validation for empty optional fields
	if !rules.Required && strings.TrimSpace(input) == "" {
		return result
	}

	// Length validation
	if len(input) < rules.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s must be at least %d characters", rules.Name, rules.MinLength))
	}

	if rules.MaxLength > 0 && len(input) > rules.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s must not exceed %d characters", rules.Name, rules.MaxLength))
	}

	// Pattern validation
	if rules.Pattern != nil && !rules.Pattern.MatchString(input) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s format is invalid", rules.Name))
	}

	// Character whitelist validation
	if rules.AllowedChars != "" {
		for _, char := range input {
			if !strings.ContainsRune(rules.AllowedChars, char) {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("%s contains invalid characters", rules.Name))
				break
			}
		}
	}

	// Sanitization
	if rules.Sanitize {
		result.Sanitized = vs.SanitizeInput(input)
	}

	// Check for potential injection attacks
	if vs.containsSQLInjection(input) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s contains potentially malicious content", rules.Name))
		
		// Log security event
		if vs.securityService != nil {
			details := map[string]interface{}{
				"field": rules.Name,
				"input": input,
				"type":  "sql_injection_attempt",
			}
			vs.securityService.LogSecurityEvent("input_validation", "high", "", "", details)
		}
	}

	if vs.containsXSS(input) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s contains potentially malicious content", rules.Name))
		
		// Log security event
		if vs.securityService != nil {
			details := map[string]interface{}{
				"field": rules.Name,
				"input": input,
				"type":  "xss_attempt",
			}
			vs.securityService.LogSecurityEvent("input_validation", "medium", "", "", details)
		}
	}

	return result
}

// SanitizeInput sanitizes input to prevent XSS and other attacks
func (vs *ValidationService) SanitizeInput(input string) string {
	// HTML escape
	sanitized := html.EscapeString(input)
	
	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
	
	// Normalize whitespace
	sanitized = vs.normalizeWhitespace(sanitized)
	
	return sanitized
}

// ValidateFilePath validates and sanitizes file paths to prevent directory traversal
func (vs *ValidationService) ValidateFilePath(path string) ValidationResult {
	result := ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Sanitized: path,
	}

	// Check for directory traversal patterns
	if vs.containsDirectoryTraversal(path) {
		result.Valid = false
		result.Errors = append(result.Errors, "Path contains directory traversal patterns")
		
		// Log security event
		if vs.securityService != nil {
			details := map[string]interface{}{
				"path": path,
				"type": "directory_traversal_attempt",
			}
			vs.securityService.LogSecurityEvent("input_validation", "high", "", "", details)
		}
		return result
	}

	// Clean and validate the path
	cleanPath := filepath.Clean(path)
	
	// Ensure path doesn't go outside allowed directory
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, "/../") {
		result.Valid = false
		result.Errors = append(result.Errors, "Path attempts to access parent directories")
		return result
	}

	result.Sanitized = cleanPath
	return result
}

// ValidateEmail validates email format
func (vs *ValidationService) ValidateEmail(email string) ValidationResult {
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	
	rule := ValidationRule{
		Name:      "Email",
		Pattern:   emailPattern,
		MinLength: 5,
		MaxLength: 254,
		Required:  true,
		Sanitize:  true,
	}
	
	return vs.ValidateAndSanitize(email, rule)
}

// ValidatePassword validates password strength
func (vs *ValidationService) ValidatePassword(password string) ValidationResult {
	result := ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Sanitized: password, // Don't sanitize passwords
	}

	if len(password) < 8 {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must be at least 8 characters long")
	}

	if len(password) > 128 {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must not exceed 128 characters")
	}

	// Check for at least one uppercase, lowercase, digit, and special character
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one uppercase letter")
	}

	if !hasLower {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one lowercase letter")
	}

	if !hasDigit {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one digit")
	}

	if !hasSpecial {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one special character")
	}

	return result
}

// ValidateURL validates and sanitizes URLs
func (vs *ValidationService) ValidateURL(urlStr string) ValidationResult {
	result := ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Sanitized: urlStr,
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "Invalid URL format")
		return result
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		result.Valid = false
		result.Errors = append(result.Errors, "URL must use http or https scheme")
		return result
	}

	// Check for malicious patterns in URL
	if vs.containsXSS(urlStr) {
		result.Valid = false
		result.Errors = append(result.Errors, "URL contains potentially malicious content")
		return result
	}

	result.Sanitized = parsedURL.String()
	return result
}

// containsSQLInjection checks for SQL injection patterns
func (vs *ValidationService) containsSQLInjection(input string) bool {
	input = strings.ToLower(input)
	
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
		"' or ''='",
		"x' or 1=1 or 'x'='y",
		"' or 'a'='a",
		"') or ('1'='1",
		"' or sleep(",
		"' or benchmark(",
		"' waitfor delay",
	}

	for _, pattern := range patterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}

	return false
}

// containsXSS checks for XSS patterns
func (vs *ValidationService) containsXSS(input string) bool {
	input = strings.ToLower(input)
	
	patterns := []string{
		"<script",
		"</script>",
		"javascript:",
		"onload=",
		"onerror=",
		"onclick=",
		"onmouseover=",
		"<iframe",
		"<object",
		"<embed",
		"eval(",
		"expression(",
		"vbscript:",
		"data:text/html",
		"<svg",
		"<img",
		"src=javascript:",
		"href=javascript:",
	}

	for _, pattern := range patterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}

	return false
}

// containsDirectoryTraversal checks for directory traversal patterns
func (vs *ValidationService) containsDirectoryTraversal(path string) bool {
	path = strings.ToLower(path)
	
	patterns := []string{
		"../",
		"..\\",
		"..%2f",
		"..%5c",
		"%2e%2e%2f",
		"%2e%2e%5c",
		"....//",
		"....\\\\",
	}

	for _, pattern := range patterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

// normalizeWhitespace normalizes whitespace in input
func (vs *ValidationService) normalizeWhitespace(input string) string {
	// Replace multiple whitespace with single space
	re := regexp.MustCompile(`\s+`)
	normalized := re.ReplaceAllString(input, " ")
	
	// Trim leading and trailing whitespace
	return strings.TrimSpace(normalized)
}

// CreateParameterizedQuery creates a parameterized query to prevent SQL injection
func (vs *ValidationService) CreateParameterizedQuery(query string, params ...interface{}) (string, []interface{}) {
	// This is a simple example - in practice, you'd use your ORM's parameterized queries
	// GORM automatically handles parameterization, but this shows the concept
	
	// Count placeholders
	placeholderCount := strings.Count(query, "?")
	if placeholderCount != len(params) {
		// Log potential SQL injection attempt
		if vs.securityService != nil {
			details := map[string]interface{}{
				"query":             query,
				"placeholder_count": placeholderCount,
				"param_count":       len(params),
				"type":              "sql_injection_attempt",
			}
			vs.securityService.LogSecurityEvent("sql_injection", "high", "", "", details)
		}
	}
	
	return query, params
}

// ValidateJSONInput validates JSON input structure
func (vs *ValidationService) ValidateJSONInput(input string, maxDepth, maxKeys int) ValidationResult {
	result := ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Sanitized: input,
	}

	// Basic length check
	if len(input) > 1024*1024 { // 1MB limit
		result.Valid = false
		result.Errors = append(result.Errors, "JSON input too large")
		return result
	}

	// Check for potential JSON injection patterns
	if strings.Contains(input, "__proto__") || strings.Contains(input, "constructor") {
		result.Valid = false
		result.Errors = append(result.Errors, "JSON contains potentially dangerous properties")
		
		if vs.securityService != nil {
			details := map[string]interface{}{
				"input": input,
				"type":  "json_injection_attempt",
			}
			vs.securityService.LogSecurityEvent("input_validation", "medium", "", "", details)
		}
	}

	return result
}

// GetCommonValidationRules returns commonly used validation rules
func (vs *ValidationService) GetCommonValidationRules() map[string]ValidationRule {
	return map[string]ValidationRule{
		"username": {
			Name:         "Username",
			Pattern:      regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`),
			MinLength:    3,
			MaxLength:    30,
			Required:     true,
			Sanitize:     true,
			AllowedChars: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-",
		},
		"name": {
			Name:      "Name",
			Pattern:   regexp.MustCompile(`^[a-zA-Z\s'-]{1,100}$`),
			MinLength: 1,
			MaxLength: 100,
			Required:  true,
			Sanitize:  true,
		},
		"phone": {
			Name:      "Phone",
			Pattern:   regexp.MustCompile(`^\+?[1-9]\d{1,14}$`),
			MinLength: 10,
			MaxLength: 15,
			Required:  false,
			Sanitize:  true,
		},
		"token": {
			Name:         "Token",
			Pattern:      regexp.MustCompile(`^[a-zA-Z0-9]{32,128}$`),
			MinLength:    32,
			MaxLength:    128,
			Required:     true,
			Sanitize:     false,
			AllowedChars: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		},
	}
}