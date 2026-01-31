package session

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/X0Ken/openai-gateway/pkg/database"
)

// Manager handles session business logic
type Manager struct {
	db            *database.DB
	idleTimeout   time.Duration
	cleanupTicker *time.Ticker
	stopCh        chan struct{}
}

// NewManager creates a new session manager
func NewManager(db *database.DB, idleTimeoutMinutes int) *Manager {
	return &Manager{
		db:          db,
		idleTimeout: time.Duration(idleTimeoutMinutes) * time.Minute,
		stopCh:      make(chan struct{}),
	}
}

// Start begins the session cleanup loop
func (m *Manager) Start() {
	m.cleanupTicker = time.NewTicker(5 * time.Minute)
	go m.cleanupLoop()
}

// Stop stops the session cleanup loop
func (m *Manager) Stop() {
	if m.cleanupTicker != nil {
		m.cleanupTicker.Stop()
	}
	close(m.stopCh)
}

// cleanupLoop periodically cleans up expired sessions
func (m *Manager) cleanupLoop() {
	for {
		select {
		case <-m.cleanupTicker.C:
			m.CleanupExpired()
		case <-m.stopCh:
			return
		}
	}
}

// CleanupExpired removes expired sessions
func (m *Manager) CleanupExpired() error {
	return m.db.DeleteExpiredSessions(int(m.idleTimeout.Minutes()))
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(id int64) (*database.Session, error) {
	return m.db.GetSession(id)
}

// GetSessionByUser retrieves the most recent session for a user
func (m *Manager) GetSessionByUser(userID int64) (*database.Session, error) {
	return m.db.GetSessionByUser(userID)
}

// ListSessions retrieves all sessions
func (m *Manager) ListSessions() ([]*database.Session, error) {
	return m.db.ListSessions()
}

// DeleteSession deletes a session
func (m *Manager) DeleteSession(id int64) error {
	return m.db.DeleteSession(id)
}

// Handler handles HTTP requests for session management
type Handler struct {
	manager *Manager
}

// NewHandler creates a new session handler
func NewHandler(manager *Manager) *Handler {
	return &Handler{manager: manager}
}

// RegisterRoutes registers session routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/sessions", h.List)
	r.GET("/sessions/:id", h.Get)
	r.DELETE("/sessions/:id", h.Delete)
}

// List handles session listing
func (h *Handler) List(c *gin.Context) {
	sessions, err := h.manager.ListSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// Get handles retrieving a single session
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	session, err := h.manager.GetSession(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// Delete handles session deletion
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	if err := h.manager.DeleteSession(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
