-- Optional cover image for event cards (URL path from our media API or external https URL).
ALTER TABLE events ADD COLUMN IF NOT EXISTS featured_image VARCHAR(1024) DEFAULT '';
