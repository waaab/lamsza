-- Daily quote: calendar day when the quote is shown on the homepage (Europe/Budapest “today” match).
ALTER TABLE mondasok ADD COLUMN IF NOT EXISTS display_date DATE;

UPDATE mondasok
SET display_date = COALESCE((created_at AT TIME ZONE 'UTC')::date, CURRENT_DATE)
WHERE display_date IS NULL;

UPDATE mondasok SET display_date = CURRENT_DATE WHERE display_date IS NULL;

ALTER TABLE mondasok ALTER COLUMN display_date SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_mondasok_display_date ON mondasok (display_date);
