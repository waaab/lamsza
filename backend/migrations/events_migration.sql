CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    location_id INTEGER REFERENCES locations(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    start_time TIME,
    end_date DATE NOT NULL,
    end_time TIME,
    event_type VARCHAR(50) DEFAULT 'other',
    organizer VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed some initial events
INSERT INTO events (location_id, title, description, start_date, start_time, end_date, end_time, event_type, organizer)
VALUES
(1, 'Csíksomlyói búcsú', 'A legnagyobb magyar katolikus búcsújáró ünnep.', '2026-06-06', '11:00', '2026-06-06', '18:00', 'cultural', 'Gyulafehérvári Római Katolikus Érsekség'),
(1, 'Hargita Rallye', 'Hagyományos aszfaltos autóverseny.', '2026-05-15', '09:00', '2026-05-15', '17:00', 'sports', 'HMT'),
(2, 'Székelyudvarhelyi Városnapok', 'Koncertek, vásár, szórakozás.', '2026-05-22', '10:00', '2026-05-24', '22:00', 'cultural', 'Polgármesteri Hivatal');
