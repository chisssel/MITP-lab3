# Build stage
FROM rust:1.75-alpine AS builder
WORKDIR /app

RUN apk add --no-cache musl-dev

COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release && rm -rf src

COPY src ./src
RUN cargo build --release

# Final stage - debian:bookworm-slim
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/rust-stats /usr/local/bin/rust-stats

EXPOSE 4000

CMD ["rust-stats"]
