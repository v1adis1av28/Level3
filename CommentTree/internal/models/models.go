package models

type CreateRequest struct {
	ParrentId int    `json:"parrent_id"`
	Username  string `json:"username"`
	Text      string `json:"text"`
}
