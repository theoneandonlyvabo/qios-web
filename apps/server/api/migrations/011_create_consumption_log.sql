-- 011_create_consumption_log.sql
-- Log pemakaian bahan baku per transaksi CONFIRMED.
-- Di-populate otomatis saat kasir konfirmasi order (via transaction service).

CREATE TABLE IF NOT EXISTS consumption_log (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES pos_orders(id) ON DELETE CASCADE,
    business_id    UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    product_id     UUID REFERENCES products(id) ON DELETE SET NULL,
    product_name   VARCHAR(255) NOT NULL,
    ingredient     VARCHAR(255) NOT NULL,
    quantity_used  NUMERIC(10,3) NOT NULL CHECK (quantity_used > 0),
    unit           VARCHAR(50),
    confirmed_at   TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_consumption_log_business_id    ON consumption_log(business_id);
CREATE INDEX IF NOT EXISTS idx_consumption_log_transaction_id ON consumption_log(transaction_id);
CREATE INDEX IF NOT EXISTS idx_consumption_log_confirmed_at   ON consumption_log(confirmed_at);
CREATE INDEX IF NOT EXISTS idx_consumption_log_ingredient     ON consumption_log(business_id, ingredient);
