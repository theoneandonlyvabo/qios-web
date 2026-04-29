-- 005_create_transaction_categories.sql
-- Kategori transaksi: ada default (system-level) dan custom per user (C-08).
-- user_id NULL = kategori default milik sistem.

CREATE TABLE IF NOT EXISTS transaction_categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL = sistem default
    name        VARCHAR(100) NOT NULL,
    type        VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tx_categories_user_id ON transaction_categories(user_id);

-- Seed default categories
INSERT INTO transaction_categories (user_id, name, type, is_default) VALUES
    -- Income
    (NULL, 'Penjualan',         'income',  TRUE),
    (NULL, 'Pendapatan Lain',   'income',  TRUE),
    (NULL, 'Modal',             'income',  TRUE),
    -- Expense
    (NULL, 'Bahan Baku',        'expense', TRUE),
    (NULL, 'Gaji & Upah',       'expense', TRUE),
    (NULL, 'Sewa',              'expense', TRUE),
    (NULL, 'Utilitas',          'expense', TRUE),
    (NULL, 'Marketing',         'expense', TRUE),
    (NULL, 'Operasional',       'expense', TRUE),
    (NULL, 'Pengeluaran Lain',  'expense', TRUE)
ON CONFLICT DO NOTHING;
