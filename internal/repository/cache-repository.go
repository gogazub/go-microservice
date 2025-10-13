package repository

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/gogazub/myapp/internal/model"
)

type CacheRepository struct {
	mu    sync.RWMutex
	cache map[string]*model.Order
}

// Кэш репозиторий при создании заполняется данными из БД
func NewCacheRepository() *CacheRepository {
	cache := make(map[string]*model.Order)
	return &CacheRepository{
		cache: cache,
	}
}

// Заполнить мапу значениями из БД
func (repo *CacheRepository) LoadFromDB(psqlRepo Repository) {
	orders, err := psqlRepo.GetAll(context.Background())
	if err != nil {
		log.Println("Error loading orders from DB:", err)
	} else {
		for _, order := range orders {
			repo.cache[order.OrderUID] = order
		}
	}
}

// Сохранить OrderModel в кэше по OrderUID
func (r *CacheRepository) Save(ctx context.Context, order *model.Order) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache[order.OrderUID] = order
	return nil
}

// Получить данные о заказе из кэша по uid заказа
func (r *CacheRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.cache[id]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

func (r *CacheRepository) GetAll(ctx context.Context) ([]*model.Order, error) {
	// Быстрый отказ, если контекст уже отменен, чтобы не лочить mutex лишний раз
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Предвыделим память, чтобы избежать лишних аллокаций
	orders := make([]*model.Order, 0, len(r.cache))

	i := 0
	for _, order := range r.cache {
		if i%1000 == 0 { // Вместо select с  <-ctx.Done будем редко проверять контекст
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		orders = append(orders, order)
		i++
	}
	return orders, nil
}
