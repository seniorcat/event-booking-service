FROM debian:stretch-slim

WORKDIR /app

# бинарь сервиса
COPY deploy/bin/booking-service /app

# миграции (если будут)
COPY deploy/migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["./booking-service"]
