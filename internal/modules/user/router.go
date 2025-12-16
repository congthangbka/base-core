package user

import (
	"gorm.io/gorm"

	"github.com/example/clean-architecture/internal/modules/user/handler"
	"github.com/example/clean-architecture/internal/modules/user/repository"
	"github.com/example/clean-architecture/internal/modules/user/service"
	"github.com/example/clean-architecture/internal/modules/user/validator"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all routes for the user module
func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	// Initialize dependencies
	userRepo := repository.NewUserRepository(db)
	baseUserService := service.NewUserService(userRepo)
	// Wrap with metrics instrumentation
	userService := service.NewInstrumentedUserService(baseUserService)
	userValidator := validator.NewUserValidator()
	userHandler := handler.NewUserHandler(userService, userValidator)

	// Define routes
	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("", userHandler.GetAll)
			users.GET("/:id", userHandler.GetByID)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}
	}
}

