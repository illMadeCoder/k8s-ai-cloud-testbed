# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY src/go.mod src/go.sum* ./

# Download dependencies
RUN go mod download

# Copy source
COPY src/*.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o log-generator .

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /app/log-generator .

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/log-generator"]
