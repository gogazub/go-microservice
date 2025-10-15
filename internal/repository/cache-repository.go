package repository

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/gogazub/myapp/internal/model"
)

const maxCacheSize = 1000

type CacheRepository struct {
	mu    sync.RWMutex
	cache map[string]*model.Order

	listMu sync.RWMutex
	list   list.List
}

// Кэш репозиторий при создании заполняется данными из БД
func NewCacheRepository() *CacheRepository {
	cache := make(map[string]*model.Order)
	return &CacheRepository{
		cache: cache,
		list:  *list.New(),
	}
}

// Заполнить мапу значениями из БД
func (repo *CacheRepository) LoadFromDB(psqlRepo Repository) {
	orders, err := psqlRepo.GetAll(context.Background())

	if err != nil {
		// Можно добавить флаг на пропуск логирования ctx.Err.
		// То есть пропускать ошибки, по типу таймаута
		log.Println("Error loading orders from DB:", err)
	} else {
		for _, order := range orders {
			if repo.list.Len() > 1000 {
				break
			}
			repo.cache[order.OrderUID] = order
			repo.list.PushBack(order.OrderUID)
		}
	}
}

// Сохранить OrderModel в кэше по OrderUID
func (r *CacheRepository) Save(ctx context.Context, order *model.Order) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Забываем самые долгохранящиеся элементы, если места нет
	if r.list.Len() > maxCacheSize {
		r.freeLRU()
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

// Освобождает место в кеше
func (cr *CacheRepository) freeLRU() {
	if cr.list.Len() < 10 {
		return
	}
	cr.listMu.Lock()
	defer cr.listMu.Unlock()

	// Освобождаем сразу 10 мест, чтобы не вызывать очистку с локом мьютекса много раз
	for i := 0; i < 10; i++ {
		idx := cr.list.Front().Value.(string)
		cr.list.Remove(cr.list.Front())
		delete(cr.cache, idx)
	}
}

func (cr *CacheRepository) Size() int {
	return cr.list.Len()
}
