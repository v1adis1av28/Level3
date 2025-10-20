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

var (
	CREATED_STATUS   string = "scheduled"
	SENT_STATUS      string = "sent"
	FAILED_STATUS    string = "failed"
	RETRYING_STATUS  string = "retrying"
	CANCELLED_STATUS string = "cancelled"
)

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

	err = SetupDelayedQueues(ch)
	if err != nil {
		zlog.Logger.Err(err).Msg("Failed to declare queue")
		ch.Close()
		conn.Close()
		return nil
	}

	publisher := rabbitmq.NewPublisher(ch, "notifications_exchange")

	service := &NotificationService{
		DB:        db,
		Producer:  ch,
		Publisher: publisher,
		Conn:      conn,
		Mutex:     &sync.Mutex{},
	}
	return service
}

func SetupDelayedQueues(ch *rabbitmq.Channel) error {
	err := ch.ExchangeDeclare(
		"notifications_exchange",
		"direct",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(
		"notifications_dlx",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		"notifications_queue",
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		amqp091.Table{
			"x-dead-letter-exchange":    "notifications_dlx",
			"x-dead-letter-routing-key": "retry",
		},
	)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		"notifications_retry",
		true,
		false,
		false,
		false,
		amqp091.Table{
			"x-dead-letter-exchange":    "notifications_exchange",
			"x-dead-letter-routing-key": "notifications",
			"x-message-ttl":             30000,
		},
	)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		"notifications_queue",
		"notifications",
		"notifications_exchange",
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		"notifications_retry",
		"retry",
		"notifications_dlx",
		false,
		nil,
	)

	return err
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
	var parsedTime time.Time
	var err error

	if duration, err := time.ParseDuration(notification.TimeToSend); err == nil {
		parsedTime = time.Now().UTC().Add(duration)
		zlog.Logger.Info().
			Str("duration", notification.TimeToSend).
			Time("calculated_time", parsedTime).
			Msg("Время рассчитано из duration")
	} else {
		parsedTime, err = time.Parse(time.RFC3339, notification.TimeToSend)
		if err != nil {
			parsedTime, err = time.Parse("2006-01-02 15:04:05", notification.TimeToSend)
			if err != nil {
				parsedTime, err = time.Parse("2006-01-02T15:04:05", notification.TimeToSend)
				if err != nil {
					return fmt.Errorf("неверный формат времени: %v. Используйте RFC3339 или duration (5m, 1h)", err)
				}
			}
		}
		parsedTime = parsedTime.UTC()
	}

	now := time.Now().UTC()
	parsedTime = parsedTime.Add(time.Hour * -3)

	if parsedTime.Before(now) {
		return fmt.Errorf("send time can`t be in past")
	}
	if notification.Text == "" {
		return fmt.Errorf("notification text can`t be empty")
	}

	var notificationID int
	err = ns.DB.QueryRowContext(
		context.Background(),
		`INSERT INTO NOTIFICATIONS (TEXT, STATUS, SEND_TIME, USER_ID, CREATED_AT, RETRY_COUNT, MAX_RETRIES) 
		 VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING ID`,
		notification.Text,
		CREATED_STATUS,
		parsedTime,
		notification.UserId,
		time.Now(),
		0,
		3,
	).Scan(&notificationID)
	if err != nil {
		return fmt.Errorf("ошибка сохранения в БД: %v", err)
	}

	msg := models.NotificationMessage{
		NotificationID: notificationID,
		UserID:         notification.UserId,
		Text:           notification.Text,
		SendTime:       parsedTime,
		Attempt:        1,
		MaxAttempts:    3,
	}

	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("json marshaling error: %v", err)
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
		return fmt.Errorf("sending in RabbitMQ error: %v", err)
	}

	zlog.Logger.Info().
		Int("id", notificationID).
		Str("text", notification.Text).
		Int("user_id", notification.UserId).
		Time("send_time", parsedTime).
		Msg("notification created and sent in queue")

	return nil
}

func (ns *NotificationService) GetNotification(id int) (*models.NotificationDB, error) {
	var notification models.NotificationDB
	ns.Mutex.Lock()
	defer ns.Mutex.Unlock()

	err := ns.DB.QueryRowContext(
		context.Background(),
		"SELECT ID, TEXT, STATUS, SEND_TIME, USER_ID, CREATED_AT,  RETRY_COUNT, MAX_RETRIES FROM NOTIFICATIONS WHERE ID = $1",
		id,
	).Scan(
		&notification.ID,
		&notification.Text,
		&notification.Status,
		&notification.TimeToSend,
		&notification.UserId,
		&notification.CreatedAt,
		&notification.RetryCount,
		&notification.MaxRetries,
	)
	if err != nil {
		return nil, fmt.Errorf("error while searching notifications: %v", err)
	}
	return &notification, nil
}

func (ns *NotificationService) IsNotificationExist(id int) error {
	var exist bool
	err := ns.DB.QueryRowContext(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM NOTIFICATIONS WHERE ID = $1)",
		id,
	).Scan(&exist)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("notification with this id doesn't exist")
	}
	return nil
}

func (ns *NotificationService) DeleteNotification(id int) error {
	ns.Mutex.Lock()
	defer ns.Mutex.Unlock()

	_, err := ns.DB.ExecContext(
		context.Background(),
		"UPDATE NOTIFICATIONS SET STATUS = $1 WHERE ID = $2",
		CANCELLED_STATUS,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}
