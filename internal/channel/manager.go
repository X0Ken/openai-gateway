package channel

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/X0Ken/openai-gateway/pkg/database"
)

// Manager handles channel business logic
type Manager struct {
	db *database.DB
}

// NewManager creates a new channel manager
func NewManager(db *database.DB) *Manager {
	return &Manager{db: db}
}

// CreateRequest represents a channel creation request
type CreateRequest struct {
	Name    string `json:"name" binding:"required"`
	BaseURL string `json:"base_url" binding:"required"`
	APIKey  string `json:"api_key" binding:"required"`
	Weight  int    `json:"weight"`
	Enabled bool   `json:"enabled"`
}

// UpdateRequest represents a channel update request
type UpdateRequest struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Weight  int    `json:"weight"`
	Enabled *bool  `json:"enabled"`
}

// Create creates a new channel
func (m *Manager) Create(req *CreateRequest) (*database.Channel, error) {
	if req.Weight <= 0 {
		req.Weight = 10
	}

	channel := &database.Channel{
		Name:    req.Name,
		BaseURL: req.BaseURL,
		APIKey:  req.APIKey,
		Weight:  req.Weight,
		Enabled: req.Enabled,
	}

	if err := m.db.CreateChannel(channel); err != nil {
		return nil, err
	}

	return channel, nil
}

// Get retrieves a channel by ID
func (m *Manager) Get(id int64) (*database.Channel, error) {
	return m.db.GetChannel(id)
}

// List retrieves all channels
func (m *Manager) List() ([]*database.Channel, error) {
	return m.db.ListChannels()
}

// ListEnabled retrieves all enabled channels
func (m *Manager) ListEnabled() ([]*database.Channel, error) {
	return m.db.ListEnabledChannels()
}

// Update updates a channel
func (m *Manager) Update(id int64, req *UpdateRequest) (*database.Channel, error) {
	channel, err := m.db.GetChannel(id)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}

	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.BaseURL != "" {
		channel.BaseURL = req.BaseURL
	}
	if req.APIKey != "" {
		channel.APIKey = req.APIKey
	}
	if req.Weight > 0 {
		channel.Weight = req.Weight
	}
	if req.Enabled != nil {
		channel.Enabled = *req.Enabled
	}

	if err := m.db.UpdateChannel(channel); err != nil {
		return nil, err
	}

	return channel, nil
}

// Delete deletes a channel
func (m *Manager) Delete(id int64) error {
	return m.db.DeleteChannel(id)
}

// Handler handles HTTP requests for channel management
type Handler struct {
	manager *Manager
}

// NewHandler creates a new channel handler
func NewHandler(manager *Manager) *Handler {
	return &Handler{manager: manager}
}

// RegisterRoutes registers channel routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/channels", h.Create)
	r.GET("/channels", h.List)
	r.GET("/channels/:id", h.Get)
	r.PUT("/channels/:id", h.Update)
	r.DELETE("/channels/:id", h.Delete)
}

// Create handles channel creation
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.manager.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, channel)
}

// List handles channel listing
func (h *Handler) List(c *gin.Context) {
	channels, err := h.manager.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channels)
}

// Get handles retrieving a single channel
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel ID"})
		return
	}

	channel, err := h.manager.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if channel == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// Update handles channel updates
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel ID"})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.manager.Update(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// Delete handles channel deletion
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel ID"})
		return
	}

	if err := h.manager.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
