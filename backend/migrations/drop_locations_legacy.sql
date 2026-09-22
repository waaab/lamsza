-- Frozen pre-settlements copy of the old `locations` table.
-- Live reads use the `locations` view (counties UNION settlements).
DROP TABLE IF EXISTS locations_legacy;
