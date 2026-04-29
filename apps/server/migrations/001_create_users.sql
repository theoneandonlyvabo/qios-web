-- 001_create_users.sql
-- Tabel utama untuk semua akun (owner bisnis).
-- Google OAuth didukung: password_hash nullable untuk akun OAuth-only.

CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT,                          -- NULL kalau login via Google OAuth
    full_name   VARCHAR(255) NOT NULL,
    google_id   VARCHAR(255) UNIQUE,             -- NULL kalau registrasi email/password
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    is_suspended BOOLEAN NOT NULL DEFAULT FALSE, -- A-03: admin bisa suspend
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);
