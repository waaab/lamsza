-- Replace entries.type varchar with type_id FK to entry_types.
-- Applied on backend start by handlers.MigrateEntryTypes (idempotent).

INSERT INTO entry_types (name) VALUES
    ('Szolgáltatás'),
    ('Cég'),
    ('Egyéb')
ON CONFLICT (name) DO NOTHING;

ALTER TABLE entries ADD COLUMN IF NOT EXISTS type_id INTEGER;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'entries' AND column_name = 'type'
    ) THEN
        UPDATE entries SET type = 'Szolgáltatás'
         WHERE lower(trim(type)) IN ('service', 'szolgáltatás')
            OR lower(trim(type)) = lower('Szolgáltatás');
        UPDATE entries SET type = 'Cég'
         WHERE lower(trim(type)) IN ('cég', 'ceg', 'company')
            OR lower(trim(type)) = lower('Cég');
        UPDATE entries SET type = 'Egyéb'
         WHERE lower(trim(type)) IN ('entry', 'egyéb', 'egyeb', 'other')
            OR lower(trim(type)) = lower('Egyéb');
        UPDATE entries SET type = 'Szolgáltatás' WHERE trim(COALESCE(type, '')) = '';

        UPDATE entries e
        SET type_id = et.id
        FROM entry_types et
        WHERE e.type_id IS NULL AND et.name = e.type;

        UPDATE entries
        SET type_id = (SELECT id FROM entry_types WHERE name = 'Szolgáltatás' LIMIT 1)
        WHERE type_id IS NULL;

        ALTER TABLE entries DROP COLUMN type;
    END IF;
END $$;

UPDATE entries
SET type_id = (SELECT id FROM entry_types WHERE name = 'Szolgáltatás' LIMIT 1)
WHERE type_id IS NULL;

ALTER TABLE entries ALTER COLUMN type_id SET NOT NULL;
ALTER TABLE entries DROP CONSTRAINT IF EXISTS entries_type_id_fkey;
ALTER TABLE entries ADD CONSTRAINT entries_type_id_fkey
    FOREIGN KEY (type_id) REFERENCES entry_types(id) ON DELETE RESTRICT;
CREATE INDEX IF NOT EXISTS idx_entries_type_id ON entries (type_id);
