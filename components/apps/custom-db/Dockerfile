# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY src/go.mod src/go.sum* ./
RUN go mod download

COPY src/*.go ./

ARG VERSION=dev
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o custom-db .

# Runtime stage
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/custom-db /app/custom-db

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/app/custom-db"]
