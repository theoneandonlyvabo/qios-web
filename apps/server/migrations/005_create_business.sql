-- 005_create_businesses.sql
-- Satu bisnis per owner (1:1).
-- Dibuat saat owner submit form daftar QIOS — atomic dengan insert users.

-- xendit_status melacak siklus aktivasi sub-account Xendit (xenPlatform MANAGED).
--   PENDING    : sub-account Xendit belum dibuat
--   REGISTERED : sub-account ada, KYC belum selesai
--   ACTIVE     : fully operational, bisa transaksi
--   SUSPENDED  : akun dibekukan Xendit (fraud/review)

CREATE TABLE IF NOT EXISTS businesses (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    qm_id              VARCHAR(20)  NOT NULL UNIQUE,         -- format QM-000001, generate di application layer
    user_id            UUID         NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    business_name      VARCHAR(255) NOT NULL,
    phone              VARCHAR(32),
    address            TEXT,
    city               VARCHAR(100),
    country            VARCHAR(100),

    xendit_account_id  VARCHAR(255),
    xendit_api_key     TEXT,
    xendit_secret_key  TEXT,
    xendit_status      VARCHAR(20)  NOT NULL DEFAULT 'PENDING'
                           CHECK (xendit_status IN ('PENDING', 'REGISTERED', 'ACTIVE', 'SUSPENDED')),

    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_businesses_user_id ON businesses(user_id);
CREATE INDEX IF NOT EXISTS idx_businesses_qm_id   ON businesses(qm_id);