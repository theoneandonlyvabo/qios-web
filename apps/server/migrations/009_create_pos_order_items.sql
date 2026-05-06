-- 009_create_pos_order_items.sql
-- Item per order. product_name dan unit_price disimpan sebagai snapshot
-- supaya data historis tidak berubah kalau produk diedit atau dihapus.

CREATE TABLE IF NOT EXISTS pos_order_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pos_order_id UUID NOT NULL REFERENCES pos_orders(id) ON DELETE CASCADE,
    product_id   UUID REFERENCES products(id) ON DELETE SET NULL,
    product_name VARCHAR(255) NOT NULL,             -- snapshot saat transaksi
    unit_price   BIGINT NOT NULL CHECK (unit_price >= 0), -- snapshot saat transaksi
    quantity     INTEGER NOT NULL CHECK (quantity > 0),
    subtotal     BIGINT GENERATED ALWAYS AS (quantity * unit_price) STORED
);

CREATE INDEX IF NOT EXISTS idx_pos_order_items_pos_order_id ON pos_order_items(pos_order_id);
CREATE INDEX IF NOT EXISTS idx_pos_order_items_product_id   ON pos_order_items(product_id);