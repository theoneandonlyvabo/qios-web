-- 017_rename_xendit_status_to_merchant_status.sql
-- Rename xendit_status → merchant_status di tabel businesses.
-- Semantik sama (PENDING/REGISTERED/ACTIVE/SUSPENDED), nama lebih netral.

ALTER TABLE businesses RENAME COLUMN xendit_status TO merchant_status;
