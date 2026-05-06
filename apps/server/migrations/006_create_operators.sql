-- 006_create_operators.sql
-- Akun kasir per bisnis. Tidak support Google OAuth, hanya email/password.

CREATE TABLE IF NOT EXISTS operators (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id   UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_operators_business_id ON operators(business_id);
CREATE INDEX IF NOT EXISTS idx_operators_email       ON operators(email);