# ==========================================
# STAGE 1: BUILDER - ЭТАП СБОРКИ
# ==========================================

FROM golang:1.25-alpine AS builder

WORKDIR /build

# Установка инструментов для сборки
RUN apk add --no-cache git gcc musl-dev make

# Ключевой момент: сначала копируем только зависимости!
# Это позволяет Docker кешировать слой с go mod download
COPY go.mod go.sum ./

# Загружаем зависимости (закешируется, если go.mod не менялся)
RUN go mod download

# Теперь копируем весь код
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o deploy/bin/booking-service \
    ./cmd/server/main.go

# ==========================================
# STAGE 2: RUNTIME - ФИНАЛЬНЫЙ ОБРАЗ
# ==========================================

FROM alpine:latest

# Минимальные зависимости для запуска
RUN apk --no-cache add ca-certificates tzdata

# Создаем непривилегированного пользователя (безопасность!)
RUN addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

# Копируем только нужное из builder stage
COPY --from=builder /build/deploy/bin/booking-service .
COPY --from=builder /build/deploy/migrations ./migrations

# Меняем владельца файлов
RUN chown -R appuser:appgroup /app

# Переключаемся на непривилегированного пользователя
USER appuser


EXPOSE 8080   
EXPOSE 50051  

# Команда запуска приложения
ENTRYPOINT ["./booking-service"]
