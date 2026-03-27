#######################################################
# Stage 1: Build binary
#######################################################
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev \
    ca-certificates \
    tzdata

WORKDIR /build

# Copy go.mod và go.sum trước để cache dependency
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy toàn bộ source vào container build
COPY . .

# Build binary tối ưu cho production
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o app \
    ./cmd/server

#######################################################
# Stage 2: Runtime image
#######################################################
FROM alpine:3.19

RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl

# Thiết lập timezone
ENV TZ=Asia/Ho_Chi_Minh

# Tạo user không phải root
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary và tài nguyên cần thiết
COPY --from=builder /build/app .
COPY --from=builder /build/docs ./docs

# Phân quyền cho user không phải root
RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

# Healthcheck đơn giản (có thể chỉnh route cho đúng với app)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["./app"]

