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
