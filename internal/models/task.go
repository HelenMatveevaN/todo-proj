package models

import "time"

// Task - сердце приложения
type Task struct {
	ID        int       `db:"id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	IsDone    bool      `db:"is_done"`
	CreatedAt time.Time `db:"created_at"`
}