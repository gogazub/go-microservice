package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gogazub/myapp/internal/model"
	svc "github.com/gogazub/myapp/internal/service"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers  []string
	Topic    string
	GroupID  string
	MinBytes int
	MaxBytes int
}

type Consumer struct {
	reader  *kafka.Reader
	service svc.Service
	config  Config
}

func NewConsumer(config Config, service svc.Service) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Brokers,
		Topic:    config.Topic,
		GroupID:  config.GroupID,
		MinBytes: config.MinBytes,
		MaxBytes: config.MaxBytes,
		MaxWait:  1 * time.Second,
	})

	return &Consumer{
		reader:  reader,
		service: service,
		config:  config,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Printf("Starting Kafka consumer for topic: %s", c.config.Topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping consumer...")
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			if err := c.processMessage(msg); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}
	}
}

// Обработка сообщения из кафки
func (c *Consumer) processMessage(msg kafka.Message) error {
	log.Printf("Received message: topic=%s partition=%d offset=%d",
		msg.Topic, msg.Partition, msg.Offset)

	var order model.Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		return fmt.Errorf("failed to unmarshal order: %w", err)
	}

	// Валидация UUID для OrderUID
	if _, err := uuid.Parse(order.OrderUID); err != nil {
		return fmt.Errorf("invalid UUID format for OrderUID: %s, error: %w", order.OrderUID, err)
	}
	log.Printf("Processing order: %s", order.OrderUID)

	if err := c.service.SaveOrder(&order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	log.Printf("Successfully processed and saved order: %s", order.OrderUID)
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
