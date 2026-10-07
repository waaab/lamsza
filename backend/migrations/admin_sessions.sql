-- admin_sessions - the lamsza-admin app's session store (BOG-45).
--
-- Separate from `sessions` on purpose. Both apps used to mint into `sessions`
-- under one cookie name (`lamsza_session`), so a token handed out by the public
-- site was accepted by the admin API on :3000. Two stores make the boundary
-- something the database holds rather than something a header has to prove.
--
-- This file lives in the `lamsza` repo because this backend owns the schema of
-- the shared `lamsza` database; the admin process runs no DDL at all. Nothing
-- in this repo reads the table - `lamsza-admin/backend/internal/auth` does.
--
-- `auth.Migrate()` runs the same statements on boot; this is the explicit form.

CREATE TABLE IF NOT EXISTS admin_sessions (
    token_hash CHAR(64) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_sessions_user ON admin_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_expires ON admin_sessions(expires_at);
