-- Add slug, county_slug, crest, parent_id to locations table.
-- These columns are used by admin_locations, public handlers, weather, and search.
-- Slug and county_slug are computed from name/county by the backend on create/update.

ALTER TABLE locations
  ADD COLUMN IF NOT EXISTS slug VARCHAR(120),
  ADD COLUMN IF NOT EXISTS county_slug VARCHAR(120),
  ADD COLUMN IF NOT EXISTS crest TEXT,
  ADD COLUMN IF NOT EXISTS parent_id INTEGER REFERENCES locations(id);

-- Backfill slug and county_slug for existing rows (matches backend utils.Slugify logic)
-- Backend: lower, replace Hungarian accents, replace non-alphanumeric with dash, trim
UPDATE locations
SET slug = trim(both '-' from regexp_replace(lower(translate(name, 'áéíóöőúüűÁÉÍÓÖŐÚÜŰ', 'aeiooouuuAEIOOOUUU')), '[^a-z0-9]+', '-', 'g'))
WHERE slug IS NULL AND name IS NOT NULL;
UPDATE locations
SET county_slug = trim(both '-' from regexp_replace(lower(translate(county, 'áéíóöőúüűÁÉÍÓÖŐÚÜŰ', 'aeiooouuuAEIOOOUUU')), '[^a-z0-9]+', '-', 'g'))
WHERE county_slug IS NULL AND county IS NOT NULL;

-- Insert county records for Székelyföld if not present (Hungarian names for consistency)
INSERT INTO locations (name, name_ro, county, county_slug, type, slug)
SELECT 'Hargita', 'Harghita', 'Hargita', 'hargita', 'megye', 'hargita'
WHERE NOT EXISTS (SELECT 1 FROM locations WHERE type = 'megye' AND (LOWER(COALESCE(county_slug,'')) = 'hargita' OR LOWER(COALESCE(county,'')) IN ('hargita','harghita')));
INSERT INTO locations (name, name_ro, county, county_slug, type, slug)
SELECT 'Kovászna', 'Covasna', 'Kovászna', 'kovaszna', 'megye', 'kovaszna'
WHERE NOT EXISTS (SELECT 1 FROM locations WHERE type = 'megye' AND (LOWER(COALESCE(county_slug,'')) = 'kovaszna' OR LOWER(COALESCE(county,'')) IN ('kovászna','kovaszna','covasna')));
INSERT INTO locations (name, name_ro, county, county_slug, type, slug)
SELECT 'Maros', 'Mureș', 'Maros', 'maros', 'megye', 'maros'
WHERE NOT EXISTS (SELECT 1 FROM locations WHERE type = 'megye' AND (LOWER(COALESCE(county_slug,'')) = 'maros' OR LOWER(COALESCE(county,'')) IN ('maros','mureș','mures')));
