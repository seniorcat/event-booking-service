# booking-service-go

Учебный backend-сервис для бронирования событий (Go + PostgreSQL).

## Цели проекта
- Практика работы с Go без ORM (pgx).
- Освоение структуры проекта с разделением по доменам (user, event, booking).
- Работа с GitHub: ветки, PR, code review.

## Структура проекта
- `cmd/server` — точка входа
- `internal/db` — работа с PostgreSQL
- `internal/user` — домен пользователя
- `internal/event` — домен событий
- `internal/booking` — домен бронирований
- `pkg/` — вспомогательные утилиты

## Быстрый старт
```bash
# клонировать репозиторий
git clone git@github.com:<username>/booking-service-go.git
cd booking-service-go

# запустить сервер (stub)
go run ./cmd/server
