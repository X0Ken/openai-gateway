package database

import (
	"database/sql"
	"fmt"
	"time"
)

// User represents an API key holder
type User struct {
	ID        int64     `json:"id"`
	APIKey    string    `json:"api_key"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUser creates a new user
func (db *DB) CreateUser(user *User) error {
	result, err := db.Exec(
		"INSERT INTO users (api_key, name) VALUES (?, ?)",
		user.APIKey, user.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.ID, _ = result.LastInsertId()
	return nil
}

// GetUser retrieves a user by ID
func (db *DB) GetUser(id int64) (*User, error) {
	var user User

	err := db.QueryRow(
		"SELECT id, api_key, name, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.APIKey, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByAPIKey retrieves a user by API key
func (db *DB) GetUserByAPIKey(apiKey string) (*User, error) {
	var user User

	err := db.QueryRow(
		"SELECT id, api_key, name, created_at, updated_at FROM users WHERE api_key = ?",
		apiKey,
	).Scan(&user.ID, &user.APIKey, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by API key: %w", err)
	}

	return &user, nil
}

// ListUsers retrieves all users
func (db *DB) ListUsers() ([]*User, error) {
	rows, err := db.Query("SELECT id, api_key, name, created_at, updated_at FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User

		if err := rows.Scan(&user.ID, &user.APIKey, &user.Name, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, &user)
	}

	return users, nil
}

// UpdateUser updates a user
func (db *DB) UpdateUser(user *User) error {
	_, err := db.Exec(
		"UPDATE users SET api_key = ?, name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		user.APIKey, user.Name, user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user by ID
func (db *DB) DeleteUser(id int64) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
