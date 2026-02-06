FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o order-service .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -D appuser
WORKDIR /app
COPY --from=builder /app/order-service .
RUN chown -R appuser:appgroup /app
USER 1000
EXPOSE 8080
CMD ["./order-service"]
