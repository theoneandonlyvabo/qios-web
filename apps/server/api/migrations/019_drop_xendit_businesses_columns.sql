-- 019_drop_xendit_businesses_columns.sql
-- Drop kolom Xendit dari tabel businesses yang tidak dipakai.

ALTER TABLE businesses
    DROP COLUMN IF EXISTS xendit_account_id,
    DROP COLUMN IF EXISTS xendit_api_key,
    DROP COLUMN IF EXISTS xendit_secret_key;
