package server

import (
	"fmt"
	"log"
	"time"

	"github.com/X0Ken/openai-gateway/internal/admin"
	"github.com/X0Ken/openai-gateway/internal/api"
	"github.com/X0Ken/openai-gateway/internal/auth"
	"github.com/X0Ken/openai-gateway/internal/channel"
	"github.com/X0Ken/openai-gateway/internal/config"
	"github.com/X0Ken/openai-gateway/internal/metrics"
	"github.com/X0Ken/openai-gateway/internal/model"
	"github.com/X0Ken/openai-gateway/internal/router"
	"github.com/X0Ken/openai-gateway/internal/session"
	"github.com/X0Ken/openai-gateway/internal/web"
	"github.com/X0Ken/openai-gateway/pkg/database"
	"github.com/X0Ken/openai-gateway/pkg/health"
	"github.com/gin-gonic/gin"
)

// Run starts the HTTP server with all components
func Run() error {
	// Load configuration
	cfgSvc, err := config.NewService("config.yaml")
	if err != nil {
		log.Printf("Warning: failed to load config file, using defaults: %v", err)
	}
	cfg := cfgSvc.Get()

	// Initialize database
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		return err
	}
	defer db.Close()

	// Initialize managers
	channelMgr := channel.NewManager(db)
	sessionMgr := session.NewManager(db, cfg.Session.IdleTimeout)
	sessionMgr.Start()
	defer sessionMgr.Stop()

	// Initialize router engine
	routerEngine := router.NewEngine(db)

	// Initialize health checker
	healthChecker := health.NewChecker(
		time.Duration(cfg.HealthCheck.Interval)*time.Second,
		time.Duration(cfg.HealthCheck.Timeout)*time.Second,
	)
	healthChecker.Start()
	defer healthChecker.Stop()

	// Setup Gin
	r := gin.Default()

	// Apply metrics middleware
	r.Use(metrics.Middleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Metrics endpoint
	if cfg.Metrics.Enabled {
		r.GET("/metrics", metrics.Handler())
	}

	// Initialize auth middleware
	authMiddleware := auth.NewMiddleware(db)

	// OpenAI API routes
	apiHandler := api.NewHandler(routerEngine, channelMgr, db)
	openaiGroup := r.Group("/v1")
	apiHandler.RegisterRoutes(openaiGroup, authMiddleware)

	// Admin API routes
	adminHandler := admin.NewHandler(channelMgr, sessionMgr, db)
	adminGroup := r.Group("/api")
	adminHandler.RegisterRoutes(adminGroup)

	// Model management routes
	modelHandler := model.NewHandler(db)
	modelGroup := r.Group("/api")
	modelHandler.RegisterRoutes(modelGroup)

	// Web UI
	webHandler := web.NewHandler()
	webHandler.RegisterRoutes(r)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	return r.Run(addr)
}
