// Package repository хранит структуры для хранения данных. В БД/Кеше
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

// ICacheRepository интерфейс кеш репозитория
type ICacheRepository interface {
	LoadFromDB(psqlRepo IDBRepository) error
	Save(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	GetAll(ctx context.Context) ([]*model.Order, error)
}

type cacheEntry struct {
	elem  *list.Element
	order *model.Order
}

// CacheRepository реализция интерфейса. TODO: сделать неэспортируемой эту структуру, а также service и dbrepo
type CacheRepository struct {
	mu    sync.RWMutex
	cache map[string]*cacheEntry

	list *list.List
}

// NewCacheRepository Конструктор. Кэш репозиторий при создании заполняется данными из БД
func NewCacheRepository() *CacheRepository {
	cache := make(map[string]*cacheEntry)
	return &CacheRepository{
		cache: cache,
		list:  list.New(),
	}
}

// LoadFromDB Заполнить мапу значениями из БД
func (r *CacheRepository) LoadFromDB(psqlRepo IDBRepository) error {
	orders, err := psqlRepo.GetAll(context.Background())

	if err != nil {
		return fmt.Errorf("load from db error:%w", err)
	}

	for _, order := range orders {

		err = r.Save(context.Background(), order)
		if err != nil {
			// Логируем ошибку и OrderLog - облегченная модель заказа
			r.logOrder(fmt.Sprintf("save order error:%v", err), model.GetOrderLog(order))
		}
		if r.Size() >= maxCacheSize {
			break
		}
	}

	return nil
}

// Save добавить OrderModel в кэш
func (r *CacheRepository) Save(ctx context.Context, order *model.Order) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("save error:%w", err)
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

// GetByID Попробовать достать model.Order из кеша. В случае, если элемент есть в кеше, продлевает его жизнь в LRU
func (r *CacheRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("getByID error:%w", err)
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

// GetAll возвращает массив всех model.Order, которые хранятся в кеше
func (r *CacheRepository) GetAll(ctx context.Context) ([]*model.Order, error) {
	// Быстрый отказ, если контекст уже отменен, чтобы не лочить mutex лишний раз
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("getAll error:%w", err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Предвыделим память, чтобы избежать лишних аллокаций
	orders := make([]*model.Order, 0, len(r.cache))

	i := 0
	for _, order := range r.cache {
		if i%1000 == 0 { // Вместо select с  <-ctx.Done будем редко проверять контекст
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("getAll error:%w", err)
			}
		}
		orders = append(orders, order.order)
		i++
	}
	return orders, nil
}

// Size текущий размер кеша
func (r *CacheRepository) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.list.Len()
}

func (r *CacheRepository) logOrder(msg string, order model.OrderLog) {
	log.Printf("%s\norder:%v", msg, order)
}

// Удаляет n самых старых элементов (с головы списка). Может потребоваться для оптимизации
// func (r *CacheRepository) evictOldest(n int) {
// 	for i := 0; i < n; i++ {
// 		front := r.list.Front()
// 		if front == nil {
// 			return
// 		}
// 		key := front.Value.(string)
// 		r.list.Remove(front)
// 		delete(r.cache, key)
// 	}
// }
