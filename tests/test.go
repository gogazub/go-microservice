package tests

import (
	"context"
	"sort"

	"github.com/gogazub/myapp/internal/model"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

// В этом файле хранятся общие части для всех тестов.

// --- MockService ---
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

// --- StubService ---
type StubService struct {
	Err error
}

func (s *StubService) SaveOrder(ctx context.Context, order *model.Order) error {
	return s.Err
}
func (s *StubService) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	return nil, s.Err
}

// --- StubReader ---

type StubReader struct {
	msg kafka.Message
	err error
}

func (s *StubReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return s.msg, s.err
}
func (s *StubReader) Close() error {
	return s.err
}

// --- utinls ---

func fakeOrder(id string) *model.Order {
	return &model.Order{OrderUID: id}
}

func idsFromOrders(orders []*model.Order) []string {
	out := make([]string, 0, len(orders))
	for _, o := range orders {
		out = append(out, o.OrderUID)
	}
	sort.Strings(out)
	return out
}

// небольшая утилита, чтобы не тянуть strconv
func strconvI(i int) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	var b [32]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = digits[i%10]
		i /= 10
	}
	return string(b[pos:])
}
