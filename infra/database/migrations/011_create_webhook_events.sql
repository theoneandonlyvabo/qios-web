-- 011_create_webhook_events.sql
-- Log setiap notifikasi masuk dari Midtrans, terlepas dari status validasinya.
-- Berguna untuk audit trail dan replay kalau ada missed webhook.

CREATE TABLE IF NOT EXISTS webhook_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source          VARCHAR(50) NOT NULL DEFAULT 'midtrans',
    order_id        VARCHAR(255),
    status_code     VARCHAR(10),
    signature_valid BOOLEAN NOT NULL DEFAULT FALSE,
    payload         JSONB NOT NULL,
    processed_at    TIMESTAMPTZ,                    -- NULL = belum diproses
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhook_events_order_id   ON webhook_events(order_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_created_at ON webhook_events(created_at);