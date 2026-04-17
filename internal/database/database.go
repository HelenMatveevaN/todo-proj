package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"     

	"github.com/jackc/pgx/v4/pgxpool"
	pb "todo-proj/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"todo-proj/internal/models"
)

var ErrEmptyTitle = errors.New("название задачи не может быть пустым")

func CreateTask(db *pgxpool.Pool, title string) (models.Task, error) {
	//1. Сохраняем данные в базу
    var t models.Task
    err := db.QueryRow(context.Background(), 
        "INSERT INTO tasks (title) VALUES ($1) RETURNING id, title, is_done, created_at", 
        title).Scan(&t.ID, &t.Title, &t.IsDone, &t.CreatedAt)

	if err != nil {
		return t, fmt.Errorf("ошибка вставки в БД: %w", err)
	}

	// Выносим адрес в переменную
	addr := os.Getenv("NOTIFIER_ADDR")
	if addr == "" {
	    addr = "notifier:50051" // Значение по умолчанию для Docker
	}

	// Запускаем отправку уведомления в фоне
	go func(taskTitle string, taskID int) {	
		// Создаем отдельный контекст для фоновой задачи, 
		// чтобы defer cancel() из родительской функции его не убил
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			slog.Error("не удалось установить соединение с Notifier", "error", err)
			return
		}
		defer conn.Close()

		client := pb.NewNotifierClient(conn)
		_, err = client.SendNotification(ctx, &pb.NotificationRequest{
			TaskTitle: taskTitle,
			Message:   "Ура! Новая задача создана в системе.",
		})

		if err != nil {
			slog.Error("ошибка отправки gRPC уведомления", "error", err)
		} else {
			slog.Info("задача создана и уведомление отправлено", "id", taskID)
		}
	}(t.Title, t.ID) // Пробрасываем данные в горутину

	slog.Info("задача создана", "id", t.ID)
    return t, nil
}

// достает все задачи из базу и возвращает их в виде слайса
func GetTasks(db *pgxpool.Pool) ([]models.Task, error) {
	//выполняем запрос
	rows, err := db.Query(context.Background(), "SELECT id, title, is_done, created_at FROM tasks ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("не удалось получить задачи: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {
		var t models.Task
		//Scan копирует данные из колонок таблицы в поля структуры
		err := rows.Scan(&t.ID, &t.Title, &t.IsDone, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка при чтении строки: %w", err)
		}
		tasks = append(tasks, t) //добавляем задачу в список
	}

	slog.Info("получен список задач", "count", len(tasks))
	return tasks, nil
}

func GetTaskByID(db *pgxpool.Pool, id int) (models.Task, error) {
	var t models.Task
	query := `SELECT id, title, is_done, created_at FROM tasks WHERE id = $1`

	err := db.QueryRow(context.Background(), query, id).Scan(&t.ID, &t.Title, &t.IsDone, &t.CreatedAt)
	if err != nil {
		return t, err
	}

	slog.Info("задача выбрана", "id", id, "title", t.Title)
	return t, nil
}

func DeleteTask(db *pgxpool.Pool, id int) error {
	query := `DELETE FROM tasks WHERE id = $1`

	_, err := db.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("не удалось удалить задачу с id %d: %w", id, err)
	}

	slog.Info("задача удалена", "id", id)
	return nil
}

func GetTasksByStatus(db *pgxpool.Pool, IsDone bool) ([]models.Task, error) {
	query := `SELECT id, title, is_done, created_at FROM tasks WHERE is_done = $1`

	rows, err := db.Query(context.Background(), query, IsDone)
	if err != nil {
		return nil, fmt.Errorf("ошибка фильтрации: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.IsDone, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	slog.Info("фильтрация по статусу", "status", IsDone, "count", len(tasks))
	return tasks, nil
}

func UpdateTaskStatus(db *pgxpool.Pool, id int, IsDone bool) error {
	query := `UPDATE tasks SET is_done = $1 WHERE id = $2`

	result, err := db.Exec(context.Background(), query, IsDone, id)
	if err != nil {
		return fmt.Errorf("не удалось обновить задачу %d: %w", id, err)
	}

	//проверка: а была ли такая задача?
	rowAffected := result.RowsAffected()
	if rowAffected == 0 {
		return fmt.Errorf("задача с id %d не найдена", id)
	}

	slog.Info("статус задачи изменен", "id", id, "is_done", IsDone)
	return nil
}

