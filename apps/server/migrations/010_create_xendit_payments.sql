-- 010_create_xendit_payments.sql
-- Record setiap transaksi Xendit yang terhubung ke pos_order.
-- raw_payload menyimpan webhook payload asli untuk audit.
-- Status mengikuti event Xendit (qr_code, invoice, dll) — disimpan apa adanya
-- supaya logic interpretasi ada di service layer, bukan di constraint DB.

-- Cleanup tabel Midtrans peninggalan iterasi sebelumnya (pre-Xendit decision).
-- Aman karena belum ada data production saat migrasi ini di-run.
DROP TABLE IF EXISTS midtrans_payments;

CREATE TABLE IF NOT EXISTS xendit_payments (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pos_order_id       UUID NOT NULL REFERENCES pos_orders(id) ON DELETE CASCADE,
    xendit_account_id  VARCHAR(255) NOT NULL,        -- sub-account merchant (for-user-id)
    xendit_invoice_id  VARCHAR(255),                 -- ID dari Xendit invoice atau QR
    xendit_charge_id   VARCHAR(255),                 -- ID transaksi spesifik (qr_code id, dll)
    payment_method     VARCHAR(50),                  -- 'QRIS', 'EWALLET', 'VIRTUAL_ACCOUNT', dll
    amount             BIGINT NOT NULL CHECK (amount > 0),
    status             VARCHAR(30) NOT NULL DEFAULT 'PENDING'
                           CHECK (status IN (
                               'PENDING', 'PAID', 'SETTLED',
                               'FAILED', 'EXPIRED', 'REFUNDED'
                           )),
    raw_payload        JSONB,                        -- raw webhook payload terakhir
    paid_at            TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_xendit_payments_pos_order_id      ON xendit_payments(pos_order_id);
CREATE INDEX IF NOT EXISTS idx_xendit_payments_xendit_invoice_id ON xendit_payments(xendit_invoice_id);
CREATE INDEX IF NOT EXISTS idx_xendit_payments_xendit_charge_id  ON xendit_payments(xendit_charge_id);
CREATE INDEX IF NOT EXISTS idx_xendit_payments_status            ON xendit_payments(status);
CREATE INDEX IF NOT EXISTS idx_xendit_payments_account_id        ON xendit_payments(xendit_account_id);
