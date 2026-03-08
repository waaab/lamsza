-- Add missing columns that backend handlers reference but the live DB lacks.

-- entry_categories.slug: used by HandleAdminEntryCategories POST/PUT
ALTER TABLE entry_categories ADD COLUMN IF NOT EXISTS slug VARCHAR(120) UNIQUE;

-- news_feeds.county_slug: used by HandleCountyNews WHERE clause
ALTER TABLE news_feeds ADD COLUMN IF NOT EXISTS county_slug VARCHAR(120);
