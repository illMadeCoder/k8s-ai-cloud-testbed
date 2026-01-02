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

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/station-monitor .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./station-monitor"]
