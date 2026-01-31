package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Model represents a logical model name that users can request
type Model struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	ChannelsCount int64     `json:"channels_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateModel creates a new model
func (db *DB) CreateModel(model *Model) error {
	result, err := db.Exec(
		"INSERT INTO models (name) VALUES (?)",
		model.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	model.ID, _ = result.LastInsertId()
	return nil
}

// GetModel retrieves a model by ID
func (db *DB) GetModel(id int64) (*Model, error) {
	var model Model

	err := db.QueryRow(
		"SELECT id, name, created_at, updated_at FROM models WHERE id = ?",
		id,
	).Scan(&model.ID, &model.Name, &model.CreatedAt, &model.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	return &model, nil
}

// GetModelByName retrieves a model by name
func (db *DB) GetModelByName(name string) (*Model, error) {
	var model Model

	err := db.QueryRow(
		"SELECT id, name, created_at, updated_at FROM models WHERE name = ?",
		name,
	).Scan(&model.ID, &model.Name, &model.CreatedAt, &model.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get model by name: %w", err)
	}

	return &model, nil
}

// ListModels retrieves all models
func (db *DB) ListModels() ([]*Model, error) {
	rows, err := db.Query(`
		SELECT m.id, m.name, m.created_at, m.updated_at, COUNT(mc.id) as channels_count
		FROM models m
		LEFT JOIN model_channels mc ON m.id = mc.model_id
		GROUP BY m.id
		ORDER BY m.name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer rows.Close()

	var models []*Model
	for rows.Next() {
		var model Model
		if err := rows.Scan(&model.ID, &model.Name, &model.CreatedAt, &model.UpdatedAt, &model.ChannelsCount); err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		models = append(models, &model)
	}

	return models, nil
}

// UpdateModel updates a model's name
func (db *DB) UpdateModel(model *Model) error {
	_, err := db.Exec(
		"UPDATE models SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		model.Name, model.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}

	return nil
}

// DeleteModel deletes a model by ID
func (db *DB) DeleteModel(id int64) error {
	_, err := db.Exec("DELETE FROM models WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}
	return nil
}
