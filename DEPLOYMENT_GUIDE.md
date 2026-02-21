# Инструкция по развёртыванию проекта Event Booking Service

## Содержание
1. [Требования к окружению](#требования-к-окружению)
2. [Структура docker-compose](#структура-docker-compose)
3. [Подготовка и сборка](#подготовка-и-сборка)
4. [Запуск локального окружения](#запуск-локального-окружения)
5. [Проверка работоспособности](#проверка-работоспособности)
6. [Работа с миграциями](#работа-с-миграциями)
7. [Запуск тестов](#запуск-тестов)
8. [Просмотр логов](#просмотр-логов)
9. [Остановка и очистка](#остановка-и-очистка)
10. [Диагностика проблем](#диагностика-проблем)

---

## Требования к окружению

Для развёртывания проекта необходимы следующие инструменты:

- **Docker** версии 20.10 или выше
- **Docker Compose** версии 2.0 или выше
- **Git** (для клонирования репозитория)
- **curl** или **wget** (для проверки эндпоинтов)

### Проверка версии Docker
```bash
docker --version
# Ожидаемый вывод: Docker version 20.10.x или выше
```

### Проверка версии Docker Compose
```bash
docker compose version
# Ожидаемый вывод: Docker Compose version v2.x.x или выше
```

---

## Структура docker-compose

Проект использует docker-compose для оркестрации следующих сервисов:

| Сервис | Описание | Порты | Назначение |
|--------|----------|-------|------------|
| **postgres** | База данных PostgreSQL 16 | `5434:5432` | Хранение данных приложения |
| **redis** | Кэширующий слой Redis | `6379:6379` | Кэширование событий и бронирований |
| **booking-service** | Основное приложение | `8080:8080` (HTTP)<br>`50051:50051` (gRPC) | REST API и gRPC сервер |
| **prometheus** | Система мониторинга | `9090:9090` | Сбор метрик приложения |
| **grafana** | Визуализация метрик | `3000:3000` | Дашборды и аналитика |

Все сервисы находятся в сети `monitoring-network` для взаимодействия друг с другом.

---

## Подготовка и сборка

### Шаг 1: Клонирование репозитория (если необходимо)
```bash
git clone git@github.com:seniorcat/event-booking-service.git
cd event-booking-service
```

### Шаг 2: Сборка Docker-образа приложения

Перед запуском docker-compose необходимо собрать образ приложения:

```bash
docker build -t booking-service:latest -f deploy/docker/Dockerfile .
```

**Важно!** Убедитесь, что сборка завершилась успешно без ошибок. Образ с тегом `booking-service:latest` будет использоваться в docker-compose.

---

## Запуск локального окружения

### Полный запуск всех сервисов

Запуск всех сервисов в фоновом режиме:

```bash
docker compose -f deploy/local/docker-compose.yaml up -d
```

**Расшифровка команды:**
- `docker compose` — команда Docker Compose v2
- `-f deploy/local/docker-compose.yaml` — путь к файлу конфигурации
- `up` — команда запуска сервисов
- `-d` — запуск в детached-режиме (в фоновом режиме)

### Проверка статуса контейнеров

После запуска проверьте, что все сервисы работают:

```bash
docker compose -f deploy/local/docker-compose.yaml ps
```

Ожидаемый вывод:
```
NAME                  STATUS                              PORTS
eventdb               Up (healthy)                        0.0.0.0:5434->5432/tcp
eventredis            Up (healthy)                        0.0.0.0:6379->6379/tcp
booking-service       Up (healthy)                        0.0.0.0:8080->8080/tcp, 0.0.0.0:50051->50051/tcp
prometheus            Up                                  0.0.0.0:9090->9090/tcp
grafana               Up                                  0.0.0.0:3000->3000/tcp
```

**Обратите внимание:** Все сервисы должны иметь статус `Up` или `Up (healthy)`.

---

## Проверка работоспособности

После запуска сервисов необходимо проверить их работоспособность через health-check эндпоинты.

### 1. Проверка базовой доступности сервиса

```bash
curl http://localhost:8080/ping
```

**Ожидаемый ответ:**
```
pong
```

Если получили `pong`, значит сервис запущен и отвечает на запросы.

### 2. Проверка здоровья сервиса

```bash
curl http://localhost:8080/health
```

**Ожидаемый JSON-ответ:**
```json
{
  "status": "healthy",
  "service": "event-booking",
  "timestamp": "2026-02-21T12:28:00Z"
}
```

### 3. Проверка готовности приложения (включая подключение к БД)

```bash
curl http://localhost:8080/ready
```

**Ожидаемый JSON-ответ:**
```json
{
  "status": "ready",
  "service": "event-booking",
  "timestamp": "2026-02-21T12:28:00Z",
  "checks": {
    "database": "connected"
  }
}
```

**Важно!** Если эндпоинт `/ready` возвращает ошибку, проверьте логи сервиса — возможно, проблемы с подключением к базе данных.

### 4. Проверка подключения к PostgreSQL

Проверка доступности базы данных напрямую:

```bash
docker exec -it eventdb pg_isready -U user -d eventdb
```

**Ожидаемый вывод:**
```
eventdb:5432 - accepting connections
```

### 5. Проверка подключения к Redis

```bash
docker exec -it eventredis redis-cli ping
```

**Ожидаемый вывод:**
```
PONG
```

### 6. Проверка Prometheus

Откройте в браузере:
```
http://localhost:9090
```

Должна открыться веб-интерфейс Prometheus.

### 7. Проверка Grafana

Откройте в браузере:
```
http://localhost:3000
```

**Данные для входа:**
- Логин: `admin`
- Пароль: `admin`

---

## Работа с миграциями

### Автоматическое применение миграций

Миграции применяются автоматически при старте сервиса, если в конфиге установлено:

```yaml
database:
  auto_migrate: true
```

Это значение уже установлено в `deploy/local/config.dev.yaml`.

### Ручное применение миграций (при необходимости)

Если нужно применить миграции отдельно или откатить их:

#### Установить goose (утилита для миграций)

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

#### Применить все миграции (Up)

```bash
docker exec -it eventdb goose postgres "user:password@eventdb:5432/eventdb?sslmode=disable" up /app/deploy/migrations
```

#### Откатить последнюю миграцию (Down)

```bash
docker exec -it eventdb goose postgres "user:password@eventdb:5432/eventdb?sslmode=disable" down /app/deploy/migrations
```

#### Посмотреть статус миграций

```bash
docker exec -it eventdb goose postgres "user:password@eventdb:5432/eventdb?sslmode=disable" status /app/deploy/migrations
```

### Проверка применения миграций

```bash
docker exec -it eventdb psql -U user -d eventdb -c "\dt"
```

**Ожидаемый вывод (таблицы базы данных):**
```
        List of relations
 Schema |   Name    | Type  |  Owner   
--------+-----------+-------+----------
 public | bookings  | table | user
 public | events    | table | user
 public | users     | table | user
(3 rows)
```

---

## Запуск тестов

Проект поддерживает два типа тестов: unit-тесты и интеграционные тесты.

### 1. Запуск unit-тестов (без контейнеров)

Unit-тесты запускаются локально через Go:

```bash
go test ./...
```

Для получения подробного вывода:

```bash
go test -v ./...
```

### 2. Запуск интеграционных тестов (через Docker)

Интеграционные тесты требуют отдельную тестовую базу данных.

#### Шаг 1: Запуск тестового окружения

```bash
docker compose -f int-tests/docker-compose-test.yaml up -d
```

#### Шаг 2: Проверка готовности тестовой БД

```bash
docker compose -f int-tests/docker-compose-test.yaml ps
```

Ожидаемый статус: `Up (healthy)` для сервиса `postgres-test`.

#### Шаг 3: Применение миграций в тестовую БД

```bash
docker exec -it eventdb-test goose postgres "user:password@localhost:5432/eventdb_test?sslmode=disable" up /app/deploy/migrations
```

#### Шаг 4: Запуск интеграционных тестов

```bash
CONFIG_PATH=int-tests/config.local.yaml go test ./internal/cache/... -v
```

#### Шаг 5: Остановка тестового окружения

```bash
docker compose -f int-tests/docker-compose-test.yaml down
```

### Запуск тестов с покрытием

```bash
go test -cover ./...
```

Для детального отчёта по покрытию:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

## Просмотр логов

### Просмотр логов всех сервисов

```bash
docker compose -f deploy/local/docker-compose.yaml logs
```

### Просмотр логов в реальном времени (follow)

```bash
docker compose -f deploy/local/docker-compose.yaml logs -f
```

### Просмотр логов конкретного сервиса

```bash
# Логи сервиса бронирования
docker compose -f deploy/local/docker-compose.yaml logs -f booking-service

# Логи PostgreSQL
docker compose -f deploy/local/docker-compose.yaml logs -f postgres

# Логи Redis
docker compose -f deploy/local/docker-compose.yaml logs -f redis
```

### Просмотр последних N строк логов

```bash
# Последние 50 строк логов booking-service
docker compose -f deploy/local/docker-compose.yaml logs --tail=50 booking-service
```

### Просмотр логов с временными метками

```bash
docker compose -f deploy/local/docker-compose.yaml logs -t
```

### Анализ логов на предмет ошибок

```bash
docker compose -f deploy/local/docker-compose.yaml logs | grep -i error
```

### Вход в контейнер для отладки

```bash
# Вход в контейнер с приложением
docker exec -it booking-service sh

# Вход в контейнер PostgreSQL
docker exec -it eventdb psql -U user -d eventdb

# Вход в контейнер Redis
docker exec -it eventredis redis-cli
```

---

## Остановка и очистка

### Остановка всех сервисов (с сохранением данных)

```bash
docker compose -f deploy/local/docker-compose.yaml stop
```

### Остановка и удаление контейнеров (с сохранением volumes)

```bash
docker compose -f deploy/local/docker-compose.yaml down
```

### Полная очистка (удаление контейнеров, volumes и данных)

**Внимание!** Эта команда удалит все данные из базы данных и Redis.

```bash
docker compose -f deploy/local/docker-compose.yaml down -v
```

### Удаление неиспользуемых образов

```bash
docker image prune -a
```

### Удаление всех неиспользуемых ресурсов

```bash
docker system prune -a --volumes
```

**Предупреждение:** Эта команда удалит все остановленные контейнеры, неиспользуемые сети, dangling images и неиспользуемые volumes.

---

## Диагностика проблем

### Проблема: Сервис не запускается

**Симптом:** Контейнер имеет статус `Exited` или `Restarting`

**Решение:**
```bash
# Проверьте логи контейнера
docker compose -f deploy/local/docker-compose.yaml logs booking-service

# Проверьте статус всех сервисов
docker compose -f deploy/local/docker-compose.yaml ps
```

**Возможные причины:**
1. Не собран Docker-образ — выполните `docker build -t booking-service:latest -f deploy/docker/Dockerfile .`
2. Ошибка в конфигурационном файле — проверьте `deploy/local/config.dev.yaml`
3. Проблемы с зависимостями (PostgreSQL/Redis не готовы) — проверьте health-check статус

---

### Проблема: Сервис отвечает ошибкой на /ready

**Симптом:** Эндпоинт `/ready` возвращает `503 Service Unavailable`

**Решение:**
```bash
# Проверьте логи на предмет ошибок подключения к БД
docker compose -f deploy/local/docker-compose.yaml logs booking-service | grep database

# Проверьте статус PostgreSQL
docker exec -it eventdb pg_isready -U user -d eventdb

# Проверьте, применились ли миграции
docker exec -it eventdb psql -U user -d eventdb -c "\dt"
```

---

### Проблема: Ошибки миграций

**Симптом:** В логах видно сообщения об ошибках миграций

**Решение:**
```bash
# Сбросьте состояние миграций (внимание: удалит данные!)
docker compose -f deploy/local/docker-compose.yaml down -v
docker compose -f deploy/local/docker-compose.yaml up -d

# Подождите, пока сервис применит миграции автоматически
sleep 10
docker compose -f deploy/local/docker-compose.yaml logs booking-service
```

---

### Проблема: Порты уже заняты

**Симптом:** Ошибка при запуске `docker compose up` о занятых портах

**Решение:**
```bash
# Найдите процесс, занимающий порт (например, 8080)
lsof -i :8080

# Или проверьте через netstat
netstat -tulpn | grep 8080

# Остановите конфликтующий процесс или измените порт в docker-compose.yaml
```

---

### Проблема: Нет доступа к Grafana/Prometheus

**Симптом:** Браузер не открывает `http://localhost:3000` или `http://localhost:9090`

**Решение:**
```bash
# Проверьте, запущены ли сервисы
docker compose -f deploy/local/docker-compose.yaml ps

# Проверьте логи сервисов
docker compose -f deploy/local/docker-compose.yaml logs grafana
docker compose -f deploy/local/docker-compose.yaml logs prometheus

# Проверьте, проброшены ли порты
docker port eventbooking-grafana-1
docker port eventbooking-prometheus-1
```

---

### Проблема: Недостаточно места на диске

**Симптом:** Ошибка `no space left on device` при работе с Docker

**Решение:**
```bash
# Проверьте использование диска Docker
docker system df

# Очистите неиспользуемые данные
docker system prune -a --volumes

# Проверьте свободное место на диске
df -h
```

---

## Полезные команды и сценарии

### Перезапуск конкретного сервиса

```bash
docker compose -f deploy/local/docker-compose.yaml restart booking-service
```

### Обновление кода и перезапуск

```bash
# 1. Соберите новый образ
docker build -t booking-service:latest -f deploy/docker/Dockerfile .

# 2. Перезапустите сервис
docker compose -f deploy/local/docker-compose.yaml up -d --force-recreate booking-service

# 3. Проверьте логи
docker compose -f deploy/local/docker-compose.yaml logs -f booking-service
```

### Просмотр ресурсов, потребляемых сервисами

```bash
docker stats
```

### Вывод информации о контейнере

```bash
docker inspect booking-service
```

### Копирование файлов из контейнера

```bash
docker cp booking-service:/app/config.yaml ./config.backup.yaml
```

---

## Рекомендации по работе

### Всегда проверяйте health-check эндпоинты

Перед началом работы с API убедитесь, что сервис полностью готов:

```bash
# 1. Базовая проверка
curl http://localhost:8080/ping

# 2. Проверка здоровья
curl http://localhost:8080/health

# 3. Проверка готовности (важно!)
curl http://localhost:8080/ready
```

### Мониторинг логов во время разработки

Запустите логи в отдельном терминале:

```bash
docker compose -f deploy/local/docker-compose.yaml logs -f
```

### Резервное копирование данных

```bash
# Экспорт базы данных
docker exec eventdb pg_dump -U user eventdb > backup.sql

# Импорт базы данных
docker exec -i eventdb psql -U user eventdb < backup.sql
```

### Используйте Make-файл для автоматизации

Если в проекте есть Makefile, используйте его для частых команд:

```bash
make dev-up    # запуск окружения
make dev-down  # остановка окружения
make logs      # просмотр логов
make test      # запуск тестов
```

---

## Контакты и поддержка

При возникновении проблем:

1. Проверьте логи всех сервисов
2. Убедитесь, что все health-check эндпоинты возвращают успешные ответы
3. Проверьте, что docker-compose файлы не были изменены
4. Попробуйте полностью пересоздать окружение: `docker compose down -v && docker compose up -d`

---

## Дополнительная документация

- [Основной README проекта](README.md)
- [Документация API (Swagger)](docs/swagger.yaml)
- [SQL-миграции](deploy/migrations/)
- [Конфигурация](deploy/local/config.dev.yaml)

---

**Дата создания:** 21 февраля 2026  
**Версия:** 1.0  