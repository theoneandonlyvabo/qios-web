-- 008_create_pos_orders.sql
-- Order dari interface kasir. order_id dikirim ke Midtrans sebagai payment reference
-- untuk matching webhook notifikasi ke order yang benar.

CREATE TABLE IF NOT EXISTS pos_orders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    operator_id UUID REFERENCES operators(id) ON DELETE SET NULL,
    order_id    VARCHAR(100) NOT NULL UNIQUE,       -- format: QIOS-YYYYMMDD-xxxx
    total_amount BIGINT NOT NULL CHECK (total_amount > 0),
    status      VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'paid', 'failed', 'expired')),
    note        TEXT,
    paid_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pos_orders_business_id ON pos_orders(business_id);
CREATE INDEX IF NOT EXISTS idx_pos_orders_order_id    ON pos_orders(order_id);
CREATE INDEX IF NOT EXISTS idx_pos_orders_status      ON pos_orders(status);
CREATE INDEX IF NOT EXISTS idx_pos_orders_created_at  ON pos_orders(created_at);