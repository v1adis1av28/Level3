package models

import "time"

type Event struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	Capacity         int    `json:"capacity"` //вместимость
	ConfirmationNeed bool   `json:"confirmation_need"`
	CreatedAt        time.Time
}

type Book struct {
	Id          int
	CreatedAt   time.Time
	IsConfirmed bool
}

type BookPayload struct {
	EventId int `json:"event_id omitempty"`
	BookId  int `json:"book_id"`
}
