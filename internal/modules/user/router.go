package user

import (
	"gorm.io/gorm"

	"llm-aggregator/internal/container"
	"llm-aggregator/internal/modules/user/handler"
	"llm-aggregator/internal/modules/user/repository"
	"llm-aggregator/internal/modules/user/service"
	"llm-aggregator/internal/modules/user/validator"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all routes for the user module
// r should be a router group (e.g., /api/v1) not the root router
// container is the module container for inter-module communication
// Returns the user service so it can be registered in the container
func RegisterRoutes(r gin.IRouter, db *gorm.DB, container *container.ModuleContainer) service.UserService {
	// Initialize dependencies
	userRepo := repository.NewUserRepository(db)
	baseUserService := service.NewUserService(userRepo)
	// Wrap with metrics instrumentation
	userService := service.NewInstrumentedUserService(baseUserService)
	userValidator := validator.NewUserValidator()
	userHandler := handler.NewUserHandler(userService, userValidator)

	// Create adapter for inter-module communication
	userAdapter := service.NewUserServiceAdapter(userService)
	container.SetUserVerifier(userAdapter)
	container.SetUserGetter(userAdapter)

	// Define routes - r is already /api/v1 group, so just add /users
	users := r.Group("/users")
	{
		users.POST("", userHandler.Create)
		users.GET("", userHandler.GetAll)
		users.GET("/:id", userHandler.GetByID)
		users.PUT("/:id", userHandler.Update)
		users.DELETE("/:id", userHandler.Delete)
	}

	// Return the service so it can be registered in the container
	return userService
}
