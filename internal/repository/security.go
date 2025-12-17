package repository

import (
	"time"

	"dashgo/internal/model"

	"gorm.io/gorm"
)

type SecurityRepository struct {
	db *gorm.DB
}

func NewSecurityRepository(db *gorm.DB) *SecurityRepository {
	return &SecurityRepository{db: db}
}

// CreateSecurityEvent creates a new security event
func (r *SecurityRepository) CreateSecurityEvent(event *model.SecurityEvent) error {
	return r.db.Create(event).Error
}

// GetSecurityEvents retrieves security events with filtering
func (r *SecurityRepository) GetSecurityEvents(eventType, severity string, since time.Time, limit int) ([]model.SecurityEvent, error) {
	var events []model.SecurityEvent
	query := r.db.Model(&model.SecurityEvent{})

	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if !since.IsZero() {
		query = query.Where("timestamp >= ?", since)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("timestamp DESC").Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

// CountSecurityEvents counts security events matching criteria
func (r *SecurityRepository) CountSecurityEvents(eventType, severity string, since time.Time) (int64, error) {
	var count int64
	query := r.db.Model(&model.SecurityEvent{})

	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if !since.IsZero() {
		query = query.Where("timestamp >= ?", since)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// DeleteOldSecurityEvents deletes security events older than the specified time
func (r *SecurityRepository) DeleteOldSecurityEvents(before time.Time) error {
	return r.db.Where("timestamp < ?", before).Delete(&model.SecurityEvent{}).Error
}

// CreateAuthFailureAttempt records an authentication failure attempt
func (r *SecurityRepository) CreateAuthFailureAttempt(attempt *model.AuthFailureAttempt) error {
	return r.db.Create(attempt).Error
}

// GetAuthFailureAttempts retrieves authentication failure attempts
func (r *SecurityRepository) GetAuthFailureAttempts(sourceIP, email string, since time.Time) ([]model.AuthFailureAttempt, error) {
	var attempts []model.AuthFailureAttempt
	query := r.db.Model(&model.AuthFailureAttempt{})

	if sourceIP != "" {
		query = query.Where("source_ip = ?", sourceIP)
	}
	if email != "" {
		query = query.Where("email = ?", email)
	}
	if !since.IsZero() {
		query = query.Where("attempt_time >= ?", since)
	}

	if err := query.Order("attempt_time DESC").Find(&attempts).Error; err != nil {
		return nil, err
	}

	return attempts, nil
}

// CountAuthFailureAttempts counts authentication failure attempts
func (r *SecurityRepository) CountAuthFailureAttempts(sourceIP, email string, since time.Time) (int64, error) {
	var count int64
	query := r.db.Model(&model.AuthFailureAttempt{})

	if sourceIP != "" {
		query = query.Where("source_ip = ?", sourceIP)
	}
	if email != "" {
		query = query.Where("email = ?", email)
	}
	if !since.IsZero() {
		query = query.Where("attempt_time >= ?", since)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// CreateSecurityAlert creates a security alert configuration
func (r *SecurityRepository) CreateSecurityAlert(alert *model.SecurityAlert) error {
	return r.db.Create(alert).Error
}

// GetSecurityAlerts retrieves security alert configurations
func (r *SecurityRepository) GetSecurityAlerts(eventType, severity string, enabled *bool) ([]model.SecurityAlert, error) {
	var alerts []model.SecurityAlert
	query := r.db.Model(&model.SecurityAlert{})

	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	if err := query.Find(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}

// UpdateSecurityAlert updates a security alert configuration
func (r *SecurityRepository) UpdateSecurityAlert(alert *model.SecurityAlert) error {
	return r.db.Save(alert).Error
}

// DeleteSecurityAlert deletes a security alert configuration
func (r *SecurityRepository) DeleteSecurityAlert(id int64) error {
	return r.db.Delete(&model.SecurityAlert{}, id).Error
}