package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/X0Ken/openai-gateway/internal/router"
	"github.com/X0Ken/openai-gateway/pkg/database"
	"github.com/gin-gonic/gin"
)

func setupTestHandler(t *testing.T) (*Handler, *database.DB, func()) {
	dbPath := "/tmp/test_chat.db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

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

	routerEngine := router.NewEngine(db)
	handler := NewHandler(routerEngine, nil, db)

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return handler, db, cleanup
}

func TestChatCompletionRequestStreamField(t *testing.T) {
	// Test that Stream field is properly parsed
	gin.SetMode(gin.TestMode)

	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create a mock backend server
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that stream is set to true in the forwarded request
		var req ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !req.Stream {
			t.Error("Expected stream to be true in forwarded request")
		}

		// Return SSE response
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: {\"id\":\"test\",\"object\":\"chat.completion.chunk\"}\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer mockBackend.Close()

	// Update channel to use mock backend
	handler.db.UpdateChannel(&database.Channel{
		ID:      1,
		Name:    "test-chan",
		BaseURL: mockBackend.URL,
		APIKey:  "sk-test",
		Weight:  10,
		Enabled: true,
	})

	// Create request with stream=true
	reqBody := ChatCompletionRequest{
		Model:    "gpt-3.5-turbo",
		Messages: []ChatCompletionMessage{{Role: "user", Content: "test"}},
		Stream:   true,
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer test-key")

	// Set user ID in context
	c.Set("user_id", int64(1))

	handler.ChatCompletions(c)

	// Check response headers
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected Content-Type text/event-stream, got %s", w.Header().Get("Content-Type"))
	}

	// Check response body contains SSE format
	body := w.Body.String()
	if !strings.Contains(body, "data:") {
		t.Error("Expected SSE format with 'data:' prefix")
	}
}

func TestChatCompletionRequestNonStream(t *testing.T) {
	// Test that non-streaming mode still works
	gin.SetMode(gin.TestMode)

	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create a mock backend server
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ChatCompletionResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-3.5-turbo",
			Choices: []Choice{
				{
					Index: 0,
					Message: ChatCompletionMessage{
						Role:    "assistant",
						Content: "Hello!",
					},
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		})
	}))
	defer mockBackend.Close()

	// Update channel to use mock backend
	handler.db.UpdateChannel(&database.Channel{
		ID:      1,
		Name:    "test-chan",
		BaseURL: mockBackend.URL,
		APIKey:  "sk-test",
		Weight:  10,
		Enabled: true,
	})

	// Create request without stream field (should default to false)
	reqBody := ChatCompletionRequest{
		Model:    "gpt-3.5-turbo",
		Messages: []ChatCompletionMessage{{Role: "user", Content: "test"}},
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer test-key")

	// Set user ID in context
	c.Set("user_id", int64(1))

	handler.ChatCompletions(c)

	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check response is JSON
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected JSON response, got Content-Type: %s", contentType)
	}

	// Parse response
	var resp ChatCompletionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if resp.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", resp.ID)
	}
}

func TestChatCompletionRequestStreamFalse(t *testing.T) {
	// Test that stream=false explicitly works
	gin.SetMode(gin.TestMode)

	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create a mock backend server
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that stream is false in the forwarded request
		var req ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Stream {
			t.Error("Expected stream to be false in forwarded request")
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ChatCompletionResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-3.5-turbo",
			Choices: []Choice{
				{
					Index: 0,
					Message: ChatCompletionMessage{
						Role:    "assistant",
						Content: "Hello!",
					},
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		})
	}))
	defer mockBackend.Close()

	// Update channel to use mock backend
	handler.db.UpdateChannel(&database.Channel{
		ID:      1,
		Name:    "test-chan",
		BaseURL: mockBackend.URL,
		APIKey:  "sk-test",
		Weight:  10,
		Enabled: true,
	})

	// Create request with stream=false
	reqBody := ChatCompletionRequest{
		Model:    "gpt-3.5-turbo",
		Messages: []ChatCompletionMessage{{Role: "user", Content: "test"}},
		Stream:   false,
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer test-key")

	// Set user ID in context
	c.Set("user_id", int64(1))

	handler.ChatCompletions(c)

	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check response is JSON (not SSE)
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected JSON response for stream=false, got Content-Type: %s", contentType)
	}
}
