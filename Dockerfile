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


# Порт приложения
EXPOSE 7540

# Команда запуска
CMD ["./main"]