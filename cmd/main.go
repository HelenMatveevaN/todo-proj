package main

import (
	"context"
	//"fmt"
	"log"
	"log/slog" // Новый стандарт Go 1.21
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo-proj/internal/handlers"
	"todo-proj/internal/service"
	"todo-proj/internal/config"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Настройка логирования (JSON формат удобен для Docker)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Инициализация конфига (Singleton из файла config.go)
	cfg := config.GetConfig()

	// 3. Инициализация БД  (теперь с retry)
	dbpool := setupDatabase(cfg.DatabaseURL)
	defer dbpool.Close()

	// 3.1 Инициализация Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisHost,
	})

	// 3.2 Проверка Redis (Ping)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("Redis не доступен", "err", err)
	} else {
		slog.Info("успешное подключение к Redis")
	}
	cancel()

	// 3.3 Инициализация RabbitMQ через новый пакет
    rabbit, err := service.NewRabbitMQ(cfg.RabbitURL, "tasks_events")
    if err != nil {
        log.Fatalf("не удалось инициализировать RabbitMQ: %v", err)
    }
    defer rabbit.Close()	

	// 4. Передаем обернутый брокер в сервис
	taskSvc := service.NewTaskService(dbpool, rdb, rabbit)

	h := &handlers.Handler{Service: taskSvc}
	router := handlers.NewRouter(h) // Все middleware уже внутри этого роутера!

	// 5. Запуск сервера
	srv := &http.Server{
		Addr:    cfg.HTTP.Port,
		Handler: router, // Передаем chi.Mux напрямую
	    ReadTimeout:  cfg.HTTP.Timeout,
	    WriteTimeout: cfg.HTTP.Timeout,
	}

	go func() {
		slog.Info("Cервер запущен", "addr", cfg.HTTP.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ошибка сервера", "err", err)
			os.Exit(1) // Если сервер не встал, лучше завершить процесс
		}
	}()

	// 6. Graceful Shutdown
	waitForShutdown(srv)	
}

func setupDatabase(connStr string) *pgxpool.Pool {
	var pool *pgxpool.Pool
	var err error

	maxRetries := 5
	delay := 2 * time.Second
	
	for i := 1; i <= maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		pool, err = pgxpool.Connect(ctx, connStr)
		if err == nil {
			err = pool.Ping(ctx) // Проверяем реальную связь

			if err == nil {
				cancel() // Успех!
				slog.Info("успешное подключение к Postgres")
				return pool
			}
		}

		// Если мы здесь, значит была ошибка (Connect или Ping)
		cancel() // <--- Закрываем вручную при неудаче перед time.Sleep
		slog.Warn("База еще не готова", 
			"attempt", i, 
			"max_attempts", maxRetries, 
			"err", err)

		if i < maxRetries {
			time.Sleep(delay)
		}
	}

	log.Fatalf("не удалось подключиться к БД: %v", err)
	return nil
}

func waitForShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("завершение работы сервера...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("ошибка при остановке", "err", err)
	}
	slog.Info("сервер остановлен")
}