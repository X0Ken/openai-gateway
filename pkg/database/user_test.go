package database

import (
	"os"
	"testing"
)

func TestUserCRUD(t *testing.T) {
	dbPath := "/tmp/test_user.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create
	user := &User{
		APIKey: "test-api-key",
		Name:   "Test User",
	}

	if err := db.CreateUser(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("User ID should be set after creation")
	}

	// Get by API key
	retrieved, err := db.GetUserByAPIKey(user.APIKey)
	if err != nil {
		t.Fatalf("Failed to get user by API key: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Retrieved user should not be nil")
	}
	if retrieved.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, retrieved.Name)
	}

	// List
	users, err := db.ListUsers()
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
}
