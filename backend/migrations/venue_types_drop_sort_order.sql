-- Venue types: manual order removed; lists use label (name) order from the API.
DROP INDEX IF EXISTS idx_venue_types_sort;
ALTER TABLE venue_types DROP COLUMN IF EXISTS sort_order;
