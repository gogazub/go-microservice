package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gogazub/myapp/internal/repository/model"
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
	handler Handler
	config  Config
}

type Handler interface {
	HandleOrder(order *model.Order) error
}

func NewConsumer(config Config, handler Handler) *Consumer {
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
		handler: handler,
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

func (c *Consumer) processMessage(msg kafka.Message) error {
	log.Printf("Received message: topic=%s partition=%d offset=%d",
		msg.Topic, msg.Partition, msg.Offset)

	var order model.Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		return fmt.Errorf("failed to unmarshal order: %w", err)
	}

	log.Printf("Processing order: %s", order.OrderUID)

	if err := c.handler.HandleOrder(&order); err != nil {
		return fmt.Errorf("handler failed for order %s: %w", order.OrderUID, err)
	}

	log.Printf("Successfully processed order: %s", order.OrderUID)
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
