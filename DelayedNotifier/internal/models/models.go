package models

import (
	"time"
)

type Notification struct {
	Text       string `json:"text"`
	TimeToSend string `json:"sendTime"`
	UserId     int    `json:"userId"`
}

type NotificationDB struct {
	ID         int       `json:"id"`
	Text       string    `json:"text"`
	Status     string    `json:"status"` // scheduled, sent, failed, cancelled, retrying
	TimeToSend time.Time `json:"sendTime"`
	UserId     int       `json:"userId"`
	CreatedAt  time.Time `json:"createdAt"`
	SentAt     time.Time `json:"sentAt"`
	RetryCount int       `json:"retryCount"`
	MaxRetries int       `json:"maxRetries"`
	NextRetry  time.Time `json:"nextRetry"`
}

type NotificationMessage struct {
	NotificationID int       `json:"notificationId"`
	UserID         int       `json:"userId"`
	Text           string    `json:"text"`
	SendTime       time.Time `json:"sendTime"`
	Attempt        int       `json:"attempt"`
	MaxAttempts    int       `json:"maxAttempts"`
}
