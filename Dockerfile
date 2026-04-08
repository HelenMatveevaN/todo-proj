FROM golang:1.21-alpine

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь остальной код
COPY . .

# Собираем бинарный файл
RUN go build -o main ./cmd/main.go

# Запуск
CMD ["./main"]