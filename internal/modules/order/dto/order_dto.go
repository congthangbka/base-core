package dto

type CreateOrderRequest struct {
	UserID      string  `json:"userId" binding:"required" validate:"required"`
	ProductName string  `json:"productName" binding:"required,min=1,max=255" validate:"required,min=1,max=255"`
	Quantity    int     `json:"quantity" binding:"required,min=1" validate:"required,min=1"`
	Amount      float64 `json:"amount" binding:"required,min=0" validate:"required,min=0"`
}

type UpdateOrderRequest struct {
	ProductName string  `json:"productName" binding:"omitempty,min=1,max=255" validate:"omitempty,min=1,max=255"`
	Quantity    *int    `json:"quantity" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	Amount      *float64 `json:"amount" binding:"omitempty,min=0" validate:"omitempty,min=0"`
	Status      *int    `json:"status" binding:"omitempty,oneof=1 2 3" validate:"omitempty,oneof=1 2 3"`
}

type OrderResponse struct {
	ID          string  `json:"id"`
	UserID      string  `json:"userId"`
	UserName    string  `json:"userName"`    // From UserService
	UserEmail   string  `json:"userEmail"`    // From UserService
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	Amount      float64 `json:"amount"`
	Status      int     `json:"status"`
	StatusText  string  `json:"statusText"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type OrderPagingRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`
	UserID     string `form:"userId" binding:"omitempty" validate:"omitempty"`
	ProductName string `form:"productName" binding:"omitempty" validate:"omitempty"`
	Status     *int   `form:"status" binding:"omitempty,oneof=1 2 3" validate:"omitempty,oneof=1 2 3"`
}

type OrderPagingResponse struct {
	Data       []OrderResponse `json:"data"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	Total      int64           `json:"total"`
	TotalPages int             `json:"totalPages"`
}

