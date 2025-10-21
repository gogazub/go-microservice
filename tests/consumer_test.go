package tests

import (
	"context"
	"testing"

	"github.com/gogazub/myapp/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) SaveOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockService) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Order), args.Error(1)
}

type StubService struct {
	Err error
}

func (s *StubService) SaveOrder(ctx context.Context, order *model.Order) error {
	return s.Err
}

func (s *StubService) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	return nil, s.Err
}

func TestConsumer(t *testing.T) {
	// Чтобы тесты были понятны, пишем тесты по методу Arrage-Act-Assert.

	// Аноноимные функции замыкают ресурсы TestConsumer.
	// Так что здесь можно провести Arrange.

	// Закидываем битый json. SaveOrder не вызывается; возращается err == "bad json"
	t.Run("ProcessMessage/bad json", func(t *testing.T) {

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
