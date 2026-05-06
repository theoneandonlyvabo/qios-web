-- 004_create_businesses.sql
-- Satu bisnis per user (1:1). Dibuat saat owner selesai onboarding.

CREATE TABLE IF NOT EXISTS businesses (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    slug         VARCHAR(255) NOT NULL UNIQUE,      -- untuk URL/referensi internal
    timezone     VARCHAR(100) NOT NULL DEFAULT 'Asia/Jakarta',
    currency     VARCHAR(10)  NOT NULL DEFAULT 'IDR',
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_businesses_user_id ON businesses(user_id);
CREATE INDEX IF NOT EXISTS idx_businesses_slug    ON businesses(slug);