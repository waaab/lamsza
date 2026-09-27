ALTER TABLE entries ADD COLUMN IF NOT EXISTS hours_enabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE entries ADD COLUMN IF NOT EXISTS delivery_enabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE entries ADD COLUMN IF NOT EXISTS social_links JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE entries
SET hours_enabled = true
WHERE hours_enabled = false
  AND hours IS NOT NULL
  AND hours::text NOT IN ('{}', 'null');
