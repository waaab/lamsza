-- Optional per-day program for events (see events.start_date/end_date for fallback).
CREATE TABLE IF NOT EXISTS event_schedule_days (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    schedule_date DATE NOT NULL,
    notes TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE (event_id, schedule_date)
);

CREATE INDEX IF NOT EXISTS idx_event_schedule_days_event ON event_schedule_days(event_id);

CREATE TABLE IF NOT EXISTS event_schedule_activities (
    id SERIAL PRIMARY KEY,
    event_day_id INTEGER NOT NULL REFERENCES event_schedule_days(id) ON DELETE CASCADE,
    starts_at TIME,
    ends_at TIME,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_event_schedule_activities_day ON event_schedule_activities(event_day_id);
