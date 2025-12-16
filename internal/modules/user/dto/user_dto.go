package dto

type CreateUserRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=255" validate:"required,min=1,max=255"`
	Email string `json:"email" binding:"required,email" validate:"required,email"`
}

type UpdateUserRequest struct {
	Name   string `json:"name" binding:"omitempty,min=1,max=255" validate:"omitempty,min=1,max=255"`
	Email  string `json:"email" binding:"omitempty,email" validate:"omitempty,email"`
	Status *int   `json:"status" binding:"omitempty,oneof=0 1" validate:"omitempty,oneof=0 1"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type PagingRequest struct {
	Page  int    `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`
	Name  string `form:"name" binding:"omitempty" validate:"omitempty"`
	Email string `form:"email" binding:"omitempty" validate:"omitempty"`
}

// UserPagingResponse is a pagination response specific to User module
type UserPagingResponse struct {
	Data       []UserResponse `json:"data"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	Total      int64          `json:"total"`
	TotalPages int            `json:"totalPages"`
}
