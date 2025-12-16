package common

import (
	"fmt"
	"sync"
)

// AppResponse is the unified response structure for all API endpoints
// This follows Go best practices: type-safe, clear structure, production-ready
type AppResponse struct {
	IsSuccess  bool        `json:"isSuccess"`
	Data       interface{} `json:"data,omitempty"`
	Error      *ErrorInfo  `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// ErrorInfo contains error details in a type-safe way
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// SuccessResponseDoc is used for Swagger documentation to show clean success responses
// @Description Success response with data
type SuccessResponseDoc struct {
	IsSuccess bool        `json:"isSuccess" example:"true"`
	Data      interface{} `json:"data"`
}

// SuccessResponseWithPaginationDoc is used for Swagger documentation to show success responses with pagination
// @Description Success response with data and pagination
type SuccessResponseWithPaginationDoc struct {
	IsSuccess  bool        `json:"isSuccess" example:"true"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// ErrorResponseDoc is used for Swagger documentation to show clean error responses
// @Description Error response
type ErrorResponseDoc struct {
	IsSuccess bool      `json:"isSuccess" example:"false"`
	Error     ErrorInfo `json:"error"`
}

// ErrorInfoDoc is used for Swagger documentation examples
// @Description Error information with code and message
type ErrorInfoDoc struct {
	Code    string `json:"code" example:"USER_NOT_FOUND"`
	Message string `json:"message" example:"User not found"`
}

// SimpleSuccessResponseDoc is used for Swagger documentation for endpoints that return only success status
// @Description Simple success response without data
type SimpleSuccessResponseDoc struct {
	IsSuccess bool `json:"isSuccess" example:"true"`
}

// Message map for error code to message mapping (thread-safe)
var (
	messageMap      map[string]string
	messageMapMutex sync.RWMutex
)

// IsProductionMode determines if the application is in production mode
// In production, detailed error messages are hidden for security
var IsProductionMode = false

// SetMessageMap initializes the error code to message mapping
// This should be called during application startup
// Example: common.SetMessageMap(map[string]string{"USER_NOT_FOUND": "User not found"})
func SetMessageMap(messages map[string]string) {
	messageMapMutex.Lock()
	defer messageMapMutex.Unlock()
	messageMap = messages
}

// getMessage retrieves the message for an error code (thread-safe)
func getMessage(code string) string {
	messageMapMutex.RLock()
	defer messageMapMutex.RUnlock()

	if msg, ok := messageMap[code]; ok {
		return msg
	}
	return fmt.Sprintf("Unknown error code: %s", code)
}

// SuccessResponse creates a success response with data
func SuccessResponse(data interface{}) *AppResponse {
	return &AppResponse{
		IsSuccess: true,
		Data:      data,
	}
}

// SuccessResponseWithPagination creates a success response with pagination
func SuccessResponseWithPagination(data interface{}, page, pageSize int, total int64) *AppResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &AppResponse{
		IsSuccess: true,
		Data:      data,
		Pagination: &Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

// FailResponse creates a failure response with error code
// The message is automatically retrieved from messageMap
func FailResponse(code string) *AppResponse {
	return &AppResponse{
		IsSuccess: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: getMessage(code),
		},
	}
}

// FailResponseWithMessage creates a failure response with custom message
// Use this when you need to override the default message from messageMap
func FailResponseWithMessage(code, message string) *AppResponse {
	return &AppResponse{
		IsSuccess: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
}

// FailResponseWithData creates a failure response with additional data
// Useful for validation errors where you need to return field-specific errors
func FailResponseWithData(code string, data interface{}) *AppResponse {
	return &AppResponse{
		IsSuccess: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: getMessage(code),
		},
		Data: data, // Additional error data (e.g., validation errors)
	}
}

// InternalErrorResponse creates an internal server error response
// In production mode, detailed error messages are hidden
func InternalErrorResponse(err error) *AppResponse {
	if IsProductionMode {
		return &AppResponse{
			IsSuccess: false,
			Error: &ErrorInfo{
				Code:    ErrorCodeInternalError,
				Message: "An internal error occurred",
			},
		}
	}

	message := "Internal server error"
	if err != nil {
		message = err.Error()
	}

	return &AppResponse{
		IsSuccess: false,
		Error: &ErrorInfo{
			Code:    ErrorCodeInternalError,
			Message: message,
		},
	}
}

// ServiceErrorResponse converts ServiceError to AppResponse
// This is the recommended way to handle service layer errors
func ServiceErrorResponse(svcErr *ServiceError) *AppResponse {
	if IsProductionMode && svcErr.Code == ErrorCodeInternalError {
		return &AppResponse{
			IsSuccess: false,
			Error: &ErrorInfo{
				Code:    svcErr.Code,
				Message: "An internal error occurred",
			},
		}
	}

	return &AppResponse{
		IsSuccess: false,
		Error: &ErrorInfo{
			Code:    svcErr.Code,
			Message: svcErr.Message,
		},
	}
}
