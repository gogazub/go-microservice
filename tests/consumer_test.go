package tests

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gogazub/myapp/internal/consumer"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Tests ---
func TestConsumer(t *testing.T) {
	// Чтобы тесты были понятны, пишем тесты по методу Arrage-Act-Assert.

	// Анонимные функции замыкают ресурсы TestConsumer.
	// Так что здесь можно провести Arrange.

	// Arrange
	stubReader := &StubReader{}
	stubService := &StubService{}
	//mockService := &MockService{}
	c := consumer.NewConsumer(stubService, stubReader)
	defer func() {
		err := c.Close()
		if err != nil {
			log.Printf("consumer close error:%s", err.Error())
		}
	}()
	// Закидываем битый json. SaveOrder не вызывается; возращается err == "bad json"
	t.Run("ProcessMessage/bad json", func(t *testing.T) {
		// Arrange
		invalidJSON := kafka.Message{Value: []byte(`{"OrderUID":123`)}
		stubReader.msg = invalidJSON
		stubReader.err = nil

		// Act
		err := c.ProcessMessageTest(context.Background(), invalidJSON)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad json")

	})

	// Можно использовать validator без Mock, потому что валидатор создается в конструкторе NewConsumer
	// Нарушение правил валидации. Возвращает ValidationErrors. SaveOrder не вызывается
	t.Run("ProcessMessage/validate error", func(t *testing.T) {

		mockSvc := &MockService{}
		c = consumer.NewConsumer(mockSvc, stubReader)
		fakeOrder := fakeOrder("123")
		fakeOrder.SmID = -1

		marshaled, err := json.Marshal(fakeOrder)
		require.NoError(t, err)

		invalidMsg := kafka.Message{Value: marshaled}

		err = c.ProcessMessageTest(context.Background(), invalidMsg)
		var verr validator.ValidationErrors
		assert.ErrorAs(t, err, &verr)
		mockSvc.AssertNotCalled(t, "SaveOrder")
	})

	// Json с лишними полями. возвращает ошибку, SaveOrder не вызывается
	t.Run("ProcessMessage/unknown fields", func(t *testing.T) {
		mockSvc := &MockService{}
		c = consumer.NewConsumer(mockSvc, stubReader)

		fakeOrder := fakeValidOrder("123")

		var orderMap map[string]interface{}
		require.NoError(t, json.Unmarshal(mustJSON(t, fakeOrder), &orderMap))

		orderMap["extra_field"] = "unexpected"
		marshaled, err := json.Marshal(orderMap)
		require.NoError(t, err)

		invalidMsg := kafka.Message{Value: marshaled}

		err = c.ProcessMessageTest(context.Background(), invalidMsg)

		require.Error(t, err)
		require.Contains(t, err.Error(), `unknown field "extra_field"`)
		mockSvc.AssertNotCalled(t, "SaveOrder")
	})

	// Пустой json. Возвращает ошибку. SaveOrder не вызывается
	t.Run("ProcessMessage/empty json", func(t *testing.T) {
		mockSvc := &MockService{}
		c = consumer.NewConsumer(mockSvc, stubReader)

		var orderMap map[string]interface{}
		marshaled, err := json.Marshal(orderMap)
		require.NoError(t, err)

		invalidMsg := kafka.Message{Value: marshaled}

		err = c.ProcessMessageTest(context.Background(), invalidMsg)
		require.Error(t, err)
		mockSvc.AssertNotCalled(t, "SaveOrder")

	})

	// Даем корректный json. Возвращает nil. SaveOrder вызывается один раз
	t.Run("ProcessMessage/correct json", func(t *testing.T) {
		mockSvc := &MockService{}
		c = consumer.NewConsumer(mockSvc, stubReader)

		order := fakeValidOrder("1")
		marshaled, err := json.Marshal(order)
		require.NoError(t, err)

		// ожидаем один успешный вызов сохранения
		mockSvc.
			On("SaveOrder", mock.Anything, mock.AnythingOfType("*model.Order")).
			Return(nil).
			Once()

		msg := kafka.Message{Value: marshaled}

		err = c.ProcessMessageTest(context.Background(), msg)
		require.NoError(t, err)

		mockSvc.AssertExpectations(t)
		mockSvc.AssertNumberOfCalls(t, "SaveOrder", 1)
	})

	// SaveOrder возвращает ошибку. Возвращает ошибку.
	t.Run("ProcessMessage/SaveOrder return error", func(t *testing.T) {
		mockSvc := &MockService{}
		c := consumer.NewConsumer(mockSvc, stubReader)

		order := fakeValidOrder("1")
		marshaled, err := json.Marshal(order)
		require.NoError(t, err)

		mockSvc.
			On("SaveOrder", mock.Anything, mock.AnythingOfType("*model.Order")).
			Return(errors.New("db error")).
			Once()

		msg := kafka.Message{Value: marshaled}
		err = c.ProcessMessageTest(context.Background(), msg)

		require.Error(t, err)
		require.Contains(t, err.Error(), "db error")
		mockSvc.AssertExpectations(t)
		mockSvc.AssertNumberOfCalls(t, "SaveOrder", 1)
	})

	// SaveOrder возвращает nil. Возвращает nil
	t.Run("ProcessMessage/SaveOrder return nil", func(t *testing.T) {
		mockSvc := &MockService{}
		c := consumer.NewConsumer(mockSvc, stubReader)

		order := fakeValidOrder("2")
		marshaled, err := json.Marshal(order)
		require.NoError(t, err)

		mockSvc.
			On("SaveOrder", mock.Anything, mock.AnythingOfType("*model.Order")).
			Return(nil).
			Once()

		msg := kafka.Message{Value: marshaled}
		err = c.ProcessMessageTest(context.Background(), msg)

		require.NoError(t, err)
		mockSvc.AssertExpectations(t)
		mockSvc.AssertNumberOfCalls(t, "SaveOrder", 1)
	})

}
