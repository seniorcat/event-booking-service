# booking-service-go

Учебный backend-сервис для бронирования событий (Go + PostgreSQL, sqlx, DI, миграции).

## Цели проекта
- Практика проектирования слоёв: model → repository → service → handler
- Работа с PostgreSQL через `sqlx` и SQL-миграции
- Добавление тестов (unit/интеграционные) и инфраструктуры (Docker, docker-compose)

## Структура проекта
- `cmd/server` — точка входа (запуск HTTP, опциональный прогон миграций)
- `internal/config` — конфиги и DI-обёртка
- `internal/db` — провайдер подключения к БД и раннер миграций (goose)
- `internal/http` — роутер и HTTP-хендлеры
- `internal/event` — домен Event (модель, репозиторий, сервис, DI)
- `internal/booking` — домен Booking (модель, репозиторий, сервис, DI)
- `internal/user` — домен User (модель, заглушки)
- `deploy/migrations` — SQL-миграции
- `deploy/local/docker-compose.yaml` — локальный PostgreSQL
- `pkg/container` — простой DI-контейнер на базе `sarulabs/di`

## Подготовка окружения
1) Поднять PostgreSQL локально:
```bash
docker compose -f deploy/local/docker-compose.yaml up -d
```

2) Настроить конфиг (пример): `deploy/local/config.dev.yaml`
```yaml
server:
  port: 8080
  read_timeout: 5s
  write_timeout: 5s
database:
  dsn: "postgres://user:password@localhost:5432/eventdb?sslmode=disable"
  auto_migrate: true
```

3) Запуск сервера:
```bash
CONFIG_PATH=deploy/local/config.dev.yaml go run ./cmd/server
```

## Миграции
- SQL-файлы лежат в `deploy/migrations` (формат goose).
- При `database.auto_migrate: true` миграции применяются автоматически при старте.

## HTTP эндпоинты (минимум)
- `GET  /ping` → "pong"
- `GET  /health` → проверка подключения к БД

### Events
- `GET    /events` — список
- `POST   /events` — создать
- `GET    /events/{id}` — получить
- `PUT    /events/{id}` — обновить
- `DELETE /events/{id}` — удалить

Пример создания события:
```bash
curl -sS -X POST :8080/events -H 'Content-Type: application/json' \
  -d '{
    "title":"Go Meetup",
    "description":"Intro",
    "location":"Online",
    "starts_at":"2025-10-01T12:00:00Z",
    "ends_at":"2025-10-01T14:00:00Z",
    "capacity":100
  }'
```

### Bookings
- `POST   /bookings` — создать (проверяется вместимость события)
- `GET    /bookings/{id}` — получить
- `GET    /events/{id}/bookings` — список бронирований по событию
- `DELETE /bookings/{id}` — отменить (status → cancelled)

## Тесты
Запуск всех тестов:
```bash
go test ./...
```
В репозитории есть простые примеры unit-тестов для сервисов и один тест для хендлера.

## Дорожная карта (для ученика)
Ниже список задач для последовательной проработки (каждая — отдельная ветка и PR):

1) Аутентификация и пользователи
- Реализовать `POST /users/register`: валидация, хэш пароля (bcrypt), уникальный email
- Реализовать `POST /users/login`: проверка пароля, выдача JWT
- Middleware аутентификации для защищённых эндпоинтов

2) Валидация и обработка ошибок
- Единый формат ошибок JSON: `{ "status":"error", "message":"..." }`
- Валидация email, дат событий и capacity в хендлерах

3) Логирование и middleware
- Внедрить структурированный логгер (`log/slog` или `zap`) через DI
- Middleware: `recover`, `request-id`, `request-logging`

4) Интеграционные тесты
- Поднятие тестовой БД в Docker, прогоn миграций в тестах
- Интеграционные тесты для `/events` и `/bookings` (CRUD-сценарии)

5) UUID-ключи
- Перевести `users`, `events`, `bookings` на UUID (pgcrypto/uuid-ossp)
- Обновить модели/репозитории/миграции и тесты

6) Кэширование (Redis)
- Добавить Redis в docker-compose
- Кэшировать `GET /events` и `GET /events/{id}/bookings` с инвалидацией при изменениях

7) gRPC API
- Описать `BookingService` в `.proto` (CreateBooking, CancelBooking, ListBookings)
- Поднять gRPC-сервер, реализовать методы, добавить интеграционные тесты

8) Контейнеризация приложения
- Multi-stage Dockerfile для сборки бинаря
- Добавить сервис приложения в `deploy/local/docker-compose.yaml`, зависимость от Postgres

9) Метрики и наблюдаемость
- Экспорт Prometheus-метрик: latency, RPS, ошибки
- Хендлер `/metrics` и базовые графики

10) CI
- GitHub Actions: `go vet`, `golangci-lint`, тесты, сборка

## Полезные команды
```bash
# поднять локальную БД
docker compose -f deploy/local/docker-compose.yaml up -d

# запустить сервис с выбранным конфигом
CONFIG_PATH=deploy/local/config.dev.yaml go run ./cmd/server

# запустить тесты
go test ./...
```
