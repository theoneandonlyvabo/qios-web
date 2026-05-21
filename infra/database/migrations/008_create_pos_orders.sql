-- 008_create_pos_orders.sql
-- Order dari interface kasir. order_id dikirim ke Xendit sebagai payment reference
-- untuk matching webhook notifikasi ke order yang benar.
--
-- Status lowercase (peninggalan iterasi awal). Mapping ke domain enum uppercase
-- dilakukan di repository boundary — lihat domain/payment/repository.go.

CREATE TABLE IF NOT EXISTS pos_orders (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id    UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    operator_id    UUID REFERENCES operators(id) ON DELETE SET NULL,
    order_id       VARCHAR(100) NOT NULL UNIQUE,      -- format: {business_id}-{unix_ts}-{rand}
    total_amount   BIGINT NOT NULL CHECK (total_amount > 0),
    payment_method VARCHAR(20) NOT NULL DEFAULT 'CASH'
                       CHECK (payment_method IN ('CASH', 'QRIS', 'EWALLET', 'VIRTUAL_ACCOUNT')),
    status         VARCHAR(20) NOT NULL DEFAULT 'pending'
                       CHECK (status IN ('pending', 'paid', 'failed', 'expired', 'cancelled')),
    note           TEXT,
    paid_at        TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pos_orders_business_id     ON pos_orders(business_id);
CREATE INDEX IF NOT EXISTS idx_pos_orders_order_id        ON pos_orders(order_id);
CREATE INDEX IF NOT EXISTS idx_pos_orders_status          ON pos_orders(status);
CREATE INDEX IF NOT EXISTS idx_pos_orders_payment_method  ON pos_orders(payment_method);
CREATE INDEX IF NOT EXISTS idx_pos_orders_created_at      ON pos_orders(created_at);