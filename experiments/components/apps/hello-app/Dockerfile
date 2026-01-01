FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY src/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server main.go

FROM alpine:3.23
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
ENTRYPOINT ["./server"]
