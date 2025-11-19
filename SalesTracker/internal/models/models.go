package models

import "time"

type Item struct {
	ID    int       `json:"id,omitempty"`
	Price int       `json:"price"`
	Name  string    `json:"name"`
	Type  string    `json:"type"`
	Date  time.Time `json:"date,omitempty"`
}
