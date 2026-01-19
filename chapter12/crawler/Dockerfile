# Build stage
FROM golang:1.21-alpine AS builder
# Install build dependencies including SSH client libraries
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    openssh-client \
    gcc \
    musl-dev
WORKDIR /build
# Copy dependency files
COPY go.mod go.sum ./
# Download dependencies with verification
RUN go mod download && go mod verify
# Copy source code
COPY . .
# Run unit tests
RUN go test -short -v ./...
# Build crawler binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static" -X main.Version=1.0.0 -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)' \
    -tags netgo \
    -o crawler \
    ./cmd/crawler
# Runtime stage
FROM alpine:3.19
# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    openssh-client \
    bash \
    curl

# Create application user and directories
RUN addgroup -g 1000 crawler && \
    adduser -D -u 1000 -G crawler crawler && \
    mkdir -p /app/data /app/logs /app/cache /app/configs && \
    mkdir -p /home/crawler/.ssh && \
    chmod 700 /home/crawler/.ssh
WORKDIR /app
# Copy binary from builder
COPY --from=builder /build/crawler .
# Copy default configuration
COPY --from=builder /build/configs/crawler.yaml ./configs/
# Copy templates for vendor-specific parsing
COPY --from=builder /build/templates ./templates/
# Set ownership
RUN chown -R crawler:crawler /app /home/crawler
# Switch to non-root user
USER crawler
# Expose metrics port
EXPOSE 9090
# Health check with longer timeout for crawler operations
HEALTHCHECK --interval=60s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:9090/metrics || exit 1
# Environment configuration
ENV CRAWLER_DATA_DIR=/app/data \
    CRAWLER_LOG_LEVEL=info \
    CRAWLER_CACHE_DIR=/app/cache \
    CRAWLER_METRICS_PORT=9090 \
    CRAWLER_CONCURRENT_DEVICES=10
# Volumes for persistent data
VOLUME ["/app/data", "/app/cache", "/home/crawler/.ssh"]
# Labels for metadata
LABEL maintainer="infrastructure-team@example.com" \
      version="1.0.0" \
      description="Infrastructure inventory and configuration crawler"
# Default command with configuration file
ENTRYPOINT ["/app/crawler"]
CMD ["run", "--config", "/app/configs/crawler.yaml"]
