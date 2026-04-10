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
---

## Task8. Добавить healthcheck для каждого сервиса.

Добавление healthcheck для всех сервисов в Docker Compose.

## Архитектура

```
┌─────────────┐     ┌───────────────┐     ┌─────────────────┐
│   Client    │────▶│  Go Gateway   │────▶│ Python Service  │
│  :8080      │     │   :8080      │     │    :5000        │
└─────────────┘     │  (healthy)    │     │  (healthy)      │
                    └───────────────┘     └─────────────────┘
                           │
                           ▼
                    ┌─────────────────┐
                    │  Rust Service   │
                    │    :4000        │
                    │  (healthy)      │
                    └─────────────────┘
```

## Структура проекта

```
├── go-server/           # API Gateway (Go)
│   ├── main.go
│   ├── main_test.go
│   └── Dockerfile       # debian:bookworm-slim (для curl)
├── python-service/      # Users API (Python/Flask)
│   ├── app.py
│   ├── test_app.py
│   └── Dockerfile       # python:3.12-slim + curl
├── rust-service/        # Stats API (Rust)
│   ├── src/main.rs
│   ├── Cargo.toml
│   └── Dockerfile       # debian:bookworm-slim + curl
└── docker-compose.yml   # healthcheck для всех сервисов
```

## Healthcheck в Docker Compose

```yaml
services:
  go-gateway:
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 5s
    depends_on:
      python-service:
        condition: service_healthy
      rust-service:
        condition: service_healthy
```

### Параметры healthcheck:

| Параметр | Описание |
|----------|----------|
| `test` | Команда проверки здоровья |
| `interval` | Интервал между проверками (30s) |
| `timeout` | Таймаут ожидания ответа (10s) |
| `retries` | Количество попыток перед неудачей (3) |
| `start_period` | Период инициализации (5s) |

## Запуск

```bash
docker-compose up --build -d
```

## Проверка статуса

```bash
# Список всех контейнеров
docker-compose ps

# Проверка здоровья конкретного сервиса
docker inspect go-gateway --format='{{.State.Health.Status}}'

# Подробная информация о healthcheck
docker inspect go-gateway | Select-String Health
```

## Зависимости между сервисами

Благодаря `condition: service_healthy`:
- **go-gateway** запускается только после того, как `python-service` и `rust-service` станут healthy
- Это гарантирует доступность бэкенд-сервисов до старта API Gateway

## Health Endpoints

| Сервис | Endpoint | Ответ |
|--------|----------|-------|
| Go Gateway | `http://localhost:8080/health` | `{"status":"ok"}` |
| Python Service | `http://localhost:5000/health` | `{"status":"ok"}` |
| Rust Service | `http://localhost:4000/health` | `{"status":"ok"}` |

## API Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Users (через Python)
curl http://localhost:8080/api/users/

# Stats (через Rust)
curl http://localhost:8080/api/stats
```

## Сборка Docker образов с curl

### Go (debian вместо scratch)
```dockerfile
FROM golang:1.22-alpine AS builder
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends curl
COPY --from=builder /app/server /server
CMD ["/server"]
```

### Python
```dockerfile
FROM python:3.12-slim
RUN apt-get update && apt-get install -y --no-install-recommends curl
```

### Rust
```dockerfile
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl
```

## Статусы контейнеров

```
NAME           STATUS
go-gateway     healthy / unhealthy / starting
python-users   healthy / unhealthy / starting
rust-stats     healthy / unhealthy / starting
```

## Тестирование

### Запуск всех тестов

```bash
# Go
cd go-server && go test -v ./...

# Python
cd python-service && pytest test_app.py -v

# Rust
cd rust-service && cargo test
```

### Сводка по тестам

| Сервис | Тестов | Методология |
|--------|--------|-------------|
| Go Gateway | 33 | Table-driven |
| Python Service | 38 | Parameterized |
| Rust Service | 23 | Unit tests |
| **Итого** | **94** | - |

### Go Tests (table-driven)

**Файл:** `go-server/main_test.go`

| Тест | Описание |
|------|----------|
| `TestHealthEndpoint` | Базовые проверки health endpoint |
| `TestRoutePatterns` | Проверка маршрутизации |
| `TestProxyPathMapping` | Проверка маппинга путей |
| `TestHTTPMethods` | Проверка HTTP методов |
| `TestContentType` | Проверка Content-Type |
| `TestPathPrefixMatching` | Проверка совпадения префиксов |
| `TestHealthCheckEndpoint` | Расширенные проверки health |
| `TestHealthCheckResponseFormat` | Формат JSON ответа |
| `TestHealthCheckCurlCompatible` | Совместимость с curl |

### Python Tests (parameterized)

**Файл:** `python-service/test_app.py`

| Тест | Описание |
|------|----------|
| `test_get_endpoints` | Проверка доступных endpoints |
| `test_content_type` | Проверка Content-Type |
| `test_response_json_keys` | Проверка ключей JSON |
| `test_get_user_by_id` | Получение пользователя по ID |
| `test_user_data` | Проверка данных пользователя |
| `test_methods` | Проверка HTTP методов |
| `test_health_returns_ok` | Проверка health |
| `TestHealthCheck` | Класс тестов healthcheck |
| `TestHealthCheckDockerCompose` | Docker Compose тесты |

### Rust Tests (unit tests)

**Файл:** `rust-service/src/main.rs`

| Тест | Описание |
|------|----------|
| `test_app_state_new` | Создание нового состояния |
| `test_app_state_increment` | Инкремент счётчика |
| `test_healthcheck_response_format` | Формат ответа health |
| `test_healthcheck_is_valid_json` | Валидный JSON |
| `test_healthcheck_json_structure` | Структура JSON |
| `test_healthcheck_status_value` | Значение status |
| `test_healthcheck_curl_compatible` | Совместимость с curl |
| `test_healthcheck_no_extra_fields` | Без лишних полей |
| `test_health_endpoint_in_routes` | Маршрутизация |
| `test_health_response_immutable` | Иммутабельность |
| `test_health_response_not_affected_by_state` | Независимость от state |

### Тестирование в Docker

Для Python доступен отдельный Dockerfile с тестами:

```bash
docker build -f python-service/Dockerfile.test -t python-service-test python-service
```

### Ожидаемые результаты

```
Go:     ok   33 tests passed
Python: 38 passed in 0.xxs
Rust:   ok. 23 passed
```
---

## Task10. Использовать переменные окружения для конфигурации.

Конфигурирование микросервисов через переменные окружения для гибкого развертывания.

## Архитектура

```
┌─────────────┐     ┌───────────────┐     ┌─────────────────┐
│   Client    │────▶│  Go Gateway   │────▶│ Python Service  │
│  :8080      │     │   :8080      │     │    :5000        │
└─────────────┘     └───────────────┘     └─────────────────┘
                           │
                           ▼
              ┌─────────────────┐
              │  Rust Service   │
              │    :4000        │
              └─────────────────┘
```

## Переменные окружения

### Go Gateway

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `GATEWAY_PORT` | Порт HTTP сервера | `8080` |
| `PYTHON_SERVICE_URL` | URL Python сервиса | `http://python-service:5000` |
| `RUST_SERVICE_URL` | URL Rust сервиса | `http://rust-service:4000` |
| `USERS_PATH` | Путь к users API | `/api/users` |
| `STATS_PATH` | Путь к stats API | `/api/stats` |

### Python Service

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `FLASK_HOST` | Хост для привязки | `0.0.0.0` |
| `FLASK_PORT` | Порт HTTP сервера | `5000` |
| `FLASK_DEBUG` | Режим отладки | `false` |
| `SERVICE_NAME` | Имя сервиса | `python-users` |
| `SERVICE_VERSION` | Версия сервиса | `1.0` |

### Rust Service

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `RUST_HOST` | Хост для привязки | `0.0.0.0` |
| `RUST_PORT` | Порт HTTP сервера | `4000` |
| `SERVICE_NAME` | Имя сервиса | `rust-stats` |
| `SERVICE_VERSION` | Версия сервиса | `1.0` |

## Структура проекта

```
├── go-server/           # API Gateway (Go)
│   ├── main.go          # Переменные окружения через os.Getenv
│   ├── main_test.go
│   └── Dockerfile
├── python-service/      # Users API (Python/Flask)
│   ├── app.py           # Переменные окружения через os.environ.get
│   ├── test_app.py
│   ├── requirements.txt
│   └── Dockerfile
├── rust-service/        # Stats API (Rust)
│   ├── src/main.rs      # Переменные окружения через env::var
│   ├── Cargo.toml
│   └── Dockerfile
├── .env                 # Файл с переменными окружения
└── docker-compose.yml  # Передача переменных в контейнеры
```

## Файл .env

```bash
# Go Gateway
GATEWAY_PORT=8080
PYTHON_SERVICE_URL=http://python-service:5000
RUST_SERVICE_URL=http://rust-service:4000

# Python Service
FLASK_HOST=0.0.0.0
FLASK_PORT=5000
FLASK_DEBUG=false
SERVICE_NAME=python-users
SERVICE_VERSION=1.0

# Rust Service
RUST_HOST=0.0.0.0
RUST_PORT=4000
RUST_SERVICE_NAME=rust-stats
RUST_SERVICE_VERSION=1.0
```

## Запуск

### Сборка и запуск:
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
# {"status":"ok","service":"gateway"}
```

### Users (Python)
```bash
curl http://localhost:8080/api/users/
# [{"email":"alice@example.com","id":1,"name":"Alice"},...]

curl http://localhost:8080/api/users/1/
# {"email":"alice@example.com","id":1,"name":"Alice"}
```

### Stats (Rust)
```bash
curl http://localhost:8080/api/stats
# {"average_per_second":0.01,"total_requests":5,"uptime_seconds":360}
```

## Тестирование

### Запуск всех тестов:

```bash
# Go
cd go-server && go test -v ./...

# Python
cd python-service && python -m pytest test_app.py -v

# Rust
cd rust-service && cargo test
```

### Сводка по тестам:

| Сервис | Тестов | Методология |
|--------|--------|-------------|
| Go Gateway | 41 | Table-driven |
| Python Service | 32 | Parameterized |
| Rust Service | 31 | Unit tests |
| **Итого** | **104** | - |

### Примеры Go тестов (environment variables)

```go
func TestGetEnv(t *testing.T) {
    result := getEnv("TEST_KEY", "default")
    if result != "default" {
        t.Errorf("Expected 'default', got '%s'", result)
    }
}

func TestGetEnvInt(t *testing.T) {
    result := getEnvInt("TEST_PORT", 8080)
    if result != 8080 {
        t.Errorf("Expected 8080, got %d", result)
    }
}
```

### Примеры Rust тестов (environment variables)

```rust
#[test]
fn test_get_env_with_default() {
    let value = get_env("NONEXISTENT", "default");
    assert_eq!(value, "default");
}

#[test]
fn test_get_env_u16_with_default() {
    let value = get_env_u16("NONEXISTENT", 4000);
    assert_eq!(value, 4000);
}
```

## Конфигурация в Docker Compose

```yaml
services:
  go-gateway:
    image: go-gateway
    ports:
      - "${GATEWAY_PORT:-8080}:8080"
    environment:
      - GATEWAY_PORT=${GATEWAY_PORT:-8080}
      - PYTHON_SERVICE_URL=${PYTHON_SERVICE_URL}
      - RUST_SERVICE_URL=${RUST_SERVICE_URL}
    depends_on:
      python-service:
        condition: service_healthy
      rust-service:
        condition: service_healthy

  python-service:
    image: python-service
    environment:
      - FLASK_HOST=${FLASK_HOST:-0.0.0.0}
      - FLASK_PORT=${FLASK_PORT:-5000}
      - FLASK_DEBUG=${FLASK_DEBUG:-false}
      - SERVICE_NAME=${SERVICE_NAME:-python-users}
      - SERVICE_VERSION=${SERVICE_VERSION:-1.0}

  rust-service:
    image: rust-service
    environment:
      - RUST_HOST=${RUST_HOST:-0.0.0.0}
      - RUST_PORT=${RUST_PORT:-4000}
      - SERVICE_NAME=${RUST_SERVICE_NAME:-rust-stats}
      - SERVICE_VERSION=${RUST_SERVICE_VERSION:-1.0}
```

## Преимущества использования переменных окружения

1. **Гибкость** - легко менять конфигурацию без перекомпиляции
2. **Безопасность** - секреты не хранятся в коде
3. **Переносимость** - один код для разных окружений (dev, staging, prod)
4. **12-Factor App** - соответствие методологии облачных приложений
5. **Docker** - нативная интеграция с контейнеризацией

## Размеры образов

| Сервис | Размер | База |
|--------|--------|------|
| Go Gateway | ~5 MB | scratch |
| Python Service | ~132 MB | python:3.12-slim |
| Rust Service | ~89 MB | debian:bookworm-slim |
