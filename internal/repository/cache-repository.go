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

type ICacheRepository interface {
	LoadFromDB(psqlRepo IDBRepository)
	Save(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	GetAll(ctx context.Context) ([]*model.Order, error)
}

type cacheEntry struct {
	elem  *list.Element
	order *model.Order
}

type CacheRepository struct {
	mu    sync.RWMutex
	cache map[string]*cacheEntry

	list *list.List
}

// Кэш репозиторий при создании заполняется данными из БД
func NewCacheRepository() *CacheRepository {
	cache := make(map[string]*cacheEntry)
	return &CacheRepository{
		cache: cache,
		list:  list.New(),
	}
}

// Заполнить мапу значениями из БД
func (repo *CacheRepository) LoadFromDB(psqlRepo IDBRepository) {
	orders, err := psqlRepo.GetAll(context.Background())

	if err != nil {
		// Можно добавить флаг на пропуск логирования ctx.Err.
		// То есть пропускать ошибки, по типу таймаута
		log.Println("Error loading orders from DB:", err)
	} else {
		for _, order := range orders {

			repo.Save(context.Background(), order)
			if repo.Size() >= maxCacheSize {
				break
			}
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

	if ent, ok := r.cache[order.OrderUID]; ok {
		// обновляем значение и освежаем позицию
		ent.order = order
		r.list.MoveToBack(ent.elem)
	} else {
		// новый
		e := r.list.PushBack(order.OrderUID)
		r.cache[order.OrderUID] = &cacheEntry{elem: e, order: order}

	}
	for r.list.Len() > maxCacheSize {
		front := r.list.Front()
		if front == nil {
			break
		}
		key := front.Value.(string)
		r.list.Remove(front)
		delete(r.cache, key)
	}
	return nil
}

// Получить данные о заказе из кэша по uid заказа. Также освежаем элемент в LRU
func (r *CacheRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	ent, exists := r.cache[id]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}

	r.list.MoveToBack(ent.elem)
	return ent.order, nil
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
		orders = append(orders, order.order)
		i++
	}
	return orders, nil
}

func (r *CacheRepository) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.list.Len()
}

// Удаляет n самых старых элементов (с головы списка).
func (r *CacheRepository) evictOldest(n int) {
	for i := 0; i < n; i++ {
		front := r.list.Front()
		if front == nil {
			return
		}
		key := front.Value.(string)
		r.list.Remove(front)
		delete(r.cache, key)
	}
}
