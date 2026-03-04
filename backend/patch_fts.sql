ALTER TABLE services 
ADD COLUMN IF NOT EXISTS search_vector tsvector 
GENERATED ALWAYS AS (
    setweight(to_tsvector('simple', coalesce(name, '')), 'A') || 
    setweight(to_tsvector('simple', coalesce(category, '')), 'B') || 
    setweight(to_tsvector('simple', coalesce(notes, '')), 'C') || 
    setweight(to_tsvector('simple', coalesce(address, '')), 'D')
) STORED;

CREATE INDEX IF NOT EXISTS idx_services_search ON services USING GIN(search_vector);
