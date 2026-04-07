package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

// Task - сердце приложения
type Task struct {
	ID        int       `db:"id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	IsDone    bool      `db:"is_done"`
	CreatedAt time.Time `db:"created_at"`
}

func main() {
	//загрузка .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	//строка подключения к бд
	//connStr := "postgres://postgres:postgres@localhost:5432/todo_db"
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL не задана в .env")
	}

	//пул соединений (+контекст, для создания контроля времени)
	dbpool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	//проверка связи
	err = dbpool.Ping(context.Background())
	if err != nil {
		log.Fatalf("База не отвечает: %v", err)
	}

	fmt.Println("Ура! Мы успешно подключились к Postgres на Go!")

	err = CreateTask(dbpool, "Помыть пол через неделю")
	if err != nil {
		log.Println(err)
	}

	tasks, err := GetTasks(dbpool)
	if err != nil {
		log.Fatalf("Ошибка% %v", err)
	}

	fmt.Println("--- Список ваших задач из базы: ---")
	for _, t := range tasks {
		status := "X"
		if t.IsDone {
			status = "OK!"
		}
		fmt.Printf("[%d] %s %s\n", t.ID, status, t.Title)
	}
}

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
func GetTasks(db *pgxpool.Pool) ([]Task, error) {
	//выполняем запрос
	rows, err := db.Query(context.Background(), "SELECT id, title, is_done FROM tasks")
	if err != nil {
		return nil, fmt.Errorf("не удалось получить задачи: %w", err)
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var t Task
		//Scan копирует данные из колонок таблицы в поля структуры
		err := rows.Scan(&t.ID, &t.Title, &t.IsDone)
		if err != nil {
			return nil, fmt.Errorf("ошибка при чтении строки: %w", err)
		}
		tasks = append(tasks, t) //добавляем задачу в список

	}
	return tasks, nil
}
