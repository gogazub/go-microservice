package model

import "time"

type Order struct {
	OrderUID          string    `json:"order_uid" db:"order_uid" validate:"required"`
	TrackNumber       string    `json:"track_number" db:"track_number" validate:"-"`
	Entry             string    `json:"entry" db:"entry" validate:"-"`
	Locale            string    `json:"locale" db:"locale" validate:"-"`
	InternalSignature string    `json:"internal_signature" db:"internal_signature" validate:"-"`
	CustomerID        string    `json:"customer_id" db:"customer_id" validate:"required"`
	DeliveryService   string    `json:"delivery_service" db:"delivery_service" validate:"-"`
	Shardkey          string    `json:"shardkey" db:"shardkey" validate:"-"`
	SmID              int       `json:"sm_id" db:"sm_id" validate:"gte=0"`
	DateCreated       time.Time `json:"date_created" db:"date_created" validate:"required"`
	OofShard          string    `json:"oof_shard" db:"oof_shard" validate:"required"`

	Delivery Delivery `json:"delivery" validate:"required"`
	Payment  Payment  `json:"payment"  validate:"required"`
	Items    []Item   `json:"items"    validate:"required,min=1,dive"`
}

type Delivery struct {
	DeliveryID int    `json:"-" db:"delivery_id" validate:"-"`
	OrderUID   string `json:"-" db:"order_uid"    validate:"-"`
	Name       string `json:"name" db:"name"       validate:"required"`
	Phone      string `json:"phone" db:"phone"     validate:"required"`
	Zip        string `json:"zip" db:"zip"         validate:"required"`
	City       string `json:"city" db:"city"       validate:"required"`
	Address    string `json:"address" db:"address" validate:"required"`
	Region     string `json:"region" db:"region"   validate:"required"`
	Email      string `json:"email" db:"email"     validate:"required,email"`
}

type Payment struct {
	PaymentID    int     `json:"-" db:"payment_id" validate:"-"`
	OrderUID     string  `json:"-" db:"order_uid"   validate:"-"`
	Transaction  string  `json:"transaction" db:"transaction" validate:"required"`
	RequestID    string  `json:"request_id" db:"request_id"   validate:"omitempty"`
	Currency     string  `json:"currency" db:"currency"       validate:"required"`
	Provider     string  `json:"provider" db:"provider"       validate:"required"`
	Amount       float64 `json:"amount" db:"amount"           validate:"gte=0"`
	PaymentDt    int64   `json:"payment_dt" db:"payment_dt"   validate:"gte=0"`
	Bank         string  `json:"bank" db:"bank"               validate:"required"`
	DeliveryCost float64 `json:"delivery_cost" db:"delivery_cost" validate:"gte=0"`
	GoodsTotal   int     `json:"goods_total" db:"goods_total"     validate:"gte=0"`
	CustomFee    float64 `json:"custom_fee" db:"custom_fee"       validate:"gte=0"`
}

type Item struct {
	ItemID      int     `json:"-" db:"item_id"         validate:"-"`
	OrderUID    string  `json:"-" db:"order_uid"       validate:"-"`
	ChrtID      int64   `json:"chrt_id" db:"chrt_id"   validate:"required,gte=1"`
	TrackNumber string  `json:"track_number" db:"track_number" validate:"required"`
	Price       float64 `json:"price" db:"price"       validate:"gte=0"`
	Rid         string  `json:"rid" db:"rid"           validate:"-"`
	Name        string  `json:"name" db:"name"         validate:"-"`
	Sale        int     `json:"sale" db:"sale"         validate:"gte=0,lte=100"`
	Size        string  `json:"size" db:"size"         validate:"omitempty"`
	TotalPrice  float64 `json:"total_price" db:"total_price" validate:"gte=0"`
	NmID        int64   `json:"nm_id" db:"nm_id"       validate:"gte=0"`
	Brand       string  `json:"brand" db:"brand"       validate:"-"`
	Status      int     `json:"status" db:"status"     validate:"gte=0"`
}
