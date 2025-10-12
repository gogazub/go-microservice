package repository

import (
	"context"

	"github.com/gogazub/myapp/internal/model"
)

type Repository interface {
	Save(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	GetAll(ctx context.Context) ([]*model.Order, error)
}
