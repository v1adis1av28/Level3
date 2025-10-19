package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/v1adis1av28/level3/DelayedNotifier/internal/models"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"
)

var CREATED_STATUS string = "created"

type NotificationService struct {
	DB        *dbpg.DB
	Producer  *rabbitmq.Channel
	Publisher *rabbitmq.Publisher
	Conn      *rabbitmq.Connection
	Mutex     *sync.Mutex
}

func NewNotificationService(db *dbpg.DB) *NotificationService {
	conn, err := rabbitmq.Connect("amqp://guest:guest@rabbitmq/", 3, 5*time.Second)
	if err != nil {
		zlog.Logger.Err(err).Msg("Failed to connect to RabbitMQ")
		return nil
	}

	ch, err := conn.Channel()
	if err != nil {
		zlog.Logger.Err(err).Msg("Failed to create channel")
		conn.Close()
		return nil
	}

	_, err = ch.QueueDeclare(
		"notifications", // name
		true,            // durable - сохранять после перезагрузки
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		zlog.Logger.Err(err).Msg("Failed to declare queue")
		ch.Close()
		conn.Close()
		return nil
	}

	publisher := rabbitmq.NewPublisher(ch, "")

	return &NotificationService{
		DB:        db,
		Producer:  ch,
		Publisher: publisher,
		Conn:      conn,
		Mutex:     &sync.Mutex{},
	}
}

func (ns *NotificationService) Close() {
	if ns.Producer != nil {
		ns.Producer.Close()
	}
	if ns.Conn != nil {
		ns.Conn.Close()
	}
}

func (ns *NotificationService) CreateNotification(notification *models.Notification) error {
	parsedTime, err := time.Parse(time.RFC3339, notification.TimeToSend)
	if err != nil {
		parsedTime, err = time.Parse("2006-01-02 15:04:05", notification.TimeToSend)
		if err != nil {
			parsedTime, err = time.Parse("2006-01-02T15:04:05", notification.TimeToSend)
			if err != nil {
				return fmt.Errorf("неверный формат времени: %v", err)
			}
		}
	}

	now := time.Now()
	if parsedTime.Before(now) {
		return fmt.Errorf("время отправки не может быть в прошлом")
	}

	if notification.Text == "" {
		return fmt.Errorf("текст уведомления не может быть пустым")
	}
	if notification.UserId <= 0 {
		return fmt.Errorf("неверный ID пользователя")
	}

	_, err = ns.DB.ExecContext(
		context.Background(),
		"INSERT INTO NOTIFICATIONS (TEXT, STATUS, SENDTIME, USERID) VALUES($1, $2, $3, $4);",
		notification.Text,
		CREATED_STATUS,
		parsedTime,
		notification.UserId,
	)
	if err != nil {
		return fmt.Errorf("ошибка сохранения в БД: %v", err)
	}

	messageBody, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("ошибка сериализации уведомления: %v", err)
	}

	err = ns.Publisher.Publish(
		messageBody,
		"notifications",
		"application/json",
		rabbitmq.PublishingOptions{
			Headers: amqp091.Table{
				"delivery-mode": 2,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("ошибка отправки в RabbitMQ: %v", err)
	}

	zlog.Logger.Info().
		Str("text", notification.Text).
		Int("user_id", notification.UserId).
		Time("send_time", parsedTime).
		Msg("Уведомление создано и отправлено в очередь")

	return nil
}

func (ns *NotificationService) GetNotification(id int) (*models.NotificationDB, error) {
	var notification models.NotificationDB
	ns.Mutex.Lock()
	err := ns.DB.QueryRowContext(context.Background(), "SELECT N.ID, N.TEXT, N.STATUS, N.SENDTIME, N.USERID FROM NOTIFICATIONS AS N WHERE ID = $1;", id).Scan(&notification.ID, &notification.Text, &notification.Status, &notification.TimeToSend, &notification.UserId)
	ns.Mutex.Unlock()
	if err != nil {
		return nil, fmt.Errorf("error while searching notifications")
	}
	return &notification, nil
}

func (ns *NotificationService) IsNotificationExist(id int) error {
	var exist bool
	err := ns.DB.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT N.ID, N.TEXT, N.STATUS, N.SENDTIME, N.USERID FROM NOTIFICATIONS AS N WHERE ID = $1);", id).Scan(&exist)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("notification with this id doesn`t exist")
	}
	return nil
}

func (ns *NotificationService) DeleteNotification(id int) error {
	ns.Mutex.Lock()
	_, err := ns.DB.ExecContext(context.Background(), "DELETE FROM NOTIFICATIONS WHERE ID = $1;", id)
	ns.Mutex.Unlock()
	if err != nil {
		return err
	}
	return nil
}
