-- Gallery photos for directory entries (url, alt, title, description, width, height).
-- Applied on backend start by handlers.MigrateEntryAvailability (idempotent).

ALTER TABLE entries
    ADD COLUMN IF NOT EXISTS photos JSONB NOT NULL DEFAULT '[]'::jsonb;
