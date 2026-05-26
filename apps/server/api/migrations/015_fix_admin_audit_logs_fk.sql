-- 015_fix_admin_audit_logs_fk.sql
-- Correction: admin_audit_logs.admin_id must reference admin_users, not users.
-- The original migration 009 was created before admin_users (migration 010).

ALTER TABLE admin_audit_logs
    DROP CONSTRAINT IF EXISTS admin_audit_logs_admin_id_fkey;

ALTER TABLE admin_audit_logs
    ADD CONSTRAINT admin_audit_logs_admin_id_fkey
        FOREIGN KEY (admin_id) REFERENCES admin_users(id) ON DELETE SET NULL;
