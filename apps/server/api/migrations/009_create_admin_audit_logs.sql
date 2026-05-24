-- 009_create_admin_audit_logs.sql
-- Setiap aksi admin tercatat: suspend user, ubah plan, akses data, dll.
-- admin_id NULL kalau aksi dilakukan oleh sistem otomatis.

CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    target_type VARCHAR(50) NOT NULL,               -- e.g. 'user', 'subscription', 'plan'
    target_id   UUID,
    action      VARCHAR(100) NOT NULL,              -- e.g. 'suspend_user', 'change_plan'
    meta        JSONB,                              -- { before: {}, after: {} }
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_admin_id   ON admin_audit_logs(admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_target     ON admin_audit_logs(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_created_at ON admin_audit_logs(created_at);
