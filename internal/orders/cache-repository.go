package orders

import (
	"fmt"
	"log"
	"sync"
)

type CacheRepository struct {
	mu    sync.RWMutex
	cache map[string]*ModelOrder
}

// Кэш репозиторий при создании заполняется данными из БД
func NewCacheRepository(psqlRepo Repository) *CacheRepository {
	cache := make(map[string]*ModelOrder)

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
func (r *CacheRepository) Save(order *ModelOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache[order.OrderUID] = order
	return nil
}

// Получить данные о заказе из кэша по uid заказа
func (r *CacheRepository) GetByID(id string) (*ModelOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, exists := r.cache[id]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

func (r *CacheRepository) GetAll() ([]*ModelOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orders []*ModelOrder
	for _, order := range r.cache {
		orders = append(orders, order)
	}
	return orders, nil
}
