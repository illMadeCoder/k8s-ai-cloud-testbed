# Build stage
FROM golang:1.22-alpine AS builder

ARG VERSION=dev
ARG BUILD_TIME=unknown

WORKDIR /app
COPY src/ .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o /app/server .

# Runtime stage - using distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080

ENTRYPOINT ["./server"]
