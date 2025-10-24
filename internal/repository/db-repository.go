package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/gogazub/myapp/internal/model"
)

// IDBRepository интерфейс БД репозитория
type IDBRepository interface {
	Save(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	GetAll(ctx context.Context) ([]*model.Order, error)
}

// DBRepository реализация БД репозитория.
type DBRepository struct {
	db *sql.DB
}

// NewOrderRepository конструктор. Создает объект репозитория по переданному sql подключению
func NewOrderRepository(db *sql.DB) *DBRepository {
	return &DBRepository{db: db}
}

// Save сохраняет заказ вместе с зависимыми сущностями
func (r *DBRepository) Save(ctx context.Context, order *model.Order) error {
	// Прокидываем контекст
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.Printf("Rollback error:%s", err.Error())
		}
	}()

	if err := r.saveOrder(tx, order); err != nil {
		return err
	}
	if err := r.saveDelivery(tx, order); err != nil {
		return err
	}
	if err := r.savePayment(tx, order); err != nil {
		return err
	}
	if err := r.saveItems(tx, order); err != nil {
		return err
	}

	return tx.Commit()
}

// GetByID возвращает заказ по ID
func (r *DBRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	var order model.Order

	if err := r.loadOrder(ctx, &order, id); err != nil {
		return nil, err
	}
	if err := r.loadDelivery(ctx, &order); err != nil {
		return nil, err
	}
	if err := r.loadPayment(ctx, &order); err != nil {
		return nil, err
	}
	if err := r.loadItems(ctx, &order); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return &order, nil
}

// GetAll создает массив []*model.Order по данным из Postgres. TODO: поставить ограничение
func (r *DBRepository) GetAll(ctx context.Context) ([]*model.Order, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT order_uid FROM orders`)
	if err != nil {
		return nil, fmt.Errorf("get all orders: %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("rows close error:%s", err.Error())
		}
	}()

	orders := make([]*model.Order, 0, 256) // предвыделение

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		// Дальше тоже пробрасываем контекст
		order, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

//
// ---------------- PRIVATE (orders) ----------------
//

func (r *DBRepository) saveOrder(tx *sql.Tx, o *model.Order) error {
	_, err := tx.Exec(`
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
	`, o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature,
		o.CustomerID, o.DeliveryService, o.Shardkey, o.SmID, o.DateCreated, o.OofShard)

	if err != nil {
		return fmt.Errorf("saveOrder: %w", err)
	}
	return nil
}

func (r *DBRepository) loadOrder(ctx context.Context, o *model.Order, id string) error {
	return r.db.QueryRowContext(ctx, `
		SELECT order_uid, track_number, entry, locale, internal_signature,
		       customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1
	`, id).Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale,
		&o.InternalSignature, &o.CustomerID, &o.DeliveryService,
		&o.Shardkey, &o.SmID, &o.DateCreated, &o.OofShard)
}

//
// ---------------- PRIVATE (delivery) ----------------
//

func (r *DBRepository) saveDelivery(tx *sql.Tx, o *model.Order) error {
	_, err := tx.Exec(`
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
	`, o.OrderUID, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip,
		o.Delivery.City, o.Delivery.Address, o.Delivery.Region, o.Delivery.Email)

	if err != nil {
		return fmt.Errorf("saveDelivery: %w", err)
	}
	return nil
}

func (r *DBRepository) loadDelivery(ctx context.Context, o *model.Order) error {
	return r.db.QueryRowContext(ctx, `
		SELECT delivery_id, order_uid, name, phone, zip, city, address, region, email
		FROM deliveries WHERE order_uid = $1
	`, o.OrderUID).Scan(&o.Delivery.DeliveryID, &o.Delivery.OrderUID, &o.Delivery.Name,
		&o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City, &o.Delivery.Address,
		&o.Delivery.Region, &o.Delivery.Email)
}

//
// ---------------- PRIVATE (payment) ----------------
//

func (r *DBRepository) savePayment(tx *sql.Tx, o *model.Order) error {
	_, err := tx.Exec(`
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
	`, o.OrderUID, o.Payment.Transaction, o.Payment.RequestID, o.Payment.Currency,
		o.Payment.Provider, o.Payment.Amount, o.Payment.PaymentDt, o.Payment.Bank,
		o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee)

	if err != nil {
		return fmt.Errorf("savePayment: %w", err)
	}
	return nil
}

func (r *DBRepository) loadPayment(ctx context.Context, o *model.Order) error {
	return r.db.QueryRowContext(ctx, `
		SELECT payment_id, order_uid, transaction, request_id, currency, provider,
		       amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payments WHERE order_uid = $1
	`, o.OrderUID).Scan(&o.Payment.PaymentID, &o.Payment.OrderUID, &o.Payment.Transaction,
		&o.Payment.RequestID, &o.Payment.Currency, &o.Payment.Provider,
		&o.Payment.Amount, &o.Payment.PaymentDt, &o.Payment.Bank,
		&o.Payment.DeliveryCost, &o.Payment.GoodsTotal, &o.Payment.CustomFee)
}

//
// ---------------- PRIVATE (items) ----------------
//

func (r *DBRepository) saveItems(tx *sql.Tx, o *model.Order) error {
	_, err := tx.Exec(`DELETE FROM items WHERE order_uid = $1`, o.OrderUID)
	if err != nil {
		return fmt.Errorf("deleteItems: %w", err)
	}

	for _, it := range o.Items {
		_, err = tx.Exec(`
			INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, o.OrderUID, it.ChrtID, it.TrackNumber, it.Price, it.Rid, it.Name,
			it.Sale, it.Size, it.TotalPrice, it.NmID, it.Brand, it.Status)
		if err != nil {
			return fmt.Errorf("insertItem: %w", err)
		}
	}
	return nil
}

func (r *DBRepository) loadItems(ctx context.Context, o *model.Order) error {
	rows, err := r.db.QueryContext(ctx, `
		SELECT item_id, order_uid, chrt_id, track_number, price, rid, name,
		       sale, size, total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1
	`, o.OrderUID)
	if err != nil {
		return fmt.Errorf("loadItems: %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("rows close error:%s", err.Error())
		}
	}()

	for rows.Next() {
		var it model.Item
		if err := rows.Scan(&it.ItemID, &it.OrderUID, &it.ChrtID, &it.TrackNumber, &it.Price,
			&it.Rid, &it.Name, &it.Sale, &it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status); err != nil {
			return fmt.Errorf("scanItem: %w", err)
		}
		o.Items = append(o.Items, it)
	}
	return nil
}
