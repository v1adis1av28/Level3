package models

type Item struct {
	ID    int    `json:"id omitempty"`
	Price int    `json:"price"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}
