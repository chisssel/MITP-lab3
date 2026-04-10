# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN go mod init go-server
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .

# Final stage - scratch (smallest possible image)
FROM scratch
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
