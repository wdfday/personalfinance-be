.PHONY: help swagger build run start dev clean test install-tools up

# Variables
BINARY_NAME=be
MAIN_FILE=cmd/server/main.go
DOCS_DIR=./docs
GOPATH=$(shell go env GOPATH)
SWAG=$(GOPATH)/bin/swag

help: ## Hiển thị trợ giúp
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install-tools: ## Cài đặt các công cụ cần thiết
	@echo "Installing swag..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✓ Tools installed successfully"

swagger: ## Generate Swagger documentation
	@echo "📚 Generating Swagger documentation..."
	@if [ ! -f "$(SWAG)" ]; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@$(SWAG) init -g $(MAIN_FILE) --output $(DOCS_DIR) --parseDependency --parseInternal
	@echo "✅ Swagger documentation generated successfully"

build: swagger ## Build với Swagger generation
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/server
	@echo "✓ Build completed successfully"

run: build ## Build và chạy server
	@echo "Starting server..."
	@./$(BINARY_NAME)

start: ## Chạy server trực tiếp (không build)
	@echo "Starting server with go run..."
	@go run $(MAIN_FILE)

up: ## Start database và chạy server
	@echo "Starting PostgreSQL..."
	@docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@echo "Starting server..."
	@go run $(MAIN_FILE)

dev: swagger ## Development mode với hot reload (cần cài air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Installing..."; \
		go install github.com/air-verse/air@latest		air; \
	fi

clean: ## Xóa file build và cache
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean
	@echo "✓ Cleaned"

test: ## Chạy tất cả tests
	@echo "Running all tests..."
	@go test ./...

test-verbose: ## Chạy tests với output chi tiết
	@echo "Running tests with verbose output..."
	@go test -v ./...

test-cover: ## Chạy tests với coverage report
	@echo "Running tests with coverage..."
	@go test -cover ./...

test-cover-html: ## Tạo HTML coverage report
	@echo "Generating HTML coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

test-race: ## Chạy tests với race detector
	@echo "Running tests with race detector..."
	@go test -race ./...

test-bench: ## Chạy benchmark tests
	@echo "Running benchmark tests..."
	@go test -bench=. -benchmem ./...

test-watch: ## Chạy tests liên tục (cần cài gotestsum)
	@if command -v gotestsum > /dev/null; then \
		gotestsum --watch ./...; \
	else \
		echo "gotestsum not installed. Run: go install gotest.tools/gotestsum@latest"; \
	fi

format: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: brew install golangci-lint"; \
	fi

docker-build: swagger ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t personalfinance-dss:latest .
	@echo "✓ Docker image built"

docker-up: ## Start với Docker Compose
	@echo "Starting Docker containers..."
	@docker-compose up -d
	@echo "✓ Containers started"

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down
	@echo "✓ Containers stopped"

# Shortcut commands
b: build ## Shortcut for build
r: run ## Shortcut for run
s: swagger ## Shortcut for swagger

