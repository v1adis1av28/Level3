package models

type LoginRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type Item struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

type Payload struct {
	Username string
	Role     string
}
