-- 012_create_pos_sessions.sql
-- Track sesi aktif operator. Owner bisa lihat siapa yang sedang login
-- dan force-end session tertentu.
--
-- ended_at NULL = sesi masih aktif.
-- last_active_at diupdate setiap kali operator membuat/mengubah order.

CREATE TABLE IF NOT EXISTS pos_sessions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id    UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    business_id    UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    started_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at       TIMESTAMPTZ,
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pos_sessions_operator_id  ON pos_sessions(operator_id);
CREATE INDEX IF NOT EXISTS idx_pos_sessions_business_id  ON pos_sessions(business_id);
CREATE INDEX IF NOT EXISTS idx_pos_sessions_ended_at     ON pos_sessions(ended_at);
