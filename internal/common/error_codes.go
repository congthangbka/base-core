package common

// Error codes used throughout the application
// These codes can be shared with Frontend for consistent error handling
const (
	// General errors
	ErrorCodeInternalError    = "INTERNAL_ERROR"
	ErrorCodeBadRequest       = "BAD_REQUEST"
	ErrorCodeNotFound         = "NOT_FOUND"
	ErrorCodeUnauthorized     = "UNAUTHORIZED"
	ErrorCodeForbidden         = "FORBIDDEN"
	ErrorCodeValidationError  = "VALIDATION_ERROR"
	ErrorCodeInvalid           = "INVALID"
	ErrorCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrorCodeRequestTimeout    = "REQUEST_TIMEOUT"

	// User-related errors
	ErrorCodeEmailExists      = "EMAIL_EXISTS"
	ErrorCodeUserNotFound      = "USER_NOT_FOUND"
	ErrorCodeUserAlreadyExists = "USER_ALREADY_EXISTS"
	ErrorCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrorCodeUserInactive      = "USER_INACTIVE"

	// Database errors
	ErrorCodeDatabaseError     = "DATABASE_ERROR"
	ErrorCodeRecordNotFound    = "RECORD_NOT_FOUND"
	ErrorCodeDuplicateEntry    = "DUPLICATE_ENTRY"
	ErrorCodeConstraintViolation = "CONSTRAINT_VIOLATION"
)

// ErrorCodeDescriptions provides default descriptions for error codes
// Frontend can use this to display user-friendly messages
var ErrorCodeDescriptions = map[string]string{
	// General errors
	ErrorCodeInternalError:    "An internal server error occurred",
	ErrorCodeBadRequest:       "Invalid request",
	ErrorCodeNotFound:         "Resource not found",
	ErrorCodeUnauthorized:     "Unauthorized access",
	ErrorCodeForbidden:        "Access forbidden",
	ErrorCodeValidationError:  "Validation failed",
	ErrorCodeInvalid:          "Invalid input",
	ErrorCodeRateLimitExceeded: "Rate limit exceeded",
	ErrorCodeRequestTimeout:   "Request timeout",

	// User-related errors
	ErrorCodeEmailExists:      "Email already exists",
	ErrorCodeUserNotFound:     "User not found",
	ErrorCodeUserAlreadyExists: "User already exists",
	ErrorCodeInvalidCredentials: "Invalid credentials",
	ErrorCodeUserInactive:     "User account is inactive",

	// Database errors
	ErrorCodeDatabaseError:     "Database error occurred",
	ErrorCodeRecordNotFound:    "Record not found",
	ErrorCodeDuplicateEntry:    "Duplicate entry",
	ErrorCodeConstraintViolation: "Constraint violation",
}

