package common

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RespondSuccess sends a success response (200 OK)
func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse(data))
}

// RespondCreated sends a created response (201 Created)
func RespondCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, SuccessResponse(data))
}

// RespondSuccessWithPagination sends a success response with pagination
func RespondSuccessWithPagination(c *gin.Context, data interface{}, page, pageSize int, total int64) {
	c.JSON(http.StatusOK, SuccessResponseWithPagination(data, page, pageSize, total))
}

// RespondFail sends a failure response with error code
// The HTTP status code is determined by the error code
func RespondFail(c *gin.Context, code string) {
	statusCode := mapErrorCodeToHTTPStatus(code)
	c.JSON(statusCode, FailResponse(code))
}

// RespondFailWithMessage sends a failure response with custom message
func RespondFailWithMessage(c *gin.Context, code, message string) {
	statusCode := mapErrorCodeToHTTPStatus(code)
	c.JSON(statusCode, FailResponseWithMessage(code, message))
}

// RespondFailWithData sends a failure response with additional data
func RespondFailWithData(c *gin.Context, code string, data interface{}) {
	statusCode := mapErrorCodeToHTTPStatus(code)
	c.JSON(statusCode, FailResponseWithData(code, data))
}

// RespondServiceError handles ServiceError and sends appropriate response
// This is the recommended way to handle errors from service layer
func RespondServiceError(c *gin.Context, err error) {
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) {
		// Unknown error, return internal server error
		c.JSON(http.StatusInternalServerError, InternalErrorResponse(err))
		return
	}

	statusCode := mapErrorCodeToHTTPStatus(svcErr.Code)
	c.JSON(statusCode, ServiceErrorResponse(svcErr))
}

// RespondBadRequest sends a 400 Bad Request response
func RespondBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, FailResponseWithMessage(ErrorCodeBadRequest, message))
}

// RespondNotFound sends a 404 Not Found response
func RespondNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, FailResponseWithMessage(ErrorCodeNotFound, message))
}

// RespondInternalError sends a 500 Internal Server Error response
func RespondInternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, InternalErrorResponse(err))
}

// RespondUnauthorized sends a 401 Unauthorized response
func RespondUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, FailResponseWithMessage(ErrorCodeUnauthorized, message))
}

// RespondForbidden sends a 403 Forbidden response
func RespondForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, FailResponseWithMessage(ErrorCodeForbidden, message))
}

// mapErrorCodeToHTTPStatus maps error codes to HTTP status codes
// This follows REST API best practices
func mapErrorCodeToHTTPStatus(code string) int {
	switch code {
	case ErrorCodeNotFound, ErrorCodeUserNotFound, ErrorCodeRecordNotFound:
		return http.StatusNotFound
	case ErrorCodeBadRequest, ErrorCodeInvalid, ErrorCodeValidationError,
		ErrorCodeEmailExists, ErrorCodeUserAlreadyExists, ErrorCodeDuplicateEntry,
		ErrorCodeConstraintViolation:
		return http.StatusBadRequest
	case ErrorCodeUnauthorized, ErrorCodeInvalidCredentials:
		return http.StatusUnauthorized
	case ErrorCodeForbidden, ErrorCodeUserInactive:
		return http.StatusForbidden
	case ErrorCodeRateLimitExceeded:
		return http.StatusTooManyRequests
	case ErrorCodeRequestTimeout:
		return http.StatusGatewayTimeout
	case ErrorCodeInternalError, ErrorCodeDatabaseError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
