package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"

	"todo-proj/internal/database"
)

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

	err = database.CreateTask(dbpool, "Помыть пол через неделю")
	if err != nil {
		log.Println(err)
	}

	err = database.DeleteTask(dbpool, 2)
	if err != nil {
		log.Println(err)
	}

	err = database.UpdateTaskStatus(dbpool, 7, true)
	if err != nil {
		log.Fatalf("Ошибка% %v", err)
	}

	tasks, err := database.GetTasks(dbpool)
	if err != nil {
		log.Fatalf("Ошибка% %v", err)
	}

	fmt.Println("--- Список ваших задач из базы: ---")
	for _, t := range tasks {
		status := "-"
		if t.IsDone {
			status = "+"
		}
		fmt.Printf("[%d] %s %s\n", t.ID, status, t.Title)
	}
}


