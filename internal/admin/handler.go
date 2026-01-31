package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/X0Ken/openai-gateway/internal/channel"
	"github.com/X0Ken/openai-gateway/internal/session"
	"github.com/X0Ken/openai-gateway/pkg/database"
)

// Handler handles admin API requests
type Handler struct {
	channelMgr *channel.Manager
	sessionMgr *session.Manager
	db         *database.DB
}

// NewHandler creates a new admin handler
func NewHandler(channelMgr *channel.Manager, sessionMgr *session.Manager, db *database.DB) *Handler {
	return &Handler{
		channelMgr: channelMgr,
		sessionMgr: sessionMgr,
		db:         db,
	}
}

// RegisterRoutes registers admin routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Channel management
	r.POST("/channels", h.CreateChannel)
	r.GET("/channels", h.ListChannels)
	r.GET("/channels/:id", h.GetChannel)
	r.PUT("/channels/:id", h.UpdateChannel)
	r.DELETE("/channels/:id", h.DeleteChannel)

	// User management
	r.POST("/users", h.CreateUser)
	r.GET("/users", h.ListUsers)
	r.GET("/users/:id", h.GetUser)
	r.DELETE("/users/:id", h.DeleteUser)

	// Session management
	r.GET("/sessions", h.ListSessions)
	r.DELETE("/sessions/:id", h.DeleteSession)
}

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	APIKey string `json:"api_key" binding:"required"`
	Name   string `json:"name"`
}

// CreateUser creates a new user
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &database.User{
		APIKey: req.APIKey,
		Name:   req.Name,
	}

	if err := h.db.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// ListUsers lists all users
func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.db.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetUser gets a user by ID
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.db.GetUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := h.db.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateChannel creates a new channel
func (h *Handler) CreateChannel(c *gin.Context) {
	var req channel.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.channelMgr.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ch)
}

// ListChannels lists all channels
func (h *Handler) ListChannels(c *gin.Context) {
	channels, err := h.channelMgr.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channels)
}

// GetChannel gets a channel by ID
func (h *Handler) GetChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel ID"})
		return
	}

	ch, err := h.channelMgr.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if ch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	c.JSON(http.StatusOK, ch)
}

// UpdateChannel updates a channel
func (h *Handler) UpdateChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel ID"})
		return
	}

	var req channel.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.channelMgr.Update(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ch)
}

// DeleteChannel deletes a channel
func (h *Handler) DeleteChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel ID"})
		return
	}

	if err := h.channelMgr.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListSessions lists all sessions
func (h *Handler) ListSessions(c *gin.Context) {
	sessions, err := h.sessionMgr.ListSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// DeleteSession deletes a session
func (h *Handler) DeleteSession(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	if err := h.sessionMgr.DeleteSession(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
