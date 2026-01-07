package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"llm-aggregator/internal/common"
	"llm-aggregator/internal/container"
	"llm-aggregator/internal/entity"
	"llm-aggregator/internal/interfaces"
	"llm-aggregator/internal/modules/order/dto"
	"llm-aggregator/internal/modules/order/repository"
)

type OrderService interface {
	Create(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateOrderRequest) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*dto.OrderResponse, error)
	GetAll(ctx context.Context, req *dto.OrderPagingRequest) (*dto.OrderPagingResponse, error)
	GetByUserID(ctx context.Context, userID string, page, limit int) (*dto.OrderPagingResponse, error)
}

type orderService struct {
	repo      repository.OrderRepository
	container *container.ModuleContainer
	db        *gorm.DB
}

func NewOrderService(repo repository.OrderRepository, container *container.ModuleContainer) OrderService {
	return &orderService{
		repo:      repo,
		container: container,
	}
}

// NewOrderServiceWithDB creates a new order service with database connection for transaction support.
// Use this when you need transaction support in service methods.
func NewOrderServiceWithDB(repo repository.OrderRepository, container *container.ModuleContainer, db *gorm.DB) OrderService {
	return &orderService{
		repo:      repo,
		container: container,
		db:        db,
	}
}

func (s *orderService) Create(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// Verify user exists and check if user is active
	// This combines verification and status check in one call to avoid duplicate lookups
	user, err := s.getUserForValidation(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if user != nil && user.Status == 0 {
		return nil, common.NewServiceError(
			common.ErrInvalid,
			"User is inactive",
			common.ErrorCodeInvalid,
		)
	}

	// Create new order
	order := &entity.Order{
		ID:          uuid.New().String(),
		UserID:      req.UserID,
		ProductName: req.ProductName,
		Quantity:    req.Quantity,
		Amount:      req.Amount,
		Status:      entity.OrderStatusPending,
	}

	// Create order (with transaction support if db is available)
	if s.db != nil {
		// Use transaction for atomic operation
		err := common.TransactionWithContext(ctx, s.db, func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			return txRepo.Create(ctx, order)
		})
		if err != nil {
			return nil, common.HandleRepositoryError(err, "", "", "Failed to create order")
		}
	} else {
		// Fallback to non-transactional create
		if err := s.repo.Create(ctx, order); err != nil {
			return nil, common.HandleRepositoryError(err, "", "", "Failed to create order")
		}
	}

	return s.toOrderResponse(ctx, order)
}

func (s *orderService) Update(ctx context.Context, id string, req *dto.UpdateOrderRequest) error {
	// Check if order exists
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return common.HandleRepositoryError(err, "Order not found", common.ErrorCodeNotFound, "Failed to get order")
	}

	// Update fields
	if req.ProductName != "" {
		order.ProductName = req.ProductName
	}
	if req.Quantity != nil {
		order.Quantity = *req.Quantity
	}
	if req.Amount != nil {
		order.Amount = *req.Amount
	}
	if req.Status != nil {
		order.Status = *req.Status
	}

	if err := s.repo.Update(ctx, order); err != nil {
		return common.HandleRepositoryError(err, "Order not found", common.ErrorCodeNotFound, "Failed to update order")
	}

	return nil
}

func (s *orderService) Delete(ctx context.Context, id string) error {
	// Check if order exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return common.HandleRepositoryError(err, "Order not found", common.ErrorCodeNotFound, "Failed to get order")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return common.HandleRepositoryError(err, "Order not found", common.ErrorCodeNotFound, "Failed to delete order")
	}

	return nil
}

func (s *orderService) GetByID(ctx context.Context, id string) (*dto.OrderResponse, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, common.HandleRepositoryError(err, "Order not found", common.ErrorCodeNotFound, "Failed to get order")
	}

	return s.toOrderResponse(ctx, order)
}

func (s *orderService) GetAll(ctx context.Context, req *dto.OrderPagingRequest) (*dto.OrderPagingResponse, error) {
	// Set defaults using common helper
	req.Page, req.Limit = common.ValidatePagination(req.Page, req.Limit, common.DefaultPaginationLimit)

	// Get orders with filters
	orders, total, err := s.repo.FindAllWithFilters(ctx, req.UserID, req.ProductName, req.Status, req.Page, req.Limit)
	if err != nil {
		return nil, common.NewServiceError(err, "Failed to get orders", common.ErrorCodeInternalError)
	}

	// Convert to response using common helper
	orderResponses, err := s.convertOrdersToResponses(ctx, orders)
	if err != nil {
		return nil, err
	}

	return &dto.OrderPagingResponse{
		Data:       orderResponses,
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: common.CalculateTotalPages(total, req.Limit),
	}, nil
}

func (s *orderService) GetByUserID(ctx context.Context, userID string, page, limit int) (*dto.OrderPagingResponse, error) {
	// Set defaults using common helper
	page, limit = common.ValidatePagination(page, limit, common.DefaultPaginationLimit)

	// Verify user exists before fetching orders (inter-module communication)
	if err := s.verifyUserExists(ctx, userID); err != nil {
		return nil, err
	}

	// Get orders by user ID
	orders, total, err := s.repo.FindByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, common.NewServiceError(err, "Failed to get orders", common.ErrorCodeInternalError)
	}

	// Convert to response using common helper
	orderResponses, err := s.convertOrdersToResponses(ctx, orders)
	if err != nil {
		return nil, err
	}

	return &dto.OrderPagingResponse{
		Data:       orderResponses,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: common.CalculateTotalPages(total, limit),
	}, nil
}

// toOrderResponse converts an Order entity to OrderResponse DTO.
// It also populates user information (name, email) from User module via inter-module communication.
func (s *orderService) toOrderResponse(ctx context.Context, order *entity.Order) (*dto.OrderResponse, error) {
	response := &dto.OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		ProductName: order.ProductName,
		Quantity:    order.Quantity,
		Amount:      order.Amount,
		Status:      order.Status,
		StatusText:  s.getStatusText(order.Status),
		CreatedAt:   order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   order.UpdatedAt.Format(time.RFC3339),
	}

	// Populate user information using type-safe inter-module interface
	// This is an example of inter-module communication: Order module calls User module
	if s.container.UserGetter != nil {
		user, err := s.container.UserGetter.GetUserByID(ctx, order.UserID)
		if err == nil && user != nil {
			response.UserName = user.Name
			response.UserEmail = user.Email
		}
		// Note: We silently ignore errors here to avoid breaking the response
		// if user service is temporarily unavailable
	}

	return response, nil
}

// verifyUserExists checks if a user exists using UserVerifier from container.
// Returns an error if user is not found or if verification fails.
// This is a lightweight check that only verifies existence without fetching user data.
func (s *orderService) verifyUserExists(ctx context.Context, userID string) error {
	// Skip verification if UserVerifier is not available
	if s.container.UserVerifier == nil {
		return nil
	}

	// Use type-safe interface to verify user
	return s.container.UserVerifier.VerifyUserExists(ctx, userID)
}

// getUserForValidation gets user information for validation purposes.
// This method combines verification and data retrieval, returning user info if available.
// Returns (user, nil) if user exists, (nil, error) if user not found or retrieval fails.
// Returns (nil, nil) if UserGetter is not available (graceful degradation).
func (s *orderService) getUserForValidation(ctx context.Context, userID string) (*interfaces.UserInfo, error) {
	// If UserGetter is not available, fall back to verification only
	if s.container.UserGetter == nil {
		// Try to verify at least
		if err := s.verifyUserExists(ctx, userID); err != nil {
			return nil, err
		}
		return nil, nil // User exists but we can't get details
	}

	// Get user information (this also verifies existence)
	user, err := s.container.UserGetter.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// convertOrdersToResponses converts a slice of Order entities to OrderResponse DTOs.
// This helper method eliminates code duplication in GetAll and GetByUserID.
func (s *orderService) convertOrdersToResponses(ctx context.Context, orders []entity.Order) ([]dto.OrderResponse, error) {
	orderResponses := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		orderResp, err := s.toOrderResponse(ctx, &order)
		if err != nil {
			return nil, err
		}
		orderResponses[i] = *orderResp
	}
	return orderResponses, nil
}

func (s *orderService) getStatusText(status int) string {
	switch status {
	case entity.OrderStatusPending:
		return "pending"
	case entity.OrderStatusCompleted:
		return "completed"
	case entity.OrderStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}
