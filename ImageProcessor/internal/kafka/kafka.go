package kafka

import (
	"log"

	"github.com/IBM/sarama"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/processor"
)

type Producer struct {
	producer sarama.SyncProducer
}

func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{producer: producer}, nil
}

func (p *Producer) SendMessage(topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Message sent to partition %d at offset %d\n", partition, offset)
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

type Consumer struct {
	consumer  sarama.Consumer
	processor *processor.ImageProcessor
}

func NewConsumer(brokers []string, processor *processor.ImageProcessor) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer:  consumer,
		processor: processor,
	}, nil
}

func (c *Consumer) StartConsuming(topic string) error {
	partitionConsumer, err := c.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				log.Printf("Received message from Kafka: %s\n", string(msg.Value))
				err := c.processor.ProcessImage(msg.Value)
				if err != nil {
					log.Printf("Error processing image: %v\n", err)
				}
			case err := <-partitionConsumer.Errors():
				log.Printf("Error from Kafka consumer: %v\n", err)
			}
		}
	}()

	return nil
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
