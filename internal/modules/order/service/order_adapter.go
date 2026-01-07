package service

import (
	"context"

	"llm-aggregator/internal/interfaces"
)

// orderServiceAdapter adapts OrderService to implement inter-module interfaces.
// This allows other modules to use OrderService without circular dependencies.
type orderServiceAdapter struct {
	service OrderService
}

// NewOrderServiceAdapter creates a new adapter that implements inter-module interfaces.
func NewOrderServiceAdapter(service OrderService) *orderServiceAdapter {
	return &orderServiceAdapter{service: service}
}

// GetByID implements interfaces.OrderService
func (a *orderServiceAdapter) GetByID(ctx context.Context, id string) (interface{}, error) {
	return a.service.GetByID(ctx, id)
}

// Ensure orderServiceAdapter implements the interface
var _ interfaces.OrderService = (*orderServiceAdapter)(nil)

