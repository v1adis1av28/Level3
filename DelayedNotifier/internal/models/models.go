package models

type Notification struct {
	Text string `json:"text"`
	//Status     string `json:"status"`
	TimeToSend string `json:"sendTime"`
	UserId     int    `json:"userId"`
}

type NotificationDB struct {
	ID         int    `json:"id"`
	Text       string `json:"text"`
	Status     string `json:"status"`
	TimeToSend string `json:"sendTime"`
	UserId     int    `json:"userId"`
}
