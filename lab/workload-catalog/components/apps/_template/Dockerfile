# Example multi-stage Dockerfile for a Go application
# Modify as needed for your language/framework

# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY src/ .

# If using Go modules
# RUN go mod download
# RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# For simple single-file apps
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server main.go

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

ENTRYPOINT ["./server"]
