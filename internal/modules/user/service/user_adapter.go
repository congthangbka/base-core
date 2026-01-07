package service

import (
	"context"
	"errors"

	"llm-aggregator/internal/common"
	"llm-aggregator/internal/interfaces"
)

// userServiceAdapter adapts UserService to implement inter-module interfaces
// This allows other modules to use UserService without circular dependencies
type userServiceAdapter struct {
	service UserService
}

// NewUserServiceAdapter creates a new adapter that implements inter-module interfaces.
// The adapter implements both UserVerifier and UserGetter interfaces.
func NewUserServiceAdapter(service UserService) *userServiceAdapter {
	return &userServiceAdapter{service: service}
}

// VerifyUserExists implements interfaces.UserVerifier
func (a *userServiceAdapter) VerifyUserExists(ctx context.Context, userID string) error {
	_, err := a.service.GetByID(ctx, userID)
	if err != nil {
		// Check if it's a "not found" error
		if errors.Is(err, common.ErrNotFound) {
			return common.NewServiceError(
				common.ErrInvalid,
				"User not found",
				common.ErrorCodeUserNotFound,
			)
		}
		// Check if it's already a ServiceError with USER_NOT_FOUND code
		if svcErr, ok := err.(*common.ServiceError); ok && svcErr.Code == common.ErrorCodeUserNotFound {
			return err
		}
		// For other errors (internal errors), wrap and return
		return common.NewServiceError(err, "Failed to verify user", common.ErrorCodeInternalError)
	}
	return nil
}

// GetUserByID implements interfaces.UserGetter
func (a *userServiceAdapter) GetUserByID(ctx context.Context, userID string) (*interfaces.UserInfo, error) {
	userResp, err := a.service.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &interfaces.UserInfo{
		ID:     userResp.ID,
		Name:   userResp.Name,
		Email:  userResp.Email,
		Status: userResp.Status,
	}, nil
}

// Ensure userServiceAdapter implements both interfaces
var (
	_ interfaces.UserVerifier = (*userServiceAdapter)(nil)
	_ interfaces.UserGetter   = (*userServiceAdapter)(nil)
)
