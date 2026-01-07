package validator

import (
	"errors"

	"llm-aggregator/internal/modules/order/dto"
)

type OrderValidator struct{}

func NewOrderValidator() *OrderValidator {
	return &OrderValidator{}
}

func (v *OrderValidator) ValidateCreateRequest(req *dto.CreateOrderRequest) error {
	if req.UserID == "" {
		return errors.New("user ID is required")
	}

	if req.ProductName == "" {
		return errors.New("product name is required")
	}

	if len(req.ProductName) > 255 {
		return errors.New("product name must be less than 255 characters")
	}

	if req.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	if req.Amount < 0 {
		return errors.New("amount must be greater than or equal to 0")
	}

	return nil
}

func (v *OrderValidator) ValidateUpdateRequest(req *dto.UpdateOrderRequest) error {
	if req.ProductName != "" && len(req.ProductName) > 255 {
		return errors.New("product name must be less than 255 characters")
	}

	if req.Quantity != nil && *req.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	if req.Amount != nil && *req.Amount < 0 {
		return errors.New("amount must be greater than or equal to 0")
	}

	if req.Status != nil {
		validStatuses := map[int]bool{1: true, 2: true, 3: true}
		if !validStatuses[*req.Status] {
			return errors.New("status must be 1 (pending), 2 (completed), or 3 (cancelled)")
		}
	}

	return nil
}

