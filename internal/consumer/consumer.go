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
func (c *Consumer) Start(ctx context.Context) error {

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("stop consumer")
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.handleError("reading message error", err)
				continue
			}
			// Если не получилось обработать сообщение, то просто логируем ошибку.
			if err := c.processMessage(ctx, msg); err != nil {
				c.handleError("processing message error", err)
			}
		}
	}
}

// Обработка сообщения из кафки
func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {

	var order model.Order
	decoder := json.NewDecoder(bytes.NewReader(msg.Value))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&order); err != nil {
		return fmt.Errorf("processing message error: %w", err)
	}

	// Валидация через validator
	// Можно добавить валидацию с бизнес логикой. Например, что cost == сумме всех item
	if err := c.validate.Struct(order); err != nil {
		return fmt.Errorf("processing message error:%w", err)
	}

	if err := c.service.SaveOrder(ctx, &order); err != nil {
		return fmt.Errorf("processing message error:%w", err)
	}

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

func (c *Consumer) handleError(msg string, err error) {
	log.Printf("%s:%v", msg, err)
}
