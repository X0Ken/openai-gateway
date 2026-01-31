package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/X0Ken/openai-gateway/internal/auth"
	"github.com/X0Ken/openai-gateway/internal/channel"
	"github.com/X0Ken/openai-gateway/internal/metrics"
	"github.com/X0Ken/openai-gateway/internal/router"
	"github.com/X0Ken/openai-gateway/pkg/database"
	"github.com/gin-gonic/gin"
)

// Handler handles OpenAI API requests
type Handler struct {
	router     *router.Engine
	channelMgr *channel.Manager
	db         *database.DB
}

// NewHandler creates a new API handler
func NewHandler(router *router.Engine, channelMgr *channel.Manager, db *database.DB) *Handler {
	return &Handler{
		router:     router,
		channelMgr: channelMgr,
		db:         db,
	}
}

// RegisterRoutes registers OpenAI API routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMiddleware *auth.Middleware) {
	// OpenAI compatible endpoints
	r.GET("/models", h.ListModels)

	authenticated := r.Group("/")
	authenticated.Use(authMiddleware.RequireAuth())
	{
		authenticated.POST("/chat/completions", h.ChatCompletions)
	}
}

// ChatCompletionRequest represents an OpenAI chat completion request
type ChatCompletionRequest struct {
	Model    string                  `json:"model" binding:"required"`
	Messages []ChatCompletionMessage `json:"messages" binding:"required"`
	Stream   bool                    `json:"stream,omitempty"`
}

// ChatCompletionMessage represents a message in the conversation
type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents an OpenAI chat completion response
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a completion choice
type Choice struct {
	Index   int                   `json:"index"`
	Message ChatCompletionMessage `json:"message"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletions handles chat completion requests
func (h *Handler) ChatCompletions(c *gin.Context) {
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Route to best channel
	routeResult, err := h.router.Route(userID, req.Model)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	// Handle streaming vs non-streaming
	if req.Stream {
		// Streaming mode
		start := time.Now()
		err := h.forwardStreamRequest(c, routeResult.Channel, routeResult.BackendModelName, &req)
		duration := time.Since(start)

		// Update metrics
		metrics.RecordChannelLatency(routeResult.Channel.Name, req.Model, duration)

		if err != nil {
			metrics.RecordChannelError(routeResult.Channel.Name)
			h.db.UpdateChannelMetrics(routeResult.Channel.ID, duration.Seconds(), false)
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		h.db.UpdateChannelMetrics(routeResult.Channel.ID, duration.Seconds(), true)
	} else {
		// Non-streaming mode
		start := time.Now()
		resp, err := h.forwardRequest(routeResult.Channel, routeResult.BackendModelName, &req)
		duration := time.Since(start)

		// Update metrics
		metrics.RecordChannelLatency(routeResult.Channel.Name, req.Model, duration)

		if err != nil {
			metrics.RecordChannelError(routeResult.Channel.Name)
			h.db.UpdateChannelMetrics(routeResult.Channel.ID, duration.Seconds(), false)
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		h.db.UpdateChannelMetrics(routeResult.Channel.ID, duration.Seconds(), true)
		c.JSON(http.StatusOK, resp)
	}
}

// forwardRequest forwards the request to the backend channel
func (h *Handler) forwardRequest(channel *database.Channel, backendModelName string, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Prepare request body with backend-specific model name
	forwardReq := *req
	forwardReq.Model = backendModelName
	body, err := json.Marshal(forwardReq)
	if err != nil {
		return nil, err
	}

	// Create request
	url := channel.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+channel.APIKey)

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("backend error: %s", string(body))
	}

	// Parse response
	var result ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// forwardStreamRequest forwards the request to the backend channel and streams the response
func (h *Handler) forwardStreamRequest(c *gin.Context, channel *database.Channel, backendModelName string, req *ChatCompletionRequest) error {
	// Prepare request body with backend-specific model name and stream enabled
	forwardReq := *req
	forwardReq.Model = backendModelName
	forwardReq.Stream = true
	body, err := json.Marshal(forwardReq)
	if err != nil {
		return err
	}

	// Create request
	url := channel.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+channel.APIKey)

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("backend error: %s", string(body))
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Stream the response
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Forward the line to the client
		c.Writer.Write([]byte(line))
		c.Writer.Flush()
	}

	return nil
}

// Model represents an OpenAI model
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ListModelsResponse represents the models list response
type ListModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// ListModels handles the models list endpoint
func (h *Handler) ListModels(c *gin.Context) {
	// Get all logical models from the database
	modelsList, err := h.db.ListModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	var models []Model
	for _, m := range modelsList {
		models = append(models, Model{
			ID:      m.Name,
			Object:  "model",
			Created: m.CreatedAt.Unix(),
			OwnedBy: "openai-gateway",
		})
	}

	c.JSON(http.StatusOK, ListModelsResponse{
		Object: "list",
		Data:   models,
	})
}
