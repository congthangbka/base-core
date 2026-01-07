package router

import (
	"time"

	"gorm.io/gorm"

	"llm-aggregator/internal/common"
	"llm-aggregator/internal/config"
	"llm-aggregator/internal/container"
	"llm-aggregator/internal/middleware"
	orderModule "llm-aggregator/internal/modules/order"
	userModule "llm-aggregator/internal/modules/user"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	// DefaultMaxMultipartMemory is the default maximum size for multipart form data (10MB)
	DefaultMaxMultipartMemory = 10 << 20
)

func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Set max request body size from config
	maxMemory := int64(cfg.ServerLimits.MaxRequestSizeMB) << 20
	if maxMemory == 0 {
		maxMemory = DefaultMaxMultipartMemory
	}
	r.MaxMultipartMemory = maxMemory

	// Apply global middleware (order matters!)
	r.Use(middleware.SecurityHeaders()) // Security headers first

	// CORS configuration based on environment
	corsOrigins := middleware.ParseAllowedOrigins(cfg.Server.CORSOrigins)
	if cfg.App.IsProduction {
		r.Use(middleware.CORSWithProductionConfig(corsOrigins, true))
	} else {
		r.Use(middleware.CORS())
	}
	r.Use(middleware.RequestID())                                                                         // Must be second to generate request ID
	r.Use(middleware.RateLimitWithConfig(cfg.ServerLimits.RateLimitRPS, cfg.ServerLimits.RateLimitBurst)) // Rate limiting from config
	r.Use(middleware.Timeout(time.Duration(cfg.ServerLimits.RequestTimeoutSeconds) * time.Second))        // Request timeout from config

	// Request validation middleware
	maxRequestSize := int64(cfg.ServerLimits.MaxRequestSizeMB) << 20
	if maxRequestSize == 0 {
		maxRequestSize = DefaultMaxMultipartMemory
	}
	r.Use(middleware.RequestSizeValidation(maxRequestSize)) // Validate request size
	r.Use(middleware.ContentTypeValidation())               // Validate Content-Type header

	r.Use(middleware.Metrics()) // Metrics before logging for accurate timing
	r.Use(middleware.Logging())
	r.Use(middleware.Recovery())

	// Health check endpoints (skip middleware for faster response)
	healthGroup := r.Group("")
	{
		// Readiness probe - checks if service is ready to accept traffic
		// @Summary     Readiness probe
		// @Description Check if the service is ready to accept traffic (database connection required)
		// @Tags        health
		// @Accept      json
		// @Produce     json
		// @Success     200  {object} map[string]interface{} "Service is ready"
		// @Failure     503  {object} map[string]interface{} "Service is not ready"
		// @Router      /health/ready [get]
		healthGroup.GET("/health/ready", func(c *gin.Context) {
			// Check database connection
			sqlDB, err := db.DB()
			if err != nil {
				c.JSON(503, gin.H{
					"status":  "not_ready",
					"message": "Database connection error",
					"error":   err.Error(),
				})
				return
			}

			if err := sqlDB.Ping(); err != nil {
				c.JSON(503, gin.H{
					"status":  "not_ready",
					"message": "Database ping failed",
					"error":   err.Error(),
				})
				return
			}

			c.JSON(200, gin.H{
				"status":  "ready",
				"message": "Service is ready to accept traffic",
			})
		})

		// Liveness probe - checks if service is alive
		// @Summary     Liveness probe
		// @Description Check if the service is alive (does not check database)
		// @Tags        health
		// @Accept      json
		// @Produce     json
		// @Success     200  {object} map[string]interface{} "Service is alive"
		// @Router      /health/live [get]
		healthGroup.GET("/health/live", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "alive",
				"message": "Service is running",
			})
		})

		// Full health check with database status
		// @Summary     Health check
		// @Description Check if the API is healthy and database is connected
		// @Tags        health
		// @Accept      json
		// @Produce     json
		// @Success     200  {object} map[string]interface{} "Health status with database info"
		// @Failure     503  {object} map[string]interface{} "Service unavailable"
		// @Router      /health [get]
		healthGroup.GET("/health", func(c *gin.Context) {
			// Check database connection
			sqlDB, err := db.DB()
			if err != nil {
				c.JSON(503, gin.H{
					"status":  "unhealthy",
					"message": "Database connection error",
					"error":   err.Error(),
				})
				return
			}

			if err := sqlDB.Ping(); err != nil {
				c.JSON(503, gin.H{
					"status":  "unhealthy",
					"message": "Database ping failed",
					"error":   err.Error(),
				})
				return
			}

			stats := sqlDB.Stats()
			c.JSON(200, gin.H{
				"status": "healthy",
				"database": gin.H{
					"status":           "connected",
					"open_connections": stats.OpenConnections,
					"in_use":           stats.InUse,
					"idle":             stats.Idle,
				},
			})
		})
	}

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Version endpoint
	// @Summary     Get API version
	// @Description Returns the API version information
	// @Tags        system
	// @Accept      json
	// @Produce     json
	// @Success     200  {object} map[string]interface{} "Version information"
	// @Router      /version [get]
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version":     "1.0.0",
			"name":        "LLM Aggregator API",
			"description": "LLM Aggregator - A production-ready Golang API following Clean Architecture and DDD principles",
			"build_time":  "", // Can be set during build: -ldflags "-X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
			"git_commit":  "", // Can be set during build: -ldflags "-X main.gitCommit=$(git rev-parse HEAD)"
		})
	})

	// Create module container for inter-module communication
	moduleContainer := container.NewModuleContainer()

	// API v1 group - create once and pass to modules
	apiV1 := r.Group("/api/v1")
	{
		// Error codes endpoint
		apiV1.GET("/error-codes", common.GetErrorCodes)

		// Register module routes with the apiV1 group
		// Modules register their inter-module interfaces in the container
		// Order: Register User module first (no dependencies)
		userModule.RegisterRoutes(apiV1, db, moduleContainer)

		// Register Order module (depends on UserVerifier/UserGetter)
		// OrderService is automatically registered in the container by RegisterRoutes
		orderModule.RegisterRoutes(apiV1, db, moduleContainer)
	}

	return r
}
