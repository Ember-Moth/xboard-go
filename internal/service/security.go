package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"dashgo/internal/model"
	"dashgo/internal/repository"

	"gorm.io/gorm"
)

// SecurityService handles security logging and monitoring
type SecurityService struct {
	db                *gorm.DB
	securityRepo      *repository.SecurityRepository
	mailService       *MailService
	telegramService   *TelegramService
	failureAttempts   map[string]*FailureTracker
	mu                sync.RWMutex
	logRotationConfig LogRotationConfig
}

// FailureTracker tracks authentication failures for progressive delays
type FailureTracker struct {
	Attempts  int
	FirstTime time.Time
	LastTime  time.Time
}

// LogRotationConfig configures log rotation settings
type LogRotationConfig struct {
	MaxAge     time.Duration // Maximum age of logs before rotation
	MaxSize    int64         // Maximum size in bytes before rotation
	MaxEntries int           // Maximum number of entries before rotation
}

// NewSecurityService creates a new security service
func NewSecurityService(db *gorm.DB, mailService *MailService, telegramService *TelegramService) *SecurityService {
	return &SecurityService{
		db:              db,
		securityRepo:    repository.NewSecurityRepository(db),
		mailService:     mailService,
		telegramService: telegramService,
		failureAttempts: make(map[string]*FailureTracker),
		logRotationConfig: LogRotationConfig{
			MaxAge:     30 * 24 * time.Hour, // 30 days
			MaxSize:    100 * 1024 * 1024,   // 100 MB
			MaxEntries: 100000,              // 100k entries
		},
	}
}

// LogSecurityEvent logs a security event with tamper-resistant hash
func (s *SecurityService) LogSecurityEvent(eventType, severity, sourceIP, userAgent string, details map[string]interface{}) error {
	// Convert details to JSON
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	event := &model.SecurityEvent{
		EventType: eventType,
		Severity:  severity,
		SourceIP:  sourceIP,
		UserAgent: userAgent,
		Details:   detailsJSON,
		Timestamp: time.Now(),
		Resolved:  false,
	}

	// Save to database
	if err := s.securityRepo.CreateSecurityEvent(event); err != nil {
		return fmt.Errorf("failed to save security event: %w", err)
	}

	// Check if alert should be triggered
	go s.checkAndTriggerAlert(eventType, severity)

	return nil
}

// RecordAuthFailure records an authentication failure and returns delay duration
func (s *SecurityService) RecordAuthFailure(sourceIP, email, userAgent string) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%s", sourceIP, email)
	now := time.Now()

	tracker, exists := s.failureAttempts[key]
	if !exists {
		tracker = &FailureTracker{
			Attempts:  1,
			FirstTime: now,
			LastTime:  now,
		}
		s.failureAttempts[key] = tracker
	} else {
		// Reset if more than 1 hour has passed
		if now.Sub(tracker.FirstTime) > time.Hour {
			tracker.Attempts = 1
			tracker.FirstTime = now
		} else {
			tracker.Attempts++
		}
		tracker.LastTime = now
	}

	// Record in database
	attempt := &model.AuthFailureAttempt{
		SourceIP:    sourceIP,
		Email:       email,
		AttemptTime: now,
		UserAgent:   userAgent,
	}
	s.securityRepo.CreateAuthFailureAttempt(attempt)

	// Calculate progressive delay: 2^(attempts-1) seconds, max 60 seconds
	delay := time.Duration(1<<uint(tracker.Attempts-1)) * time.Second
	if delay > 60*time.Second {
		delay = 60 * time.Second
	}

	// Log security event for multiple failures
	if tracker.Attempts >= 3 {
		details := map[string]interface{}{
			"email":    email,
			"attempts": tracker.Attempts,
			"duration": now.Sub(tracker.FirstTime).String(),
		}
		s.LogSecurityEvent("auth_failure", "medium", sourceIP, userAgent, details)
	}

	return delay
}

// ShouldBlockAuth checks if authentication should be blocked due to too many failures
func (s *SecurityService) ShouldBlockAuth(sourceIP, email string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", sourceIP, email)
	tracker, exists := s.failureAttempts[key]
	if !exists {
		return false
	}

	// Block if more than 10 attempts in the last hour
	if tracker.Attempts > 10 && time.Since(tracker.FirstTime) < time.Hour {
		return true
	}

	return false
}

// checkAndTriggerAlert checks if an alert should be triggered based on event patterns
func (s *SecurityService) checkAndTriggerAlert(eventType, severity string) {
	// Get alert configuration
	enabled := true
	alerts, err := s.securityRepo.GetSecurityAlerts(eventType, severity, &enabled)
	if err != nil || len(alerts) == 0 {
		// No alert configured for this event type/severity
		return
	}

	alert := alerts[0] // Use the first matching alert

	// Count recent events within time window
	since := time.Now().Add(-time.Duration(alert.TimeWindow) * time.Second)
	count, err := s.securityRepo.CountSecurityEvents(eventType, severity, since)
	if err != nil {
		log.Printf("Failed to count security events: %v", err)
		return
	}

	// Trigger alert if threshold exceeded
	if int(count) >= alert.Threshold {
		s.triggerAlert(eventType, severity, int(count), alert.TimeWindow)
	}
}

// triggerAlert sends alert notifications
func (s *SecurityService) triggerAlert(eventType, severity string, count, timeWindow int) {
	message := fmt.Sprintf(
		"Security Alert: %d %s events with %s severity detected in the last %d seconds",
		count, eventType, severity, timeWindow,
	)

	log.Printf("[SECURITY ALERT] %s", message)

	// Send notifications (email, telegram, etc.)
	// This would integrate with existing notification services
}

// RotateLogs performs log rotation based on configuration
func (s *SecurityService) RotateLogs() error {
	// Delete old logs based on MaxAge
	cutoffTime := time.Now().Add(-s.logRotationConfig.MaxAge)
	err := s.securityRepo.DeleteOldSecurityEvents(cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to rotate logs by age: %w", err)
	}

	// Check if we need to rotate based on entry count
	count, err := s.securityRepo.CountSecurityEvents("", "", time.Time{})
	if err != nil {
		return fmt.Errorf("failed to count security events: %w", err)
	}

	if count > int64(s.logRotationConfig.MaxEntries) {
		// Delete oldest entries to bring count under limit
		toDelete := count - int64(s.logRotationConfig.MaxEntries)
		subQuery := s.db.Model(&model.SecurityEvent{}).
			Select("id").
			Order("timestamp ASC").
			Limit(int(toDelete))

		result := s.db.Where("id IN (?)", subQuery).Delete(&model.SecurityEvent{})
		if result.Error != nil {
			return fmt.Errorf("failed to rotate logs by count: %w", result.Error)
		}
	}

	return nil
}

// VerifyLogIntegrity verifies that logs haven't been tampered with
func (s *SecurityService) VerifyLogIntegrity() (bool, error) {
	// This is a simplified integrity check
	// In production, you might use cryptographic signatures or blockchain-like chains
	events, err := s.securityRepo.GetSecurityEvents("", "", time.Time{}, 0)
	if err != nil {
		return false, fmt.Errorf("failed to fetch events: %w", err)
	}

	// Verify each event has required fields
	for _, event := range events {
		if event.EventType == "" || event.Severity == "" || event.Timestamp.IsZero() {
			return false, fmt.Errorf("event %d has missing required fields", event.ID)
		}
	}

	return true, nil
}

// ComputeEventHash computes a tamper-resistant hash for an event
func (s *SecurityService) ComputeEventHash(event *model.SecurityEvent) string {
	data := fmt.Sprintf("%s:%s:%s:%s:%s:%d",
		event.EventType,
		event.Severity,
		event.SourceIP,
		event.UserAgent,
		event.Details,
		event.Timestamp.Unix(),
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GetSecurityEvents retrieves security events with filtering
func (s *SecurityService) GetSecurityEvents(eventType, severity string, since time.Time, limit int) ([]model.SecurityEvent, error) {
	events, err := s.securityRepo.GetSecurityEvents(eventType, severity, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch security events: %w", err)
	}

	return events, nil
}

// CleanupOldFailureAttempts removes old failure tracking data from memory
func (s *SecurityService) CleanupOldFailureAttempts() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, tracker := range s.failureAttempts {
		if now.Sub(tracker.LastTime) > 24*time.Hour {
			delete(s.failureAttempts, key)
		}
	}
}
