-- backend/schema.sql

CREATE TABLE IF NOT EXISTS locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    county VARCHAR(255),
    type VARCHAR(50),
    CONSTRAINT unique_location UNIQUE (name, county, type)
);

CREATE TABLE IF NOT EXISTS service_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    location_id INTEGER REFERENCES locations(id),
    category_id INTEGER REFERENCES service_categories(id),
    category VARCHAR(100),
    name VARCHAR(255) NOT NULL,
    url VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    notes TEXT,
    is_magyar_language BOOLEAN DEFAULT true,
    tags TEXT,
    CONSTRAINT unique_service UNIQUE (name, location_id)
);

CREATE TABLE IF NOT EXISTS mondasok (
    id SERIAL PRIMARY KEY,
    text TEXT NOT NULL,
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS quick_links (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    url VARCHAR(255) NOT NULL UNIQUE,
    bg_color VARCHAR(50) DEFAULT '#ffffff',
    match_category VARCHAR(100) 
);

CREATE TABLE IF NOT EXISTS news_feeds (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    feed_url VARCHAR(255) NOT NULL UNIQUE,
    bg_color VARCHAR(50) DEFAULT '#ffebd6'
);

-- Optional: Insert basic seed data so the API returns something immediately
INSERT INTO locations (id, name, county, type) VALUES 
(1, 'Csíkszereda', 'Hargita', 'város'),
(2, 'Székelyudvarhely', 'Hargita', 'város')
ON CONFLICT DO NOTHING;

INSERT INTO services (location_id, category, name, phone, address, notes, is_magyar_language) VALUES 
(1, 'egeszsegugy', 'Csíkszeredai Megyei Sürgősségi Kórház', '0266 324 193', 'Vakáció u. 1-3.', 'Sürgősség éjjel-nappal nyitva.', true),
(2, 'egeszsegugy', 'Dr. Papp Zoltán - Fogorvos', '0744 123 456', 'Kossuth Lajos u. 10.', 'Bejelentkezés szükséges.', true),
(1, 'oktatas', 'Márton Áron Főgimnázium', '0266 311 294', 'Márton Áron u. 72.', 'Elit iskola.', true),
(2, 'oktatas', 'Tamási Áron Gimnázium', '0266 218 194', 'Márton Áron tér 4.', 'Bentlakás van.', true),
(1, 'mesteremberek', 'Nagy István - Villanyszerelő', '0755 987 654', 'Kiszállás megyeszerte', 'Gyors, megbízható.', true),
(1, 'hivatalok', 'Csíkszereda Városháza', '0266 315 120', 'Vár tér 1.', 'Ügyfélfogadás hétköznap 8-14.', true),
(1, 'egyeb', 'Góbé Termékek - Helyi bolt', '0722 000 111', 'Központ', 'Hagyományos ízek.', true)
ON CONFLICT DO NOTHING;

INSERT INTO quick_links (title, url, match_category) VALUES
('RMDSZ', 'https://rmdsz.ro', ''),
('eMAG', 'https://www.emag.ro', ''),
('CFR Menetrend', 'https://www.cfrcalatori.ro', ''),
('Székelyhon', 'https://szekelyhon.ro', ''),
('Maszol', 'https://maszol.ro', ''),
('Erdélyiek FB', 'https://www.facebook.com/groups/erdelyi', ''),
('Városháza', 'https://csikszereda.ro/', 'hivatalok'),
('Meteo Romania', 'https://www.meteo.ro/', ''),
('Székelyföld Info', 'https://www.szekelyfold.info/', ''),
('Google Térkép', 'https://maps.google.com', '')
ON CONFLICT DO NOTHING;

INSERT INTO news_feeds (title, feed_url, bg_color) VALUES
('Székelyhon', 'https://szekelyhon.ro/rss/szekelyhon_hirek.xml', '#ffebd6'),
('Maszol', 'https://maszol.ro/rss', '#e6f0ff'),
('Krónika', 'https://kronika.ro/rss/kronika_hirek.xml', '#ebf5e6'),
('Erdély.ma', 'https://www.erdely.ma/rss', '#ffe6e6'),
('Bihari Napló', 'https://biharinaplo.ro/rss', '#f5e6ff')
ON CONFLICT DO NOTHING;
