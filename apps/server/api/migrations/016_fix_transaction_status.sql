-- 016_fix_transaction_status.sql
-- Ganti status pos_orders dari Xendit-era lowercase ke in-house uppercase.
-- PENDING → order dibuat, belum dikonfirmasi
-- CONFIRMED → kasir konfirmasi + payment_method di-set
-- VOIDED → dibatalkan

ALTER TABLE pos_orders DROP CONSTRAINT IF EXISTS pos_orders_status_check;

UPDATE pos_orders SET status = 'PENDING'   WHERE status = 'pending';
UPDATE pos_orders SET status = 'CONFIRMED' WHERE status = 'paid';
UPDATE pos_orders SET status = 'VOIDED'    WHERE status IN ('cancelled', 'failed', 'expired');

ALTER TABLE pos_orders
    ALTER COLUMN status SET DEFAULT 'PENDING',
    ADD CONSTRAINT pos_orders_status_check CHECK (status IN ('PENDING', 'CONFIRMED', 'VOIDED'));

-- Rename paid_at → confirmed_at (lebih akurat untuk in-house flow)
ALTER TABLE pos_orders RENAME COLUMN paid_at TO confirmed_at;
