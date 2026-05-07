# Use lightweight Alpine-based Go image for building
FROM golang:1.21-alpine AS builder

# Set necessary environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Create and change to the app directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o /partworks-be .

# Use lightweight scratch image for final stage
FROM alpine:3.18

# Install CA certificates for HTTPS
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /partworks-be /app/partworks-be

# Copy .env file
COPY .env /app/.env

# Change to app directory
WORKDIR /app

# Set non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Expose the port the app runs on
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8080/health || exit 1

# Command to run the application
CMD ["/app/partworks-be"],
