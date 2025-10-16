package service

import (
	"context"
	"fmt"
	"time"

	"github.com/v1adis1av28/level3/DelayedNotifier/internal/models"
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
	conn, err := rabbitmq.Connect("amqp://guest:guest@rabbitmq/", 3, time.Duration(time.Now().Hour()))
	if err != nil {
		zlog.Logger.Err(err)
		fmt.Println(err.Error())
		fmt.Println("Connection error")
		return nil
	}
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Error on channel")
		zlog.Logger.Err(err)
		return nil
	}
	return &NotificationService{DB: db, Producer: ch}

}

func (ns *NotificationService) CreateNotification(notification *models.Notification) error {
	//пофиксить парс времени и даты в нотификатионс и нужно сделать отправку сообщения в rabbitmq
	ns.DB.ExecContext(context.Background(), "INSERT INTO NOTIFICATIONS (TEXT,STATUS,SENDTIME,USERID) VALUES($1,$2,$3,$4)", notification.Text, notification.Status, time.Now(), notification.UserId)
	return nil
}
