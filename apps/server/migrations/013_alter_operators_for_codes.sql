-- 013_alter_operators_for_codes.sql
-- Refactor tabel operators: ganti login berbasis email menjadi operator_code + qr_token.
-- operator_code: ID login yang di-set owner, unik per bisnis (bukan global).
-- qr_token: token statis yang di-encode ke QR untuk login scan, unik global.
-- Tambah updated_at dan deleted_at untuk soft delete + audit timestamp.

-- Bersihkan unique constraint dan index lama dari kolom email.
ALTER TABLE operators DROP CONSTRAINT IF EXISTS operators_email_key;
DROP INDEX IF EXISTS idx_operators_email;

-- Tambah kolom baru. Nullable dulu supaya migrasi tidak gagal kalau ada data.
ALTER TABLE operators
    ADD COLUMN IF NOT EXISTS operator_code VARCHAR(64),
    ADD COLUMN IF NOT EXISTS qr_token      VARCHAR(128),
    ADD COLUMN IF NOT EXISTS updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS deleted_at    TIMESTAMPTZ;

-- Backfill data existing: pakai email sebagai operator_code, generate qr_token random.
-- gen_random_uuid() built-in di Postgres 13+, dipakai dua kali untuk dapat 64 hex chars.
UPDATE operators
   SET operator_code = COALESCE(operator_code, email),
       qr_token      = COALESCE(
                         qr_token,
                         replace(gen_random_uuid()::text, '-', '') ||
                         replace(gen_random_uuid()::text, '-', '')
                       )
 WHERE operator_code IS NULL OR qr_token IS NULL;

-- Setelah backfill, baru kunci NOT NULL.
ALTER TABLE operators
    ALTER COLUMN operator_code SET NOT NULL,
    ALTER COLUMN qr_token      SET NOT NULL;

-- Drop kolom email beserta semua dependency (constraint, index).
ALTER TABLE operators DROP COLUMN IF EXISTS email CASCADE;

-- Unique constraint:
--   operator_code unik per bisnis (bukan global) — owner bebas pakai "kasir-1" di tiap toko.
--   Partial index supaya operator yang sudah di-soft-delete tidak ikut diuji unik,
--   sehingga code bisa dipakai ulang setelah dihapus.
CREATE UNIQUE INDEX IF NOT EXISTS uq_operators_business_code
    ON operators(business_id, operator_code)
    WHERE deleted_at IS NULL;

-- qr_token unik global karena login scan QR tidak punya konteks business_id dulu.
CREATE UNIQUE INDEX IF NOT EXISTS uq_operators_qr_token
    ON operators(qr_token);

-- Helper index untuk lookup cepat dari handler kasir.
CREATE INDEX IF NOT EXISTS idx_operators_qr_token ON operators(qr_token);
