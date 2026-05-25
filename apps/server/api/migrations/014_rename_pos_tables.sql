-- 014_rename_pos_tables.sql
-- Rename pos_* tables and the pos_order_id FK column.
-- Run in a transaction; rollback if any step fails.
--
-- WARNING: Do NOT run against any database until application code is deployed.
-- Apply to staging first, verify row counts before/after, then production.

BEGIN;

ALTER TABLE pos_orders      RENAME TO orders;
ALTER TABLE pos_order_items RENAME TO order_items;
ALTER TABLE pos_sessions    RENAME TO order_sessions;

-- Rename the FK column that still has the old table prefix
ALTER TABLE order_items RENAME COLUMN pos_order_id TO order_id;

-- Rename indexes to match new table names (advisory — queries still work without this)
ALTER INDEX idx_pos_orders_business_id       RENAME TO idx_orders_business_id;
ALTER INDEX idx_pos_orders_order_id          RENAME TO idx_orders_order_id;
ALTER INDEX idx_pos_orders_status            RENAME TO idx_orders_status;
ALTER INDEX idx_pos_orders_payment_method    RENAME TO idx_orders_payment_method;
ALTER INDEX idx_pos_orders_created_at        RENAME TO idx_orders_created_at;
ALTER INDEX idx_pos_order_items_pos_order_id RENAME TO idx_order_items_order_id;
ALTER INDEX idx_pos_order_items_product_id   RENAME TO idx_order_items_product_id;
ALTER INDEX idx_pos_sessions_operator_id     RENAME TO idx_order_sessions_operator_id;
ALTER INDEX idx_pos_sessions_business_id     RENAME TO idx_order_sessions_business_id;
ALTER INDEX idx_pos_sessions_ended_at        RENAME TO idx_order_sessions_ended_at;

COMMIT;
