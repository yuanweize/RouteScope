package storage

import (
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

// --- Target Management ---

func (d *DB) SaveTarget(t *Target) error {
	return d.conn.Save(t).Error
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

// --- User Management (Phase 13) ---

func (d *DB) GetUser(username string) (*User, error) {
	var u User
	err := d.conn.Where("username = ?", username).First(&u).Error
	return &u, err
}

func (d *DB) SaveUser(u *User) error {
	return d.conn.Save(u).Error
}

func (d *DB) HasAnyUser() bool {
	var count int64
	d.conn.Model(&User{}).Count(&count)
	return count > 0
}
