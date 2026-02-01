package storage

import (
	"time"
)

// Target represents a monitoring destination
type Target struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time
	Name      string `gorm:"type:varchar(64);not null"`
	Address   string `gorm:"type:varchar(128);uniqueIndex;not null"` // IP or Domain
	Desc      string `gorm:"type:text"`
	Enabled   bool   `gorm:"default:true"`

	// --- Probing Configuration (Phase 13) ---
	// ProbeMode: ICMP, SSH, HTTP, IPERF3
	ProbeMode string `gorm:"type:varchar(20);default:'ICMP'"`

	// ProbeConfig (JSON stored as text for flexibility)
	// Includes URL for HTTP, Port for Iperf, Credentials for SSH
	ProbeConfig string `gorm:"type:text"`
}

// User represents a system administrator
type User struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"not null"`
	Username  string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	Password  string    `gorm:"type:varchar(128);not null"` // Hashed
}

// MonitorRecord represents a single monitoring data point (snapshot)
type MonitorRecord struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"index;not null"` // Time-series index
	Target    string    `gorm:"index;type:varchar(128);not null"`

	// Ping Metrics (Always present)
	LatencyMs  float64 `gorm:"not null"` // Average RTT in milliseconds
	PacketLoss float64 `gorm:"not null"` // Loss Percentage (0.0 - 100.0)

	// Traceroute Data (JSON Blob)
	TraceJson []byte `gorm:"type:text"`

	// Speed Test Metrics
	SpeedUp   float64 `gorm:"default:0"` // Mbps
	SpeedDown float64 `gorm:"default:0"` // Mbps
}
