package database

import (
	"fmt"
	"time"
)

// ModelChannel represents a mapping between a model and a channel
type ModelChannel struct {
	ID               int64     `json:"id"`
	ModelID          int64     `json:"model_id"`
	ChannelID        int64     `json:"channel_id"`
	BackendModelName string    `json:"backend_model_name"`
	Weight           int       `json:"weight"`
	CreatedAt        time.Time `json:"created_at"`
}

// AddModelChannel creates a new model-channel mapping
func (db *DB) AddModelChannel(mc *ModelChannel) error {
	if mc.Weight <= 0 {
		mc.Weight = 10
	}

	result, err := db.Exec(
		"INSERT INTO model_channels (model_id, channel_id, backend_model_name, weight) VALUES (?, ?, ?, ?)",
		mc.ModelID, mc.ChannelID, mc.BackendModelName, mc.Weight,
	)
	if err != nil {
		return fmt.Errorf("failed to add model channel: %w", err)
	}

	mc.ID, _ = result.LastInsertId()
	return nil
}

// ListModelChannels retrieves all model-channel mappings
func (db *DB) ListModelChannels() ([]*ModelChannel, error) {
	rows, err := db.Query("SELECT id, model_id, channel_id, backend_model_name, weight, created_at FROM model_channels")
	if err != nil {
		return nil, fmt.Errorf("failed to list model channels: %w", err)
	}
	defer rows.Close()

	var mappings []*ModelChannel
	for rows.Next() {
		var mc ModelChannel
		if err := rows.Scan(&mc.ID, &mc.ModelID, &mc.ChannelID, &mc.BackendModelName, &mc.Weight, &mc.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan model channel: %w", err)
		}
		mappings = append(mappings, &mc)
	}

	return mappings, nil
}

// GetModelChannelsByModel retrieves all channel mappings for a specific model
func (db *DB) GetModelChannelsByModel(modelID int64) ([]*ModelChannel, error) {
	rows, err := db.Query(
		"SELECT id, model_id, channel_id, backend_model_name, weight, created_at FROM model_channels WHERE model_id = ?",
		modelID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get model channels by model: %w", err)
	}
	defer rows.Close()

	var mappings []*ModelChannel
	for rows.Next() {
		var mc ModelChannel
		if err := rows.Scan(&mc.ID, &mc.ModelID, &mc.ChannelID, &mc.BackendModelName, &mc.Weight, &mc.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan model channel: %w", err)
		}
		mappings = append(mappings, &mc)
	}

	return mappings, nil
}

// GetModelChannelsByChannel retrieves all model mappings for a specific channel
func (db *DB) GetModelChannelsByChannel(channelID int64) ([]*ModelChannel, error) {
	rows, err := db.Query(
		"SELECT id, model_id, channel_id, backend_model_name, weight, created_at FROM model_channels WHERE channel_id = ?",
		channelID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get model channels by channel: %w", err)
	}
	defer rows.Close()

	var mappings []*ModelChannel
	for rows.Next() {
		var mc ModelChannel
		if err := rows.Scan(&mc.ID, &mc.ModelID, &mc.ChannelID, &mc.BackendModelName, &mc.Weight, &mc.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan model channel: %w", err)
		}
		mappings = append(mappings, &mc)
	}

	return mappings, nil
}

// RemoveModelChannel deletes a specific model-channel mapping
func (db *DB) RemoveModelChannel(modelID, channelID int64) error {
	_, err := db.Exec(
		"DELETE FROM model_channels WHERE model_id = ? AND channel_id = ?",
		modelID, channelID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove model channel: %w", err)
	}
	return nil
}

// RemoveAllModelChannelsForModel deletes all mappings for a specific model
func (db *DB) RemoveAllModelChannelsForModel(modelID int64) error {
	_, err := db.Exec("DELETE FROM model_channels WHERE model_id = ?", modelID)
	if err != nil {
		return fmt.Errorf("failed to remove model channels for model: %w", err)
	}
	return nil
}

// RemoveAllModelChannelsForChannel deletes all mappings for a specific channel
func (db *DB) RemoveAllModelChannelsForChannel(channelID int64) error {
	_, err := db.Exec("DELETE FROM model_channels WHERE channel_id = ?", channelID)
	if err != nil {
		return fmt.Errorf("failed to remove model channels for channel: %w", err)
	}
	return nil
}
