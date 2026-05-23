-- 006_create_operators.sql
-- Akun kasir per bisnis. Login via operator_code (di-set owner, unik per bisnis)
-- atau scan QR (qr_token unik global). Tidak support Google OAuth.

CREATE TABLE IF NOT EXISTS operators (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id   UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    operator_code VARCHAR(64) NOT NULL,
    qr_token      VARCHAR(128) NOT NULL,
    password_hash TEXT NOT NULL,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

-- operator_code unik per bisnis (bukan global) — owner bebas pakai "kasir-1" di tiap toko.
-- Partial index: operator yang sudah di-soft-delete tidak ikut diuji unik,
-- sehingga code bisa dipakai ulang setelah dihapus.
CREATE UNIQUE INDEX IF NOT EXISTS uq_operators_business_code
    ON operators(business_id, operator_code)
    WHERE deleted_at IS NULL;

-- qr_token unik global karena login scan QR tidak punya konteks business_id dulu.
CREATE UNIQUE INDEX IF NOT EXISTS uq_operators_qr_token
    ON operators(qr_token);

CREATE INDEX IF NOT EXISTS idx_operators_business_id ON operators(business_id);
CREATE INDEX IF NOT EXISTS idx_operators_qr_token    ON operators(qr_token);