# Этап сборки (Builder)
FROM golang:1.21-alpine AS builder

# Устанавливаем необходимые системные зависимости для сборки (если понадобятся)
RUN apk add --no-cache git

WORKDIR /app

# Сначала копируем только модули, чтобы закешировать слои
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной код (включая папку static и internal)
COPY . .

# Собираем бинарник. 
# CGO_ENABLED=0 делает файл статическим (не зависит от системных библиотек)
RUN CGO_ENABLED=0 GOOS=linux go build -o todo-app ./cmd/main.go

# Финальный легковесный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем исполняемый файл из билдера
COPY --from=builder /app/todo-app .

# ВАЖНО: Копируем статику, чтобы обработчик r.Handle("/*", ...) ее нашел
COPY --from=builder /app/static ./static

# 3. Файл .env (если он нужен внутри, но в Compose мы его уже прокинули)
# COPY --from=builder /app/.env . 

# Открываем порт
EXPOSE 8080

# Запускаем
CMD ["./todo-app"]