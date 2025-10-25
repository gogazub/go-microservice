package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gogazub/myapp/internal/model"
	"github.com/gogazub/myapp/internal/service"
)

type mockDBRepo struct{ mock.Mock }

func (m *mockDBRepo) Save(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockDBRepo) GetByID(ctx context.Context, id string) (*model.Order, error) {
	args := m.Called(ctx, id)
	var o *model.Order
	if v := args.Get(0); v != nil {
		o = v.(*model.Order)
	}
	return o, args.Error(1)
}

func (m *mockDBRepo) GetAll(ctx context.Context) ([]*model.Order, error) {
	args := m.Called(ctx)
	var out []*model.Order
	if v := args.Get(0); v != nil {
		out = v.([]*model.Order)
	}
	return out, args.Error(1)
}

type mockCacheRepo struct{ mock.Mock }

func (m *mockCacheRepo) Save(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockCacheRepo) GetByID(ctx context.Context, id string) (*model.Order, error) {
	args := m.Called(ctx, id)
	var o *model.Order
	if v := args.Get(0); v != nil {
		o = v.(*model.Order)
	}
	return o, args.Error(1)
}

// ---------- SaveOrder ----------

func TestService_SaveOrder_success(t *testing.T) {
	db := new(mockDBRepo)
	cache := new(mockCacheRepo)
	s := service.NewService(db, cache)
	ctx := context.Background()
	o := FakeOrder("uid-1")

	mock.InOrder(
		db.On("Save", ctx, o).Return(nil).Once(),
		cache.On("Save", ctx, o).Return(nil).Once(),
	)

	err := s.SaveOrder(ctx, o)
	require.NoError(t, err)

	db.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestService_SaveOrder_dbError(t *testing.T) {
	db := new(mockDBRepo)
	cache := new(mockCacheRepo)
	s := service.NewService(db, cache)
	ctx := context.Background()
	o := FakeOrder("uid-1")

	db.On("Save", ctx, o).Return(errors.New("db fail")).Once()

	err := s.SaveOrder(ctx, o)
	require.Error(t, err)

	cache.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	db.AssertExpectations(t)
}

func TestService_SaveOrder_cacheError(t *testing.T) {
	db := new(mockDBRepo)
	cache := new(mockCacheRepo)
	s := service.NewService(db, cache)
	ctx := context.Background()
	o := FakeOrder("uid-1")

	mock.InOrder(
		db.On("Save", ctx, o).Return(nil).Once(),
		cache.On("Save", ctx, o).Return(errors.New("cache fail")).Once(),
	)

	err := s.SaveOrder(ctx, o)
	require.Error(t, err)

	db.AssertExpectations(t)
	cache.AssertExpectations(t)
}

// ---------- GetOrderByID ----------

func TestService_GetOrderByID_cacheHit(t *testing.T) {
	db := new(mockDBRepo)
	cache := new(mockCacheRepo)
	s := service.NewService(db, cache)
	ctx := context.Background()
	o := FakeOrder("uid-2")

	cache.On("GetByID", ctx, o.OrderUID).Return(o, nil).Once()

	got, err := s.GetOrderByID(ctx, o.OrderUID)
	require.NoError(t, err)
	require.Equal(t, o, got)

	db.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
	cache.AssertExpectations(t)
}

func TestService_GetOrderByID_cacheMiss_thenDBHit_thenCacheSave_ok(t *testing.T) {
	db := new(mockDBRepo)
	cache := new(mockCacheRepo)
	s := service.NewService(db, cache)
	ctx := context.Background()
	o := FakeOrder("uid-3")

	cache.On("GetByID", ctx, o.OrderUID).Return((*model.Order)(nil), errors.New("miss")).Once()
	db.On("GetByID", ctx, o.OrderUID).Return(o, nil).Once()
	cache.On("Save", ctx, o).Return(nil).Once()

	got, err := s.GetOrderByID(ctx, o.OrderUID)
	require.NoError(t, err)
	require.Equal(t, o, got)

	db.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestService_GetOrderByID_cacheMiss_thenDBHit_thenCacheSave_errorIgnored(t *testing.T) {
	db := new(mockDBRepo)
	cache := new(mockCacheRepo)
	s := service.NewService(db, cache)
	ctx := context.Background()
	o := FakeOrder("uid-4")

	cache.On("GetByID", ctx, o.OrderUID).Return((*model.Order)(nil), errors.New("miss")).Once()
	db.On("GetByID", ctx, o.OrderUID).Return(o, nil).Once()
	cache.On("Save", ctx, o).Return(errors.New("cache down")).Once()

	got, err := s.GetOrderByID(ctx, o.OrderUID)
	require.NoError(t, err)
	require.Equal(t, o, got)

	db.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestService_GetOrderByID_cacheMiss_andDBError(t *testing.T) {
	db := new(mockDBRepo)
	cache := new(mockCacheRepo)
	s := service.NewService(db, cache)
	ctx := context.Background()
	id := "uid-5"

	cache.On("GetByID", ctx, id).Return((*model.Order)(nil), errors.New("miss")).Once()
	db.On("GetByID", ctx, id).Return((*model.Order)(nil), errors.New("db not found")).Once()

	got, err := s.GetOrderByID(ctx, id)
	require.Nil(t, got)
	require.Error(t, err)

	cache.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	db.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestService_GetOrderByID_dbReturnsNilNoError(t *testing.T) {
	db := new(mockDBRepo)
	cache := new(mockCacheRepo)
	s := service.NewService(db, cache)
	ctx := context.Background()
	id := "uid-6"

	cache.On("GetByID", ctx, id).Return((*model.Order)(nil), errors.New("miss")).Once()
	db.On("GetByID", ctx, id).Return((*model.Order)(nil), nil).Once()

	got, err := s.GetOrderByID(ctx, id)
	assert.NoError(t, err)
	assert.Nil(t, got)

	cache.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	db.AssertExpectations(t)
	cache.AssertExpectations(t)
}
