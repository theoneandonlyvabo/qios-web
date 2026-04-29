-- 008_create_payment_records.sql
-- Menyimpan semua data transaksi Midtrans (PG-01 s/d PG-05).
-- Tidak ada data yang hilang meski webhook terlambat — setiap event disimpan.

CREATE TABLE IF NOT EXISTS payment_records (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id          UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    midtrans_order_id   VARCHAR(255) NOT NULL UNIQUE,   -- order_id yang dikirim ke Midtrans
    midtrans_tx_id      VARCHAR(255),                   -- transaction_id dari Midtrans response
    payment_url         TEXT,                           -- payment link untuk user (PG-01)
    status              VARCHAR(30) NOT NULL DEFAULT 'pending'
                            CHECK (status IN (
                                'pending', 'capture', 'settlement',
                                'deny', 'cancel', 'expire', 'failure', 'refund'
                            )),
    payment_type        VARCHAR(50),                    -- e.g. 'bank_transfer', 'gopay'
    gross_amount        BIGINT NOT NULL,
    raw_notification    JSONB,                          -- raw webhook payload untuk audit (PG-04)
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_records_invoice_id ON payment_records(invoice_id);
CREATE INDEX IF NOT EXISTS idx_payment_records_user_id ON payment_records(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_records_midtrans_order_id ON payment_records(midtrans_order_id);
CREATE INDEX IF NOT EXISTS idx_payment_records_status ON payment_records(status);

-- ------------------------------------------------------------------
-- Webhook event log — setiap notifikasi masuk dari Midtrans dicatat (PG-04)
-- Berguna untuk replay kalau ada missed webhook

CREATE TABLE IF NOT EXISTS webhook_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source          VARCHAR(50) NOT NULL DEFAULT 'midtrans',
    order_id        VARCHAR(255),
    status_code     VARCHAR(10),
    signature_valid BOOLEAN NOT NULL DEFAULT FALSE,     -- hasil validasi PG-05
    payload         JSONB NOT NULL,
    processed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhook_events_order_id ON webhook_events(order_id);
