-- backend/patch_locations_metadata.sql

ALTER TABLE locations 
ADD COLUMN IF NOT EXISTS post_code TEXT,
ADD COLUMN IF NOT EXISTS coordinates TEXT,
ADD COLUMN IF NOT EXISTS population TEXT,
ADD COLUMN IF NOT EXISTS area TEXT;
