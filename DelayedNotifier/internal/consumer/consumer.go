package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/v1adis1av28/level3/DelayedNotifier/internal/models"
	"github.com/wb-go/wbf/rabbitmq"
)

type NotificationProcessor struct {
	processedCount int
	failedCount    int
}

func (p *NotificationProcessor) Process(notification models.Notification) error {
	fmt.Printf("processed: UserID=%d, Text='%s', Time=%s\n",
		notification.UserId, notification.Text, notification.TimeToSend)

	p.processedCount++

	return nil
}

func ConsumeWithShutdown(consumer *rabbitmq.Consumer, processor *NotificationProcessor) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan []byte, 100)

	consumeErrChan := make(chan error, 1)

	go func() {
		if err := consumer.Consume(msgChan); err != nil {
			consumeErrChan <- fmt.Errorf("consumer error: %v", err)
		} else {
			consumeErrChan <- nil
		}
		close(consumeErrChan)
	}()

	processErrChan := make(chan error, 1)
	go func() {
		if err := processMessages(ctx, msgChan, processor); err != nil {
			processErrChan <- err
		} else {
			processErrChan <- nil
		}
		close(processErrChan)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("get signal %v", sig)
		cancel()

	case err := <-consumeErrChan:
		if err != nil {
			log.Printf("consumer error: %v", err)
		}
		cancel()

	case err := <-processErrChan:
		if err != nil {
			log.Printf("processing error: %v", err)
		}
		cancel()
	}

	time.Sleep(1 * time.Second)

	close(msgChan)

	<-consumeErrChan
	<-processErrChan
}

func processMessages(ctx context.Context, msgs <-chan []byte, processor *NotificationProcessor) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case jsn, ok := <-msgs:
			if !ok {
				return nil
			}
			if err := processSingleMessage(jsn, processor); err != nil {
				processor.failedCount++
				continue
			}
		}
	}
}

func processSingleMessage(jsn []byte, processor *NotificationProcessor) error {
	var notification models.Notification
	if err := json.Unmarshal(jsn, &notification); err != nil {
		return fmt.Errorf("json parsing error: %v", err)
	}

	if err := processor.Process(notification); err != nil {
		return fmt.Errorf("ошибка обработки уведомления: %v", err)
	}

	return nil
}

func Consume(consumer *rabbitmq.Consumer) {
	processor := &NotificationProcessor{}
	ConsumeWithShutdown(consumer, processor)
}
