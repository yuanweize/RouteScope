package storage

import (
	"fmt"
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
