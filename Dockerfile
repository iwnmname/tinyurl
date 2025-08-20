# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копирование файлов модулей
COPY go.mod go.sum ./
RUN go mod download

# Копирование кода
COPY . .

# Сборка сервера
RUN go build -o tinyurl-server ./cmd/server/

# Сборка CLI-клиента
RUN go build -o tinyurl-cli ./cmd/cli/

# Stage 2: Runtime
FROM alpine:3.18

WORKDIR /app

# Установка зависимостей
RUN apk --no-cache add ca-certificates

# Копирование исполняемых файлов и схемы БД
COPY --from=builder /app/tinyurl-server .
COPY --from=builder /app/tinyurl-cli .
COPY --from=builder /app/db ./db

# Создание директории для данных
VOLUME /data

# Переменные окружения по умолчанию
ENV TINYURL_DB_PATH="file:/data/tinyurl.db?cache=shared&mode=rwc&_fk=1"
ENV PORT="8080"

# Запуск
EXPOSE 8080
CMD ["./tinyurl-server"]