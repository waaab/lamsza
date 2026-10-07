-- admin_audit_log - the write trail for the lamsza-admin API (BOG-48).
--
-- Nothing recorded which admin changed what. `lamsza-admin` is the only writer
-- of the directory this site reads, so every mutating call on its API lands one
-- row here: who, when, which resource, which action, and the before/after of
-- the row it touched.
--
-- This file lives in the `lamsza` repo because this backend owns the schema of
-- the shared `lamsza` database; the admin process runs no DDL at all (BOG-39,
-- held by `lamsza-admin/backend/boot_ddl_test.go`). Nothing in this repo reads
-- or writes the table - `lamsza-admin/backend/internal/audit` does.
--
-- `auth.migrateAdminAuditLog()` runs the same statements on boot; this is the
-- explicit form.
--
-- APPEND-ONLY FROM THE APP. The admin API only ever INSERTs. There is no admin
-- route that UPDATEs or DELETEs a row here, and `internal/audit` carries a
-- source-derived test that fails if one appears. Pruning is a deliberate
-- operator action against the database, not something a request can trigger.
--
-- RETENTION AND SIZE. No automatic retention: rows live until someone prunes
-- them. The app bounds the row, not the table - `payload`, `before_state` and
-- `after_state` are each capped at 16 KiB by the writer, which replaces an
-- oversized value with {"_truncated":true,...}. A heavy admin day is a few
-- thousand rows of a few KiB, so this grows by megabytes a year, not
-- gigabytes. Prune by age when it is ever worth it:
--
--   DELETE FROM admin_audit_log WHERE occurred_at < NOW() - INTERVAL '2 years';
--
-- See lamsza-admin/docs/ARCHITECTURE.md for the full note.

CREATE TABLE IF NOT EXISTS admin_audit_log (
    id BIGSERIAL PRIMARY KEY,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Who. The email is denormalised on purpose: the row has to stay readable
    -- after the user is deleted, which is exactly when it matters most.
    actor_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    actor_email TEXT NOT NULL DEFAULT '',

    -- What. `action` is create | update | delete | other, derived from the
    -- method; `route` and `method` keep the raw call for anything the resource
    -- name loses.
    resource TEXT NOT NULL,
    resource_id TEXT NOT NULL DEFAULT '',
    action TEXT NOT NULL,
    method TEXT NOT NULL,
    route TEXT NOT NULL,
    status_code INTEGER NOT NULL DEFAULT 0,

    -- Enough to undo a mistake. `payload` is the request body as sent (with
    -- credential-shaped keys redacted); before_state/after_state are the row
    -- snapshots; `diff` is the changed columns only.
    payload JSONB,
    before_state JSONB,
    after_state JSONB,
    diff JSONB
);

CREATE INDEX IF NOT EXISTS idx_admin_audit_log_occurred ON admin_audit_log(occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_actor ON admin_audit_log(actor_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_resource ON admin_audit_log(resource, resource_id);
