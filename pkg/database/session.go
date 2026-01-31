package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Session represents a user-channel mapping for sticky routing
type Session struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	ChannelID  int64     `json:"channel_id"`
	LastUsedAt time.Time `json:"last_used_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateSession creates a new session
func (db *DB) CreateSession(session *Session) error {
	result, err := db.Exec(
		"INSERT INTO sessions (user_id, channel_id) VALUES (?, ?)",
		session.UserID, session.ChannelID,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	session.ID, _ = result.LastInsertId()
	return nil
}

// GetSession retrieves a session by ID
func (db *DB) GetSession(id int64) (*Session, error) {
	var session Session

	err := db.QueryRow(
		"SELECT id, user_id, channel_id, last_used_at, created_at FROM sessions WHERE id = ?",
		id,
	).Scan(&session.ID, &session.UserID, &session.ChannelID, &session.LastUsedAt, &session.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// GetSessionByUserAndChannel retrieves a session by user and channel
func (db *DB) GetSessionByUserAndChannel(userID, channelID int64) (*Session, error) {
	var session Session

	err := db.QueryRow(
		"SELECT id, user_id, channel_id, last_used_at, created_at FROM sessions WHERE user_id = ? AND channel_id = ?",
		userID, channelID,
	).Scan(&session.ID, &session.UserID, &session.ChannelID, &session.LastUsedAt, &session.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session by user and channel: %w", err)
	}

	return &session, nil
}

// GetSessionByUser retrieves the most recent session for a user
func (db *DB) GetSessionByUser(userID int64) (*Session, error) {
	var session Session

	err := db.QueryRow(
		"SELECT id, user_id, channel_id, last_used_at, created_at FROM sessions WHERE user_id = ? ORDER BY last_used_at DESC LIMIT 1",
		userID,
	).Scan(&session.ID, &session.UserID, &session.ChannelID, &session.LastUsedAt, &session.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session by user: %w", err)
	}

	return &session, nil
}

// ListSessions retrieves all sessions
func (db *DB) ListSessions() ([]*Session, error) {
	rows, err := db.Query("SELECT id, user_id, channel_id, last_used_at, created_at FROM sessions")
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session

		if err := rows.Scan(&session.ID, &session.UserID, &session.ChannelID, &session.LastUsedAt, &session.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// UpdateSessionLastUsed updates the last_used_at timestamp
func (db *DB) UpdateSessionLastUsed(id int64) error {
	_, err := db.Exec(
		"UPDATE sessions SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update session last used: %w", err)
	}

	return nil
}

// DeleteSession deletes a session by ID
func (db *DB) DeleteSession(id int64) error {
	_, err := db.Exec("DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions deletes sessions older than the given duration
func (db *DB) DeleteExpiredSessions(idleTimeoutMinutes int) error {
	_, err := db.Exec(
		"DELETE FROM sessions WHERE last_used_at < datetime('now', ?)",
		fmt.Sprintf("-%d minutes", idleTimeoutMinutes),
	)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}
