package tests

import (
	"context"
	"testing"

	"github.com/gogazub/myapp/internal/consumer"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

// --- Tests ---
func TestConsumer(t *testing.T) {
	// Чтобы тесты были понятны, пишем тесты по методу Arrage-Act-Assert.

	// Аноноимные функции замыкают ресурсы TestConsumer.
	// Так что здесь можно провести Arrange.

	// Arrange
	stubReader := &StubReader{}
	stubService := &StubService{}
	//mockService := &MockService{}
	consumer := consumer.NewConsumer(stubService, stubReader)
	// Закидываем битый json. SaveOrder не вызывается; возращается err == "bad json"
	t.Run("ProcessMessage/bad json", func(t *testing.T) {
		// Arrange
		invalidJSON := kafka.Message{Value: []byte(`{"OrderUID":123`)}
		stubReader.msg = invalidJSON
		stubReader.err = nil

		// Act
		err := consumer.ProcessMessageTest(context.Background(), invalidJSON)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad json")

	})

	// Можно использовать validator без Mock, потому что валидатор создается в конструкторе NewConsumer

	// Нарушение правил валидации. Возвращает ValidationErrors. SaveOrder не вызывается
	t.Run("ProcessMessage/validate error", func(t *testing.T) {

	})

	// Json с лишними полями. возвращает ошибку, SaveOrder не вызывается
	t.Run("ProcessMessage/unknown fields", func(t *testing.T) {

	})

	// Пустой json. Возвращает ошибку. SaveOrder не вызывается
	t.Run("ProcessMessage/empty json", func(t *testing.T) {

	})

	// Даем корректный json. Возвращает nil. SaveOrder вызывается один раз
	t.Run("ProcessMessage/correct json", func(t *testing.T) {

	})

	// SaveOrder возвращает ошибку. Возвращает ошибку.
	t.Run("ProcessMessage/SaveOrder return error", func(t *testing.T) {

	})

	// SaveOrder возвращает nil. Возвращает nil
	t.Run("ProcessMessage/SaveOrder return nil", func(t *testing.T) {

	})

}
