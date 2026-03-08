-- backend/schema.sql

-- Base schema for core entities used by the current application.

CREATE TABLE IF NOT EXISTS locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    name_ro VARCHAR(255),
    name_de VARCHAR(255),
    county VARCHAR(255),
    type VARCHAR(50),
    CONSTRAINT unique_location UNIQUE (name, county, type)
);

CREATE TABLE IF NOT EXISTS entry_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(120) UNIQUE
);

CREATE TABLE IF NOT EXISTS entry_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS entries (
    id SERIAL PRIMARY KEY,
    location_id INTEGER REFERENCES locations(id),
    category_id INTEGER REFERENCES entry_categories(id),
    type VARCHAR(50) NOT NULL DEFAULT 'entry',
    category VARCHAR(100),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255),
    url VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    notes TEXT,
    languages VARCHAR(10)[] DEFAULT '{"HU"}',
    CONSTRAINT unique_entry UNIQUE (name, location_id)
);

CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS entry_tags (
    entry_id INTEGER REFERENCES entries(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (entry_id, tag_id)
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
