package events

import (
	"backend/internal/db"
	"log"
)

// Migrate ensures catalog tables, event columns, and backfill from legacy event_type (idempotent).
func Migrate() {
	_, err := db.DB.Exec(`
		ALTER TABLE events ADD COLUMN IF NOT EXISTS featured_image VARCHAR(1024) DEFAULT ''`)
	if err != nil {
		log.Printf("events.Migrate (featured_image): %v", err)
	}
	_, err = db.DB.Exec(`
		ALTER TABLE events ADD COLUMN IF NOT EXISTS entry_price VARCHAR(128) DEFAULT ''`)
	if err != nil {
		log.Printf("events.Migrate (entry_price): %v", err)
	}

	_, err = db.DB.Exec(`
CREATE TABLE IF NOT EXISTS catalog_event_types (
	id SERIAL PRIMARY KEY,
	slug VARCHAR(64) NOT NULL UNIQUE,
	label_hu TEXT NOT NULL,
	sort_order INT NOT NULL DEFAULT 0
)`)
	if err != nil {
		log.Printf("events.Migrate (catalog_event_types): %v", err)
	}

	_, err = db.DB.Exec(`
CREATE TABLE IF NOT EXISTS catalog_event_subtypes (
	id SERIAL PRIMARY KEY,
	event_type_id INT NOT NULL REFERENCES catalog_event_types(id) ON DELETE CASCADE,
	slug VARCHAR(64) NOT NULL,
	label_hu TEXT NOT NULL,
	sort_order INT NOT NULL DEFAULT 0,
	UNIQUE(event_type_id, slug)
)`)
	if err != nil {
		log.Printf("events.Migrate (catalog_event_subtypes): %v", err)
	}

	_, err = db.DB.Exec(`
INSERT INTO catalog_event_types (slug, label_hu, sort_order)
SELECT v.slug, v.label_hu, v.sort_order FROM (VALUES
	('cultural', 'Kulturális', 1),
	('sports', 'Sport', 2),
	('festival', 'Fesztivál', 3),
	('religious', 'Vallási', 4),
	('other', 'Egyéb', 5)
) AS v(slug, label_hu, sort_order)
-- Only the missing rows: ON CONFLICT alone takes a sequence value per boot.
WHERE NOT EXISTS (SELECT 1 FROM catalog_event_types t WHERE t.slug = v.slug)
ON CONFLICT (slug) DO NOTHING`)
	if err != nil {
		log.Printf("events.Migrate (seed types): %v", err)
	}

	// Sub-types (examples per type), inserted only when missing: ON CONFLICT
	// alone would take a sequence value on every boot for each existing row.
	seedSub := []struct {
		typeSlug, slug, label string
		order                 int
	}{
		{"sports", "hockey", "Jégkorong", 1},
		{"sports", "football", "Futball", 2},
		{"sports", "golf", "Golf", 3},
		{"sports", "tennis", "Tenisz", 4},
		{"sports", "handball", "Kézilabda", 5},
		{"cultural", "concert", "Koncert", 1},
		{"cultural", "theatre", "Színház", 2},
		{"cultural", "exhibition", "Kiállítás", 3},
		{"cultural", "cinema", "Mozi", 4},
		{"festival", "music", "Zene", 1},
		{"festival", "folk", "Népi", 2},
		{"festival", "wine", "Bor", 3},
		{"religious", "mass", "Mise", 1},
		{"religious", "pilgrimage", "Zarándoklat", 2},
		{"other", "community", "Közösségi", 1},
		{"other", "charity", "Jótékonysági", 2},
	}
	for _, st := range seedSub {
		if _, e := db.DB.Exec(`
			INSERT INTO catalog_event_subtypes (event_type_id, slug, label_hu, sort_order)
			SELECT t.id, $2, $3, $4::int FROM catalog_event_types t
			WHERE t.slug = $1
			  AND NOT EXISTS (SELECT 1 FROM catalog_event_subtypes s WHERE s.event_type_id = t.id AND s.slug = $2)
			ON CONFLICT (event_type_id, slug) DO NOTHING`,
			st.typeSlug, st.slug, st.label, st.order); e != nil {
			log.Printf("events.Migrate (seed subtype): %v", e)
		}
	}

	_, err = db.DB.Exec(`
ALTER TABLE events ADD COLUMN IF NOT EXISTS access_type VARCHAR(32) DEFAULT 'public'`)
	if err != nil {
		log.Printf("events.Migrate (access_type): %v", err)
	}
	_, err = db.DB.Exec(`
ALTER TABLE events ADD COLUMN IF NOT EXISTS event_type_id INTEGER`)
	if err != nil {
		log.Printf("events.Migrate (event_type_id add): %v", err)
	}
	_, err = db.DB.Exec(`
ALTER TABLE events ADD COLUMN IF NOT EXISTS event_subtype_id INTEGER`)
	if err != nil {
		log.Printf("events.Migrate (event_subtype_id add): %v", err)
	}

	_, err = db.DB.Exec(`
DO $$
BEGIN
	IF EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'events' AND column_name = 'event_type'
	) THEN
		UPDATE events e
		SET event_type_id = ct.id
		FROM catalog_event_types ct
		WHERE e.event_type = ct.slug AND e.event_type_id IS NULL;
		UPDATE events
		SET event_type_id = (SELECT id FROM catalog_event_types WHERE slug = 'other' LIMIT 1)
		WHERE event_type_id IS NULL;
		ALTER TABLE events DROP COLUMN event_type;
	END IF;
END $$`)
	if err != nil {
		log.Printf("events.Migrate (backfill/drop legacy event_type): %v", err)
	}

	_, err = db.DB.Exec(`
UPDATE events SET event_type_id = (SELECT id FROM catalog_event_types WHERE slug = 'other' LIMIT 1) WHERE event_type_id IS NULL`)
	if err != nil {
		log.Printf("events.Migrate (null event_type_id): %v", err)
	}

	_, err = db.DB.Exec(`ALTER TABLE events ALTER COLUMN event_type_id SET NOT NULL`)
	if err != nil {
		log.Printf("events.Migrate (event_type_id NOT NULL): %v", err)
	}

	_, err = db.DB.Exec(`
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_event_type_id_fkey`)
	if err != nil {
		log.Printf("events.Migrate (drop fk event_type_id): %v", err)
	}
	_, err = db.DB.Exec(`
ALTER TABLE events ADD CONSTRAINT events_event_type_id_fkey
	FOREIGN KEY (event_type_id) REFERENCES catalog_event_types(id) ON DELETE RESTRICT`)
	if err != nil {
		log.Printf("events.Migrate (fk event_type_id): %v", err)
	}

	_, err = db.DB.Exec(`
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_event_subtype_id_fkey`)
	if err != nil {
		log.Printf("events.Migrate (drop fk event_subtype_id): %v", err)
	}
	_, err = db.DB.Exec(`
ALTER TABLE events ADD CONSTRAINT events_event_subtype_id_fkey
	FOREIGN KEY (event_subtype_id) REFERENCES catalog_event_subtypes(id) ON DELETE SET NULL`)
	if err != nil {
		log.Printf("events.Migrate (fk event_subtype_id): %v", err)
	}

	_, err = db.DB.Exec(`
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_access_type_check`)
	if err != nil {
		log.Printf("events.Migrate (drop access check): %v", err)
	}
	_, err = db.DB.Exec(`
ALTER TABLE events ADD CONSTRAINT events_access_type_check
	CHECK (access_type IN ('public', 'members_only', 'invitation_only'))`)
	if err != nil {
		log.Printf("events.Migrate (access_type check): %v", err)
	}

	_, err = db.DB.Exec(`UPDATE events SET access_type = 'public' WHERE access_type IS NULL OR access_type = ''`)
	if err != nil {
		log.Printf("events.Migrate (access_type default): %v", err)
	}

	_, err = db.DB.Exec(`ALTER TABLE events ADD COLUMN IF NOT EXISTS attraction_id INTEGER REFERENCES attractions(id) ON DELETE SET NULL`)
	if err != nil {
		log.Printf("events.Migrate (attraction_id): %v", err)
	}
	_, err = db.DB.Exec(`ALTER TABLE events ADD COLUMN IF NOT EXISTS featured_image_copyright TEXT NOT NULL DEFAULT ''`)
	if err != nil {
		log.Printf("events.Migrate (featured_image_copyright): %v", err)
	}
}
