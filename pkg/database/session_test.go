package database

import (
	"os"
	"testing"
)

func TestSessionCRUD(t *testing.T) {
	dbPath := "/tmp/test_session.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create user and channel first
	user := &User{APIKey: "test-key", Name: "Test"}
	db.CreateUser(user)

	channel := &Channel{Name: "test-chan", BaseURL: "https://api.openai.com", APIKey: "sk-test"}
	db.CreateChannel(channel)

	// Create session
	session := &Session{
		UserID:    user.ID,
		ChannelID: channel.ID,
	}

	if err := db.CreateSession(session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ID == 0 {
		t.Error("Session ID should be set after creation")
	}

	// Get by user
	retrieved, err := db.GetSessionByUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to get session by user: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Retrieved session should not be nil")
	}
	if retrieved.ChannelID != channel.ID {
		t.Errorf("Expected channel ID %d, got %d", channel.ID, retrieved.ChannelID)
	}
}
