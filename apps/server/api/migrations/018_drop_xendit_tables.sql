-- 018_drop_xendit_tables.sql
-- Drop tabel Xendit yang tidak dipakai setelah pivot ke in-house transaction.

DROP TABLE IF EXISTS xendit_payments;
DROP TABLE IF EXISTS webhook_events;
