package service

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"todo-proj/internal/database"
	"todo-proj/internal/models"
)

//что должен уметь наш сервис?
type TaskService interface {
	Create(ctx context.Context, title string) error
	List(ctx context.Context) ([]models.Task, error)
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

func (s *taskService) Create(ctx context.Context, title string) error {
	//здесь будет бизнес-логика (валидация, доп.проверки)
	if title == "" {
		return database.ErrEmptyTitle //пример ошибки
	}
	return database.CreateTask(s.pool, title)
}

func (s *taskService) List(ctx context.Context) ([]models.Task, error) {
	return database.GetTasks(s.pool)
}

func (s *taskService) Delete(ctx context.Context, id int) error {
	return database.DeleteTask(s.pool, id)
}

func (s *taskService) UpdateStatus(ctx context.Context, id int, isDone bool) error {
	return database.UpdateTaskStatus(s.pool, id, isDone)
}