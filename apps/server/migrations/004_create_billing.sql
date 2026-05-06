-- 004_create_billing.sql
-- Sistem billing QIOS: plan tier dan subscription per user.
-- Feature flags di tabel plans menentukan akses fitur dan batas resource.
-- Seed data plans belum final — tunggu konfirmasi pricing dari board.

CREATE TABLE IF NOT EXISTS plans (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              VARCHAR(100) NOT NULL UNIQUE,  -- 'free', 'starter', 'pro'
    description       TEXT,
    price_idr         INTEGER NOT NULL DEFAULT 0,

    -- Resource limits (-1 = unlimited)
    max_operators     INTEGER NOT NULL DEFAULT 3,
    max_products      INTEGER NOT NULL DEFAULT 50,

    -- Feature flags
    can_export        BOOLEAN NOT NULL DEFAULT FALSE, -- export laporan ke CSV/PDF
    can_ai_insight    BOOLEAN NOT NULL DEFAULT FALSE, -- AI insight cards di dashboard
    can_whitelabel    BOOLEAN NOT NULL DEFAULT FALSE, -- custom branding di interface kasir

    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- TODO: seed data plans belum final, tunggu konfirmasi pricing dari board.
-- INSERT INTO plans (name, description, price_idr, max_operators, max_products, can_export, can_ai_insight, can_whitelabel) VALUES
--     ('free',    'TBD', 0,      3,  50,  FALSE, FALSE, FALSE),
--     ('starter', 'TBD', 0,      5,  200, TRUE,  FALSE, FALSE),
--     ('pro',     'TBD', 0,      -1, -1,  TRUE,  TRUE,  TRUE)
-- ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS subscriptions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id    UUID NOT NULL REFERENCES plans(id),
    status     VARCHAR(20) NOT NULL DEFAULT 'active'
                   CHECK (status IN ('active', 'expired', 'terminated')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,                          -- NULL = no expiry
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);