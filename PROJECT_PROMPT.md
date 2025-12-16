# Complete Project Prompt for Clean Architecture Go API

## Project Overview

Build a production-ready REST API in Go following Clean Architecture principles with comprehensive features including logging, metrics, error handling, and security.

## Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: Gin
- **ORM**: GORM with MySQL driver
- **Logging**: Zap (structured logging)
- **Metrics**: Prometheus
- **Documentation**: Swagger/OpenAPI
- **Validation**: go-playground/validator
- **UUID**: google/uuid
- **Rate Limiting**: golang.org/x/time/rate
- **Environment**: godotenv

## Architecture Requirements

### 1. Clean Architecture Layers

```
internal/
├── common/          # Shared utilities (errors, responses, error codes)
├── config/          # Configuration management
├── database/        # Database connection and migrations
├── entity/          # Domain entities (shared across modules)
├── logger/          # Logging system
├── metrics/         # Prometheus metrics
├── middleware/      # HTTP middleware
├── modules/         # Business modules
│   └── {module}/
│       ├── dto/     # Data Transfer Objects
│       ├── handler/ # HTTP handlers
│       ├── repository/ # Data access layer
│       ├── service/ # Business logic
│       ├── validator/ # Validation logic
│       └── router.go # Module routes
├── router/          # Main router
├── server/          # HTTP server
└── store/           # Query builder utilities
```

### 2. Core Principles

- **Separation of Concerns**: Clear boundaries between layers
- **Dependency Inversion**: Depend on interfaces, not implementations
- **Context Propagation**: Use `context.Context` throughout
- **Error Handling**: Centralized error codes and messages
- **Type Safety**: Use constants instead of magic strings

## Detailed Requirements

### 1. Error Handling System

**File**: `internal/common/error_codes.go`
- Define all error codes as constants (e.g., `ErrorCodeNotFound`, `ErrorCodeEmailExists`)
- Create `ErrorCodeDescriptions` map for error code to message mapping
- Categories: General, User-related, Database errors

**File**: `internal/common/errors.go`
- `ServiceError` struct with `Err`, `Message`, `Code` fields
- `NewServiceError()` function
- Standard errors: `ErrNotFound`, `ErrInvalid`, `ErrInternal`

**File**: `docs/error_codes.json`
- JSON file with all error codes, messages, and HTTP status codes
- Organized by categories (GENERAL, USER, DATABASE)
- Shareable with Frontend team

**File**: `docs/error_codes.md`
- Documentation for error codes
- Usage examples for Backend and Frontend

### 2. Response System

**File**: `internal/common/response.go`
- `AppResponse` struct: unified response format
  ```go
  type AppResponse struct {
      IsSuccess  bool        `json:"isSuccess"`
      Data       interface{} `json:"data,omitempty"`
      Error      *ErrorInfo  `json:"error,omitempty"`
      Pagination *Pagination `json:"pagination,omitempty"`
  }
  ```
- `ErrorInfo` struct with `Code` and `Message` (type-safe)
- `Pagination` struct for pagination metadata
- Thread-safe message map with mutex
- Production mode flag to hide error details
- Functions:
  - `SuccessResponse(data)`
  - `SuccessResponseWithPagination(data, page, pageSize, total)`
  - `FailResponse(code)`
  - `FailResponseWithMessage(code, message)`
  - `InternalErrorResponse(err)`
  - `ServiceErrorResponse(svcErr)`

**File**: `internal/common/response_helper.go`
- Helper functions for handlers:
  - `RespondSuccess(c, data)`
  - `RespondCreated(c, data)`
  - `RespondSuccessWithPagination(c, data, page, pageSize, total)`
  - `RespondServiceError(c, err)` - Auto maps ServiceError to HTTP status
  - `RespondBadRequest(c, message)`
  - `RespondNotFound(c, message)`
  - `RespondInternalError(c, err)`
  - `RespondUnauthorized(c, message)`
  - `RespondForbidden(c, message)`
- Auto HTTP status mapping from error codes

### 3. Logging System

**File**: `internal/logger/logger.go`
- Initialize Zap logger with file rotation
- Separate files: `app-YYYY-MM-DD.log` and `error-YYYY-MM-DD.log`
- JSON format for production, console for development
- Context-aware logging with Request ID

**File**: `internal/logger/file_writer.go`
- `DailyFileWriter` for automatic daily file rotation
- Thread-safe file writing
- Automatic file creation based on date

**File**: `internal/logger/compression.go`
- Background job to compress old log files (gzip)
- Configurable: compress files older than X days
- Automatic compression to `.gz` format
- Remove original file after compression

**File**: `internal/logger/cleanup.go`
- Background job to delete old log files
- Configurable retention period
- Delete both compressed and uncompressed files

**File**: `internal/logger/context.go`
- Attach Request ID to logs
- Context propagation for logging

**Configuration**:
- `LOG_DIRECTORY`: Log file directory (default: `./logs`)
- `LOG_RETENTION_DAYS`: Days to keep logs (default: 30)
- `LOG_COMPRESS_AFTER_DAYS`: Days before compression (default: 7)
- `LOG_LEVEL`: Log level (default: `info`)

### 4. Middleware

**File**: `internal/middleware/request_id.go`
- Generate unique Request ID for each request
- Attach to context and response headers

**File**: `internal/middleware/logging.go`
- Log all requests with Request ID
- Log slow requests (>1 second) as warnings
- Include: method, path, status, latency, IP, user agent

**File**: `internal/middleware/recovery.go`
- Panic recovery with stack trace logging
- Return 500 error on panic

**File**: `internal/middleware/metrics.go`
- Prometheus metrics for HTTP requests
- Track: request duration, request count, status codes

**File**: `internal/middleware/ratelimit.go`
- Rate limiting per IP address
- Configurable RPS and burst size
- Thread-safe with cleanup of old limiters
- Return 429 on rate limit exceeded

**File**: `internal/middleware/timeout.go`
- Request timeout middleware
- Configurable timeout duration
- Return 504 on timeout

**File**: `internal/middleware/cors.go`
- CORS headers configuration
- Support for all HTTP methods

**File**: `internal/middleware/security.go`
- Security headers: X-Frame-Options, X-Content-Type-Options, X-XSS-Protection, Referrer-Policy

**File**: `internal/middleware/validation.go`
- Content-Type validation for POST/PUT requests
- Ensure JSON content type

**File**: `internal/middleware/auth.go`
- Basic Auth support
- API Key Auth support
- Bearer Token Auth support

### 5. Metrics System

**File**: `internal/metrics/metrics.go`
- Prometheus metrics:
  - HTTP request metrics (duration, count, status codes)
  - Database operation metrics
  - Business operation metrics (create, update, delete, etc.)
- Expose `/metrics` endpoint

**Decorator Pattern**: Wrap service layer with metrics instrumentation
- File: `internal/modules/{module}/service/{module}_service_metrics.go`
- Track operation duration, success/failure counts, error codes

### 6. Module Structure

Each module follows this structure:

```
modules/{module}/
├── dto/
│   └── {module}_dto.go
│       - Request DTOs (Create, Update, Paging)
│       - Response DTOs ({Module}Response)
│       - Paging DTO ({Module}PagingResponse)
├── handler/
│   └── {module}_handler.go
│       - HTTP handlers with Swagger annotations
│       - Use common.Respond* helpers
│       - Validate requests
│       - Call service layer
├── repository/
│   └── {module}_repository.go
│       - Interface definition
│       - GORM implementation
│       - Use entity.Column pattern for field names
│       - Use store.QueryBuilder for complex queries
├── service/
│   ├── {module}_service.go
│   │   - Interface definition
│   │   - Business logic implementation
│   │   - Use common.ServiceError for errors
│   │   - Use error code constants
│   └── {module}_service_metrics.go
│       - Metrics decorator wrapper
├── validator/
│   └── {module}_validator.go
│       - Custom validation logic
└── router.go
    - RegisterRoutes function
    - Initialize dependencies
    - Define routes
```

### 7. Entity Pattern

**File**: `internal/entity/{entity}.go`
- Domain entity struct
- `Column` struct for field name constants
- Example:
  ```go
  type User struct {
      ID        string
      Name      string
      Email     string
      Status    int
      CreatedAt time.Time
      UpdatedAt time.Time
  }
  
  var Column = struct {
      ID        string
      Name      string
      Email     string
      Status    string
      CreatedAt string
      UpdatedAt string
  }{
      ID:        "id",
      Name:      "name",
      Email:     "email",
      Status:    "status",
      CreatedAt: "created_at",
      UpdatedAt: "updated_at",
  }
  ```
- Usage: `entity.User.Column.Email` instead of hardcoded strings

### 8. Query Builder

**File**: `internal/store/query.go`
- Fluent query builder for GORM
- Support filtering, sorting, pagination
- Type-safe field references

### 9. Configuration

**File**: `internal/config/config.go`
- Load from environment variables
- Support `.env` file
- Config structs:
  - `ServerConfig`: Port, Host
  - `DatabaseConfig`: Host, Port, User, Password, DBName, Charset
  - `LoggingConfig`: Directory, RetentionDays, CompressAfterDays, Level
  - `ServerLimitsConfig`: RequestTimeoutSeconds, RateLimitRPS, RateLimitBurst, MaxRequestSizeMB
  - `AppConfig`: IsProduction

**File**: `.env.example`
- Template with all environment variables
- Default values and descriptions

### 10. Database

**File**: `internal/database/database.go`
- GORM connection setup
- Auto-migrate support
- Connection pooling configuration

**File**: `internal/database/migration.go`
- SQL-based migration system using golang-migrate
- Functions: RunMigrations, DownMigrations, GetCurrentMigrationVersion

### 11. Router

**File**: `internal/router/router.go`
- Main router setup
- Global middleware (order matters):
  1. SecurityHeaders
  2. CORS
  3. RequestID
  4. RateLimit
  5. Timeout
  6. Metrics
  7. Logging
  8. Recovery
  9. RequestValidation
- Health check endpoint with database status
- Prometheus metrics endpoint
- Swagger UI endpoint
- Module route registration

**File**: `internal/modules/{module}/router.go`
- Module-specific route registration
- Dependency initialization
- Route grouping

### 12. Server

**File**: `internal/server/server.go`
- HTTP server configuration
- Graceful shutdown support
- Timeouts: ReadTimeout, WriteTimeout, IdleTimeout
- MaxHeaderBytes configuration

### 13. Main Application

**File**: `cmd/app/main.go`
- Application entry point
- Initialize configuration
- Initialize logger
- Start compression and cleanup jobs
- Initialize database
- Run migrations
- Initialize error message map
- Set production mode
- Create router
- Start server
- Graceful shutdown on SIGINT/SIGTERM
- Close database connection
- Swagger annotations

### 14. Swagger Documentation

- Swagger annotations in handlers
- Global annotations in main.go
- Auto-generate with `swag init`
- Serve at `/swagger/*any`

### 15. Response Format

**Success Response**:
```json
{
  "isSuccess": true,
  "data": { ... }
}
```

**Error Response**:
```json
{
  "isSuccess": false,
  "error": {
    "code": "EMAIL_EXISTS",
    "message": "Email already exists"
  }
}
```

**Pagination Response**:
```json
{
  "isSuccess": true,
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 100,
    "totalPages": 5
  }
}
```

## Implementation Guidelines

### 1. Error Handling
- Always use error code constants from `common.ErrorCode*`
- Use `common.NewServiceError()` to create errors
- Use `common.RespondServiceError()` in handlers
- Never hardcode error strings

### 2. Response Handling
- Always use `common.Respond*` helper functions
- Never create response structs directly in handlers
- Use `common.RespondServiceError()` for service errors

### 3. Field Names
- Always use `entity.{Entity}.Column.{Field}` pattern
- Never hardcode database field names

### 4. Context Propagation
- Always pass `context.Context` through layers
- Use `c.Request.Context()` in handlers
- Pass context to service and repository methods

### 5. Logging
- Use structured logging with zap
- Always include Request ID in logs
- Log errors with stack traces
- Use appropriate log levels

### 6. Module Creation
- Follow the exact structure above
- Create router.go for route registration
- Use dependency injection
- Wrap service with metrics decorator

## Environment Variables

```env
# Server
SERVER_PORT=8085
SERVER_HOST=0.0.0.0
ENV=development

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=clean_architecture
DB_CHARSET=utf8mb4

# Logging
LOG_DIRECTORY=./logs
LOG_RETENTION_DAYS=30
LOG_COMPRESS_AFTER_DAYS=7
LOG_LEVEL=info

# Server Limits
REQUEST_TIMEOUT_SECONDS=30
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200
MAX_REQUEST_SIZE_MB=10
```

## Key Features Summary

1. ✅ Clean Architecture with clear layer separation
2. ✅ Centralized error handling with error codes
3. ✅ Standardized API responses
4. ✅ Daily log rotation with compression and cleanup
5. ✅ Prometheus metrics (HTTP, database, business)
6. ✅ Comprehensive middleware (CORS, Security, Rate Limit, Timeout, etc.)
7. ✅ Swagger documentation
8. ✅ Graceful shutdown
9. ✅ Context propagation
10. ✅ Type-safe field name constants
11. ✅ Module-based routing
12. ✅ Production-ready error handling (hide details in production)
13. ✅ Thread-safe operations
14. ✅ Decorator pattern for metrics
15. ✅ Database migration system

## Example Module: User

Create a complete User module with:
- CRUD operations (Create, Read, Update, Delete)
- List with pagination and filtering
- Email uniqueness validation
- Status field (active/inactive)
- Proper error handling
- Swagger documentation
- Metrics instrumentation

## Deliverables

1. Complete Go project structure
2. All source files with proper implementation
3. `.env.example` file
4. `README.md` with setup instructions
5. `Makefile` with common commands
6. `.gitignore` file
7. `docs/error_codes.json` for Frontend
8. `docs/error_codes.md` documentation

## Code Quality

- Follow Go best practices
- Use interfaces for testability
- Proper error wrapping with `%w`
- Context cancellation support
- Thread-safe operations
- Production-ready code
- Comprehensive comments
- Type safety over convenience

## Testing Considerations

- All layers use interfaces (easily mockable)
- Service layer is pure business logic (no dependencies on HTTP)
- Repository layer is pure data access (no business logic)
- Handler layer is thin (only HTTP concerns)

---

**Note**: This prompt should generate a complete, production-ready Go API following Clean Architecture with all the features listed above.

