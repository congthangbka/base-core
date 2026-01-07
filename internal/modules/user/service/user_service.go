package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"llm-aggregator/internal/common"
	"llm-aggregator/internal/entity"
	"llm-aggregator/internal/modules/user/dto"
	"llm-aggregator/internal/modules/user/repository"
)

type UserService interface {
	Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateUserRequest) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*dto.UserResponse, error)
	GetAll(ctx context.Context, req *dto.PagingRequest) (*dto.UserPagingResponse, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Check if user with email already exists
	existingUser, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, common.ErrNotFound) {
		return nil, common.HandleRepositoryError(err, "", "", "Failed to check user existence")
	}
	if existingUser != nil {
		return nil, common.NewServiceError(common.ErrInvalid, "User with this email already exists", common.ErrorCodeEmailExists)
	}

	// Create new user
	user := &entity.User{
		ID:     uuid.New().String(),
		Name:   req.Name,
		Email:  req.Email,
		Status: 1,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, common.HandleRepositoryError(err, "", "", "Failed to create user")
	}

	return s.toUserResponse(user), nil
}

func (s *userService) Update(ctx context.Context, id string, req *dto.UpdateUserRequest) error {
	// Check if user exists
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return common.HandleRepositoryError(err, "User not found", common.ErrorCodeUserNotFound, "Failed to get user")
	}

	// Check email uniqueness if email is being updated
	if req.Email != "" && req.Email != user.Email {
		existingUser, err := s.repo.FindByEmail(ctx, req.Email)
		if err != nil && !errors.Is(err, common.ErrNotFound) {
			return common.HandleRepositoryError(err, "", "", "Failed to check email uniqueness")
		}
		if existingUser != nil {
			return common.NewServiceError(common.ErrInvalid, "Email already exists", common.ErrorCodeEmailExists)
		}
		user.Email = req.Email
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return common.HandleRepositoryError(err, "User not found", common.ErrorCodeUserNotFound, "Failed to update user")
	}

	return nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	// Check if user exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return common.HandleRepositoryError(err, "User not found", common.ErrorCodeUserNotFound, "Failed to get user")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return common.HandleRepositoryError(err, "User not found", common.ErrorCodeUserNotFound, "Failed to delete user")
	}

	return nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, common.HandleRepositoryError(err, "User not found", common.ErrorCodeUserNotFound, "Failed to get user")
	}

	return s.toUserResponse(user), nil
}

func (s *userService) GetAll(ctx context.Context, req *dto.PagingRequest) (*dto.UserPagingResponse, error) {
	// Set defaults using common helper
	req.Page, req.Limit = common.ValidatePagination(req.Page, req.Limit, common.DefaultPaginationLimitUser)

	// Get users with filters using repository method
	users, total, err := s.repo.FindAllWithFilters(ctx, req.Name, req.Email, req.Page, req.Limit)
	if err != nil {
		return nil, common.HandleRepositoryError(err, "", "", "Failed to get users")
	}

	// Convert to response using common helper
	userResponses := s.convertUsersToResponses(users)

	return &dto.UserPagingResponse{
		Data:       userResponses,
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: common.CalculateTotalPages(total, req.Limit),
	}, nil
}

// convertUsersToResponses converts a slice of User entities to UserResponse DTOs.
// This helper method eliminates code duplication.
func (s *userService) convertUsersToResponses(users []entity.User) []dto.UserResponse {
	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *s.toUserResponse(&user)
	}
	return userResponses
}

func (s *userService) toUserResponse(user *entity.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
