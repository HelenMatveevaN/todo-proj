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

// Получение одной задачи
func (h *Handler) GetTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id") // Вытаскиваем ID из ссылки
	id, err := strconv.Atoi(idStr) // Конвертируем "1" в число 1
	if err != nil {
		//http.Error(w, "Некорректный ID", http.StatusBadRequest)
		sendJSONError(w, "Некорректный ID", http.StatusBadRequest)
		return
	}
	
	//task, err := database.GetTaskByID(h.Pool, id)
	task, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		//http.Error(w, "Задача не найдена", http.StatusNotFound)
		sendJSONError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
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

	err := h.Service.Create(r.Context(), newTask.Title)
	if err != nil {
		// Проверяем, какая именно ошибка произошла
		switch err {
		case service.ErrTitleTooEmpty, service.ErrTitleTooLong:
			//http.Error(w, err.Error(), http.StatusBadRequest) // Ошибка клиента (400)
			sendJSONError(w, err.Error(), http.StatusBadRequest)
		default:
			//http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError) // Ошибка сервера (500)
			sendJSONError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
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
		if err == service.ErrTaskNotFound {
			//http.Error(w, err.Error(), http.StatusNotFound) //Отдаем 404
			sendJSONError(w, err.Error(), http.StatusNotFound) 
		} else {
			//http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			sendJSONError(w, "Ошибка сервера", http.StatusInternalServerError) 
		}
		return
	}

	w.Write([]byte("Статус обновлен"))
}

func sendJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}