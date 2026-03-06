CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_loc_hu_trgm ON locations USING gin (unaccent(lower(name_hu)) gin_trgm_ops);
CREATE INDEX idx_loc_ro_trgm ON locations USING gin (unaccent(lower(name_ro)) gin_trgm_ops);
CREATE INDEX idx_loc_de_trgm ON locations USING gin (unaccent(lower(name_de)) gin_trgm_ops);
