package router

import (
	"os"
	"testing"

	"github.com/X0Ken/openai-gateway/pkg/database"
)

func TestRoute(t *testing.T) {
	dbPath := "/tmp/test_router.db"
	defer os.Remove(dbPath)

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create test user
	user := &database.User{APIKey: "test-key", Name: "Test"}
	db.CreateUser(user)

	// Create test channel
	channel := &database.Channel{
		Name:    "test-chan",
		BaseURL: "https://api.openai.com",
		APIKey:  "sk-test",
		Weight:  10,
		Enabled: true,
	}
	db.CreateChannel(channel)

	// Create test model
	model := &database.Model{Name: "gpt-3.5-turbo"}
	db.CreateModel(model)

	// Create model-channel mapping
	mc := &database.ModelChannel{
		ModelID:          model.ID,
		ChannelID:        channel.ID,
		BackendModelName: "gpt-3.5-turbo",
		Weight:           10,
	}
	db.AddModelChannel(mc)

	engine := NewEngine(db)

	// Test routing
	result, err := engine.Route(user.ID, "gpt-3.5-turbo")
	if err != nil {
		t.Fatalf("Failed to route: %v", err)
	}

	if result.Channel == nil {
		t.Fatal("Expected channel, got nil")
	}

	if result.Channel.ID != channel.ID {
		t.Errorf("Expected channel ID %d, got %d", channel.ID, result.Channel.ID)
	}

	if !result.IsNew {
		t.Error("Expected new session")
	}

	if result.BackendModelName != "gpt-3.5-turbo" {
		t.Errorf("Expected backend model name 'gpt-3.5-turbo', got '%s'", result.BackendModelName)
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if !contains(slice, "a") {
		t.Error("Expected to contain 'a'")
	}

	if !contains(slice, "b") {
		t.Error("Expected to contain 'b'")
	}

	if contains(slice, "d") {
		t.Error("Expected not to contain 'd'")
	}
}
