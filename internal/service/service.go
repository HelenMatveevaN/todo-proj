package service

import (
	"context"
	"encoding/json" // Добавили для работы с JSON в Redis
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"

	"todo-proj/internal/database"
	"todo-proj/internal/models"
)

// Ключ для хранения списка задач в Redis
const tasksCacheKey = "tasks_list"

var (
	ErrTitleTooEmpty = errors.New("название задачи не может быть пустым")
	ErrTitleTooLong = errors.New("название задачи слишком длинное (макс. 100 символов)")
	ErrTaskNotFound = errors.New("задача не найдена")
	ErrTaskInvalidTitle = errors.New("пустой заголовок")
)

// ValidateTask — та самая функция, которую ищет тест
func ValidateTask(title string) error {
	if title == "" {
		return ErrTitleTooEmpty
	}
	if len(title) > 100 {
		return ErrTitleTooLong
	}
	return nil
}

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
	rdb *redis.Client
	broker *RabbitMQ
}

// Обновленный конструктор
func NewTaskService(pool *pgxpool.Pool, rdb *redis.Client, b *RabbitMQ) TaskService {
	return &taskService{
		pool: pool,
		rdb: rdb,
		broker: b,
	}
}

func (s *taskService) List(ctx context.Context) ([]models.Task, error) {
	// 1. Пробуем получить данные из Redis
	val, err := s.rdb.Get(ctx, tasksCacheKey).Result()
	if err == nil {
		var tasks []models.Task
		if err := json.Unmarshal([]byte(val), &tasks); err == nil {
			return tasks, nil // Успешно вернули из кеша
		}
	}

	// 2. Если в кеше нет — идем в БД
	tasks, err := database.GetTasks(s.pool)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем результат в Redis на 5 минут
	data, _ := json.Marshal(tasks)
	s.rdb.Set(ctx, tasksCacheKey, data, 5*time.Minute)

	return tasks, nil	
}

func (s *taskService) Create(ctx context.Context, title string) (models.Task, error) {
	if err := ValidateTask(title); err != nil {
		return models.Task{}, err
	}	
	if strings.TrimSpace(title) == "" {
		return models.Task{}, ErrTaskInvalidTitle
	}
	if len(title) > 100 {
		return models.Task{}, ErrTitleTooLong
	}

	task, err := database.CreateTask(s.pool, title)
	if err != nil {
		return models.Task{}, err
	}	

	// 1. Инвалидация кеша Redis
	s.rdb.Del(ctx, tasksCacheKey)

	// 2. Отправка через новый пакет
	go func() {
		if err := s.broker.PublishTaskCreated(context.Background(), task); err != nil {
			slog.Error("ошибка публикации события", "err", err)
		}
	}()    

	return task, nil
}

func (s *taskService) Delete(ctx context.Context, id int) error {
	err := database.DeleteTask(s.pool, id)
	if err == nil {
		s.rdb.Del(ctx, tasksCacheKey)
	}
	return err
}

func (s *taskService) UpdateStatus(ctx context.Context, id int, isDone bool) error {
	err := database.UpdateTaskStatus(s.pool, id, isDone)
	if err != nil {
		return ErrTaskNotFound
	}
	// Успешно обновили — сбрасываем кеш
	s.rdb.Del(ctx, tasksCacheKey)
	return nil
}

func (s *taskService) GetByID(ctx context.Context, id int) (models.Task, error) {
	return database.GetTaskByID(s.pool, id)
}