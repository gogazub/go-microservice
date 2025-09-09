package repository

import (
	"database/sql"
)

/*
type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByUID(ctx context.Context, uid string) (*model.Order, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.Order, error)
	Update(ctx context.Context, order *model.Order) error
	Delete(ctx context.Context, uid string) error
	Exists(ctx context.Context, uid string) (bool, error)
}*/

type PostgresOrderRepository struct {
	db *sql.DB
}

// func Create(ctx context.Context, order *model.Order) error {

// }
