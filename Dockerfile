# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Build arguments for versioning
ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_TIME=unknown

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with version info
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-X github.com/rally-go/rally-realtime/internal/version.Version=${VERSION} \
              -X github.com/rally-go/rally-realtime/internal/version.CommitSHA=${COMMIT_SHA} \
              -X github.com/rally-go/rally-realtime/internal/version.BuildTime=${BUILD_TIME}" \
    -o server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/server .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server
CMD ["./server"]
