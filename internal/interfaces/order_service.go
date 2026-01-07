package interfaces

import (
	"context"
)

// OrderService defines the interface for order operations across modules.
// This interface is used for inter-module communication to avoid circular dependencies.
//
// Usage example:
//   order, err := orderService.GetByID(ctx, orderID)
//   if err != nil {
//       return err
//   }
type OrderService interface {
	// GetByID retrieves an order by ID.
	// Returns order data if found, error if not found or retrieval fails.
	GetByID(ctx context.Context, id string) (interface{}, error)
}

// OrderInfo contains minimal order information needed for inter-module communication.
// This avoids importing module-specific DTOs and prevents circular dependencies.
type OrderInfo struct {
	ID          string  // Order unique identifier
	UserID      string  // User who created the order
	ProductName string  // Product name
	Quantity    int     // Order quantity
	Amount      float64 // Order amount
	Status      int     // Order status
}

