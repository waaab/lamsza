-- Classify schedule lines (opening, matches, closing, etc.). Optional end time already supported (NULL ends_at).
ALTER TABLE event_schedule_activities
    ADD COLUMN IF NOT EXISTS activity_type VARCHAR(40) NOT NULL DEFAULT 'other';

COMMENT ON COLUMN event_schedule_activities.activity_type IS 'opening | match | closing | other';
