.PHONY: help build run test clean migrate lint fmt swagger

help:
	@echo "Available commands:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make migrate     - Run database migrations"
	@echo "  make lint        - Run linter"
	@echo "  make fmt         - Format code"
	@echo "  make swagger     - Generate Swagger docs"

build:
	@echo "Building application..."
	@go build -o bin/app cmd/app/main.go

run:
	@echo "Running application..."
	@go run cmd/app/main.go

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/

migrate:
	@echo "Running migrations..."
	@go run cmd/app/main.go

lint:
	@echo "Running linter..."
	@golangci-lint run

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

swagger:
	@echo "Generating Swagger docs..."
	@swag init -g cmd/app/main.go

