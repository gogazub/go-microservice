package tests

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/gogazub/myapp/internal/model"
	"github.com/gogazub/myapp/internal/repository"
)

type Repository interface {
	Save(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	GetAll(ctx context.Context) ([]*model.Order, error)
}

/*
 1. GetByID
    a) поиск несуществующего
    b) поиск существующего
    c) поиск в пустом
 2. GetAll
    a) мапа пустая
    b) мапа имеет один объект
    c) мапа имеет несколько объектов
 3. Save
    a) сохранить новый элемент
    b) сохранить существующий элемент
 4. Timeout
*/
func TestCacheRepo(t *testing.T) {
	repo := repository.NewCacheRepository()

	t.Run("GetByID/not found in empty cache", func(t *testing.T) {
		ctx := context.Background()
		o, err := repo.GetByID(ctx, "123")
		if o != nil || err == nil {
			t.Fatalf("want (nil, err) on missing id; got (%v, %v)", o, err)
		}
	})

	t.Run("Save/new element then GetByID", func(t *testing.T) {
		ctx := context.Background()
		in := fakeOrder("o1")

		if err := repo.Save(ctx, in); err != nil {
			t.Fatalf("save: %v", err)
		}
		out, err := repo.GetByID(ctx, "o1")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if out == nil || out.OrderUID != "o1" {
			t.Fatalf("got %+v; want id=o1", out)
		}
	})

	t.Run("Save/existing element overwritten", func(t *testing.T) {
		ctx := context.Background()
		first := fakeOrder("same")
		second := fakeOrder("same")

		if err := repo.Save(ctx, first); err != nil {
			t.Fatalf("save first: %v", err)
		}
		if err := repo.Save(ctx, second); err != nil {
			t.Fatalf("save second: %v", err)
		}
		got, err := repo.GetByID(ctx, "same")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got == nil || got.OrderUID != "same" {
			t.Fatalf("wrong element after overwrite: %+v", got)
		}
	})

	t.Run("GetAll/empty", func(t *testing.T) {
		repo2 := repository.NewCacheRepository()
		ctx := context.Background()

		list, err := repo2.GetAll(ctx)
		if err != nil {
			t.Fatalf("getall: %v", err)
		}
		if len(list) != 0 {
			t.Fatalf("want 0, got %d", len(list))
		}
	})

	t.Run("GetAll/one element", func(t *testing.T) {
		repo2 := repository.NewCacheRepository()
		ctx := context.Background()

		_ = repo2.Save(ctx, fakeOrder("x1"))

		list, err := repo2.GetAll(ctx)
		if err != nil {
			t.Fatalf("getall: %v", err)
		}
		if got := idsFromOrders(list); len(got) != 1 || got[0] != "x1" {
			t.Fatalf("ids=%v; want [x1]", got)
		}
	})

	t.Run("GetAll/many elements", func(t *testing.T) {
		repo2 := repository.NewCacheRepository()
		ctx := context.Background()

		want := []string{"a", "b", "c"}
		for _, id := range want {
			_ = repo2.Save(ctx, fakeOrder(id))
		}
		list, err := repo2.GetAll(ctx)
		if err != nil {
			t.Fatalf("getall: %v", err)
		}
		got := idsFromOrders(list)
		sort.Strings(want)
		if len(got) != len(want) {
			t.Fatalf("len=%d want=%d; ids=%v", len(got), len(want), got)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("ids=%v; want %v", got, want)
			}
		}
	})

	t.Run("Context/canceled: GetByID respects ctx", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := repo.GetByID(ctx, "nope")
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("want context.Canceled, got %v", err)
		}
	})

	t.Run("Context/canceled: GetAll respects ctx", func(t *testing.T) {
		repo2 := repository.NewCacheRepository()
		for i := 0; i < 10_000; i++ {
			_ = repo2.Save(context.Background(), fakeOrder("id-"+strconvI(i)))
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := repo2.GetAll(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("want context.Canceled, got %v", err)
		}
	})

	t.Run("Context/deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
		defer cancel()

		_, err := repo.GetAll(ctx)
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			t.Fatalf("want deadline-related error, got %v", err)
		}
	})
}

func TestLRU(t *testing.T) {
	t.Run("Test/LRU save 10000 orders", func(t *testing.T) {
		r := repository.NewCacheRepository()
		ctx := context.Background()
		for i := 0; i < 10000; i++ {
			order := fakeOrder(strconvI(i))
			r.Save(ctx, order)
		}
		if r.Size() > 1000 {
			t.Fatalf("saved %d orders \n", r.Size())
		}
	})
}

func fakeOrder(id string) *model.Order {
	return &model.Order{OrderUID: id}
}

func idsFromOrders(orders []*model.Order) []string {
	out := make([]string, 0, len(orders))
	for _, o := range orders {
		out = append(out, o.OrderUID)
	}
	sort.Strings(out)
	return out
}

// небольшая утилита, чтобы не тянуть strconv
func strconvI(i int) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	var b [32]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = digits[i%10]
		i /= 10
	}
	return string(b[pos:])
}
