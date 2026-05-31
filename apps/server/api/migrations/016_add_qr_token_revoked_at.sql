-- 016_add_qr_token_revoked_at.sql
-- Adds revocation timestamp to operators so that regenerating a QR token
-- immediately invalidates the old one. Any QR login is rejected if the
-- token was issued before qr_token_revoked_at.

ALTER TABLE operators ADD COLUMN IF NOT EXISTS qr_token_revoked_at TIMESTAMPTZ;
