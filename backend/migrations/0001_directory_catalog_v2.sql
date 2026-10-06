-- 0001_directory_catalog_v2.sql
--
-- ONE-SHOT. DESTRUCTIVE. RUN BY HAND. NEVER FROM THE BOOT PATH.
--
-- This is the directory catalog v2 reset. It used to run inside
-- handlers.MigrateDirectoryCatalog() on every backend start, guarded only by the
-- site_settings.directory_catalog_v2 row in the same database it empties. If that
-- row went missing (an older dump restore, an admin delete, an earlier bail-out
-- that never wrote the flag), the next restart destroyed the whole directory. It
-- is a file now so that it can only happen when a person asks for it.
--
-- WHAT IT DELETES
--   entries, tags, entry_categories, entry_types          -- directly
--   entry_tags, entry_members, entry_reviews,
--   entry_suggestions, entry_category_links               -- via ON DELETE CASCADE
--   websites.entry_id                                     -- set to NULL
--
-- It does not touch settlements, locations, users, sessions, events, news,
-- mondasok, pages, or the website submission rows themselves.
--
-- IT CAN REFUSE TO RUN, AND THAT IS SAFE
--   websites.category_id and website_category_links reference entry_categories
--   with NO ACTION, so DELETE FROM entry_categories fails while any approved
--   website still has a category. The whole transaction then rolls back and
--   nothing is lost. Clearing those website references is a separate decision --
--   make it on purpose, in a separate statement, not by widening this file.
--
-- BEFORE YOU RUN IT
--   1. Take a dump:
--        docker exec lamsza-db pg_dump -U lamsza_user lamsza > ~/lamsza-before-reset.sql
--   2. Be sure you want an empty directory. There is no undo in this file.
--
-- HOW TO RUN IT
--      docker exec -i lamsza-db psql -U lamsza_user -d lamsza \
--        < backend/migrations/0001_directory_catalog_v2.sql
--
-- AFTER YOU RUN IT
--   Restart the lamsza backend. handlers.MigrateDirectoryCatalog() then reseeds
--   the three entry types and the two-level category tree from
--   backend/internal/utils/directory_catalog.go, and realigns the id sequences.
--   That path only inserts missing rows, so it is safe on every later start too.
--   The seed lives in Go, not in this file, so there is one source of truth for
--   the catalog and this file cannot drift from it.

BEGIN;

DELETE FROM entries;
DELETE FROM tags;
DELETE FROM entry_categories;
DELETE FROM entry_types;

-- Kept for compatibility: the two legacy migrators
-- (handlers.MigrateEntryCategories, handlers.MigrateEntryTypes) still read this
-- flag to decide they have nothing to do.
INSERT INTO site_settings (key, value) VALUES ('directory_catalog_v2', '1')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;

COMMIT;
