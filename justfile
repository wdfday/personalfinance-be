# Personal Finance DSS - justfile
# Run `just` or `just --list` to see all available commands

# Variables
binary_name := "be"
main_file := "cmd/server/main.go"
docs_dir := "./docs"

# Default recipe (runs when you type `just`)
default:
    @just --list

# Show all available commands (alias for default)
help:
    @just --list

# Install necessary tools
install-tools:
    @echo "Installing swag..."
    @go install github.com/swaggo/swag/cmd/swag@latest
    @echo "✓ Tools installed successfully"

# Generate Swagger documentation
swagger:
    #!/usr/bin/env bash
    echo "📚 Generating Swagger documentation..."
    export PATH="$PATH:$(go env GOPATH)/bin"
    if ! command -v swag &> /dev/null; then
        echo "Installing swag..."
        go install github.com/swaggo/swag/cmd/swag@latest
    fi
    swag init -g {{main_file}} --output {{docs_dir}} --parseDependency --parseInternal
    echo "✅ Swagger documentation generated successfully"

# Build with Swagger generation
build: swagger
    @echo "Building {{binary_name}}..."
    @go build -o {{binary_name}} ./cmd/server
    @echo "✓ Build completed successfully"

# Run server with hot reload (using air)
run:
    #!/usr/bin/env bash
    export PATH="$PATH:$(go env GOPATH)/bin"
    if ! command -v air &> /dev/null; then
        echo "Installing air..."
        go install github.com/air-verse/air@latest
    fi
    echo "🔥 Starting server with hot reload..."
    air

# Run server directly (without hot reload)
start:
    @echo "Starting server with go run..."
    @go run {{main_file}} serve

# Build binary
build-bin: swagger
    @echo "Building {{binary_name}}..."
    @go build -o {{binary_name}} ./cmd/server
    @echo "✓ Build completed successfully"

# Start database and run server
up:
    @echo "Starting PostgreSQL..."
    @docker-compose up -d
    @echo "Waiting for PostgreSQL to be ready..."
    @sleep 3
    @echo "Starting server..."
    @go run {{main_file}} serve

# Development mode with hot reload (requires air)
dev: swagger
    #!/usr/bin/env bash
    export PATH="$PATH:$(go env GOPATH)/bin"
    if ! command -v air &> /dev/null; then
        echo "Air not installed. Installing..."
        go install github.com/air-verse/air@latest
    fi
    echo "🔥 Starting server with hot reload..."
    air

# Clean build files and cache
clean:
    @echo "Cleaning..."
    @rm -f {{binary_name}}
    @go clean
    @echo "✓ Cleaned"

# Run all tests
test:
    @echo "Running all tests..."
    @go test ./...

# Run tests with verbose output
test-verbose:
    @echo "Running tests with verbose output..."
    @go test -v ./...

# Run tests with coverage report
test-cover:
    @echo "Running tests with coverage..."
    @go test -cover ./...


# Run tests with race detector
test-race:
    @echo "Running tests with race detector..."
    @go test -race ./...

# Run benchmark tests
test-bench:
    @echo "Running benchmark tests..."
    @go test -bench=. -benchmem ./...

# Run tests continuously (requires gotestsum)
test-watch:
    #!/usr/bin/env bash
    if command -v gotestsum > /dev/null; then
        gotestsum --watch ./...
    else
        echo "gotestsum not installed. Run: go install gotest.tools/gotestsum@latest"
        exit 1
    fi

# Run simulation tests
simulate:
    @echo "Running simulation tests..."
    @go test -v ./tests/simulation/...

# Format code
format:
    @echo "Formatting code..."
    @go fmt ./...
    @echo "✓ Code formatted"

# Run linter
lint:
    #!/usr/bin/env bash
    if command -v golangci-lint > /dev/null; then
        echo "Running linter..."
        golangci-lint run
    else
        echo "golangci-lint not installed. Run: brew install golangci-lint"
        exit 1
    fi

# Build Docker image
docker-build: swagger
    @echo "Building Docker image..."
    @docker build -t personalfinance-dss:latest .
    @echo "✓ Docker image built"

# Start with Docker Compose
docker-up:
    @echo "Starting Docker containers..."
    @docker-compose up -d
    @echo "✓ Containers started"

# Stop Docker containers
docker-down:
    @echo "Stopping Docker containers..."
    @docker-compose down
    @echo "✓ Containers stopped"

# Restart Docker containers
docker-restart: docker-down docker-up

# View Docker logs
docker-logs:
    @docker-compose logs -f

# Database migrations
migrate:
    @echo "Running database migrations..."
    @go run {{main_file}} migrate

# Database migration reset (drop all + migrate)
migrate-reset:
    @echo "⚠️ Resetting database..."
    @go run {{main_file}} migrate reset

# Seed database with test data
seed:
    @echo "Seeding database..."
    @go run {{main_file}} seed

# Seed only categories
seed-categories:
    @go run {{main_file}} seed categories

# Clean database: drop all tables + fresh migrations (NO seed)
db-clean:
    @echo "🧹 Cleaning database (drop + migrate, NO seed)..."
    @go run {{main_file}} db clean

# Full database reset: drop + migrate + seed (all in one atomic command)
db-reset:
    @echo "🔄 Running complete database reset..."
    @go run {{main_file}} db reset

# Generate Go code (swagger, etc.)
generate: swagger
    @echo "Generating Go code..."
    @go generate ./...
    @echo "✓ Code generation completed"

# Tidy Go modules
tidy:
    @echo "Tidying Go modules..."
    @go mod tidy
    @echo "✓ Modules tidied"

# Download Go dependencies
deps:
    @echo "Downloading dependencies..."
    @go mod download
    @echo "✓ Dependencies downloaded"

# Verify Go dependencies
verify:
    @echo "Verifying dependencies..."
    @go mod verify
    @echo "✓ Dependencies verified"

# Full setup (install tools, download deps, generate code)
setup: install-tools deps swagger
    @echo "✅ Setup completed successfully"

# Pre-commit checks (format, lint, test)
pre-commit: format lint test
    @echo "✅ Pre-commit checks passed"

# CI pipeline (used in GitHub Actions)
ci: format lint test-race test-cover
    @echo "✅ CI pipeline completed"

# Shortcut aliases
alias b := build
alias r := run
alias s := swagger
alias t := test
alias tv := test-verbose
alias tc := test-cover
alias f := format
alias l := lint
alias d := dev
alias sim := simulate
