-- 014_alter_pos_orders.sql
--
-- DRAFT — PENDING VANO APPROVAL sebelum di-run.
-- Jangan tambahkan ke migration runner sampai schema ini di-approve.
--
-- Context: migration 008 membuat pos_orders tapi missing beberapa kolom
-- yang dibutuhkan untuk payment domain (lihat AGENTS.md section 7).
--
-- Changes:
--   1. Tambah payment_method — required untuk distinguish cash vs digital
--   2. Tambah updated_at — konsistensi dengan semua tabel lain
--   3. Tambah status 'CANCELLED' ke CHECK constraint
--
-- Catatan: Status case (lowercase vs uppercase) TIDAK diubah di migration ini
-- karena memerlukan data migration. Keputusan tentang case standardisasi
-- dibuat terpisah setelah diskusi dengan Vano.

ALTER TABLE pos_orders
    ADD COLUMN IF NOT EXISTS payment_method VARCHAR(20) NOT NULL DEFAULT 'CASH'
        CHECK (payment_method IN ('CASH', 'QRIS', 'EWALLET', 'VIRTUAL_ACCOUNT')),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Update CHECK constraint untuk status agar include CANCELLED.
-- Postgres tidak support ALTER CONSTRAINT langsung — harus drop dan recreate.
ALTER TABLE pos_orders DROP CONSTRAINT IF EXISTS pos_orders_status_check;
ALTER TABLE pos_orders ADD CONSTRAINT pos_orders_status_check
    CHECK (status IN ('pending', 'paid', 'failed', 'expired', 'cancelled'));

-- Index untuk payment_method — dipakai di filter analytics (e.g., revenue by method).
CREATE INDEX IF NOT EXISTS idx_pos_orders_payment_method ON pos_orders(payment_method);
