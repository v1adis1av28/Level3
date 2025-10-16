package service

import (
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"
)

// здесь будут формировать сообщения для отправки в rbmq
type NotificationService struct {
	DB       *dbpg.DB
	Producer *rabbitmq.Channel
}

func NewNotificationService(db *dbpg.DB) *NotificationService {
	conn, err := rabbitmq.Connect("amqp://guest:guest@localhost:5672", 3, time.Duration(time.Now().Hour()))
	if err != nil {
		zlog.Logger.Err(err)
		return nil
	}
	ch, err := conn.Channel()
	if err != nil {
		zlog.Logger.Err(err)
		return nil
	}
	return &NotificationService{DB: db, Producer: ch}

}
