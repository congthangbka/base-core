package validator

import (
	"fmt"

	"github.com/example/clean-architecture/internal/modules/user/dto"
	"github.com/go-playground/validator/v10"
)

type UserValidator struct {
	validate *validator.Validate
}

func NewUserValidator() *UserValidator {
	v := validator.New()
	return &UserValidator{validate: v}
}

func (uv *UserValidator) ValidateCreate(req *dto.CreateUserRequest) error {
	if err := uv.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

func (uv *UserValidator) ValidateUpdate(req *dto.UpdateUserRequest) error {
	if err := uv.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

func (uv *UserValidator) ValidatePaging(req *dto.PagingRequest) error {
	if err := uv.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	return nil
}
