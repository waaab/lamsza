-- Configurable venue kinds (labels); venues.kind stores the slug.
CREATE TABLE IF NOT EXISTS venue_types (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(64) NOT NULL UNIQUE,
    label_hu VARCHAR(255) NOT NULL
);

INSERT INTO venue_types (slug, label_hu) VALUES
    ('sports_arena', 'Sportcsarnok / pálya'),
    ('indoor_hall', 'Fedett csarnok'),
    ('outdoor_area', 'Szabadtéri terület'),
    ('market_square', 'Piac / tér'),
    ('park', 'Park'),
    ('street', 'Utca / felvonulás'),
    ('mixed', 'Több helyszín'),
    ('temporary', 'Ideiglenes'),
    ('other', 'Egyéb')
ON CONFLICT (slug) DO NOTHING;
