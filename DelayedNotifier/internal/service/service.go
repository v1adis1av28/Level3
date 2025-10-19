package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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

	return &NotificationService{
		DB:        db,
		Producer:  ch,
		Publisher: publisher,
		Conn:      conn,
		Mutex:     &sync.Mutex{},
	}
}

func SetupDelayedQueues(ch *rabbitmq.Channel) error {
	// Основной обменник
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

	// DLX для повторных попыток
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

	// Основная очередь
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

	// Очередь для повторных попыток
	_, err = ch.QueueDeclare(
		"notifications_retry",
		true,
		false,
		false,
		false,
		amqp091.Table{
			"x-dead-letter-exchange":    "notifications_exchange",
			"x-dead-letter-routing-key": "notifications",
			"x-message-ttl":             30000, // 30 секунд для первой попытки
		},
	)
	if err != nil {
		return err
	}

	// Привязки
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

	// Сохраняем в БД
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
		3, // Максимум 3 попытки
	).Scan(&notificationID)
	if err != nil {
		return fmt.Errorf("ошибка сохранения в БД: %v", err)
	}

	// Создаем сообщение для RabbitMQ
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
		return fmt.Errorf("ошибка сериализации уведомления: %v", err)
	}

	// Рассчитываем задержку
	delay := time.Until(parsedTime)
	if delay < 0 {
		delay = 0
	}

	// Публикуем в RabbitMQ
	err = ns.Publisher.Publish(
		messageBody,
		"notifications_queue", // routing key
		"application/json",
		rabbitmq.PublishingOptions{
			Headers: amqp091.Table{
				"delivery-mode": 2, // persistent
			},
		},
	)
	if err != nil {
		return fmt.Errorf("ошибка отправки в RabbitMQ: %v", err)
	}

	zlog.Logger.Info().
		Int("id", notificationID).
		Str("text", notification.Text).
		Int("user_id", notification.UserId).
		Time("send_time", parsedTime).
		Msg("Уведомление создано и отправлено в очередь")

	return nil
}

// ProcessMessage обрабатывает сообщение из RabbitMQ
func (ns *NotificationService) ProcessMessage(messageBody []byte) error {
	var msg models.NotificationMessage
	if err := json.Unmarshal(messageBody, &msg); err != nil {
		return fmt.Errorf("ошибка парсинга сообщения: %v", err)
	}

	// Проверяем, не отменено ли уведомление
	var status string
	err := ns.DB.QueryRowContext(
		context.Background(),
		"SELECT STATUS FROM NOTIFICATIONS WHERE ID = $1",
		msg.NotificationID,
	).Scan(&status)
	if err != nil {
		return fmt.Errorf("ошибка проверки статуса: %v", err)
	}

	if status == CANCELLED_STATUS {
		zlog.Logger.Info().Int("id", msg.NotificationID).Msg("Уведомление отменено, пропускаем")
		return nil
	}

	// Проверяем время отправки
	if time.Now().Before(msg.SendTime) {
		// Еще рано, возвращаем в очередь
		return ns.requeueWithDelay(&msg, time.Until(msg.SendTime))
	}

	// Пытаемся отправить уведомление
	err = ns.sendNotification(&msg)
	if err != nil {
		zlog.Logger.Warn().
			Int("id", msg.NotificationID).
			Int("attempt", msg.Attempt).
			Err(err).
			Msg("Ошибка отправки уведомления")

		// Планируем повторную попытку
		msg.Attempt++
		if msg.Attempt <= msg.MaxAttempts {
			return ns.scheduleRetry(&msg)
		} else {
			// Превышено максимальное количество попыток
			return ns.markAsFailed(msg.NotificationID)
		}
	}

	// Успешная отправка
	return ns.markAsSent(msg.NotificationID)
}

// sendNotification имитирует отправку уведомления
func (ns *NotificationService) sendNotification(msg *models.NotificationMessage) error {
	// Здесь должна быть реальная логика отправки:
	// - Email
	// - SMS
	// - Push-уведомление
	// - Webhook и т.д.

	// Имитация отправки (в 20% случаев возвращаем ошибку для тестирования повторов)
	if time.Now().Unix()%5 == 0 {
		return fmt.Errorf("временная ошибка отправки")
	}

	zlog.Logger.Info().
		Int("id", msg.NotificationID).
		Int("user_id", msg.UserID).
		Int("attempt", msg.Attempt).
		Msg("Уведомление успешно отправлено")

	return nil
}

// getRetryDelay вычисляет задержку для повторной попытки (экспоненциальная)
func (ns *NotificationService) getRetryDelay(attempt int) time.Duration {
	delay := time.Duration(math.Pow(2, float64(attempt-1))) * 30 * time.Second
	if delay > 10*time.Minute {
		delay = 10 * time.Minute
	}
	return delay
}

func (ns *NotificationService) scheduleRetry(msg *models.NotificationMessage) error {
	retryDelay := ns.getRetryDelay(msg.Attempt)

	// Обновляем статус в БД
	_, err := ns.DB.ExecContext(
		context.Background(),
		`UPDATE NOTIFICATIONS SET STATUS = $1, RETRY_COUNT = $2, NEXT_RETRY = $3 WHERE ID = $4`,
		RETRYING_STATUS,
		msg.Attempt,
		time.Now().Add(retryDelay),
		msg.NotificationID,
	)
	if err != nil {
		return err
	}

	// Публикуем сообщение для повторной попытки
	messageBody, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return ns.Publisher.Publish(
		messageBody,
		"notifications_queue",
		"application/json",
		rabbitmq.PublishingOptions{
			Expiration: retryDelay,
			Headers: amqp091.Table{
				"delivery-mode": 2,
			},
		},
	)
}

func (ns *NotificationService) requeueWithDelay(msg *models.NotificationMessage, delay time.Duration) error {
	messageBody, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return ns.Publisher.Publish(
		messageBody,
		"notifications_queue",
		"application/json",
		rabbitmq.PublishingOptions{
			Expiration: delay,
			Headers: amqp091.Table{
				"delivery-mode": 2,
			},
		},
	)
}

func (ns *NotificationService) markAsSent(notificationID int) error {
	_, err := ns.DB.ExecContext(
		context.Background(),
		"UPDATE NOTIFICATIONS SET STATUS = $1, SENT_AT = $2 WHERE ID = $3",
		SENT_STATUS,
		time.Now(),
		notificationID,
	)
	return err
}

func (ns *NotificationService) markAsFailed(notificationID int) error {
	_, err := ns.DB.ExecContext(
		context.Background(),
		"UPDATE NOTIFICATIONS SET STATUS = $1 WHERE ID = $2",
		FAILED_STATUS,
		notificationID,
	)
	return err
}

func (ns *NotificationService) GetNotification(id int) (*models.NotificationDB, error) {
	var notification models.NotificationDB
	ns.Mutex.Lock()
	defer ns.Mutex.Unlock()

	err := ns.DB.QueryRowContext(
		context.Background(),
		"SELECT ID, TEXT, STATUS, SEND_TIME, USER_ID, CREATED_AT, SENT_AT, RETRY_COUNT, MAX_RETRIES, NEXT_RETRY FROM NOTIFICATIONS WHERE ID = $1",
		id,
	).Scan(
		&notification.ID,
		&notification.Text,
		&notification.Status,
		&notification.TimeToSend,
		&notification.UserId,
		&notification.CreatedAt,
		&notification.SentAt,
		&notification.RetryCount,
		&notification.MaxRetries,
		&notification.NextRetry,
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
