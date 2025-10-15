package service

import "github.com/wb-go/wbf/dbpg"

// здесь будут формировать сообщения для отправки в rbmq
type NotificationService struct {
	DB *dbpg.DB
}

func NewNotificationService(db *dbpg.DB) *NotificationService {
	return &NotificationService{DB: db}
}
