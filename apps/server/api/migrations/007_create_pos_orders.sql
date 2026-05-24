-- 007_create_pos_orders.sql
-- Order dari interface kasir.
-- Flow status: DRAFT → PENDING (checkout dimulai) → CONFIRMED / VOIDED
-- checkout_started_at dipakai untuk validasi slide-to-confirm server-side (>= 800ms).

CREATE TABLE IF NOT EXISTS pos_orders (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id          UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    operator_id          UUID REFERENCES operators(id) ON DELETE SET NULL,
    order_id             VARCHAR(100) NOT NULL UNIQUE,      -- format: {business_id}-{unix_ts}-{rand}
    total_amount         BIGINT NOT NULL CHECK (total_amount > 0),
    payment_method       VARCHAR(20) NOT NULL DEFAULT 'CASH'
                             CHECK (payment_method IN ('CASH', 'QRIS', 'EWALLET', 'VIRTUAL_ACCOUNT')),
    status               VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
                             CHECK (status IN ('DRAFT', 'PENDING', 'CONFIRMED', 'VOIDED')),
    note                 TEXT,
    checkout_started_at  TIMESTAMPTZ,
    confirmed_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pos_orders_business_id     ON pos_orders(business_id);
CREATE INDEX IF NOT EXISTS idx_pos_orders_order_id        ON pos_orders(order_id);
CREATE INDEX IF NOT EXISTS idx_pos_orders_status          ON pos_orders(status);
CREATE INDEX IF NOT EXISTS idx_pos_orders_payment_method  ON pos_orders(payment_method);
CREATE INDEX IF NOT EXISTS idx_pos_orders_created_at      ON pos_orders(created_at);
