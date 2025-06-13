FROM golang:1.24.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Копирование всего проекта
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./main .


# Основная сборка
FROM alpine:latest

# Создаем структуру директорий
RUN mkdir -p /app/web /data

# Рабочая директория
WORKDIR /app
# Сохранение при перезапуске
VOLUME /data

# Копируем бинарник и статику и бд
COPY --from=builder /app/main .
COPY --from=builder /app/web ./web


# Устанавливаем переменные окружения по умолчанию
ENV TODO_PORT=7540 \
    TODO_DBFILE=/data/scheduler.db \
    TODO_LIMIT_TASKS=50

# Порт приложения
EXPOSE $TODO_PORT

# Команда запуска
CMD ["./main"]