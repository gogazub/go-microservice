package tests

import (
	"context"
	"encoding/json"
	"sort"
	"testing"
	"time"

	"github.com/gogazub/myapp/internal/model"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func fakeValidOrder(id string) *model.Order {
	now := time.Now().UTC()

	return &model.Order{
		OrderUID:    id,
		CustomerID:  "cust-001",
		DateCreated: now,
		OofShard:    "1",
		Delivery: model.Delivery{
			Name:    "Alice",
			Phone:   "+1234567890",
			Zip:     "10001",
			City:    "NY",
			Address: "5th Avenue, 1",
			Region:  "NY",
			Email:   "alice@example.com",
		},
		Payment: model.Payment{
			Transaction:  "trx-001",
			Currency:     "USD",
			Provider:     "visa",
			Amount:       149.90,
			PaymentDt:    1712345678,
			Bank:         "Chase",
			DeliveryCost: 0,
			GoodsTotal:   1,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      1,        // required,gte=1
				TrackNumber: "TRK123", // required
				Price:       149.90,   // gte=0
				Sale:        0,        // 0..100
				TotalPrice:  149.90,   // gte=0
				NmID:        1,        // gte=0
				Status:      0,        // gte=0
			},
		},
	}
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

// Вспомогательная функция, чтобы быстро маршалить объект в []byte
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
