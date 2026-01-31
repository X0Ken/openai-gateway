package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ChannelMetrics represents performance metrics for a channel
type ChannelMetrics struct {
	ChannelID     int64     `json:"channel_id"`
	LatencyAvg    float64   `json:"latency_avg"`
	ErrorRate     float64   `json:"error_rate"`
	RequestCount  int64     `json:"request_count"`
	SuccessCount  int64     `json:"success_count"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

// GetChannelMetrics retrieves metrics for a channel
func (db *DB) GetChannelMetrics(channelID int64) (*ChannelMetrics, error) {
	var metrics ChannelMetrics

	err := db.QueryRow(
		"SELECT channel_id, latency_avg, error_rate, request_count, success_count, last_updated_at FROM channel_metrics WHERE channel_id = ?",
		channelID,
	).Scan(&metrics.ChannelID, &metrics.LatencyAvg, &metrics.ErrorRate, &metrics.RequestCount, &metrics.SuccessCount, &metrics.LastUpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get channel metrics: %w", err)
	}

	return &metrics, nil
}

// UpdateChannelMetrics updates metrics for a channel
func (db *DB) UpdateChannelMetrics(channelID int64, latency float64, success bool) error {
	// Use INSERT OR REPLACE to handle both insert and update
	_, err := db.Exec(`
		INSERT INTO channel_metrics (channel_id, latency_avg, error_rate, request_count, success_count, last_updated_at)
		VALUES (?, ?, ?, 1, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(channel_id) DO UPDATE SET
			latency_avg = (channel_metrics.latency_avg * channel_metrics.request_count + ?) / (channel_metrics.request_count + 1),
			error_rate = (channel_metrics.error_rate * channel_metrics.request_count + ?) / (channel_metrics.request_count + 1),
			request_count = channel_metrics.request_count + 1,
			success_count = channel_metrics.success_count + ?,
			last_updated_at = CURRENT_TIMESTAMP
	`, channelID, latency, boolToInt(!success), boolToInt(success), latency, boolToInt(!success), boolToInt(success))

	if err != nil {
		return fmt.Errorf("failed to update channel metrics: %w", err)
	}

	return nil
}

// ResetChannelMetrics resets metrics for a channel
func (db *DB) ResetChannelMetrics(channelID int64) error {
	_, err := db.Exec(
		"DELETE FROM channel_metrics WHERE channel_id = ?",
		channelID,
	)
	if err != nil {
		return fmt.Errorf("failed to reset channel metrics: %w", err)
	}
	return nil
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
