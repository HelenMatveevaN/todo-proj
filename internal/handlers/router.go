package handlers

import (
    "net/http"

    "github.com/go-chi/chi/v4"
    "github.com/go-chi/chi/v4/middleware"
)

// NewRouter настраивает маршруты и промежуточное ПО (middleware)
func NewRouter(h *Handler) *chi.Mux {
    r := chi.NewRouter()

    // 1. Middleware (промежуточное ПО)
    r.Use(middleware.Logger)    // Логирует запросы в консоль
    r.Use(middleware.Recoverer) // Спасает от паники (panic)

    // 2. Маршруты API
    r.Get("/health", HealthCheck) // Простая проверка доступности
    
    r.Route("/tasks", func(r chi.Router) {      
        r.Get("/", h.GetTasksHandler)           // Получить все
        r.Get("/{id}", h.GetTaskByIDHandler)    // Получить одну
        r.Post("/", h.CreateTaskHandler)        // Создать
        r.Delete("/{id}", h.DeleteTaskHandler)  // Удалить
        r.Patch("/{id}", h.UpdateTaskHandler)   // Обновить (статус или текст)
    })

    // 3. Раздача статики (Frontend)
    // Важно: папка static должна быть в корне проекта
    r.Handle("/*", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))
    
    return r
}