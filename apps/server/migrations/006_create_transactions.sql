-- 006_create_transactions.sql
-- Inti dari sistem: semua transaksi masuk dan keluar (C-07 s/d C-10).
-- source: 'manual' = input user, 'invoice' = otomatis dari pembayaran invoice.

CREATE TABLE IF NOT EXISTS transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id     UUID REFERENCES transaction_categories(id) ON DELETE SET NULL,
    type            VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
    amount          BIGINT NOT NULL CHECK (amount > 0),  -- dalam Rupiah (sen dihindari)
    description     TEXT,
    transaction_date DATE NOT NULL DEFAULT CURRENT_DATE,
    source          VARCHAR(20) NOT NULL DEFAULT 'manual'
                        CHECK (source IN ('manual', 'invoice')),
    source_id       UUID,                                -- invoice_id kalau source='invoice'
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ                          -- soft delete untuk audit trail (C-09)
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category_id);
CREATE INDEX IF NOT EXISTS idx_transactions_deleted_at ON transactions(deleted_at)
    WHERE deleted_at IS NULL;
