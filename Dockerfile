# Используем базовый образ Ubuntu
FROM ubuntu:latest

# Устанавливаем необходимые пакеты
RUN apt update && apt install -y \
    golang \
    sqlite3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Создаём рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости Go
RUN go mod download

# Копируем все
COPY . .

# Кросс-компиляция Go-программы для Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/myapp cmd/main.go

# Устанавливаем переменные окружения
ENV TODO_PORT=:8080 \
    TODO_DBFILE=../scheduler.db \
    TODO_PASSWORD=qwerty69

# Открываем порт
EXPOSE 7540

# Команда запуска
CMD ["./cmd/myapp"]
