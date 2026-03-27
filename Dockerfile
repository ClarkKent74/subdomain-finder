# Этап сборки
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Устанавливаем swag для генерации документации
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4

# Копируем зависимости отдельно — кешируется если go.mod не менялся
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Генерируем swagger документацию
RUN swag init -g cmd/server/main.go

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o subdomain-finder ./cmd/server

# Финальный образ — берём готовый образ Sudomy со всеми зависимостями
FROM screetsec/sudomy:v1.2.0-dev

WORKDIR /app

# Копируем только бинарник из этапа сборки
COPY --from=builder /app/subdomain-finder .

# Sudomy ищет engine и plugin в /app/ но они установлены в /usr/lib/sudomy/
# Создаём симлинки чтобы Sudomy нашёл свои файлы
RUN ln -sf /usr/lib/sudomy/engine /app/engine && \
    ln -sf /usr/lib/sudomy/plugin /app/plugin && \
    ln -sf /usr/lib/sudomy/lib /app/lib && \
    ln -sf /usr/lib/sudomy/templates /app/templates

# Создаём пустой sudomy.api чтобы Sudomy не падал при старте
# API ключи передаются через переменные окружения из .env
RUN echo 'SHODAN_API=""\nCENSYS_API=""\nCENSYS_SECRET=""\nVIRUSTOTAL=""\nBINARYEDGE=""\nSECURITY_TRAILS=""' \
    > /usr/lib/sudomy/sudomy.api

EXPOSE 8080

# Переопределяем entrypoint базового образа Sudomy
ENTRYPOINT ["/app/subdomain-finder"]
CMD []