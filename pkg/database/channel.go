package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Channel represents a backend channel configuration
type Channel struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	BaseURL   string    `json:"base_url"`
	APIKey    string    `json:"api_key"`
	Weight    int       `json:"weight"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateChannel creates a new channel
func (db *DB) CreateChannel(channel *Channel) error {
	result, err := db.Exec(
		"INSERT INTO channels (name, base_url, api_key, weight, enabled) VALUES (?, ?, ?, ?, ?)",
		channel.Name, channel.BaseURL, channel.APIKey, channel.Weight, channel.Enabled,
	)
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	channel.ID, _ = result.LastInsertId()
	return nil
}

// GetChannel retrieves a channel by ID
func (db *DB) GetChannel(id int64) (*Channel, error) {
	var channel Channel

	err := db.QueryRow(
		"SELECT id, name, base_url, api_key, weight, enabled, created_at, updated_at FROM channels WHERE id = ?",
		id,
	).Scan(&channel.ID, &channel.Name, &channel.BaseURL, &channel.APIKey, &channel.Weight, &channel.Enabled, &channel.CreatedAt, &channel.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	return &channel, nil
}

// GetChannelByName retrieves a channel by name
func (db *DB) GetChannelByName(name string) (*Channel, error) {
	var channel Channel

	err := db.QueryRow(
		"SELECT id, name, base_url, api_key, weight, enabled, created_at, updated_at FROM channels WHERE name = ?",
		name,
	).Scan(&channel.ID, &channel.Name, &channel.BaseURL, &channel.APIKey, &channel.Weight, &channel.Enabled, &channel.CreatedAt, &channel.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get channel by name: %w", err)
	}

	return &channel, nil
}

// ListChannels retrieves all channels
func (db *DB) ListChannels() ([]*Channel, error) {
	rows, err := db.Query("SELECT id, name, base_url, api_key, weight, enabled, created_at, updated_at FROM channels")
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		var channel Channel

		if err := rows.Scan(&channel.ID, &channel.Name, &channel.BaseURL, &channel.APIKey, &channel.Weight, &channel.Enabled, &channel.CreatedAt, &channel.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}

		channels = append(channels, &channel)
	}

	return channels, nil
}

// ListEnabledChannels retrieves all enabled channels
func (db *DB) ListEnabledChannels() ([]*Channel, error) {
	rows, err := db.Query("SELECT id, name, base_url, api_key, weight, enabled, created_at, updated_at FROM channels WHERE enabled = 1")
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled channels: %w", err)
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		var channel Channel

		if err := rows.Scan(&channel.ID, &channel.Name, &channel.BaseURL, &channel.APIKey, &channel.Weight, &channel.Enabled, &channel.CreatedAt, &channel.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}

		channels = append(channels, &channel)
	}

	return channels, nil
}

// UpdateChannel updates a channel
func (db *DB) UpdateChannel(channel *Channel) error {
	_, err := db.Exec(
		"UPDATE channels SET name = ?, base_url = ?, api_key = ?, weight = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		channel.Name, channel.BaseURL, channel.APIKey, channel.Weight, channel.Enabled, channel.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return nil
}

// DeleteChannel deletes a channel by ID
func (db *DB) DeleteChannel(id int64) error {
	_, err := db.Exec("DELETE FROM channels WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}
	return nil
}
