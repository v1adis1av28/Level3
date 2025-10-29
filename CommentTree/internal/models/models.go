package models

import "time"

type CreateRequest struct {
	ParentId int    `json:"parent_id"`
	Username string `json:"username"`
	Text     string `json:"text"`
}

type Comment struct {
	ID        int        `json:"id"`
	ParentId  int        `json:"parent_id"`
	Username  string     `json:"username"`
	Text      string     `json:"text"`
	CreatedAt time.Time  `json:"created_at"`
	Children  []*Comment `json:"children,omitempty"`
}

type CommentsResponse struct {
	Comments []*Comment `json:"comments"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	Limit    int        `json:"limit"`
}

type SearchRequest struct {
	Query string `json:"query"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}
