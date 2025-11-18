# ==================================================
# Stage 1: Builder
# ==================================================
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev \
    ca-certificates \
    tzdata

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a \
    -installsuffix cgo \
    -o app \
    ./cmd/server

# Verify the binary
RUN chmod +x /build/app

# ==================================================
# Stage 2: Runtime
# ==================================================
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl

# Set timezone
ENV TZ=Asia/Ho_Chi_Minh

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/app .

# Copy docs (Swagger)
COPY --from=builder /build/docs ./docs

# Copy migrations (optional, if you want to run migrations from container)
COPY --from=builder /build/migrations ./migrations

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./app"]
