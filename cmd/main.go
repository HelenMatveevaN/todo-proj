package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/go-chi/chi/v4"

	"todo-proj/internal/database"
	"todo-proj/internal/handlers"
	"todo-proj/internal/service"
)

func main() {
	// 1. Загрузка конфигов
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение: .env не найден, берем переменные из окружения")
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL не задана")
	}

	// 2. Подключение к БД
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbpool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer dbpool.Close()

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalf("База не отвечает: %v", err)
	}
	fmt.Println("Успешное подключение к Postgres!")

	// 3. Миграции (создание таблиц)
	if err = database.InitDatabase(dbpool); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}

	// Создаем сервис, передаем ему пул
	taskSvc := service.NewTaskService(dbpool)

	// 4. Настройка роутера
	h := &handlers.Handler{
		//Pool: dbpool
		Service: taskSvc,
	}
	
	r := chi.NewRouter()
	r.Get("/health", handlers.HealthCheck) //Маршрут для проверки

    // Middleware для логов (очень полезно при разработке)
    // r.Use(middleware.Logger)
	
	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", h.GetTasksHandler)
		r.Post("/", h.CreateTaskHandler)
		r.Delete("/{id}", h.DeleteTaskHandler)
		r.Patch("/{id}", h.UpdateTaskHandler)
	})

	// Раздаем статику из папки "static"
	r.Handle("/*", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))

	// 5. ЗАПУСК СЕРВЕРА (это всегда в самом конце)
	//fmt.Println("Сервер запущен на :8080")
	//if err := http.ListenAndServe(":8080", r); err != nil { //Запускаем сервер, он будет "висеть" и ждать запросов
	//	log.Fatalf("Ошибка запуска сервера: %v", err)
	//}

	//Настраиваем параметры сервера
	srv := &http.Server {
		Addr: ":8080",
		Handler: r,
	}

	//Запускаем сервер в отдельной го-рутине (в фоне),
	//чтобы он не блокировал основной поток
	go func() {
		fmt.Println("Сервер запущен на :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	//Канал для прослушивания сигналов прерывания от системы (например, Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM) // Ждем сигнал прерывания

	<-quit // Здесь программа замирает и ждет сигнала
	fmt.Println("Завершение работы сервера...")

	// Даем серверу 5 секунд, чтобы завершить текущие запросы
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Сервер не смог плавно завершиться: %v", err)
	}

	fmt.Println("Сервер успешно остановлен.")
}


