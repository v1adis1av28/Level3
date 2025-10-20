package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/v1adis1av28/level3/DelayedNotifier/internal/models"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"
)

func (ns *NotificationService) ProcessMessage(messageBody []byte) error {
	var msg models.NotificationMessage
	if err := json.Unmarshal(messageBody, &msg); err != nil {
		return fmt.Errorf("unmarshalimn json error: %v", err)
	}

	msg.SendTime = msg.SendTime.UTC()
	now := time.Now().UTC()

	var status string
	err := ns.DB.QueryRowContext(
		context.Background(),
		"SELECT STATUS FROM NOTIFICATIONS WHERE ID = $1",
		msg.NotificationID,
	).Scan(&status)
	if err != nil {
		return fmt.Errorf("checking status error: %v", err)
	}

	if status == CANCELLED_STATUS {
		return nil
	}

	if now.Before(msg.SendTime.Add(-1 * time.Second)) {
		timeUntilSend := time.Until(msg.SendTime)
		zlog.Logger.Info().
			Int("id", msg.NotificationID).
			Dur("time_until_send", timeUntilSend).
			Msg("time to send not come on")

		delay := 5 * time.Second
		if timeUntilSend < delay {
			delay = timeUntilSend
		}
		return ns.RequeueWithDelay(&msg, delay)
	}

	zlog.Logger.Info().
		Int("id", msg.NotificationID).
		Int("attempt", msg.Attempt).
		Time("send_time", msg.SendTime).
		Time("now", now).
		Msg("message succesfully sent")

	err = ns.SendNotification(&msg)
	if err != nil {
		zlog.Logger.Warn().
			Int("id", msg.NotificationID).
			Int("attempt", msg.Attempt).
			Err(err).
			Msg("error on sending notification")

		msg.Attempt++
		if msg.Attempt <= msg.MaxAttempts {
			return ns.scheduleRetry(&msg)
		} else {
			return ns.markAsFailed(msg.NotificationID)
		}
	}

	return ns.markAsSent(msg.NotificationID)
}

func (ns *NotificationService) SendNotification(msg *models.NotificationMessage) error {
	if time.Now().Unix()%5 == 0 {
		return fmt.Errorf("временная ошибка отправки")
	}

	zlog.Logger.Info().
		Int("id", msg.NotificationID).
		Int("user_id", msg.UserID).
		Int("attempt", msg.Attempt).
		Msg("notifications was sent")

	return nil
}

func (ns *NotificationService) GetRetryDelay(attempt int) time.Duration {
	delay := time.Duration(math.Pow(2, float64(attempt-1))) * 30 * time.Second
	if delay > 10*time.Minute {
		delay = 10 * time.Minute
	}
	return delay
}

func (ns *NotificationService) scheduleRetry(msg *models.NotificationMessage) error {
	retryDelay := ns.GetRetryDelay(msg.Attempt)

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

func (ns *NotificationService) RequeueWithDelay(msg *models.NotificationMessage, delay time.Duration) error {
	messageBody, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return ns.Publisher.Publish(
		messageBody,
		"notifications",
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
