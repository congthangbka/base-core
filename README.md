# Clean Architecture Golang Project

A production-ready Golang project following Clean Architecture and DDD principles.

## Features

- ✅ Clean Architecture - Separation of concerns with clear layers
- ✅ Graceful Shutdown - Proper signal handling and resource cleanup
- ✅ Structured Logging - Zap logger with daily rotation and compression
- ✅ Prometheus Metrics - Comprehensive metrics for monitoring
- ✅ Rate Limiting - Configurable per-IP rate limiting
- ✅ Security - Security headers, CORS, request validation
- ✅ Health Checks - Database-aware health check endpoint
- ✅ Swagger Documentation - Auto-generated API documentation
- ✅ Database Migrations - SQL-based migration system
- ✅ Authentication - Basic Auth, API Key, Bearer Token support
- ✅ Request Timeout - Configurable request timeouts
- ✅ Error Handling - Standardized error responses
- ✅ Module-based Routing - Each module manages its own routes

## Tech Stack

- **Golang** ≥ 1.21
- **Gin Framework** - HTTP web framework
- **GORM** - ORM for MySQL
- **Zap** - Structured logging
- **Prometheus** - Metrics collection
- **Swagger** - API documentation

## Project Structure

```
.
├── cmd/app/              # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection & migrations
│   ├── entity/          # Domain entities
│   ├── logger/          # Logging system
│   ├── metrics/         # Prometheus metrics
│   ├── middleware/      # HTTP middlewares
│   ├── modules/         # Business modules
│   │   └── user/        # User module
│   │       ├── handler/ # HTTP handlers
│   │       ├── service/ # Business logic
│   │       ├── repository/ # Data access
│   │       ├── dto/     # Data transfer objects
│   │       └── validator/ # Input validation
│   ├── router/          # HTTP router
│   ├── server/          # HTTP server
│   └── store/           # Query builder
├── migrations/          # Database migrations (optional)
└── Makefile            # Build commands
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- MySQL 8.0+

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd clean-architecture
```

2. **Install dependencies**
```bash
go mod download
```

3. **Configure environment**
```bash
cp .env.example .env
# Edit .env with your settings
```

4. **Run the application**
```bash
make run
# or
go run cmd/app/main.go
```

## API Documentation

### Swagger UI

After starting the server, access Swagger UI at:
```
http://localhost:8085/swagger/index.html
```

To generate Swagger docs:
```bash
make swagger
# or
swag init -g cmd/app/main.go
```

### API Endpoints

#### User Module

- `POST /api/v1/users` - Create user
- `GET /api/v1/users` - Get all users (with pagination)
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

#### System Endpoints

- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /swagger/*` - Swagger documentation

### Query Parameters

For `GET /api/v1/users`:
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)
- `name` - Filter by name (LIKE)
- `email` - Filter by email (LIKE)

## Configuration

### Environment Variables

See `.env.example` for all available options:

**Server:**
- `SERVER_PORT` - Server port (default: 8085)
- `SERVER_HOST` - Server host (default: 0.0.0.0)

**Database:**
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name

**Logging:**
- `LOG_DIRECTORY` - Log directory (default: ./logs)
- `LOG_RETENTION_DAYS` - Days to keep logs (default: 30)
- `LOG_COMPRESS_AFTER_DAYS` - Days before compression (default: 7)
- `LOG_LEVEL` - Log level (debug, info, warn, error)

**Server Limits:**
- `REQUEST_TIMEOUT_SECONDS` - Request timeout (default: 30)
- `RATE_LIMIT_RPS` - Rate limit requests/second (default: 100)
- `RATE_LIMIT_BURST` - Rate limit burst size (default: 200)
- `MAX_REQUEST_SIZE_MB` - Max request size in MB (default: 10)

## Make Commands

```bash
make build       # Build the application
make run         # Run the application
make test        # Run tests
make clean       # Clean build artifacts
make migrate     # Run database migrations
make lint        # Run linter
make fmt         # Format code
make swagger     # Generate Swagger docs
```

## Testing

```bash
# Run all tests
make test

# Run with coverage
go test -v -race -coverprofile=coverage.out ./...
```

## Monitoring

### Prometheus Metrics

Access metrics at:
```
http://localhost:8085/metrics
```

### Health Check

```bash
curl http://localhost:8085/health
```

## Security Features

- Security headers (XSS protection, frame options, etc.)
- CORS configuration
- Rate limiting
- Request size limits
- Input validation
- Authentication middleware support (Basic, API Key, Bearer Token)

## Database Migrations

### Using Migration System

1. Create migrations directory:
```bash
mkdir -p migrations
```

2. Create migration files:
```
migrations/
├── 001_create_users_table.up.sql
└── 001_create_users_table.down.sql
```

3. Enable in `cmd/app/main.go` (see comments in code)

## Authentication

### Basic Auth
```go
r.Use(middleware.BasicAuth())
```

### API Key Auth
```go
validKeys := []string{"your-api-key"}
r.Use(middleware.APIKeyAuth(validKeys))
```

### Bearer Token Auth
```go
validateToken := func(token string) (bool, error) {
    return isValidToken(token), nil
}
r.Use(middleware.BearerTokenAuth(validateToken))
```

## Code Quality

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
make lint
```

### Formatting

```bash
# Format code
make fmt
```

## License

MIT
