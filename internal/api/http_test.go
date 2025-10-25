package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gogazub/myapp/internal/model"
	"github.com/gogazub/myapp/tests"
)

type mockService struct{ mock.Mock }

func (m *mockService) SaveOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}
func (m *mockService) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	args := m.Called(ctx, id)
	var o *model.Order
	if v := args.Get(0); v != nil {
		o = v.(*model.Order)
	}
	return o, args.Error(1)
}

// ---- handleGetOrderByID ----

func TestHandleGetOrderByID_Success(t *testing.T) {
	ms := new(mockService)
	s := NewServer(ms)

	want := tests.FakeOrder("uid-1")

	ms.On("GetOrderByID", mock.Anything, "uid-1").
		Return(want, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/orders/uid-1", nil)
	rr := httptest.NewRecorder()
	s.handleGetOrderByID(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var got model.Order
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
	require.Equal(t, want.OrderUID, got.OrderUID)

	ms.AssertExpectations(t)
}

func TestHandleGetOrderByID_NotFound(t *testing.T) {
	ms := new(mockService)
	s := NewServer(ms)

	ms.On("GetOrderByID", mock.Anything, "missing").
		Return((*model.Order)(nil), assertAnError()).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/orders/missing", nil)
	rr := httptest.NewRecorder()

	s.handleGetOrderByID(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	ms.AssertExpectations(t)
}

func TestHandleGetOrderByID_EmptyID(t *testing.T) {
	ms := new(mockService)
	s := NewServer(ms)

	ms.On("GetOrderByID", mock.Anything, "").
		Return((*model.Order)(nil), assertAnError()).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/orders/", nil)
	rr := httptest.NewRecorder()

	s.handleGetOrderByID(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	ms.AssertExpectations(t)
}

func assertAnError() error { return errAny }

var errAny = &anyError{}

type anyError struct{}

func (e *anyError) Error() string { return "any error" }
