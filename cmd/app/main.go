// @title           LLM Aggregator API
// @version         1.0
// @description     LLM Aggregator - A production-ready Golang API following Clean Architecture and DDD principles.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8085
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"llm-aggregator/internal/common"
	"llm-aggregator/internal/config"
	"llm-aggregator/internal/database"
	"llm-aggregator/internal/logger"
	"llm-aggregator/internal/router"
	"llm-aggregator/internal/server"

	_ "llm-aggregator/docs" // Swagger documentation
)

// @Summary     Health check
// @Description Check if the API is healthy
// @Tags        health
// @Accept      json
// @Produce     json
// @Success     200  {object}  map[string]interface{}
// @Router      /health [get]
func main() {
	// Load configuration first (needed for logging config)
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// Initialize response system
	common.IsProductionMode = cfg.App.IsProduction

	// Initialize error message mapping using centralized error codes
	// Frontend can use docs/error_codes.json for reference
	common.SetMessageMap(common.ErrorCodeDescriptions)

	// Initialize logger with config
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	if err := logger.Init(env, cfg.Logging.Directory, cfg.Logging.Level); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer func() {
		logger.Sync()
		logger.Close()
	}()

	// Start log compression job (compress files older than X days)
	logger.StartCompressionJob(cfg.Logging.Directory, cfg.Logging.CompressAfterDays)
	logger.GetLogger().Info("Log compression job started",
		zap.String("directory", cfg.Logging.Directory),
		zap.Int("compress_after_days", cfg.Logging.CompressAfterDays),
	)

	// Start log cleanup job (delete files older than retention days)
	logger.StartCleanupJob(cfg.Logging.Directory, cfg.Logging.RetentionDays)
	logger.GetLogger().Info("Log cleanup job started",
		zap.String("directory", cfg.Logging.Directory),
		zap.Int("retention_days", cfg.Logging.RetentionDays),
	)

	// Initialize database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.GetLogger().Fatal("Failed to connect to database", zap.Error(err))
	}

	// Auto migrate
	if err := database.AutoMigrate(db); err != nil {
		logger.GetLogger().Fatal("Failed to auto migrate", zap.Error(err))
	}

	// Initialize router
	r := router.NewRouter(db, cfg)

	// Log Swagger availability
	logger.GetLogger().Info("Swagger documentation available",
		zap.String("url", fmt.Sprintf("http://%s:%s/swagger/index.html", cfg.Server.Host, cfg.Server.Port)))

	// Start server
	srv := server.NewServer(cfg.Server, r)
	logger.GetLogger().Info("Server starting", zap.String("port", cfg.Server.Port))

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			logger.GetLogger().Error("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.GetLogger().Info("Shutting down server...")

	// The context is used to inform the server it has 30 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close database connection
	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			logger.GetLogger().Error("Error closing database", zap.Error(err))
		} else {
			logger.GetLogger().Info("Database connection closed")
		}
	}

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		logger.GetLogger().Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.GetLogger().Info("Server exited")
}
