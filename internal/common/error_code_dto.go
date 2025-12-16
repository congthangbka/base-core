package common

// ErrorCodeInfo represents an error code with its details
type ErrorCodeInfo struct {
	Code       string `json:"code" example:"USER_NOT_FOUND"`
	Message    string `json:"message" example:"User not found"`
	HTTPStatus int    `json:"httpStatus" example:"404"`
}

// ErrorCodeCategory represents a category of error codes
type ErrorCodeCategory struct {
	Category   string                   `json:"category" example:"USER"`
	ErrorCodes map[string]ErrorCodeInfo `json:"errorCodes"`
}

// ErrorCodesResponse represents the response for error codes endpoint
type ErrorCodesResponse struct {
	Categories map[string]map[string]ErrorCodeInfo `json:"categories"`
}

// GetAllErrorCodes returns all error codes with their details
func GetAllErrorCodes() ErrorCodesResponse {
	return ErrorCodesResponse{
		Categories: map[string]map[string]ErrorCodeInfo{
			"GENERAL": {
				"INTERNAL_ERROR": {
					Code:       ErrorCodeInternalError,
					Message:    ErrorCodeDescriptions[ErrorCodeInternalError],
					HTTPStatus: 500,
				},
				"BAD_REQUEST": {
					Code:       ErrorCodeBadRequest,
					Message:    ErrorCodeDescriptions[ErrorCodeBadRequest],
					HTTPStatus: 400,
				},
				"NOT_FOUND": {
					Code:       ErrorCodeNotFound,
					Message:    ErrorCodeDescriptions[ErrorCodeNotFound],
					HTTPStatus: 404,
				},
				"UNAUTHORIZED": {
					Code:       ErrorCodeUnauthorized,
					Message:    ErrorCodeDescriptions[ErrorCodeUnauthorized],
					HTTPStatus: 401,
				},
				"FORBIDDEN": {
					Code:       ErrorCodeForbidden,
					Message:    ErrorCodeDescriptions[ErrorCodeForbidden],
					HTTPStatus: 403,
				},
				"VALIDATION_ERROR": {
					Code:       ErrorCodeValidationError,
					Message:    ErrorCodeDescriptions[ErrorCodeValidationError],
					HTTPStatus: 400,
				},
				"INVALID": {
					Code:       ErrorCodeInvalid,
					Message:    ErrorCodeDescriptions[ErrorCodeInvalid],
					HTTPStatus: 400,
				},
				"RATE_LIMIT_EXCEEDED": {
					Code:       ErrorCodeRateLimitExceeded,
					Message:    ErrorCodeDescriptions[ErrorCodeRateLimitExceeded],
					HTTPStatus: 429,
				},
				"REQUEST_TIMEOUT": {
					Code:       ErrorCodeRequestTimeout,
					Message:    ErrorCodeDescriptions[ErrorCodeRequestTimeout],
					HTTPStatus: 504,
				},
			},
			"USER": {
				"EMAIL_EXISTS": {
					Code:       ErrorCodeEmailExists,
					Message:    ErrorCodeDescriptions[ErrorCodeEmailExists],
					HTTPStatus: 400,
				},
				"USER_NOT_FOUND": {
					Code:       ErrorCodeUserNotFound,
					Message:    ErrorCodeDescriptions[ErrorCodeUserNotFound],
					HTTPStatus: 404,
				},
				"USER_ALREADY_EXISTS": {
					Code:       ErrorCodeUserAlreadyExists,
					Message:    ErrorCodeDescriptions[ErrorCodeUserAlreadyExists],
					HTTPStatus: 400,
				},
				"INVALID_CREDENTIALS": {
					Code:       ErrorCodeInvalidCredentials,
					Message:    ErrorCodeDescriptions[ErrorCodeInvalidCredentials],
					HTTPStatus: 401,
				},
				"USER_INACTIVE": {
					Code:       ErrorCodeUserInactive,
					Message:    ErrorCodeDescriptions[ErrorCodeUserInactive],
					HTTPStatus: 403,
				},
			},
			"DATABASE": {
				"DATABASE_ERROR": {
					Code:       ErrorCodeDatabaseError,
					Message:    ErrorCodeDescriptions[ErrorCodeDatabaseError],
					HTTPStatus: 500,
				},
				"RECORD_NOT_FOUND": {
					Code:       ErrorCodeRecordNotFound,
					Message:    ErrorCodeDescriptions[ErrorCodeRecordNotFound],
					HTTPStatus: 404,
				},
				"DUPLICATE_ENTRY": {
					Code:       ErrorCodeDuplicateEntry,
					Message:    ErrorCodeDescriptions[ErrorCodeDuplicateEntry],
					HTTPStatus: 400,
				},
				"CONSTRAINT_VIOLATION": {
					Code:       ErrorCodeConstraintViolation,
					Message:    ErrorCodeDescriptions[ErrorCodeConstraintViolation],
					HTTPStatus: 400,
				},
			},
		},
	}
}
