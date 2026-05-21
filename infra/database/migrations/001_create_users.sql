-- 001_create_users.sql
-- Akun owner bisnis. Support email/password dan Google OAuth.

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT,                             -- NULL kalau login via Google OAuth
    full_name     VARCHAR(255) NOT NULL,
    phone         VARCHAR(20),                      -- opsional, bisa diisi setelah register
    google_id     VARCHAR(255) UNIQUE,              -- NULL kalau registrasi email/password
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    is_suspended  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email     ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);