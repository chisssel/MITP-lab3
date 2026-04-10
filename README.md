# Лабораторная работа №11. Контейнеризация мультиязычных приложений
**Студент:** *Платов Артем Русланович*\
**Группа:** *220032-11*\
**Вариант:** *6*\
**Сложность:** *Средняя*
---

## Task6. Настроить сеть между контейнерами.

Контейнеризованное приложение с API Gateway на Go, сервисом пользователей на Python и сервисом статистики на Rust.

## Архитектура

```
┌─────────────┐     ┌───────────────┐     ┌─────────────────┐
│   Client    │────▶│  Go Gateway   │────▶│ Python Service  │
│  :8080      │     │   :8080      │     │    :5000        │
└─────────────┘     └───────────────┘     │  (/api/users)    │
                           │             └─────────────────┘
                           │
                           ▼             ┌─────────────────┐
                           │             │  Rust Service   │
                           └────────────▶│    :4000        │
                                         │  (/api/stats)   │
                                         └─────────────────┘
```

## Структура проекта

```
├── go-server/           # API Gateway (Go)
│   ├── main.go
│   ├── main_test.go
│   └── Dockerfile
├── python-service/      # Users API (Python/Flask)
│   ├── app.py
│   ├── test_app.py
│   └── Dockerfile
├── rust-service/        # Stats API (Rust/tiny_http)
│   ├── src/main.rs
│   ├── Cargo.toml
│   └── Dockerfile
└── docker-compose.yml
```

## Запуск

### Сборка и запуск всех сервисов:
```bash
docker-compose up --build -d
```

### Просмотр логов:
```bash
docker-compose logs -f
```

### Остановка:
```bash
docker-compose down
```

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### Users (Python)
```bash
# Список пользователей
curl http://localhost:8080/api/users/
# [{"email":"alice@example.com","id":1,"name":"Alice"},...]

# Конкретный пользователь
curl http://localhost:8080/api/users/1/
# {"email":"alice@example.com","id":1,"name":"Alice"}
```

### Stats (Rust)
```bash
curl http://localhost:8080/api/stats
# {"average_per_second":0.01,"total_requests":5,"uptime_seconds":360}
```

## Размеры образов

| Сервис | Размер | База |
|--------|--------|------|
| Go Gateway | ~5 MB | scratch |
| Python Service | ~132 MB | python:3.12-slim |
| Rust Service | ~89 MB | debian:bookworm-slim |

## Docker Multi-Stage Build

### Go (статическая компиляция → scratch)
```dockerfile
FROM golang:1.22-alpine AS builder
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .
FROM scratch
COPY --from=builder /app/server /server
```

### Rust (musl target → debian:bookworm-slim)
```dockerfile
FROM rust:1.75-alpine AS builder
RUN rustup target add x86_64-unknown-linux-musl
RUN cargo build --release --target x86_64-unknown-linux-musl
FROM debian:bookworm-slim
COPY --from=builder /app/target/release/rust-stats /usr/local/bin/rust-stats
```

### Python (slim base)
```dockerfile
FROM python:3.12-slim
RUN pip install --no-cache-dir -r requirements.txt
```

## Тестирование

### Go tests (table-driven)
```bash
cd go-server && go test -v ./...
```

### Python tests (parameterized)
```bash
cd python-service && python -m pytest test_app.py -v
```

### Rust tests (unit tests)
```bash
cd rust-service && cargo test
```

## Сеть Docker

Контейнеры общаются через внутреннюю сеть `microservices-network`:
- `go-gateway` → `python-service:5000`
- `go-gateway` → `rust-service:4000`

Внешние порты:
- `8080` - Go Gateway
- `5000` - Python Service (внутренний)
- `4000` - Rust Service (внутренний)