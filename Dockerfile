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
    -ldflags="-w -s -X 'custodian/internal/cmd.version=${VERSION}' -X 'custodian/internal/cmd.commit=${COMMIT}' -X 'custodian/internal/cmd.date=${DATE}'" \
    -o custodian \
    ./cmd/custodian

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates git openssh-client

# Create non-root user
RUN addgroup -g 1001 custodian && \
    adduser -D -s /bin/sh -u 1001 -G custodian custodian

# Copy the binary
COPY --from=builder /app/custodian /usr/local/bin/custodian

# Set permissions
RUN chmod +x /usr/local/bin/custodian

# Create directories for templates and output
RUN mkdir -p /app/templates /app/output && \
    chown -R custodian:custodian /app

USER custodian
WORKDIR /app

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD custodian --version || exit 1

ENTRYPOINT ["custodian"]