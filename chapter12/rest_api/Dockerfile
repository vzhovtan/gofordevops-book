# Build stage
FROM golang:1.21-alpine AS builder
# Install build essentials
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev
WORKDIR /build
# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify
# Copy application source
COPY . .
# Run tests during build
RUN go test -v ./...
# Build optimized binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -tags netgo \
    -installsuffix netgo \
    -o api-server \
    ./cmd/api-server
# Runtime stage
FROM alpine:3.19
# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata
# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser
WORKDIR /app
# Copy binary and configuration
COPY --from=builder /build/api-server .
COPY --from=builder /build/configs ./configs
# Create directories for runtime data
RUN mkdir -p /app/data /app/logs && \
    chown -R appuser:appgroup /app
# Switch to non-root user
USER appuser
# Expose API port
EXPOSE 8080
# Health check endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
# Environment variables with defaults
ENV API_PORT=8080 \
    LOG_LEVEL=info \
    DATA_DIR=/app/data

# Volume for persistent data
VOLUME ["/app/data"]

# Start application
ENTRYPOINT ["/app/api-server"]
CMD ["--config", "/app/configs/production.yaml"]
