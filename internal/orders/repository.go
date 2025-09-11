package orders

import (
	"database/sql"
	"fmt"
)

type Repository interface {
	Save(order *ModelOrder) error
	GetByID(id string) (*ModelOrder, error)
	GetAll() ([]*ModelOrder, error)
}

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Save сохраняет заказ вместе с зависимыми сущностями
func (r *OrderRepository) Save(order *ModelOrder) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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
func (r *OrderRepository) GetByID(id string) (*ModelOrder, error) {
	var order ModelOrder

	if err := r.loadOrder(&order, id); err != nil {
		return nil, err
	}
	if err := r.loadDelivery(&order); err != nil {
		return nil, err
	}
	if err := r.loadPayment(&order); err != nil {
		return nil, err
	}
	if err := r.loadItems(&order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) GetAll() ([]*ModelOrder, error) {
	rows, err := r.db.Query(`SELECT order_uid FROM orders`)
	if err != nil {
		return nil, fmt.Errorf("get all orders: %w", err)
	}
	defer rows.Close()

	var orders []*ModelOrder
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		order, err := r.GetByID(id)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

//
// ---------------- PRIVATE (orders) ----------------
//

func (r *OrderRepository) saveOrder(tx *sql.Tx, o *ModelOrder) error {
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

func (r *OrderRepository) loadOrder(o *ModelOrder, id string) error {
	return r.db.QueryRow(`
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

func (r *OrderRepository) saveDelivery(tx *sql.Tx, o *ModelOrder) error {
	_, err := tx.Exec(`
		INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (order_uid) DO UPDATE SET
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

func (r *OrderRepository) loadDelivery(o *ModelOrder) error {
	return r.db.QueryRow(`
		SELECT delivery_id, order_uid, name, phone, zip, city, address, region, email
		FROM deliveries WHERE order_uid = $1
	`, o.OrderUID).Scan(&o.Delivery.DeliveryID, &o.Delivery.OrderUID, &o.Delivery.Name,
		&o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City, &o.Delivery.Address,
		&o.Delivery.Region, &o.Delivery.Email)
}

//
// ---------------- PRIVATE (payment) ----------------
//

func (r *OrderRepository) savePayment(tx *sql.Tx, o *ModelOrder) error {
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

func (r *OrderRepository) loadPayment(o *ModelOrder) error {
	return r.db.QueryRow(`
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

func (r *OrderRepository) saveItems(tx *sql.Tx, o *ModelOrder) error {
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

func (r *OrderRepository) loadItems(o *ModelOrder) error {
	rows, err := r.db.Query(`
		SELECT item_id, order_uid, chrt_id, track_number, price, rid, name,
		       sale, size, total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1
	`, o.OrderUID)
	if err != nil {
		return fmt.Errorf("loadItems: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ItemID, &it.OrderUID, &it.ChrtID, &it.TrackNumber, &it.Price,
			&it.Rid, &it.Name, &it.Sale, &it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status); err != nil {
			return fmt.Errorf("scanItem: %w", err)
		}
		o.Items = append(o.Items, it)
	}
	return nil
}
