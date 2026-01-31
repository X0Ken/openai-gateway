package database

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	// Create temporary database
	dbPath := "/tmp/test_gateway.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify tables exist
	tables := []string{"channels", "users", "sessions", "channel_metrics"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("Table %s does not exist: %v", table, err)
		}
		if name != table {
			t.Errorf("Expected table name %s, got %s", table, name)
		}
	}
}

func TestMigrations(t *testing.T) {
	dbPath := "/tmp/test_migrations.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test that we can insert into channels
	_, err = db.Exec("INSERT INTO channels (name, base_url, api_key, models) VALUES (?, ?, ?, ?)",
		"test-channel", "https://api.openai.com", "sk-test", `["gpt-3.5-turbo"]`)
	if err != nil {
		t.Errorf("Failed to insert into channels: %v", err)
	}

	// Test that we can insert into users
	_, err = db.Exec("INSERT INTO users (api_key, name) VALUES (?, ?)",
		"test-key", "Test User")
	if err != nil {
		t.Errorf("Failed to insert into users: %v", err)
	}
}
