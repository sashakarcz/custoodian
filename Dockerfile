# Build stage
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Install buf
RUN go install github.com/bufbuild/buf/cmd/buf@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate protobuf code
RUN buf generate

# Build arguments
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X 'custoodian/internal/cmd.version=${VERSION}' -X 'custoodian/internal/cmd.commit=${COMMIT}' -X 'custoodian/internal/cmd.date=${DATE}'" \
    -o custoodian \
    ./cmd/custoodian

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates git openssh-client

# Create non-root user
RUN addgroup -g 1001 custoodian && \
    adduser -D -s /bin/sh -u 1001 -G custoodian custoodian

# Copy the binary
COPY --from=builder /app/custoodian /usr/local/bin/custoodian

# Set permissions
RUN chmod +x /usr/local/bin/custoodian

# Create directories for templates and output
RUN mkdir -p /app/templates /app/output && \
    chown -R custoodian:custoodian /app

USER custoodian
WORKDIR /app

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD custoodian --version || exit 1

ENTRYPOINT ["custoodian"]