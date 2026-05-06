-- 007_create_products.sql
-- Katalog produk per bisnis. Soft delete supaya data historis transaksi tetap akurat.
-- total_sold di-increment setiap transaksi paid — dipakai untuk sorting popular.

CREATE TABLE IF NOT EXISTS products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    price       BIGINT NOT NULL CHECK (price >= 0),
    category    VARCHAR(100),
    description TEXT,
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    total_sold  INTEGER NOT NULL DEFAULT 0,
    deleted_at  TIMESTAMPTZ,                        -- soft delete
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_business_id ON products(business_id);
CREATE INDEX IF NOT EXISTS idx_products_category    ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_deleted_at  ON products(deleted_at) WHERE deleted_at IS NULL;