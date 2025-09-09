package repository

import (
	"context"

	"github.com/gogazub/myapp/internal/repository/model"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByUID(ctx context.Context, uid string) (*model.Order, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.Order, error)
	Update(ctx context.Context, order *model.Order) error
	Delete(ctx context.Context, uid string) error
	Exists(ctx context.Context, uid string) (bool, error)
}
