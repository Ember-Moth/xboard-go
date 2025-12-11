package model

import "time"

// AgentVersion Agent 版本配置
type AgentVersion struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Version      string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"version"`
	DownloadURL  string    `gorm:"type:varchar(500);not null" json:"download_url"`
	SHA256       string    `gorm:"type:varchar(64);not null" json:"sha256"`
	FileSize     int64     `gorm:"not null" json:"file_size"`
	Strategy     string    `gorm:"type:varchar(20);not null;default:'manual'" json:"strategy"` // auto, manual
	ReleaseNotes string    `gorm:"type:text" json:"release_notes"`
	IsLatest     bool      `gorm:"default:false" json:"is_latest"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AgentVersion) TableName() string {
	return "agent_versions"
}

// AgentUpdateLog Agent 更新日志
type AgentUpdateLog struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	HostID       int64     `gorm:"not null;index" json:"host_id"`
	FromVersion  string    `gorm:"type:varchar(50);not null" json:"from_version"`
	ToVersion    string    `gorm:"type:varchar(50);not null" json:"to_version"`
	Status       string    `gorm:"type:varchar(20);not null" json:"status"` // success, failed, rollback
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (AgentUpdateLog) TableName() string {
	return "agent_update_logs"
}
