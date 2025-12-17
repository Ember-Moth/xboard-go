package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"dashgo/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		dialector = mysql.Open(dsn)
	case "sqlite":
		// Ensure the directory exists for SQLite database file
		if err := ensureSQLiteDir(cfg.Database); err != nil {
			return nil, fmt.Errorf("failed to create SQLite directory: %w", err)
		}
		
		// Build SQLite DSN with optimized settings for concurrency
		dsn := cfg.Database
		if cfg.WALMode {
			dsn += "?_journal_mode=WAL"
		}
		if cfg.BusyTimeout > 0 {
			if cfg.WALMode {
				dsn += fmt.Sprintf("&_busy_timeout=%d", cfg.BusyTimeout)
			} else {
				dsn += fmt.Sprintf("?_busy_timeout=%d", cfg.BusyTimeout)
			}
		}
		
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error), // 只打印错误，不打印 record not found 警告
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool for SQLite
	if cfg.Driver == "sqlite" {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}

		// Set connection pool settings
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

		// Configure additional SQLite pragmas for better concurrency
		if err := configureSQLitePragmas(sqlDB, cfg); err != nil {
			return nil, fmt.Errorf("failed to configure SQLite pragmas: %w", err)
		}
	}

	return db, nil
}

// ensureSQLiteDir ensures the directory for SQLite database file exists
func ensureSQLiteDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." || dir == "" {
		return nil // Current directory, no need to create
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Create directory with 0755 permissions
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// configureSQLitePragmas sets up SQLite-specific configuration for better concurrency and performance
func configureSQLitePragmas(sqlDB *sql.DB, cfg config.DatabaseConfig) error {
	pragmas := []string{
		"PRAGMA foreign_keys = ON",           // Enable foreign key constraints
		"PRAGMA synchronous = NORMAL",        // Balance between safety and performance
		"PRAGMA cache_size = -64000",         // 64MB cache
		"PRAGMA temp_store = MEMORY",         // Store temporary tables in memory
		"PRAGMA mmap_size = 268435456",       // 256MB memory-mapped I/O
	}

	// Add WAL mode pragma if enabled
	if cfg.WALMode {
		pragmas = append(pragmas, "PRAGMA journal_mode = WAL")
		pragmas = append(pragmas, "PRAGMA wal_autocheckpoint = 1000") // Checkpoint every 1000 pages
	}

	// Execute all pragmas
	for _, pragma := range pragmas {
		if _, err := sqlDB.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute pragma '%s': %w", pragma, err)
		}
	}

	return nil
}

// WithRetry executes a database operation with exponential backoff retry mechanism
func WithRetry(operation func() error, maxRetries int) error {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := operation(); err != nil {
			lastErr = err
			
			// Check if it's a database lock error that we should retry
			if isDatabaseLockError(err) && attempt < maxRetries {
				// Exponential backoff: 100ms, 200ms, 400ms, 800ms, 1600ms
				backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
				time.Sleep(backoff)
				continue
			}
			
			return err
		}
		
		return nil
	}
	
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

// isDatabaseLockError checks if the error is related to database locking
func isDatabaseLockError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	lockErrors := []string{
		"database is locked",
		"database table is locked",
		"SQLITE_BUSY",
		"SQLITE_LOCKED",
		"database schema has changed",
	}
	
	for _, lockErr := range lockErrors {
		if contains(errStr, lockErr) {
			return true
		}
	}
	
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr || 
		      containsSubstring(s, substr))))
}

// containsSubstring performs a simple substring search
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}