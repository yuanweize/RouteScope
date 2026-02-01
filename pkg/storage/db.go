package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	conn *gorm.DB
}

// NewDB initializes the SQLite database
func NewDB(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db dir: %w", err)
	}

	// GORM Config
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // Quiet logs
	}

	// Open DB
	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	// Enable WAL Mode for better concurrency
	// "PRAGMA journal_mode=WAL;"
	if res := db.Exec("PRAGMA journal_mode=WAL;"); res.Error != nil {
		log.Printf("Warning: Failed to enable WAL mode: %v", res.Error)
	}

	// Auto Migrate
	if err := db.AutoMigrate(&MonitorRecord{}, &Target{}, &User{}); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return &DB{conn: db}, nil
}

// Close closes the underlying db connection (optional, GORM manages pool)
func (d *DB) Close() error {
	sqlDB, err := d.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
