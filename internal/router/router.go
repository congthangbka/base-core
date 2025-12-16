package router

import (
	"time"

	"gorm.io/gorm"

	"github.com/example/clean-architecture/internal/common"
	"github.com/example/clean-architecture/internal/config"
	"github.com/example/clean-architecture/internal/middleware"
	userModule "github.com/example/clean-architecture/internal/modules/user"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Set max request body size (10MB)
	r.MaxMultipartMemory = 10 << 20 // 10 MB

	// Apply global middleware (order matters!)
	r.Use(middleware.SecurityHeaders())                                                                   // Security headers first
	r.Use(middleware.CORS())                                                                              // CORS before other middleware
	r.Use(middleware.RequestID())                                                                         // Must be second to generate request ID
	r.Use(middleware.RateLimitWithConfig(cfg.ServerLimits.RateLimitRPS, cfg.ServerLimits.RateLimitBurst)) // Rate limiting from config
	r.Use(middleware.Timeout(time.Duration(cfg.ServerLimits.RequestTimeoutSeconds) * time.Second))        // Request timeout from config
	r.Use(middleware.Metrics())                                                                           // Metrics before logging for accurate timing
	r.Use(middleware.Logging())
	r.Use(middleware.Recovery())

	// Health check with database status
	// @Summary     Health check
	// @Description Check if the API is healthy and database is connected
	// @Tags        health
	// @Accept      json
	// @Produce     json
	// @Success     200  {object} map[string]interface{} "Health status with database info"
	// @Failure     503  {object} map[string]interface{} "Service unavailable"
	// @Router      /health [get]
	r.GET("/health", func(c *gin.Context) {
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

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Error codes endpoint
	r.GET("/api/v1/error-codes", common.GetErrorCodes)

	// Register module routes
	userModule.RegisterRoutes(r, db)

	return r
}
