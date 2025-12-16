# Error Codes Reference

This document contains all error codes used by the API. Frontend developers can use this to implement consistent error handling.

## File Location

- **Go Constants**: `internal/common/error_codes.go`
- **JSON Reference**: `docs/error_codes.json`
- **This Document**: `docs/error_codes.md`

## Usage

### For Backend Developers

```go
import "github.com/example/clean-architecture/internal/common"

// Use constants instead of hardcoded strings
return common.NewServiceError(err, "User not found", common.ErrorCodeUserNotFound)
```

### For Frontend Developers

Use `docs/error_codes.json` to:
1. Generate TypeScript enums/types
2. Create error message mappings
3. Implement consistent error handling

Example TypeScript:
```typescript
enum ErrorCode {
  INTERNAL_ERROR = "INTERNAL_ERROR",
  NOT_FOUND = "NOT_FOUND",
  EMAIL_EXISTS = "EMAIL_EXISTS",
  // ... more codes
}

const errorMessages: Record<ErrorCode, string> = {
  [ErrorCode.INTERNAL_ERROR]: "An internal server error occurred",
  [ErrorCode.NOT_FOUND]: "Resource not found",
  [ErrorCode.EMAIL_EXISTS]: "Email already exists",
  // ... more messages
};
```

## Error Code Categories

### General Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INTERNAL_ERROR` | 500 | An internal server error occurred |
| `BAD_REQUEST` | 400 | Invalid request |
| `NOT_FOUND` | 404 | Resource not found |
| `UNAUTHORIZED` | 401 | Unauthorized access |
| `FORBIDDEN` | 403 | Access forbidden |
| `VALIDATION_ERROR` | 400 | Validation failed |
| `INVALID` | 400 | Invalid input |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit exceeded |
| `REQUEST_TIMEOUT` | 504 | Request timeout |

### User-Related Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `EMAIL_EXISTS` | 400 | Email already exists |
| `USER_NOT_FOUND` | 404 | User not found |
| `USER_ALREADY_EXISTS` | 400 | User already exists |
| `INVALID_CREDENTIALS` | 401 | Invalid credentials |
| `USER_INACTIVE` | 403 | User account is inactive |

### Database Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `DATABASE_ERROR` | 500 | Database error occurred |
| `RECORD_NOT_FOUND` | 404 | Record not found |
| `DUPLICATE_ENTRY` | 400 | Duplicate entry |
| `CONSTRAINT_VIOLATION` | 400 | Constraint violation |

## Response Format

All error responses follow this format:

```json
{
  "isSuccess": false,
  "error": {
    "code": "EMAIL_EXISTS",
    "message": "Email already exists"
  }
}
```

## Adding New Error Codes

1. Add constant to `internal/common/error_codes.go`
2. Add description to `ErrorCodeDescriptions` map
3. Update `docs/error_codes.json`
4. Update this document
5. Add HTTP status mapping in `response_helper.go` if needed

## Best Practices

1. **Always use constants** - Never hardcode error code strings
2. **Consistent naming** - Use UPPER_SNAKE_CASE
3. **Descriptive codes** - Make codes self-explanatory
4. **Group by domain** - Organize codes by feature/module
5. **Document changes** - Update all three files when adding codes

