-- 021_add_pos_order_draft_status_and_checkout_tracking.sql
-- Tambah status DRAFT sebagai state awal order (order sedang diisi, belum checkout).
-- Flow baru: DRAFT → PENDING (checkout dimulai) → CONFIRMED / VOIDED
--
-- Juga tambah checkout_started_at untuk validasi slide-to-confirm server-side:
-- server catat waktu saat operator mulai checkout, lalu validasi >= 800ms
-- saat confirm diterima.

ALTER TABLE pos_orders DROP CONSTRAINT IF EXISTS pos_orders_status_check;

ALTER TABLE pos_orders
    ALTER COLUMN status SET DEFAULT 'DRAFT',
    ADD CONSTRAINT pos_orders_status_check
        CHECK (status IN ('DRAFT', 'PENDING', 'CONFIRMED', 'VOIDED'));

-- Migrasi baris existing: PENDING tetap PENDING (sudah di-checkout flow lama)
-- Tidak ada data yang perlu di-backfill ke DRAFT.

ALTER TABLE pos_orders
    ADD COLUMN IF NOT EXISTS checkout_started_at TIMESTAMPTZ;
