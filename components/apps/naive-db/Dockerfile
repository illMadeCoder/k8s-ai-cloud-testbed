# Build stage
FROM rust:1.84-alpine AS builder

RUN apk add --no-cache musl-dev

WORKDIR /app

# Dependency caching: build deps with dummy main first
COPY src/Cargo.toml src/Cargo.lock* ./
RUN mkdir src && echo 'fn main() {}' > src/main.rs
RUN cargo build --release
RUN rm -rf src

# Copy real source and rebuild
COPY src/src/ src/
RUN touch src/main.rs && cargo build --release

# Runtime stage
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/target/release/naive-db /naive-db

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/naive-db"]
