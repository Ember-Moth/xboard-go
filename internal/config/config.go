package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Node     NodeConfig     `yaml:"node"`
	Mail     MailConfig     `yaml:"mail"`
	Telegram TelegramConfig `yaml:"telegram"`
	Admin    AdminConfig    `yaml:"admin"`
}

type AdminConfig struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

type AppConfig struct {
	Name   string `yaml:"name"`
	Mode   string `yaml:"mode"` // debug, release
	Listen string `yaml:"listen"`
}

type DatabaseConfig struct {
	Driver   string `yaml:"driver"` // mysql, sqlite
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	
	// SQLite specific configuration
	MaxOpenConns    int    `yaml:"max_open_conns"`     // Maximum open connections
	MaxIdleConns    int    `yaml:"max_idle_conns"`     // Maximum idle connections
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`  // Connection maximum lifetime in seconds
	WALMode         bool   `yaml:"wal_mode"`           // Enable WAL mode for better concurrency
	BusyTimeout     int    `yaml:"busy_timeout"`       // Busy timeout in milliseconds
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireHour int    `yaml:"expire_hour"`
}

type NodeConfig struct {
	Token        string `yaml:"token"`         // Node communication token
	PushInterval int    `yaml:"push_interval"` // seconds
	PullInterval int    `yaml:"pull_interval"` // seconds
	EnableSync   bool   `yaml:"enable_sync"`   // 是否启用主动节点同步（默认 false，使用 Agent 模式）
}

type MailConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	FromName   string `yaml:"from_name"`
	FromAddr   string `yaml:"from_addr"`
	Encryption string `yaml:"encryption"` // ssl, tls, none
}

type TelegramConfig struct {
	BotToken    string `yaml:"bot_token"`
	ChatID      string `yaml:"chat_id"`
	SecretToken string `yaml:"secret_token"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.App.Listen == "" {
		cfg.App.Listen = ":8080"
	}
	if cfg.JWT.ExpireHour == 0 {
		cfg.JWT.ExpireHour = 24
	}
	if cfg.Node.PushInterval == 0 {
		cfg.Node.PushInterval = 60
	}
	if cfg.Node.PullInterval == 0 {
		cfg.Node.PullInterval = 60
	}
	
	// Set SQLite defaults
	if cfg.Database.Driver == "sqlite" {
		if cfg.Database.MaxOpenConns == 0 {
			cfg.Database.MaxOpenConns = 25
		}
		if cfg.Database.MaxIdleConns == 0 {
			cfg.Database.MaxIdleConns = 5
		}
		if cfg.Database.ConnMaxLifetime == 0 {
			cfg.Database.ConnMaxLifetime = 300 // 5 minutes
		}
		if cfg.Database.BusyTimeout == 0 {
			cfg.Database.BusyTimeout = 30000 // 30 seconds
		}
		// WAL mode is enabled by default for better concurrency
		cfg.Database.WALMode = true
	}

	return &cfg, nil
}
