-- 015_alter_xendit_payments_qr_string.sql
--
-- Migration 010 created xendit_payments tetapi belum menyimpan QR string mentah
-- yang dipakai frontend kasir untuk render QR. qr_string genuinely missing —
-- bukan duplikat dari kolom lain dan tidak masuk ke raw_payload secara konsisten.
--
-- Tambah kolom qr_string TEXT. Nullable karena tidak semua payment_method
-- menghasilkan QR (EWALLET / VIRTUAL_ACCOUNT pakai instruksi lain).

ALTER TABLE xendit_payments
    ADD COLUMN IF NOT EXISTS qr_string TEXT;
