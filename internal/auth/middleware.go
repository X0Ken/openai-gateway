package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/X0Ken/openai-gateway/pkg/database"
)

// Middleware provides authentication middleware
type Middleware struct {
	db *database.DB
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(db *database.DB) *Middleware {
	return &Middleware{db: db}
}

// RequireAuth middleware ensures the request has a valid API key
func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing API key"})
			c.Abort()
			return
		}

		user, err := m.db.GetUserByAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		// Store user ID in context for later use
		c.Set("user_id", user.ID)
		c.Set("user", user)
		c.Next()
	}
}

// OptionalAuth middleware extracts API key if present but doesn't require it
func (m *Middleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			c.Next()
			return
		}

		user, err := m.db.GetUserByAPIKey(apiKey)
		if err != nil {
			c.Next()
			return
		}

		if user != nil {
			c.Set("user_id", user.ID)
			c.Set("user", user)
		}

		c.Next()
	}
}

// extractAPIKey extracts the API key from the Authorization header
func extractAPIKey(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}

	// Expect "Bearer <api_key>"
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(int64)
	return id, ok
}

// GetUser retrieves the user from the context
func GetUser(c *gin.Context) (*database.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	u, ok := user.(*database.User)
	return u, ok
}
