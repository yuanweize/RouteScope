package storage

import (
	"time"
)

// Target represents a monitoring destination
type Target struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `gorm:"type:varchar(64);not null" json:"name"`
	Address   string    `gorm:"type:varchar(128);uniqueIndex;not null" json:"address"` // IP or Domain
	Desc      string    `gorm:"type:text" json:"desc"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`

	// --- Probing Configuration (Phase 13) ---
	// ProbeMode: ICMP, SSH, HTTP, IPERF3
	ProbeType string `gorm:"column:probe_type;type:varchar(20);default:'MODE_ICMP'" json:"probe_type"`

	// ProbeConfig (JSON stored as text for flexibility)
	// Includes URL for HTTP, Port for Iperf, Credentials for SSH
	ProbeConfig string `gorm:"column:probe_config;type:text" json:"probe_config"`
}

// User represents a system administrator
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	Username  string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"type:varchar(128);not null" json:"-"` // Hashed
}

// MonitorRecord represents a single monitoring data point (snapshot)
type MonitorRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"index;not null" json:"created_at"` // Time-series index
	Target    string    `gorm:"index;type:varchar(128);not null" json:"target"`

	// Ping Metrics (Always present)
	LatencyMs  float64 `gorm:"not null" json:"latency_ms"`  // Average RTT in milliseconds
	PacketLoss float64 `gorm:"not null" json:"packet_loss"` // Loss Percentage (0.0 - 100.0)

	// Traceroute Data (JSON Blob)
	TraceJson []byte `gorm:"type:text" json:"trace_json,omitempty"`

	// Speed Test Metrics
	SpeedUp   float64 `gorm:"default:0" json:"speed_up"`   // Mbps
	SpeedDown float64 `gorm:"default:0" json:"speed_down"` // Mbps
}

const (
	ProbeModeICMP  = "MODE_ICMP"
	ProbeModeHTTP  = "MODE_HTTP"
	ProbeModeSSH   = "MODE_SSH"
	ProbeModeIPERF = "MODE_IPERF"
)
