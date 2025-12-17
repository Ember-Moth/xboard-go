package model

import (
	"encoding/json"
	"time"
)

// SecurityEvent represents a security event in the system
type SecurityEvent struct {
	ID        int64           `gorm:"primaryKey;column:id" json:"id"`
	EventType string          `gorm:"column:event_type;size:50;not null" json:"event_type"`   // "auth_failure", "webhook_forge", "port_scan"
	Severity  string          `gorm:"column:severity;size:20;not null" json:"severity"`       // "low", "medium", "high", "critical"
	SourceIP  string          `gorm:"column:source_ip;size:45" json:"source_ip"`
	UserAgent string          `gorm:"column:user_agent;size:500" json:"user_agent"`
	Details   json.RawMessage `gorm:"column:details;type:text" json:"details"`
	Timestamp time.Time       `gorm:"column:timestamp;not null" json:"timestamp"`
	Resolved  bool            `gorm:"column:resolved;default:false" json:"resolved"`
	CreatedAt int64           `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt int64           `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (SecurityEvent) TableName() string {
	return "v2_security_events"
}

// AuthFailureAttempt tracks authentication failure attempts for rate limiting
type AuthFailureAttempt struct {
	ID          int64     `gorm:"primaryKey;column:id" json:"id"`
	SourceIP    string    `gorm:"column:source_ip;size:45;not null;index" json:"source_ip"`
	Email       string    `gorm:"column:email;size:64;index" json:"email"`
	AttemptTime time.Time `gorm:"column:attempt_time;not null" json:"attempt_time"`
	UserAgent   string    `gorm:"column:user_agent;size:500" json:"user_agent"`
	CreatedAt   int64     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (AuthFailureAttempt) TableName() string {
	return "v2_auth_failure_attempts"
}

// SecurityAlert represents a security alert configuration
type SecurityAlert struct {
	ID          int64  `gorm:"primaryKey;column:id" json:"id"`
	EventType   string `gorm:"column:event_type;size:50;not null" json:"event_type"`
	Severity    string `gorm:"column:severity;size:20;not null" json:"severity"`
	Enabled     bool   `gorm:"column:enabled;default:true" json:"enabled"`
	Threshold   int    `gorm:"column:threshold;default:1" json:"threshold"`        // Number of events to trigger alert
	TimeWindow  int    `gorm:"column:time_window;default:300" json:"time_window"` // Time window in seconds
	NotifyEmail bool   `gorm:"column:notify_email;default:false" json:"notify_email"`
	NotifyTG    bool   `gorm:"column:notify_tg;default:false" json:"notify_tg"`
	CreatedAt   int64  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (SecurityAlert) TableName() string {
	return "v2_security_alerts"
}