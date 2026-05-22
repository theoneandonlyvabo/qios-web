-- 013_add_qris_string_to_businesses.sql
-- QRIS string milik merchant — diperoleh dari bank/PJSP provider mereka, bukan dari Xendit.
-- Nullable: tidak semua merchant menggunakan QRIS.

ALTER TABLE businesses ADD COLUMN IF NOT EXISTS qris_string TEXT;
