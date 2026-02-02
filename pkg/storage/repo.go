package storage

import (
	"fmt"
	"os"
	"time"
)

// SaveRecord persists a monitoring record
func (d *DB) SaveRecord(r *MonitorRecord) error {
	return d.conn.Create(r).Error
}

// GetHistory fetches records for a specific target within a time range.
// Optimization: We exclude TraceJson to reduce I/O for general charts.
func (d *DB) GetHistory(target string, start, end time.Time) ([]MonitorRecord, error) {
	var records []MonitorRecord

	err := d.conn.Model(&MonitorRecord{}).
		Select("id, created_at, target, latency_ms, packet_loss, speed_up, speed_down"). // Exclude TraceJson
		Where("target = ? AND created_at BETWEEN ? AND ?", target, start, end).
		Order("created_at asc").
		Find(&records).Error

	return records, err
}

// GetRecordDetail fetches the full record including TraceJson by ID
func (d *DB) GetRecordDetail(id uint) (*MonitorRecord, error) {
	var r MonitorRecord
	err := d.conn.First(&r, id).Error
	return &r, err
}

// GetLatestRecord fetches the most recent record for a target
func (d *DB) GetLatestRecord(target string) (*MonitorRecord, error) {
	var r MonitorRecord
	err := d.conn.
		Where("target = ?", target).
		Order("created_at desc").
		Limit(1).
		First(&r).Error
	return &r, err
}

// GetLatestTrace fetches the most recent record that includes traceroute data
func (d *DB) GetLatestTrace(target string) (*MonitorRecord, error) {
	var r MonitorRecord
	err := d.conn.
		Where("target = ? AND trace_json IS NOT NULL AND trace_json != ''", target).
		Order("created_at desc").
		Limit(1).
		First(&r).Error
	return &r, err
}

// --- Target Management ---

// CreateTarget inserts a new target. Returns error if address already exists.
func (d *DB) CreateTarget(t *Target) error {
	return d.conn.Create(t).Error
}

// UpdateTarget updates an existing target by ID.
// Only updates non-zero fields to avoid overwriting with defaults.
func (d *DB) UpdateTarget(t *Target) error {
	if t.ID == 0 {
		return fmt.Errorf("cannot update target without ID")
	}
	return d.conn.Model(t).Updates(t).Error
}

// SaveTarget creates or updates a target based on whether ID is set.
// WARNING: For updates, prefer UpdateTarget to avoid GORM Save() pitfalls.
func (d *DB) SaveTarget(t *Target) error {
	if t.ID == 0 {
		return d.conn.Create(t).Error
	}
	return d.conn.Save(t).Error
}

// GetTargetByID retrieves a target by its ID
func (d *DB) GetTargetByID(id uint) (*Target, error) {
	var t Target
	err := d.conn.First(&t, id).Error
	return &t, err
}

func (d *DB) DeleteTarget(id uint) error {
	return d.conn.Delete(&Target{}, id).Error
}

func (d *DB) GetTargets(onlyEnabled bool) ([]Target, error) {
	var targets []Target
	query := d.conn.Model(&Target{})
	if onlyEnabled {
		query = query.Where("enabled = ?", true)
	}
	err := query.Find(&targets).Error
	return targets, err
}

// UpdateTargetError updates the last_error and last_error_at fields for a target
func (d *DB) UpdateTargetError(address string, errMsg string) error {
	now := time.Now()
	return d.conn.Model(&Target{}).
		Where("address = ?", address).
		Updates(map[string]interface{}{
			"last_error":    errMsg,
			"last_error_at": now,
		}).Error
}

// ClearTargetError clears the error fields for a target (on successful probe)
func (d *DB) ClearTargetError(address string) error {
	return d.conn.Model(&Target{}).
		Where("address = ?", address).
		Updates(map[string]interface{}{
			"last_error":    "",
			"last_error_at": nil,
		}).Error
}

// --- User Management (Phase 13) ---

func (d *DB) GetUser(username string) (*User, error) {
	var u User
	err := d.conn.Where("username = ?", username).First(&u).Error
	return &u, err
}

// GetFirstUser returns the first user in the database (single-user system)
func (d *DB) GetFirstUser() (*User, error) {
	var u User
	err := d.conn.First(&u).Error
	return &u, err
}

// UpdateUserPassword updates the password for a specific user ID
// This avoids GORM's Save() which can INSERT if ID is missing
func (d *DB) UpdateUserPassword(userID uint, hashedPassword string) error {
	return d.conn.Model(&User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

func (d *DB) SaveUser(u *User) error {
	return d.conn.Save(u).Error
}

func (d *DB) HasAnyUser() bool {
	var count int64
	d.conn.Model(&User{}).Count(&count)
	return count > 0
}

// --- Database Management ---

// DatabaseStats returns database statistics
type DatabaseStats struct {
	SizeBytes      int64  `json:"size_bytes"`
	SizeHuman      string `json:"size_human"`
	RecordCount    int64  `json:"record_count"`
	TargetCount    int64  `json:"target_count"`
	OldestRecord   string `json:"oldest_record,omitempty"`
	NewestRecord   string `json:"newest_record,omitempty"`
	RetentionDays  int    `json:"retention_days"`
}

// GetDatabaseStats returns statistics about the database
func (d *DB) GetDatabaseStats(dbPath string, retentionDays int) (*DatabaseStats, error) {
	stats := &DatabaseStats{RetentionDays: retentionDays}

	// File size
	if fi, err := os.Stat(dbPath); err == nil {
		stats.SizeBytes = fi.Size()
		stats.SizeHuman = formatBytes(fi.Size())
	}

	// Record count
	d.conn.Model(&MonitorRecord{}).Count(&stats.RecordCount)

	// Target count
	d.conn.Model(&Target{}).Count(&stats.TargetCount)

	// Oldest and newest records
	var oldest, newest MonitorRecord
	if d.conn.Model(&MonitorRecord{}).Order("created_at ASC").First(&oldest).Error == nil {
		stats.OldestRecord = oldest.CreatedAt.Format(time.RFC3339)
	}
	if d.conn.Model(&MonitorRecord{}).Order("created_at DESC").First(&newest).Error == nil {
		stats.NewestRecord = newest.CreatedAt.Format(time.RFC3339)
	}

	return stats, nil
}

// CleanOldRecords deletes records older than the specified number of days
func (d *DB) CleanOldRecords(days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	result := d.conn.Where("created_at < ?", cutoff).Delete(&MonitorRecord{})
	return result.RowsAffected, result.Error
}

// VacuumDatabase runs VACUUM to reclaim space
func (d *DB) VacuumDatabase() error {
	return d.conn.Exec("VACUUM").Error
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
