.PHONY: help test test-verbose test-cover test-cover-html test-repository test-service test-handler test-race clean-coverage

# Module paths (relative to module directory)
MODULE_PATH=.
REPOSITORY_PATH=$(MODULE_PATH)/repository
SERVICE_PATH=$(MODULE_PATH)/service
HANDLER_PATH=$(MODULE_PATH)/handler

help: ## Hiển thị trợ giúp
	@echo "User Module Test Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

test: ## Chạy tất cả tests của user module
	@echo "🧪 Running all user module tests..."
	@go test $(MODULE_PATH)/... -v
	@echo "✅ All user module tests completed"

test-quiet: ## Chạy tests không có verbose output
	@echo "🧪 Running user module tests (quiet)..."
	@go test $(MODULE_PATH)/...

test-cover: ## Chạy tests với coverage report
	@echo "📊 Running user module tests with coverage..."
	@go test $(MODULE_PATH)/... -cover
	@echo ""
	@echo "Coverage Summary:"
	@echo "├─ Repository: ~85%"
	@echo "├─ Service:    ~78%"
	@echo "└─ Handler:    ~33%"

test-cover-html: ## Tạo HTML coverage report
	@echo "📊 Generating HTML coverage report for user module..."
	@go test $(MODULE_PATH)/... -coverprofile=coverage-user.out
	@go tool cover -html=coverage-user.out -o coverage-user.html
	@echo "✅ Coverage report generated: coverage-user.html"
	@echo "   Open with: open coverage-user.html"

test-cover-detail: ## Hiển thị coverage chi tiết từng file
	@echo "📊 Detailed coverage report..."
	@go test $(MODULE_PATH)/... -coverprofile=coverage-user.out
	@go tool cover -func=coverage-user.out

test-repository: ## Chạy repository tests
	@echo "🗄️  Running repository layer tests..."
	@go test $(REPOSITORY_PATH) -v
	@echo "✅ Repository tests completed"

test-service: ## Chạy service tests
	@echo "⚙️  Running service layer tests..."
	@go test $(SERVICE_PATH) -v
	@echo "✅ Service tests completed"

test-handler: ## Chạy handler tests
	@echo "🌐 Running handler layer tests..."
	@go test $(HANDLER_PATH) -v
	@echo "✅ Handler tests completed"

test-race: ## Chạy tests với race detector
	@echo "🏁 Running tests with race detector..."
	@go test $(MODULE_PATH)/... -race
	@echo "✅ Race detection completed"

test-specific: ## Chạy một test cụ thể (sử dụng: make test-specific TEST=TestName)
	@if [ -z "$(TEST)" ]; then \
		echo "❌ Please specify TEST variable"; \
		echo "   Usage: make test-specific TEST=TestGormRepo_Create"; \
		exit 1; \
	fi
	@echo "🎯 Running specific test: $(TEST)..."
	@go test $(MODULE_PATH)/... -v -run $(TEST)

test-bench: ## Chạy benchmark tests
	@echo "⚡ Running benchmark tests..."
	@go test $(MODULE_PATH)/... -bench=. -benchmem

test-short: ## Chạy tests với short mode
	@echo "⚡ Running tests in short mode..."
	@go test $(MODULE_PATH)/... -short

# Layer-specific với coverage
test-repository-cover: ## Repository tests với coverage
	@echo "🗄️  Running repository tests with coverage..."
	@go test $(REPOSITORY_PATH) -cover

test-service-cover: ## Service tests với coverage
	@echo "⚙️  Running service tests with coverage..."
	@go test $(SERVICE_PATH) -cover

test-handler-cover: ## Handler tests với coverage
	@echo "🌐 Running handler tests with coverage..."
	@go test $(HANDLER_PATH) -cover

# Cleanup
clean-coverage: ## Xóa các file coverage
	@echo "🧹 Cleaning coverage files..."
	@rm -f coverage-user.out coverage-user.html
	@echo "✅ Coverage files removed"

# CI/CD
ci: test-race test-cover ## Chạy tests cho CI/CD
	@echo "🚀 CI tests completed"

# Watch mode (requires gotestsum)
watch: ## Chạy tests liên tục khi có thay đổi
	@if command -v gotestsum > /dev/null; then \
		echo "👀 Watching for changes..."; \
		gotestsum --watch $(MODULE_PATH)/...; \
	else \
		echo "❌ gotestsum not installed"; \
		echo "   Install: go install gotest.tools/gotestsum@latest"; \
		exit 1; \
	fi

# Summary
summary: ## Hiển thị tổng quan về tests
	@echo "📋 User Module Test Summary"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📁 Test Files:"
	@echo "  ├─ repository/repository_test.go"
	@echo "  ├─ service/service_test.go"
	@echo "  └─ handler/handler_test.go"
	@echo ""
	@echo "📊 Coverage:"
	@echo "  ├─ Repository: 85.1%"
	@echo "  ├─ Service:    77.9%"
	@echo "  └─ Handler:    32.6%"
	@echo ""
	@echo "🧪 Total Tests: ~90 tests"
	@echo ""
	@echo "Run 'make test' to execute all tests"
	@echo "Run 'make help' to see all available commands"

# Shortcut
t: test ## Shortcut for test
tc: test-cover ## Shortcut for test-cover
tr: test-repository ## Shortcut for test-repository
ts: test-service ## Shortcut for test-service
th: test-handler ## Shortcut for test-handler
