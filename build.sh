#!/bin/bash

# Build script với tự động generate Swagger
# Usage: ./build.sh [options]

set -e

BINARY_NAME="be"
DOCS_DIR="./docs"
GOPATH=$(go env GOPATH)
SWAG="$GOPATH/bin/swag"

echo "🚀 Personal Finance DSS - Build Script"
echo "======================================"

# Function to check if swag is installed
check_swag() {
    if [ ! -f "$SWAG" ]; then
        echo "⚠️  swag không được cài đặt. Đang cài đặt..."
        go install github.com/swaggo/swag/cmd/swag@latest
        echo "✓ swag đã được cài đặt"
    fi
}

# Function to generate swagger docs
generate_swagger() {
    echo ""
    echo "📝 Generating Swagger documentation..."
    "$SWAG" init -g cmd/server/main.go --output "$DOCS_DIR" --parseDependency --parseInternal
    
    if [ $? -eq 0 ]; then
        echo "✓ Swagger documentation generated successfully"
    else
        echo "❌ Failed to generate Swagger documentation"
        exit 1
    fi
}

# Function to build the application
build_app() {
    echo ""
    echo "🔨 Building application..."
    go build -o "$BINARY_NAME" ./cmd/server
    
    if [ $? -eq 0 ]; then
        echo "✓ Build completed successfully: $BINARY_NAME"
    else
        echo "❌ Build failed"
        exit 1
    fi
}

# Function to run the application
run_app() {
    echo ""
    echo "🏃 Starting server..."
    echo "======================================"
    ./"$BINARY_NAME"
}

# Parse command line arguments
RUN_AFTER_BUILD=false
SKIP_SWAGGER=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -r|--run)
            RUN_AFTER_BUILD=true
            shift
            ;;
        -s|--skip-swagger)
            SKIP_SWAGGER=true
            shift
            ;;
        -h|--help)
            echo "Usage: ./build.sh [options]"
            echo ""
            echo "Options:"
            echo "  -r, --run           Run server after build"
            echo "  -s, --skip-swagger  Skip Swagger generation"
            echo "  -h, --help          Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Main execution
check_swag

if [ "$SKIP_SWAGGER" = false ]; then
    generate_swagger
else
    echo "⏭️  Skipping Swagger generation"
fi

build_app

if [ "$RUN_AFTER_BUILD" = true ]; then
    run_app
else
    echo ""
    echo "✅ Done! Run './$BINARY_NAME' to start the server"
    echo "📚 Swagger docs: http://localhost:8080/swagger/index.html"
fi

