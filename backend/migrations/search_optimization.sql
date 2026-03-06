-- 1. Add denormalized columns to 'entries'
ALTER TABLE entries 
ADD COLUMN IF NOT EXISTS loc_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS loc_name_ro VARCHAR(255),
ADD COLUMN IF NOT EXISTS loc_name_de VARCHAR(255),
ADD COLUMN IF NOT EXISTS cat_name VARCHAR(100),
ADD COLUMN IF NOT EXISTS tag_names TEXT;

-- 2. Create Postgres slugify function (matching Go version)
CREATE OR REPLACE FUNCTION pg_slugify(text) RETURNS text AS $$
DECLARE
    s text;
BEGIN
    s := lower($1);
    -- Simple Hungarian replacement
    s := translate(s, 'áéíóöőúüű', 'aeiooouuu');
    -- Replace non-alphanumeric with dash
    s := regexp_replace(s, '[^a-z0-9]+', '-', 'g');
    -- Trim dashes
    s := trim(both '-' from s);
    RETURN s;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- 3. DROP old search_vector if exists
ALTER TABLE entries DROP COLUMN IF EXISTS search_vector;

-- 4. Create Weighted search_vector
-- Weights: A (Entry), B (Location), C (Category/Tags), D (Notes/Address)
ALTER TABLE entries 
ADD COLUMN search_vector tsvector 
GENERATED ALWAYS AS (
    setweight(to_tsvector('simple', pg_slugify(name)), 'A') || 
    setweight(to_tsvector('simple', coalesce(pg_slugify(loc_name), '')), 'B') || 
    setweight(to_tsvector('simple', coalesce(pg_slugify(loc_name_ro), '')), 'B') || 
    setweight(to_tsvector('simple', coalesce(pg_slugify(loc_name_de), '')), 'B') || 
    setweight(to_tsvector('simple', coalesce(pg_slugify(cat_name), '')), 'C') || 
    setweight(to_tsvector('simple', coalesce(pg_slugify(tag_names), '')), 'C') || 
    setweight(to_tsvector('simple', coalesce(pg_slugify(notes), '')), 'D') || 
    setweight(to_tsvector('simple', coalesce(pg_slugify(address), '')), 'D')
) STORED;

-- 5. Create GIN index
CREATE INDEX IF NOT EXISTS idx_entries_search_vector ON entries USING GIN(search_vector);

-- 6. Trigger for Locations sync
CREATE OR REPLACE FUNCTION sync_entry_location()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE entries
    SET loc_name = NEW.name,
        loc_name_ro = NEW.name_ro,
        loc_name_de = NEW.name_de
    WHERE location_id = NEW.id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_entry_location ON locations;
CREATE TRIGGER trg_sync_entry_location
AFTER UPDATE OF name, name_ro, name_de ON locations
FOR EACH ROW EXECUTE FUNCTION sync_entry_location();

-- 7. Trigger for Categories sync
CREATE OR REPLACE FUNCTION sync_entry_category()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE entries
    SET cat_name = NEW.name
    WHERE category_id = NEW.id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_entry_category ON entry_categories;
CREATE TRIGGER trg_sync_entry_category
AFTER UPDATE OF name ON entry_categories
FOR EACH ROW EXECUTE FUNCTION sync_entry_category();

-- 8. Trigger for Tags sync
CREATE OR REPLACE FUNCTION sync_entry_tags()
RETURNS TRIGGER AS $$
DECLARE
    e_id INT;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        e_id := OLD.entry_id;
    ELSE
        e_id := NEW.entry_id;
    END IF;

    UPDATE entries
    SET tag_names = (
        SELECT string_agg(t.name, ' ')
        FROM tags t
        JOIN entry_tags et ON t.id = et.tag_id
        WHERE et.entry_id = e_id
    )
    WHERE id = e_id;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_entry_tags ON entry_tags;
CREATE TRIGGER trg_sync_entry_tags
AFTER INSERT OR UPDATE OR DELETE ON entry_tags
FOR EACH ROW EXECUTE FUNCTION sync_entry_tags();

-- 9. Initial Sync
UPDATE entries e
SET loc_name = l.name,
    loc_name_ro = l.name_ro,
    loc_name_de = l.name_de
FROM locations l
WHERE e.location_id = l.id;

UPDATE entries e
SET cat_name = ec.name
FROM entry_categories ec
WHERE e.category_id = ec.id;

UPDATE entries e
SET tag_names = (
    SELECT string_agg(t.name, ' ')
    FROM tags t
    JOIN entry_tags et ON t.id = et.tag_id
    WHERE et.entry_id = e.id
);
