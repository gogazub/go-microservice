package tests

import (
	"testing"
)

type mockRepo struct {
}

func newMockRepo(t *testing.T) *mockRepo {
	return &mockRepo{}
}

func makeOrder() interface{} {
	return nil
}

// ---------- Save ----------

func TestDBRepository_Save(t *testing.T) {
	t.Run("success: upsert всех сущностей в одной транзакции -> commit", func(t *testing.T) {
		// TODO
	})

	t.Run("error: orders upsert падает -> rollback и ошибка наружу", func(t *testing.T) {
		// TODO
	})
}

// ---------- GetByID ----------

func TestDBRepository_GetByID(t *testing.T) {
	t.Run("success: грузим order + delivery + payment + items", func(t *testing.T) {
		// TODO
	})

	t.Run("not found: orders возвращает sql.ErrNoRows", func(t *testing.T) {
		// TODO
	})
}

// ---------- GetAll ----------

func TestDBRepository_GetAll(t *testing.T) {
	t.Run("success: читаем список id и для каждого вызываем GetByID", func(t *testing.T) {
		// TODO
	})

	t.Run("error: SELECT order_uid падает", func(t *testing.T) {
		// TODO
	})
}
