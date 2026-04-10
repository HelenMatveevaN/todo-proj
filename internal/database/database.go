package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"

	"todo-proj/internal/models"
)

var ErrEmptyTitle = errors.New("название задачи не может быть пустым")

func CreateTask(db *pgxpool.Pool, title string) (models.Task, error) {
    var t models.Task
    err := db.QueryRow(context.Background(), 
        "INSERT INTO tasks (title) VALUES ($1) RETURNING id, title, is_done, created_at", 
        title).Scan(&t.ID, &t.Title, &t.IsDone, &t.CreatedAt)

    if err == nil {
        fmt.Printf("✅ Задача создана: ID=%d, Title=%s\n", t.ID, t.Title)
    }
    return t, err
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

    fmt.Printf("📋 Получен список задач: всего %d шт.\n", len(tasks))
	return tasks, nil
}

func GetTaskByID(db *pgxpool.Pool, id int) (models.Task, error) {
	var t models.Task
	query := `SELECT id, title, is_done, created_at FROM tasks WHERE id = $1`

	err := db.QueryRow(context.Background(), query, id).Scan(&t.ID, &t.Title, &t.IsDone, &t.CreatedAt)
	if err != nil {
		return t, err
	}

	fmt.Printf("🔍 Задача выбрана: ID=%d, Title=%s\n", id, t.Title) // Добавили детали в лог
	return t, nil
}

func DeleteTask(db *pgxpool.Pool, id int) error {
	query := `DELETE FROM tasks WHERE id = $1`

	_, err := db.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("не удалось удалить задачу с id %d: %w", id, err)
	}

	fmt.Printf("Задача с ID %d удалена \n", id)
	return nil
}

func GetTasksByStatus(db *pgxpool.Pool, IsDone bool) ([]models.Task, error) {
	query := `SELECT id, title, is_done FROM tasks WHERE is_done = $1`

	rows, err := db.Query(context.Background(), query, IsDone)
	if err != nil {
		return nil, fmt.Errorf("ошибка фильтрации: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.IsDone); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	fmt.Printf("📂 Фильтрация по статусу (is_done=%v): найдено %d шт.\n", IsDone, len(tasks))
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

	fmt.Printf("Статус задачи %d изменен на %v\n", id, IsDone)
	return nil
}

func InitDatabase(pool *pgxpool.Pool) error {
	query := `
	create table if not exists tasks (
		id 			SERIAL	primary key,
		title 		text	not null,
		content		text,
		is_done		boolean	default false,
		created_at	timestamp	default now()
	);`

	_, err := pool.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("не удалось инициализировать базу: %w", err)
	}

	fmt.Println("База данных проверена и готова к работе!")
	return nil
}


