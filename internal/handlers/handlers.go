package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v4"

	"todo-proj/internal/service"
)

type Handler struct {
	Service service.TaskService //зависим от интерфейса
}

type Response struct {
	Data  interface{} `json:"data,omitempty"`  // Любые данные (объект или список)
	Error string      `json:"error,omitempty"` // Текст ошибки
}

// Хелперы для унификации ответов
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{Data: data})
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{Error: message})
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusOK, "API To-Do приложение работает!")
}

func (h *Handler) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.Service.List(r.Context())
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Ошибка БД")
		return
	}
	sendJSON(w, http.StatusOK, tasks)
}

func (h *Handler) GetTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id") // Вытаскиваем ID из ссылки
	id, err := strconv.Atoi(idStr) // Конвертируем "1" в число 1
	if err != nil {
		sendError(w, http.StatusBadRequest, "Некорректный ID")
		return
	}
	
	task, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		sendError(w, http.StatusNotFound, "Задача не найдена")
		return
	}
	sendJSON(w, http.StatusOK, task)
}

func (h *Handler) CreateTaskHandler (w http.ResponseWriter, r *http.Request) {
	var req struct { 
		Title string `json:"title"` 
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	task, err := h.Service.Create(r.Context(), req.Title)
	if err != nil {
		// Проверяем конкретные ошибки сервиса
		switch {
		case errors.Is(err, service.ErrTitleTooEmpty), errors.Is(err, service.ErrTaskInvalidTitle):
			// Если заголовок пустой или невалидный — 400 Bad Request
			sendError(w, http.StatusBadRequest, err.Error())
		
		case errors.Is(err, service.ErrTitleTooLong):
			// Если превышена длина — 400 Bad Request
			sendError(w, http.StatusBadRequest, err.Error())

		default:
			// Все остальные ошибки (например, упала база) — 500 Internal Server Error
			sendError(w, http.StatusInternalServerError, "Внутренняя ошибка сервера")
		}
		return
	}

	// Если ошибок нет, возвращаем созданную задачу
	sendJSON(w, http.StatusCreated, task)
}

func (h *Handler) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Неверный ID")
		return
	}

	var input struct {
		IsDone bool `json:"is_done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendError(w, http.StatusBadRequest, "Плохой JSON")
		return
	}

	err = h.Service.UpdateStatus(r.Context(), id, input.IsDone)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			sendError(w, http.StatusNotFound, err.Error())
		} else {
			sendError(w, http.StatusInternalServerError, "Не удалось обновить статус")
		}
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"message": "Статус обновлен"})
}

func (h *Handler) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id") // достаем {id} из URL
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Неверный ID")
		return
	}

	if err := h.Service.Delete(r.Context(), id); err != nil {
		sendError(w, http.StatusInternalServerError, "Ошибка удаления")
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"message": "Удалено"})
}

