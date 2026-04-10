# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN go mod init go-server
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .

# Final stage - debian:bookworm-slim (для healthcheck с curl)
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends curl && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
