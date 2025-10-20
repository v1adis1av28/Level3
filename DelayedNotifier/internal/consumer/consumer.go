package consumer

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/v1adis1av28/level3/DelayedNotifier/internal/service"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"
)

type NotificationProcessor struct {
	service        *service.NotificationService
	processedCount int
	failedCount    int
}

func NewNotificationProcessor(service *service.NotificationService) *NotificationProcessor {
	return &NotificationProcessor{
		service: service,
	}
}

func (p *NotificationProcessor) ProcessMessage(messageBody []byte) error {
	err := p.service.ProcessMessage(messageBody)
	if err != nil {
		p.failedCount++
		return fmt.Errorf("ошибка обработки уведомления: %v", err)
	}
	p.processedCount++
	return nil
}

func (p *NotificationProcessor) GetStats() (int, int) {
	return p.processedCount, p.failedCount
}

func ConsumeWithShutdown(consumer *rabbitmq.Consumer, processor *NotificationProcessor) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan []byte, 100)
	consumeErrChan := make(chan error, 1)

	go func() {
		log.Println("Starting RabbitMQ consumer...")
		if err := consumer.Consume(msgChan); err != nil {
			consumeErrChan <- fmt.Errorf("consumer error: %v", err)
		} else {
			consumeErrChan <- nil
		}
		close(consumeErrChan)
	}()

	processErrChan := make(chan error, 1)
	go func() {
		log.Println("Starting message processor...")
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
		log.Printf("Received signal: %v", sig)
		cancel()

	case err := <-consumeErrChan:
		if err != nil {
			log.Printf("Consumer error: %v", err)
		}
		cancel()

	case err := <-processErrChan:
		if err != nil {
			log.Printf("Processing error: %v", err)
		}
		cancel()
	}

	time.Sleep(1 * time.Second)
	close(msgChan)

	<-consumeErrChan
	<-processErrChan

	processed, failed := processor.GetStats()
	log.Printf("Processing stats: successful=%d, failed=%d", processed, failed)
	log.Println("Consumer stopped")
}

func processMessages(ctx context.Context, msgs <-chan []byte, processor *NotificationProcessor) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping message processing")
			return ctx.Err()

		case jsn, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				return nil
			}

			if err := processor.ProcessMessage(jsn); err != nil {
				zlog.Logger.Error().Err(err)
				continue
			}
		}
	}
}

func Consume(consumer *rabbitmq.Consumer, service *service.NotificationService) {
	processor := NewNotificationProcessor(service)
	ConsumeWithShutdown(consumer, processor)
}
