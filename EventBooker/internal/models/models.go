package models

import "time"

type Event struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity"` //вместимость
	CreatedAt   time.Time
}

type Book struct {
	Id          int
	CreatedAt   time.Time
	IsConfirmed bool
}
