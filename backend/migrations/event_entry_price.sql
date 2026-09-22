-- Optional belépő / jegyár (szabad szöveg, pl. "15", "10–15", "ingyenes"); üres = nincs megadva.
ALTER TABLE events ADD COLUMN IF NOT EXISTS entry_price VARCHAR(128) DEFAULT '';
