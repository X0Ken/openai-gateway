package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/X0Ken/openai-gateway/internal/admin"
	"github.com/X0Ken/openai-gateway/internal/api"
	"github.com/X0Ken/openai-gateway/internal/auth"
	"github.com/X0Ken/openai-gateway/internal/channel"
	"github.com/X0Ken/openai-gateway/internal/router"
	"github.com/X0Ken/openai-gateway/internal/session"
	"github.com/X0Ken/openai-gateway/pkg/database"
	"github.com/gin-gonic/gin"
)

func setupTestServer(t *testing.T) (*gin.Engine, *database.DB, func()) {
	// Create unique test database for each test
	dbPath := fmt.Sprintf("/tmp/test_integration_%d.db", os.Getpid())
	os.Remove(dbPath) // Clean up if exists

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Setup Gin
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create test user with unique key
	user := &database.User{APIKey: fmt.Sprintf("test-api-key-%d", os.Getpid()), Name: "Test User"}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create test channel
	ch := &database.Channel{
		Name:    fmt.Sprintf("test-channel-%d", os.Getpid()),
		BaseURL: "https://httpbin.org",
		APIKey:  "test-key",
		Weight:  10,
		Enabled: true,
	}
	if err := db.CreateChannel(ch); err != nil {
		t.Fatalf("Failed to create channel: %v", err)
	}

	// Create test model
	model := &database.Model{Name: "gpt-3.5-turbo"}
	if err := db.CreateModel(model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Create model-channel mapping
	mc := &database.ModelChannel{
		ModelID:          model.ID,
		ChannelID:        ch.ID,
		BackendModelName: "gpt-3.5-turbo",
		Weight:           10,
	}
	if err := db.AddModelChannel(mc); err != nil {
		t.Fatalf("Failed to add model-channel mapping: %v", err)
	}

	// Initialize components
	channelMgr := channel.NewManager(db)
	sessionMgr := session.NewManager(db, 30)
	routerEngine := router.NewEngine(db)
	authMiddleware := auth.NewMiddleware(db)

	// Setup routes
	apiHandler := api.NewHandler(routerEngine, channelMgr, db)
	openaiGroup := r.Group("/v1")
	apiHandler.RegisterRoutes(openaiGroup, authMiddleware)

	adminHandler := admin.NewHandler(channelMgr, sessionMgr, db)
	adminGroup := r.Group("/api")
	adminHandler.RegisterRoutes(adminGroup)

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return r, db, cleanup
}

func TestListModels(t *testing.T) {
	r, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/v1/models", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response api.ListModelsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Object != "list" {
		t.Errorf("Expected object 'list', got '%s'", response.Object)
	}

	if len(response.Data) == 0 {
		t.Error("Expected at least one model")
	}
}

func TestCreateChannel(t *testing.T) {
	r, _, cleanup := setupTestServer(t)
	defer cleanup()

	reqBody := channel.CreateRequest{
		Name:    "new-channel",
		BaseURL: "https://api.example.com",
		APIKey:  "sk-new-key",
		Weight:  20,
		Enabled: true,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/channels", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var response database.Channel
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Name != reqBody.Name {
		t.Errorf("Expected name '%s', got '%s'", reqBody.Name, response.Name)
	}
}

func TestListChannels(t *testing.T) {
	r, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/channels", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []*database.Channel
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(response) == 0 {
		t.Error("Expected at least one channel")
	}
}

func TestAuthentication(t *testing.T) {
	r, _, cleanup := setupTestServer(t)
	defer cleanup()

	// Test without auth
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader([]byte(`{}`)))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 without auth, got %d", w.Code)
	}

	// Test with invalid auth
	req = httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer invalid-key")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with invalid auth, got %d", w.Code)
	}
}
