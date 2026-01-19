# Build stage
FROM golang:1.21-alpine AS builder
# Install build dependencies
RUN apk add --no-cache git ca-certificates
# Set working directory
WORKDIR /build
# Copy dependency definitions
COPY go.mod go.sum ./
# Download dependencies with caching
RUN go mod download
# Copy application source
COPY . .
# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o webserver \
    ./cmd/webserver
# Runtime stage
FROM scratch
# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy binary from build stage
COPY --from=builder /build/webserver /webserver
# Expose application port
EXPOSE 8080
# Set non-root user
USER 65534:65534
# Define startup command
ENTRYPOINT ["/webserver"]
