package model

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/X0Ken/openai-gateway/pkg/database"
)

// Handler handles model management API requests
type Handler struct {
	db *database.DB
}

// NewHandler creates a new model handler
func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes registers model management routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Model CRUD
	r.POST("/models", h.CreateModel)
	r.GET("/models", h.ListModels)
	r.GET("/models/:id", h.GetModel)
	r.PUT("/models/:id", h.UpdateModel)
	r.DELETE("/models/:id", h.DeleteModel)

	// Model-Channel mappings
	r.POST("/models/:id/channels", h.AddModelChannel)
	r.GET("/models/:id/channels", h.ListModelChannels)
	r.DELETE("/models/:id/channels/:channel_id", h.RemoveModelChannel)
}

// CreateModelRequest represents a model creation request
type CreateModelRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateModel handles creating a new model
func (h *Handler) CreateModel(c *gin.Context) {
	var req CreateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model := &database.Model{
		Name: req.Name,
	}

	if err := h.db.CreateModel(model); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model)
}

// ListModels handles listing all models
func (h *Handler) ListModels(c *gin.Context) {
	models, err := h.db.ListModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models)
}

// GetModel handles retrieving a single model
func (h *Handler) GetModel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	model, err := h.db.GetModel(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if model == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	c.JSON(http.StatusOK, model)
}

// UpdateModelRequest represents a model update request
type UpdateModelRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateModel handles updating a model
func (h *Handler) UpdateModel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	var req UpdateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model, err := h.db.GetModel(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if model == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	model.Name = req.Name
	if err := h.db.UpdateModel(model); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model)
}

// DeleteModel handles deleting a model
func (h *Handler) DeleteModel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	// First remove all model-channel mappings
	if err := h.db.RemoveAllModelChannelsForModel(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Then delete the model
	if err := h.db.DeleteModel(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddModelChannelRequest represents a model-channel mapping request
type AddModelChannelRequest struct {
	ChannelID        int64  `json:"channel_id" binding:"required"`
	BackendModelName string `json:"backend_model_name" binding:"required"`
	Weight           int    `json:"weight"`
}

// AddModelChannel handles adding a channel to a model
func (h *Handler) AddModelChannel(c *gin.Context) {
	modelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	var req AddModelChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mc := &database.ModelChannel{
		ModelID:          modelID,
		ChannelID:        req.ChannelID,
		BackendModelName: req.BackendModelName,
		Weight:           req.Weight,
	}

	if err := h.db.AddModelChannel(mc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, mc)
}

// ListModelChannels handles listing all channels for a model
func (h *Handler) ListModelChannels(c *gin.Context) {
	modelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	mappings, err := h.db.GetModelChannelsByModel(modelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mappings)
}

// RemoveModelChannel handles removing a channel from a model
func (h *Handler) RemoveModelChannel(c *gin.Context) {
	modelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	channelID, err := strconv.ParseInt(c.Param("channel_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel ID"})
		return
	}

	if err := h.db.RemoveModelChannel(modelID, channelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
