-- 004_create_plans_and_subscriptions.sql
-- Plans: paket yang tersedia di platform (A-05).
-- Subscriptions: plan aktif per user, bisa di-extend atau terminate oleh admin.

CREATE TABLE IF NOT EXISTS plans (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL UNIQUE,    -- e.g. 'free', 'starter', 'pro'
    description TEXT,
    price_idr   INTEGER NOT NULL DEFAULT 0,      -- harga dalam Rupiah
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default plans INI MASIH MAU DIUBAH YAAAAAAAAAAAAAAAAAAAAAA
INSERT INTO plans (name, description, price_idr) VALUES
    ('free',    'Paket gratis dengan fitur dasar', 0),
    ('starter', 'Paket untuk bisnis yang mulai berkembang', 99000),
    ('pro',     'Paket lengkap untuk bisnis serius', 299000)
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS subscriptions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id     UUID NOT NULL REFERENCES plans(id),
    status      VARCHAR(20) NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'expired', 'terminated')),
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ,                     -- NULL = no expiry (lifetime/manual)
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
