package database

import (
	"context"
	"fmt"
	//"log"
	//"os"

	"github.com/jackc/pgx/v4/pgxpool"
	//"github.com/joho/godotenv"

	"todo-proj/internal/models"
)

func CreateTask(db *pgxpool.Pool, title string) error {
	query := `INSERT INTO tasks (title, is_done) VALUES ($1, false)`

	_, err := db.Exec(context.Background(), query, title)
	if err != nil {
		return fmt.Errorf("ошибка при создании задачи '%s': %w", title, err)
	}

	fmt.Printf("Задача '%s' успешно добавлена!\n", title)
	return nil
}

// достает все задачи из базу и возвращает их в виде слайса
func GetTasks(db *pgxpool.Pool) ([]models.Task, error) {
	//выполняем запрос
	rows, err := db.Query(context.Background(), "SELECT id, title, is_done FROM tasks")
	if err != nil {
		return nil, fmt.Errorf("не удалось получить задачи: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {
		var t models.Task
		//Scan копирует данные из колонок таблицы в поля структуры
		err := rows.Scan(&t.ID, &t.Title, &t.IsDone)
		if err != nil {
			return nil, fmt.Errorf("ошибка при чтении строки: %w", err)
		}
		tasks = append(tasks, t) //добавляем задачу в список

	}
	return tasks, nil
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



