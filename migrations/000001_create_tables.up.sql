CREATE TABLE orders (
    order_uid UUID PRIMARY KEY,
    track_number VARCHAR(64),
    entry VARCHAR(64),
    locale VARCHAR(8),
    internal_signature TEXT,
    customer_id VARCHAR(64),
    delivery_service VARCHAR(255),
    shardkey VARCHAR(64),
    sm_id INT,
    date_created TIMESTAMPTZ,
    oof_shard VARCHAR(64)
);

CREATE TABLE deliveries (
    delivery_id SERIAL PRIMARY KEY,
    order_uid UUID REFERENCES orders(order_uid),
    name VARCHAR(255),
    phone VARCHAR(15),
    zip VARCHAR(16),
    city VARCHAR(255),
    address VARCHAR(255),
    region VARCHAR(128),
    email VARCHAR(255)
);

CREATE TABLE payments (
    payment_id SERIAL PRIMARY KEY,
    order_uid UUID REFERENCES orders(order_uid),
    transaction VARCHAR(64),
    request_id VARCHAR(64),
    currency VARCHAR(8),
    provider VARCHAR(64),
    amount NUMERIC(12,2),
    payment_dt BIGINT,
    bank VARCHAR(64),
    delivery_cost NUMERIC(12,2),
    goods_total NUMERIC(12,0),
    custom_fee NUMERIC(12,2)
);

CREATE TABLE items (
    item_id SERIAL PRIMARY KEY,
    order_uid UUID REFERENCES orders(order_uid),
    chrt_id BIGINT,
    track_number VARCHAR(64),
    price NUMERIC(12,2),
    rid VARCHAR(64),
    name VARCHAR(255),
    sale INT,
    size VARCHAR(32),
    total_price NUMERIC(12,2),
    nm_id BIGINT,
    brand VARCHAR(128),
    status INT
);