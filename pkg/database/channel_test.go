package database

import (
	"os"
	"testing"
)

func TestChannelCRUD(t *testing.T) {
	dbPath := "/tmp/test_channel.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create
	channel := &Channel{
		Name:    "test-channel",
		BaseURL: "https://api.openai.com",
		APIKey:  "sk-test",
		Weight:  10,
		Enabled: true,
	}

	if err := db.CreateChannel(channel); err != nil {
		t.Fatalf("Failed to create channel: %v", err)
	}

	if channel.ID == 0 {
		t.Error("Channel ID should be set after creation")
	}

	// Get
	retrieved, err := db.GetChannel(channel.ID)
	if err != nil {
		t.Fatalf("Failed to get channel: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Retrieved channel should not be nil")
	}
	if retrieved.Name != channel.Name {
		t.Errorf("Expected name %s, got %s", channel.Name, retrieved.Name)
	}

	// Update
	channel.Weight = 20
	if err := db.UpdateChannel(channel); err != nil {
		t.Fatalf("Failed to update channel: %v", err)
	}

	retrieved, _ = db.GetChannel(channel.ID)
	if retrieved.Weight != 20 {
		t.Errorf("Expected weight 20, got %d", retrieved.Weight)
	}

	// List
	channels, err := db.ListChannels()
	if err != nil {
		t.Fatalf("Failed to list channels: %v", err)
	}
	if len(channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(channels))
	}

	// Delete
	if err := db.DeleteChannel(channel.ID); err != nil {
		t.Fatalf("Failed to delete channel: %v", err)
	}

	retrieved, _ = db.GetChannel(channel.ID)
	if retrieved != nil {
		t.Error("Channel should be deleted")
	}
}
