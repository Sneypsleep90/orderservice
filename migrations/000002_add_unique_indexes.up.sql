CREATE UNIQUE INDEX IF NOT EXISTS uq_delivery_order_uid_idx ON delivery(order_uid);

CREATE UNIQUE INDEX IF NOT EXISTS uq_payment_order_uid_idx ON payment(order_uid);

