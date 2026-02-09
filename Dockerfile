# ===== Build Stage =====
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server main.go

# ===== Runtime Stage =====
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS and tzdata for timezone
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/server .

# Copy static files
COPY --from=builder /app/static ./static

# Copy migrations (for reference, migrations run from code)
COPY --from=builder /app/migrations ./migrations

# Expose port (Render sets PORT env var)
EXPOSE 8080

# Run the binary
CMD ["./server"]
