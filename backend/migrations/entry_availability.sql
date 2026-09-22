-- Claimed badge + weekly opening / delivery hours for directory entries.
-- Applied on backend start by handlers.MigrateEntryAvailability (idempotent).

ALTER TABLE entries
    ADD COLUMN IF NOT EXISTS claimed BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE entries
    ADD COLUMN IF NOT EXISTS hours JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE entries
    ADD COLUMN IF NOT EXISTS delivery_hours JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE entries
    ADD COLUMN IF NOT EXISTS photos JSONB NOT NULL DEFAULT '[]'::jsonb;
