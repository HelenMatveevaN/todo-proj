package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo-proj/internal/database"
	"todo-proj/internal/handlers"
	"todo-proj/internal/service"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Инициализация конфига
	cfg := loadConfig()

	// 2. Инициализация БД
	dbpool := setupDatabase(cfg.dbURL)
	defer dbpool.Close()

	// 3. Сборка слоев приложения
	taskSvc := service.NewTaskService(dbpool)
	h := &handlers.Handler{Service: taskSvc}
	router := handlers.NewRouter(h) // Перенесли настройку chi внутрь

	// 4. Запуск сервера
	srv := &http.Server{
		Addr:    cfg.port,
		Handler: router,
	}

	go func() {
		log.Printf("Сервер запущен на %s", cfg.port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// 5. Ожидание завершения (Graceful Shutdown)
	waitForShutdown(srv)	
}

// --- Вспомогательные функции для чистоты main ---
type config struct {
	dbURL string
	port  string
}

func loadConfig() config {
	if err := godotenv.Load(); err != nil {
		log.Println("Инфо: .env не найден, используем системные переменные")
	}
	
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("Ошибка: DB_URL не установлена")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":8080"
	}
	return config{dbURL: dbURL, port: port}
}

func setupDatabase(connStr string) *pgxpool.Pool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Критическая ошибка подключения к БД: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("База не отвечает на Ping: %v", err)
	}

	if err := database.InitDatabase(pool); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}

	fmt.Println("Успешное подключение к Postgres!")
	return pool
}

func waitForShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Завершение работы сервера...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при остановке: %v", err)
	}
	log.Println("Сервер остановлен.")
}
