Skip to content
theoneandonlyvabo
qios-web
Repository navigation
Code
Issues
6
 (6)
Pull requests
Agents
Discussions
Actions
Projects
Wiki
Security and quality
Insights
Settings
Files
Go to file
t
T
operator content loaded
apps
client
server
bruno
cmd
config
domain
migrations
001_create_users.sql
002_create_refresh_tokens.sql
003_create_password_reset_tokens.sql
005_create_business.sql
006_create_operators.sql
007_create_products.sql
008_create_pos_orders.sql
009_create_pos_order_items.sql
010_create_xendit_payments.sql
011_create_webhook_events.sql
012_create_admin_audit_logs.sql
platform
.env.example
CLAUDE.md
Dockerfile
go.mod
go.sum
bruno
docs
infra
.gitignore
CLAUDE.md
Makefile
README.md
docker-compose.yml
folder.bru
qios-web/apps/server/migrations
/002_create_refresh_tokens.sql
theoneandonlyvabo
theoneandonlyvabo
Updated Migrations -vano
69b96ba
 · 
3 weeks ago

Code

Blame
-- 002_create_refresh_tokens.sql
-- Refresh token untuk multi-device session. Token di-hash sebelum disimpan.

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id    ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
 
operator content loaded