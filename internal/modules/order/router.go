package order

import (
	"gorm.io/gorm"

	"llm-aggregator/internal/container"
	"llm-aggregator/internal/modules/order/handler"
	"llm-aggregator/internal/modules/order/repository"
	"llm-aggregator/internal/modules/order/service"
	"llm-aggregator/internal/modules/order/validator"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all routes for the order module
// r should be a router group (e.g., /api/v1) not the root router
// container is the module container for inter-module communication
// Returns the order service so it can be registered in the container
func RegisterRoutes(r gin.IRouter, db *gorm.DB, container *container.ModuleContainer) service.OrderService {
	// Initialize dependencies
	orderRepo := repository.NewOrderRepository(db)
	// Pass container and db to service for transaction support
	orderService := service.NewOrderServiceWithDB(orderRepo, container, db)
	orderValidator := validator.NewOrderValidator()
	orderHandler := handler.NewOrderHandler(orderService, orderValidator)

	// Create adapter for inter-module communication
	orderAdapter := service.NewOrderServiceAdapter(orderService)
	container.SetOrderService(orderAdapter)

	// Define routes - r is already /api/v1 group, so just add /orders
	orders := r.Group("/orders")
	{
		orders.POST("", orderHandler.Create)
		orders.GET("", orderHandler.GetAll)
		orders.GET("/:id", orderHandler.GetByID)
		orders.PUT("/:id", orderHandler.Update)
		orders.DELETE("/:id", orderHandler.Delete)
		orders.GET("/user/:userId", orderHandler.GetByUserID)
	}

	// Return the service so it can be registered in the container
	return orderService
}
