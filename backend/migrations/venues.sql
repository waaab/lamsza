-- Venues: named places (arenas, markets, parks) within a settlement. Events can have a default venue; schedule lines can override per activity.

CREATE TABLE IF NOT EXISTS venues (
    id SERIAL PRIMARY KEY,
    settlement_id INTEGER NOT NULL REFERENCES settlements(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    name_ro VARCHAR(255) DEFAULT '',
    name_de VARCHAR(255) DEFAULT '',
    slug VARCHAR(200) NOT NULL,
    kind VARCHAR(40) NOT NULL DEFAULT 'other',
    address TEXT,
    notes TEXT,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    seating_capacity INTEGER,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT venues_settlement_slug_unique UNIQUE (settlement_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_venues_settlement ON venues(settlement_id);

COMMENT ON TABLE venues IS 'Concrete event sites (stadium, rink, square) tied to one settlement.';
COMMENT ON COLUMN venues.kind IS 'sports_arena | indoor_hall | outdoor_area | market_square | park | street | mixed | temporary | other';

ALTER TABLE events ADD COLUMN IF NOT EXISTS default_venue_id INTEGER REFERENCES venues(id) ON DELETE SET NULL;

ALTER TABLE event_schedule_activities ADD COLUMN IF NOT EXISTS venue_id INTEGER REFERENCES venues(id) ON DELETE SET NULL;

-- Example: ice rink in Kézdivásárhely (if settlement exists)
INSERT INTO venues (settlement_id, name, slug, kind)
SELECT s.id, 'Deme László Műjégpálya', 'deme-laszlo-mujegpalya', 'sports_arena'
FROM settlements s
JOIN counties c ON s.county_id = c.id
WHERE c.slug = 'kovaszna' AND s.slug = 'kezdivasarhely'
ON CONFLICT (settlement_id, slug) DO NOTHING;
