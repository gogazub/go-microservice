package repository

import (
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
func NewCacheRepository(psqlRepo Repository) *CacheRepository {
	cache := make(map[string]*model.Order)

	orders, err := psqlRepo.GetAll()
	if err != nil {
		log.Println("Error loading orders from DB:", err)
	} else {
		for _, order := range orders {
			cache[order.OrderUID] = order
		}
	}

	return &CacheRepository{
		cache: cache,
	}
}

// Сохранить OrderModel в кэше по OrderUID
func (r *CacheRepository) Save(order *model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache[order.OrderUID] = order
	return nil
}

// Получить данные о заказе из кэша по uid заказа
func (r *CacheRepository) GetByID(id string) (*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, exists := r.cache[id]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

func (r *CacheRepository) GetAll() ([]*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orders []*model.Order
	for _, order := range r.cache {
		orders = append(orders, order)
	}
	return orders, nil
}
