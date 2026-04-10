package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"

	"todo-proj/internal/database"
	"todo-proj/internal/models"
)

var (
	ErrTitleTooEmpty = errors.New("название задачи не может быть пустым")
	ErrTitleTooLong = errors.New("название задачи слишком длинное (макс. 100 символов)")
	ErrTaskNotFound = errors.New("задача не найдена")
	ErrTaskInvalidTitle = errors.New("пустой заголовок")
)

//что должен уметь наш сервис?
type TaskService interface {
	List(ctx context.Context) ([]models.Task, error)
	GetByID(ctx context.Context, id int) (models.Task, error)
	Create(ctx context.Context, title string) (models.Task, error)
	Delete(ctx context.Context, id int) error
	UpdateStatus(ctx context.Context, id int, isDone bool) error
}

//реализация сервиса
type taskService struct {
	pool *pgxpool.Pool
}

func NewTaskService(pool *pgxpool.Pool) TaskService {
	return &taskService{pool: pool}
}

func (s *taskService) Create(ctx context.Context, title string) (models.Task, error) {
	//здесь будет бизнес-логика (валидация, доп.проверки)
	if title == "" {
		return models.Task{}, ErrTitleTooEmpty
	}
	// Пример для ErrTaskInvalidTitle (например, если в названии только пробелы)
    if strings.TrimSpace(title) == "" {
        return models.Task{}, ErrTaskInvalidTitle
    }
	if len(title) > 100 {
		return models.Task{}, ErrTitleTooLong
	}

	return database.CreateTask(s.pool, title)
}

func (s *taskService) List(ctx context.Context) ([]models.Task, error) {
	return database.GetTasks(s.pool)
}

func (s *taskService) Delete(ctx context.Context, id int) error {
	err := database.DeleteTask(s.pool, id)
	if err != nil {
		//здесь можно добавить логик: если ошибка от БД говорит "не найдено"
		//но пока просто пробрасываем
		return err
	}
	return nil
}

func (s *taskService) UpdateStatus(ctx context.Context, id int, isDone bool) error {
	err := database.UpdateTaskStatus(s.pool, id, isDone)
	if err != nil {
		// Если в ошибке есть текст про "не найдена", возвращаем ErrTaskNotFound
		return ErrTaskNotFound
	}
	return nil
}

func (s *taskService) GetByID(ctx context.Context, id int) (models.Task, error) {
	return database.GetTaskByID(s.pool, id)
}