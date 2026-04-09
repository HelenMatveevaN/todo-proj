package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	//"github.com/jackc/pgx/v4/pgxpool"
	"github.com/go-chi/chi/v4"

	"todo-proj/internal/models"
	//"todo-proj/internal/database"
	"todo-proj/internal/service"
)

type Handler struct {
	//Pool *pgxpool.Pool
	Service service.TaskService //зависим от интерфейса
}

// Старая функция (оставляем ее, полезна для тестов)
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("API To-Do приложение работает!"))
}

// Метод структуры Handler
func (h *Handler) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	//tasks, err := database.GetTasks(h.Pool)

	//вызываем сервис, а не бд напрямую
	tasks, err := h.Service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks) //превращаем список задач в json
}

func (h *Handler) CreateTaskHandler (w http.ResponseWriter, r *http.Request) {
	var newTask struct {
		Title string `json:"title"`
	}

	//декодируем то, что прислал пользователь
	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	if newTask.Title == "" {
		http.Error(w, "Название задачи не может быть пустым", http.StatusBadRequest)
		return
	}

	//err := database.CreateTask(h.Pool, newTask.Title)
	err := h.Service.Create(r.Context(), newTask.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("+ Задача создана"))
}

func (h *Handler) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id") // достаем {id} из URL
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	//err = database.DeleteTask(h.Pool, id)
	err = h.Service.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // Успешно, без тела ответа
}

func (h *Handler) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	var input struct {
		IsDone bool `json:"is_done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Плохой JSON", http.StatusBadRequest)
		return
	}

	//err = database.UpdateTaskStatus(h.Pool, id, input.IsDone)
	err = h.Service.UpdateStatus(r.Context(), id, input.IsDone)
	if err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Статус обновлен"))
}