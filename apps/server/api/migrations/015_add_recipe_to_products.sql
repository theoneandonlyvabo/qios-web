-- 015_add_recipe_to_products.sql
-- Tambah kolom recipe ke products untuk consumption tracking.
-- Format: [{"ingredient": "Kopi", "quantity": 18.0, "unit": "gram"}]

ALTER TABLE products ADD COLUMN IF NOT EXISTS recipe JSONB;
