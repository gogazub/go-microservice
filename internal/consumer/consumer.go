package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gogazub/myapp/internal/model"
	svc "github.com/gogazub/myapp/internal/service"
	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers  []string
	Topic    string
	GroupID  string
	MinBytes int
	MaxBytes int
}

type IConsumer interface {
	rocessMessage(ctx context.Context, msg kafka.Message) error
}

type Consumer struct {
	reader   *kafka.Reader
	service  svc.Service
	config   Config
	validate *validator.Validate
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
		reader:   reader,
		service:  service,
		config:   config,
		validate: validator.New(),
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
			// Если не получилось обработать сообщение, то просто логируем ошибку.
			if err := c.processMessage(ctx, msg); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}
	}
}

// Обработка сообщения из кафки
func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {
	log.Printf("Received message: topic=%s partition=%d offset=%d",
		msg.Topic, msg.Partition, msg.Offset)

	var order model.Order
	decoder := json.NewDecoder(bytes.NewReader(msg.Value))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&order); err != nil {
		return fmt.Errorf("bad json: %w", err)
	}

	// Валидация через validator
	// Можно добавить валидацию с бизнес логикой. Например, что cost == сумме всех item
	if err := c.validate.Struct(order); err != nil {
		return err
	}

	log.Printf("Processing order: %s", order.OrderUID)

	if err := c.service.SaveOrder(ctx, &order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	log.Printf("Successfully processed and saved order: %s", order.OrderUID)
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) ProcessMessageTest(ctx context.Context, msg kafka.Message) error {
	return c.processMessage(ctx, msg)
}
