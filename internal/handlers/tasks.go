package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"

	"todo-proj/internal/database"
)

type Handler struct {
	Pool *pgxpool.Pool
}

// Старая функция (оставляем ее, полезна для тестов)
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("API To-Do приложение работает!"))
}

// Метод структуры Handler
func (h *Handler) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := database.GetTasks(h.Pool)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks) //превращаем список задач в json
}

func (h *Handler) CreateTaskHandler (w http.ResponseWriter, r *http.Request) {
	var newTask struct {
		Title string `json:title`
	}

	//декодируем то, что прислал пользователь
	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	err := database.CreateTask(h.Pool, newTask.Title)
	if err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("+ Задача создана"))
}