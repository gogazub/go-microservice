// Package consumer читает kafka, сохраняет валидные сообщения
package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gogazub/myapp/internal/model"
	svc "github.com/gogazub/myapp/internal/service"
	"github.com/segmentio/kafka-go"
)

// Config конфигурация consumer для подключения к kafka.
type Config struct {
	Brokers  []string
	Topic    string
	GroupID  string
	MinBytes int
	MaxBytes int
}

// IConsumer никуда ни инъектируется, так что интерфейс не нужен
// type IConsumer interface {
// 	processMessage(ctx context.Context, msg kafka.Message) error
//}

// IReader вынесен в интерфейс, для корректного DI. Это даст нам возможности для фейков при тестировании
type IReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

// Consumer получает сообщения из reader; валидирует их через validate; передает валидные сообщения в service
type Consumer struct {
	reader   IReader
	service  svc.IService
	config   Config
	validate *validator.Validate
}

// NewConsumer конструктор
func NewConsumer(service svc.IService, reader IReader) *Consumer {

	return &Consumer{
		reader:   reader,
		service:  service,
		validate: validator.New(),
	}
}

// Start запускает consumer; Начинает обрабатывать сообщения.
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
		return fmt.Errorf("process message error: %w", err)
	}

	log.Printf("Successfully processed and saved order: %s", order.OrderUID)
	return nil
}

// Close Закрывает подключение с kafka
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// ProcessMessageTest экспортируемый метод для тестирования Consumer`а
func (c *Consumer) ProcessMessageTest(ctx context.Context, msg kafka.Message) error {
	return c.processMessage(ctx, msg)
}
