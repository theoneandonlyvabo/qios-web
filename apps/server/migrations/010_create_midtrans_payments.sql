-- 010_create_midtrans_payments.sql
-- Record setiap transaksi Midtrans yang terhubung ke pos_order.
-- raw_notification menyimpan webhook payload asli untuk audit.

CREATE TABLE IF NOT EXISTS midtrans_payments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pos_order_id      UUID NOT NULL REFERENCES pos_orders(id) ON DELETE CASCADE,
    midtrans_order_id VARCHAR(255) NOT NULL UNIQUE, -- sama dengan order_id di pos_orders
    midtrans_tx_id    VARCHAR(255),
    payment_type      VARCHAR(50),                  -- e.g. 'qris', 'gopay', 'bank_transfer'
    gross_amount      BIGINT NOT NULL,
    status            VARCHAR(30) NOT NULL DEFAULT 'pending'
                          CHECK (status IN (
                              'pending', 'capture', 'settlement',
                              'deny', 'cancel', 'expire', 'failure', 'refund'
                          )),
    raw_notification  JSONB,                        -- raw webhook payload
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_midtrans_payments_pos_order_id      ON midtrans_payments(pos_order_id);
CREATE INDEX IF NOT EXISTS idx_midtrans_payments_midtrans_order_id ON midtrans_payments(midtrans_order_id);
CREATE INDEX IF NOT EXISTS idx_midtrans_payments_status            ON midtrans_payments(status);