package models

import "time"

// Task - сердце приложения
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	IsDone    bool      `json:"is_done"`
	CreatedAt time.Time `json:"created_at"`
}