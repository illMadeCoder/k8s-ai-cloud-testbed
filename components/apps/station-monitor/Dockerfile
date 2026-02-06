# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY src/go.mod src/go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY src/*.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o station-monitor .

# Runtime stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 65532 -S appgroup && \
    adduser -u 65532 -S appuser -G appgroup

WORKDIR /app

# Copy the binary from builder with proper ownership
COPY --from=builder --chown=65532:65532 /app/station-monitor .

# Ensure executable
RUN chmod +x /app/station-monitor

USER 65532:65532

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/station-monitor"]
