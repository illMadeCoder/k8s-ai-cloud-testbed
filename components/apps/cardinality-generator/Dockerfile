# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git

# Copy go.mod first for layer caching
COPY src/go.mod src/go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY src/*.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o cardinality-generator .

# Runtime stage - use distroless for minimal attack surface
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/cardinality-generator /app/cardinality-generator

# Run as non-root user
USER nonroot:nonroot

# Expose the application port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/cardinality-generator"]
