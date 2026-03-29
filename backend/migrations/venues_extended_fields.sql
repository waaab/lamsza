-- Extra fields for venues (multilingual names, geo, capacity, public description)
ALTER TABLE venues ADD COLUMN IF NOT EXISTS name_ro VARCHAR(255) DEFAULT '';
ALTER TABLE venues ADD COLUMN IF NOT EXISTS name_de VARCHAR(255) DEFAULT '';
ALTER TABLE venues ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION;
ALTER TABLE venues ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION;
ALTER TABLE venues ADD COLUMN IF NOT EXISTS seating_capacity INTEGER;
ALTER TABLE venues ADD COLUMN IF NOT EXISTS description TEXT;

COMMENT ON COLUMN venues.name IS 'Hungarian display name (primary)';
COMMENT ON COLUMN venues.name_ro IS 'Romanian name';
COMMENT ON COLUMN venues.name_de IS 'German name';
COMMENT ON COLUMN venues.seating_capacity IS 'Approximate seating / capacity (férőhely)';
