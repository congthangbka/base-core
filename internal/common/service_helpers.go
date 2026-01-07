package common

import (
	"context"
	"errors"
)

// HandleRepositoryError handles repository errors and converts them to ServiceError.
// This is a centralized error handling function to avoid code duplication.
//
// Usage:
//   entity, err := repo.FindByID(ctx, id)
//   if err != nil {
//       return nil, HandleRepositoryError(err, "Entity not found", ErrorCodeNotFound, "Failed to get entity")
//   }
//
// Parameters:
//   - err: The error from repository
//   - notFoundMessage: Message to return if error is ErrNotFound
//   - notFoundCode: Error code to return if error is ErrNotFound
//   - internalErrorMessage: Message to return for other errors
//
// Returns:
//   - *ServiceError: A properly formatted ServiceError
func HandleRepositoryError(err error, notFoundMessage, notFoundCode, internalErrorMessage string) *ServiceError {
	if err == nil {
		return nil
	}

	if errors.Is(err, ErrNotFound) {
		return NewServiceError(err, notFoundMessage, notFoundCode)
	}

	return NewServiceError(err, internalErrorMessage, ErrorCodeInternalError)
}

// HandleRepositoryErrorWithReturn handles repository errors for methods that return (T, error).
// This is similar to HandleRepositoryError but returns (nil, error) for consistency.
//
// Usage:
//   entity, err := repo.FindByID(ctx, id)
//   if err != nil {
//       return nil, HandleRepositoryErrorWithReturn(err, "Entity not found", ErrorCodeNotFound, "Failed to get entity")
//   }
func HandleRepositoryErrorWithReturn(err error, notFoundMessage, notFoundCode, internalErrorMessage string) (*ServiceError, error) {
	svcErr := HandleRepositoryError(err, notFoundMessage, notFoundCode, internalErrorMessage)
	if svcErr == nil {
		return nil, nil
	}
	return nil, svcErr
}

// ValidatePagination sets default values for pagination parameters.
// This ensures consistent pagination handling across all services.
//
// Usage:
//   page, limit := ValidatePagination(req.Page, req.Limit)
//
// Parameters:
//   - page: Current page number (will be set to 1 if <= 0)
//   - limit: Items per page (will be set to defaultLimit if <= 0)
//
// Returns:
//   - int: Validated page number
//   - int: Validated limit
func ValidatePagination(page, limit, defaultLimit int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = defaultLimit
	}
	return page, limit
}

// CalculateTotalPages calculates the total number of pages based on total items and limit.
// This ensures consistent pagination calculation across all services.
//
// Usage:
//   totalPages := CalculateTotalPages(total, limit)
//
// Parameters:
//   - total: Total number of items
//   - limit: Items per page
//
// Returns:
//   - int: Total number of pages
func CalculateTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (int(total) + limit - 1) / limit
}

// DefaultPaginationLimit is the default limit for pagination queries
const DefaultPaginationLimit = 10

// DefaultPaginationLimitUser is the default limit for user pagination (different from order)
const DefaultPaginationLimitUser = 20

// CheckEntityExists is a helper to check if an entity exists before operations.
// This pattern is commonly used before Update/Delete operations.
//
// Usage:
//   if err := CheckEntityExists(ctx, repo, id, "Entity not found", ErrorCodeNotFound); err != nil {
//       return err
//   }
//
// Note: This requires a repository with FindByID method. For more complex cases,
// use the repository directly.
type EntityRepository interface {
	FindByID(ctx context.Context, id string) (interface{}, error)
}

func CheckEntityExists(ctx context.Context, repo EntityRepository, id, notFoundMessage, notFoundCode string) error {
	_, err := repo.FindByID(ctx, id)
	if err != nil {
		return HandleRepositoryError(err, notFoundMessage, notFoundCode, "Failed to check entity existence")
	}
	return nil
}

