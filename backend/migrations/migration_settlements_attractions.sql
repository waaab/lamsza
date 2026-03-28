-- Migration: Settlements, Geo Locations, Counties, Historical Seats, Attractions
-- Run after: schema.sql, patch_locations_metadata, patch_county_seat, patch_locations_slugs
-- Run: psql $DATABASE_URL -f backend/migrations/migration_settlements_attractions.sql

BEGIN;

-- 1. Drop trigger that syncs locations to entries
DROP TRIGGER IF EXISTS trg_sync_entry_location ON locations;

-- 2. Create geo_locations (geographic coordinates)
CREATE TABLE IF NOT EXISTS geo_locations (
    id SERIAL PRIMARY KEY,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    address TEXT,
    elevation INTEGER
);

-- 3. Create historical_seats
CREATE TABLE IF NOT EXISTS historical_seats (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    name_ro VARCHAR(255),
    name_de VARCHAR(255),
    slug VARCHAR(120) NOT NULL UNIQUE,
    content TEXT
);

-- 4. Create counties
CREATE TABLE IF NOT EXISTS counties (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    name_ro VARCHAR(255),
    name_de VARCHAR(255),
    slug VARCHAR(120) NOT NULL UNIQUE,
    location_id INTEGER REFERENCES geo_locations(id),
    content TEXT
);

-- 5. County - Historical seat junction
CREATE TABLE IF NOT EXISTS county_historical_seats (
    county_id INTEGER REFERENCES counties(id) ON DELETE CASCADE,
    historical_seat_id INTEGER REFERENCES historical_seats(id) ON DELETE CASCADE,
    PRIMARY KEY (county_id, historical_seat_id)
);

-- 6. Create settlements
CREATE TABLE IF NOT EXISTS settlements (
    id SERIAL PRIMARY KEY,
    county_id INTEGER NOT NULL REFERENCES counties(id),
    name VARCHAR(255) NOT NULL,
    name_ro VARCHAR(255),
    name_de VARCHAR(255),
    slug VARCHAR(120) NOT NULL,
    type VARCHAR(50) NOT NULL,
    location_id INTEGER REFERENCES geo_locations(id),
    parent_id INTEGER REFERENCES settlements(id),
    post_code VARCHAR(20),
    population VARCHAR(50),
    area VARCHAR(50),
    crest TEXT,
    is_county_seat BOOLEAN DEFAULT false,
    content TEXT,
    UNIQUE(slug, county_id)
);

-- 7. Create attractions
CREATE TABLE IF NOT EXISTS attractions (
    id SERIAL PRIMARY KEY,
    county_id INTEGER NOT NULL REFERENCES counties(id),
    name VARCHAR(255) NOT NULL,
    name_ro VARCHAR(255),
    name_de VARCHAR(255),
    slug VARCHAR(120) NOT NULL,
    description TEXT,
    location_id INTEGER REFERENCES geo_locations(id),
    featured_image TEXT,
    content TEXT,
    UNIQUE(slug, county_id)
);

-- 8. Attraction images gallery
CREATE TABLE IF NOT EXISTS attraction_images (
    id SERIAL PRIMARY KEY,
    attraction_id INTEGER NOT NULL REFERENCES attractions(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    sort_order INTEGER DEFAULT 0
);

-- 9. Insert historical seats
INSERT INTO historical_seats (name, name_ro, slug, content) VALUES
('Csíkszék', 'Ținutul Ciuc', 'csikszek', ''),
('Udvarhelyszék', 'Ținutul Odorhei', 'udvarhelyszek', ''),
('Háromszék', 'Trei Scaune', 'haromszek', ''),
('Marosszék', 'Ținutul Mureș', 'marosszek', ''),
('Aranyosszék', 'Ținutul Arieș', 'aranyosszek', '')
ON CONFLICT (slug) DO NOTHING;

-- 10. Ensure we have 3 counties (use high ids 1000+ to avoid overlap with settlement ids)
INSERT INTO counties (id, name, name_ro, slug) VALUES
(1000, 'Hargita', 'Harghita', 'hargita'),
(1001, 'Kovászna', 'Covasna', 'kovaszna'),
(1002, 'Maros', 'Mureș', 'maros')
ON CONFLICT (slug) DO NOTHING;

SELECT setval('counties_id_seq', 1003);

-- 11. Link counties to historical seats (county id 1000, 1001, 1002)
INSERT INTO county_historical_seats (county_id, historical_seat_id)
SELECT c.id, h.id FROM counties c, historical_seats h 
WHERE c.slug = 'hargita' AND h.slug IN ('csikszek', 'udvarhelyszek')
ON CONFLICT DO NOTHING;
INSERT INTO county_historical_seats (county_id, historical_seat_id)
SELECT c.id, h.id FROM counties c, historical_seats h 
WHERE c.slug = 'kovaszna' AND h.slug = 'haromszek'
ON CONFLICT DO NOTHING;
INSERT INTO county_historical_seats (county_id, historical_seat_id)
SELECT c.id, h.id FROM counties c, historical_seats h 
WHERE c.slug = 'maros' AND h.slug = 'marosszek'
ON CONFLICT DO NOTHING;

-- Counties must exist for migration - ensure ON CONFLICT works (slug is unique)

-- 12. Create geo_locations from settlement coordinates
CREATE TEMP TABLE _geo_mapping (old_loc_id INT, geo_id INT);
INSERT INTO geo_locations (latitude, longitude)
SELECT 
    trim(split_part(trim(coordinates), ',', 1))::double precision,
    trim(split_part(trim(coordinates), ',', 2))::double precision
FROM locations
WHERE type != 'megye' AND coordinates IS NOT NULL AND trim(coordinates) != ''
  AND trim(split_part(trim(coordinates), ',', 1)) ~ '^-?[0-9]+\.?[0-9]*$'
  AND trim(split_part(trim(coordinates), ',', 2)) ~ '^-?[0-9]+\.?[0-9]*$';

WITH numbered AS (
  SELECT id, row_number() OVER (ORDER BY id) as rn FROM locations
  WHERE type != 'megye' AND coordinates IS NOT NULL AND trim(coordinates) != ''
  AND trim(split_part(trim(coordinates), ',', 1)) ~ '^-?[0-9]+\.?[0-9]*$'
  AND trim(split_part(trim(coordinates), ',', 2)) ~ '^-?[0-9]+\.?[0-9]*$'
),
geo_numbered AS (
  SELECT id, row_number() OVER (ORDER BY id) as rn FROM geo_locations
)
INSERT INTO _geo_mapping (old_loc_id, geo_id)
SELECT n.id, g.id FROM numbered n JOIN geo_numbered g ON n.rn = g.rn;

-- 13. Migrate settlements from locations (preserve id for entries/events)
-- Map county_slug to county: hargita/harghita->hargita, kovaszna/covasna->kovaszna, maros/mures->maros
INSERT INTO settlements (id, county_id, name, name_ro, name_de, slug, type, location_id, parent_id, post_code, population, area, crest, is_county_seat)
SELECT 
    l.id,
    c.id,
    l.name,
    l.name_ro,
    l.name_de,
    COALESCE(l.slug, trim(both '-' from regexp_replace(lower(translate(l.name, 'áéíóöőúüű', 'aeiooouuu')), '[^a-z0-9]+', '-', 'g'))),
    l.type,
    (SELECT geo_id FROM _geo_mapping WHERE old_loc_id = l.id),
    l.parent_id,
    l.post_code,
    l.population,
    l.area,
    l.crest,
    COALESCE(l.is_county_seat, false)
FROM locations l
JOIN counties c ON c.slug = CASE 
  WHEN LOWER(COALESCE(l.county_slug, '')) IN ('hargita', 'harghita') THEN 'hargita'
  WHEN LOWER(COALESCE(l.county_slug, '')) IN ('kovaszna', 'covasna') THEN 'kovaszna'
  WHEN LOWER(COALESCE(l.county_slug, '')) IN ('maros', 'mures', 'mureș') THEN 'maros'
  ELSE LOWER(COALESCE(l.county_slug, 'hargita'))
END
WHERE l.type != 'megye'
ON CONFLICT DO NOTHING;

SELECT setval('settlements_id_seq', (SELECT COALESCE(MAX(id), 1) FROM settlements));

-- 14. Add settlement_id to entries, migrate, switch
ALTER TABLE entries ADD COLUMN IF NOT EXISTS settlement_id INTEGER;
UPDATE entries e SET settlement_id = e.location_id WHERE e.location_id IN (SELECT id FROM settlements);
ALTER TABLE entries DROP CONSTRAINT IF EXISTS entries_location_id_fkey;
ALTER TABLE entries DROP CONSTRAINT IF EXISTS entries_unique_entry;
ALTER TABLE entries ALTER COLUMN location_id DROP NOT NULL;
UPDATE entries SET location_id = NULL WHERE settlement_id IS NOT NULL;
ALTER TABLE entries RENAME COLUMN location_id TO location_id_old;
ALTER TABLE entries RENAME COLUMN settlement_id TO location_id;
ALTER TABLE entries ADD CONSTRAINT entries_location_id_fkey FOREIGN KEY (location_id) REFERENCES settlements(id);
ALTER TABLE entries DROP COLUMN location_id_old;
ALTER TABLE entries ADD CONSTRAINT entries_unique_entry UNIQUE (name, location_id);

-- 15. Same for events
ALTER TABLE events ADD COLUMN IF NOT EXISTS settlement_id INTEGER;
UPDATE events e SET settlement_id = e.location_id WHERE e.location_id IN (SELECT id FROM settlements);
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_location_id_fkey;
UPDATE events SET location_id = NULL WHERE settlement_id IS NOT NULL;
ALTER TABLE events RENAME COLUMN location_id TO location_id_old;
ALTER TABLE events RENAME COLUMN settlement_id TO location_id;
ALTER TABLE events ADD CONSTRAINT events_location_id_fkey FOREIGN KEY (location_id) REFERENCES settlements(id);
ALTER TABLE events DROP COLUMN location_id_old;

-- 16. Update entries denormalized columns
UPDATE entries e SET loc_name = s.name, loc_name_ro = COALESCE(s.name_ro, ''), loc_name_de = COALESCE(s.name_de, '')
FROM settlements s WHERE e.location_id = s.id;

-- 17. Recreate trigger for settlements
CREATE OR REPLACE FUNCTION sync_entry_settlement()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE entries SET loc_name = NEW.name, loc_name_ro = COALESCE(NEW.name_ro, ''), loc_name_de = COALESCE(NEW.name_de, '')
    WHERE location_id = NEW.id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_sync_entry_settlement ON settlements;
CREATE TRIGGER trg_sync_entry_settlement AFTER UPDATE OF name, name_ro, name_de ON settlements
FOR EACH ROW EXECUTE FUNCTION sync_entry_settlement();

-- 18. Insert Szent Anna-tó attraction
INSERT INTO geo_locations (latitude, longitude, address) VALUES (46.1265, 25.8876, 'Szent Anna-tó, Harghita');
INSERT INTO attractions (county_id, name, name_ro, slug, description, location_id, content)
SELECT c.id, 'Szent Anna-tó', 'Lacul Sfânta Ana', 'szent-anna-to', 
  'Közép-Európa egyetlen vulkanikus tava. 946 m tengerszint feletti magasságban.',
  (SELECT id FROM geo_locations ORDER BY id DESC LIMIT 1), ''
FROM counties c WHERE c.slug = 'hargita'
ON CONFLICT DO NOTHING;

-- 19. Rename old locations table and create view for backward compatibility
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'locations') THEN
    ALTER TABLE locations RENAME TO locations_legacy;
  END IF;
END $$;

-- 20. Create locations view (counties + settlements) for backward compat
CREATE OR REPLACE VIEW locations AS
SELECT id, name, name_ro, name_de, name as county, slug as county_slug, 'megye' as type, slug,
  NULL::text as post_code, NULL::text as coordinates, NULL::text as population, NULL::text as area,
  NULL::text as crest, NULL::integer as parent_id, false as is_county_seat
FROM counties
UNION ALL
SELECT s.id, s.name, s.name_ro, s.name_de, c.name, c.slug, s.type, s.slug, s.post_code,
  (SELECT gl.latitude::text || ', ' || gl.longitude::text FROM geo_locations gl WHERE gl.id = s.location_id),
  s.population, s.area, s.crest, s.parent_id, s.is_county_seat
FROM settlements s
JOIN counties c ON s.county_id = c.id;

COMMIT;
