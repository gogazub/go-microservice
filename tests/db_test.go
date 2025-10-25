package tests

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gogazub/myapp/internal/model"
	"github.com/gogazub/myapp/internal/repository"
	"github.com/stretchr/testify/require"
)

// Используем sqlmock для мока БД
func newDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
	)
	require.NoError(t, err)

	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func q(sql string) string { return regexp.QuoteMeta(sql) }

// ---------- Save ----------
func TestDBRepository_Save(t *testing.T) {
	db, mock := newDB(t)
	repo := repository.NewOrderRepository(db)
	o := FakeValidOrder("uid-1")

	t.Run("success: upsert всех сущностей в одной транзакции -> commit", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(q(`
			INSERT INTO orders (
				order_uid, track_number, entry, locale, internal_signature,
				customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			ON CONFLICT (order_uid) DO UPDATE SET
				track_number = EXCLUDED.track_number,
				entry = EXCLUDED.entry,
				locale = EXCLUDED.locale,
				internal_signature = EXCLUDED.internal_signature,
				customer_id = EXCLUDED.customer_id,
				delivery_service = EXCLUDED.delivery_service,
				shardkey = EXCLUDED.shardkey,
				sm_id = EXCLUDED.sm_id,
				date_created = EXCLUDED.date_created,
				oof_shard = EXCLUDED.oof_shard
		`)).
			WithArgs(o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature,
				o.CustomerID, o.DeliveryService, o.Shardkey, o.SmID, o.DateCreated, o.OofShard).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(q(`
			INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT ON CONSTRAINT deliveries_order_uid_uniq DO UPDATE SET
				name = EXCLUDED.name,
				phone = EXCLUDED.phone,
				zip = EXCLUDED.zip,
				city = EXCLUDED.city,
				address = EXCLUDED.address,
				region = EXCLUDED.region,
				email = EXCLUDED.email
		`)).
			WithArgs(o.OrderUID, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip,
				o.Delivery.City, o.Delivery.Address, o.Delivery.Region, o.Delivery.Email).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(q(`
			INSERT INTO payments (order_uid, transaction, request_id, currency, provider,
				amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			ON CONFLICT (order_uid) DO UPDATE SET
				transaction = EXCLUDED.transaction,
				request_id = EXCLUDED.request_id,
				currency = EXCLUDED.currency,
				provider = EXCLUDED.provider,
				amount = EXCLUDED.amount,
				payment_dt = EXCLUDED.payment_dt,
				bank = EXCLUDED.bank,
				delivery_cost = EXCLUDED.delivery_cost,
				goods_total = EXCLUDED.goods_total,
				custom_fee = EXCLUDED.custom_fee
		`)).
			WithArgs(
				o.OrderUID, o.Payment.Transaction, o.Payment.RequestID, o.Payment.Currency,
				o.Payment.Provider, o.Payment.Amount, o.Payment.PaymentDt, o.Payment.Bank,
				o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(q(`DELETE FROM items WHERE order_uid = $1`)).
			WithArgs(o.OrderUID).
			WillReturnResult(sqlmock.NewResult(0, 2))

		insertItems := q(`
			INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`)
		for _, it := range o.Items {
			mock.ExpectExec(insertItems).
				WithArgs(
					o.OrderUID, it.ChrtID, it.TrackNumber, it.Price, it.Rid, it.Name,
					it.Sale, it.Size, it.TotalPrice, it.NmID, it.Brand, it.Status,
				).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}

		mock.ExpectCommit()

		require.NoError(t, repo.Save(context.Background(), o))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error: orders upsert падает -> rollback и ошибка наружу", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO orders").
			WillReturnError(errors.New("db fail"))
		mock.ExpectRollback()

		err := repo.Save(context.Background(), o)
		require.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// ---------------- GetByID ----------------

func expectGetByID(mock sqlmock.Sqlmock, o *model.Order) {
	mock.ExpectQuery(q(`
		SELECT order_uid, track_number, entry, locale, internal_signature,
		       customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1
	`)).
		WithArgs(o.OrderUID).
		WillReturnRows(sqlmock.NewRows([]string{
			"order_uid", "track_number", "entry", "locale", "internal_signature",
			"customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard",
		}).AddRow(
			o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature,
			o.CustomerID, o.DeliveryService, o.Shardkey, o.SmID, o.DateCreated, o.OofShard,
		))

	mock.ExpectQuery(q(`
		SELECT delivery_id, order_uid, name, phone, zip, city, address, region, email
		FROM deliveries WHERE order_uid = $1
	`)).
		WithArgs(o.OrderUID).
		WillReturnRows(sqlmock.NewRows([]string{
			"delivery_id", "order_uid", "name", "phone", "zip", "city", "address", "region", "email",
		}).AddRow(1, o.OrderUID, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip, o.Delivery.City, o.Delivery.Address, o.Delivery.Region, o.Delivery.Email))

	mock.ExpectQuery(q(`
		SELECT payment_id, order_uid, transaction, request_id, currency, provider,
		       amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payments WHERE order_uid = $1
	`)).
		WithArgs(o.OrderUID).
		WillReturnRows(sqlmock.NewRows([]string{
			"payment_id", "order_uid", "transaction", "request_id", "currency", "provider",
			"amount", "payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee",
		}).AddRow(1, o.OrderUID, o.Payment.Transaction, o.Payment.RequestID, o.Payment.Currency, o.Payment.Provider, o.Payment.Amount, o.Payment.PaymentDt, o.Payment.Bank, o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee))

	items := sqlmock.NewRows([]string{
		"item_id", "order_uid", "chrt_id", "track_number", "price", "rid", "name",
		"sale", "size", "total_price", "nm_id", "brand", "status",
	})
	for i, it := range o.Items {
		items.AddRow(i+1, o.OrderUID, it.ChrtID, it.TrackNumber, it.Price, it.Rid, it.Name, it.Sale, it.Size, it.TotalPrice, it.NmID, it.Brand, it.Status)
	}
	mock.ExpectQuery(q(`
		SELECT item_id, order_uid, chrt_id, track_number, price, rid, name,
		       sale, size, total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1
	`)).
		WithArgs(o.OrderUID).
		WillReturnRows(items)
}

func TestDBRepository_GetByID(t *testing.T) {
	db, mock := newDB(t)
	repo := repository.NewOrderRepository(db)

	t.Run("success: грузим order + delivery + payment + items", func(t *testing.T) {
		o := FakeValidOrder("uid-2")
		expectGetByID(mock, o)

		got, err := repo.GetByID(context.Background(), o.OrderUID)
		require.NoError(t, err)
		require.Equal(t, o.OrderUID, got.OrderUID)
		require.Len(t, got.Items, len(o.Items))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found: orders возвращает sql.ErrNoRows", func(t *testing.T) {
		mock.ExpectQuery("FROM orders WHERE order_uid = \\$1").
			WithArgs("missing").
			WillReturnError(sql.ErrNoRows)

		got, err := repo.GetByID(context.Background(), "missing")
		require.Error(t, err)
		require.Nil(t, got)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// ---------------- GetAll ----------------

func TestDBRepository_GetAll(t *testing.T) {
	db, mock := newDB(t)
	repo := repository.NewOrderRepository(db)

	t.Run("success: читаем список id и для каждого вызываем GetByID", func(t *testing.T) {
		ids := sqlmock.NewRows([]string{"order_uid"}).
			AddRow("uid-1").
			AddRow("uid-2")
		mock.ExpectQuery(q(`SELECT order_uid FROM orders`)).WillReturnRows(ids)

		o1 := FakeValidOrder("uid-1")
		o2 := FakeValidOrder("uid-2")
		expectGetByID(mock, o1)
		expectGetByID(mock, o2)

		list, err := repo.GetAll(context.Background())
		require.NoError(t, err)
		require.Len(t, list, 2)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error: SELECT order_uid падает", func(t *testing.T) {
		mock.ExpectQuery(`SELECT order_uid FROM orders`).
			WillReturnError(errors.New("db fail"))

		list, err := repo.GetAll(context.Background())
		require.Error(t, err)
		require.Nil(t, list)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
