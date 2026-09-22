package handlers

import (
	"backend/internal/db"
	"backend/internal/utils"
	"log"
)

// MigrateEntryTypes seeds the type catalog, backfills entries.type_id, and
// drops the legacy varchar type column (idempotent).
func MigrateEntryTypes() {
	for _, name := range []string{utils.EntryTypeService, utils.EntryTypeCeg, utils.EntryTypeEgyeb} {
		if _, err := db.DB.Exec(`INSERT INTO entry_types (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`, name); err != nil {
			log.Printf("MigrateEntryTypes (seed %s): %v", name, err)
		}
	}

	if _, err := db.DB.Exec(`ALTER TABLE entries ADD COLUMN IF NOT EXISTS type_id INTEGER`); err != nil {
		log.Printf("MigrateEntryTypes (add type_id): %v", err)
	}

	var hasTypeCol bool
	if err := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = 'entries' AND column_name = 'type'
		)`).Scan(&hasTypeCol); err != nil {
		log.Printf("MigrateEntryTypes (type column check): %v", err)
	}

	if hasTypeCol {
		stmts := []struct {
			sql  string
			args []interface{}
		}{
			{
				`UPDATE entries SET type = $1
				 WHERE lower(trim(type)) IN ('service', 'szolgáltatás')
				    OR lower(trim(type)) = lower($1)`,
				[]interface{}{utils.EntryTypeService},
			},
			{
				`UPDATE entries SET type = $1
				 WHERE lower(trim(type)) IN ('cég', 'ceg', 'company')
				    OR lower(trim(type)) = lower($1)`,
				[]interface{}{utils.EntryTypeCeg},
			},
			{
				`UPDATE entries SET type = $1
				 WHERE lower(trim(type)) IN ('entry', 'egyéb', 'egyeb', 'other')
				    OR lower(trim(type)) = lower($1)`,
				[]interface{}{utils.EntryTypeEgyeb},
			},
			{
				`UPDATE entries SET type = $1 WHERE trim(COALESCE(type, '')) = ''`,
				[]interface{}{utils.EntryTypeService},
			},
		}
		for _, s := range stmts {
			if _, err := db.DB.Exec(s.sql, s.args...); err != nil {
				log.Printf("MigrateEntryTypes (normalize varchar): %v", err)
			}
		}

		if _, err := db.DB.Exec(`
			UPDATE entries e
			SET type_id = et.id
			FROM entry_types et
			WHERE e.type_id IS NULL AND et.name = e.type`); err != nil {
			log.Printf("MigrateEntryTypes (backfill type_id): %v", err)
		}

		if _, err := db.DB.Exec(`
			UPDATE entries
			SET type_id = (SELECT id FROM entry_types WHERE name = $1 LIMIT 1)
			WHERE type_id IS NULL`, utils.EntryTypeService); err != nil {
			log.Printf("MigrateEntryTypes (default type_id): %v", err)
		}

		if _, err := db.DB.Exec(`ALTER TABLE entries DROP COLUMN type`); err != nil {
			log.Printf("MigrateEntryTypes (drop type): %v", err)
		}
	}

	if _, err := db.DB.Exec(`
		UPDATE entries
		SET type_id = (SELECT id FROM entry_types WHERE name = $1 LIMIT 1)
		WHERE type_id IS NULL`, utils.EntryTypeService); err != nil {
		log.Printf("MigrateEntryTypes (null type_id): %v", err)
	}

	if _, err := db.DB.Exec(`ALTER TABLE entries ALTER COLUMN type_id SET NOT NULL`); err != nil {
		log.Printf("MigrateEntryTypes (type_id NOT NULL): %v", err)
	}

	if _, err := db.DB.Exec(`ALTER TABLE entries DROP CONSTRAINT IF EXISTS entries_type_id_fkey`); err != nil {
		log.Printf("MigrateEntryTypes (drop fk): %v", err)
	}
	if _, err := db.DB.Exec(`
		ALTER TABLE entries ADD CONSTRAINT entries_type_id_fkey
		FOREIGN KEY (type_id) REFERENCES entry_types(id) ON DELETE RESTRICT`); err != nil {
		log.Printf("MigrateEntryTypes (add fk): %v", err)
	}

	if _, err := db.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_entries_type_id ON entries (type_id)`); err != nil {
		log.Printf("MigrateEntryTypes (index): %v", err)
	}

	log.Println("Entry types migrated to type_id FK")
}

// resolveEntryTypeID maps a client type name (including legacy aliases) to a
// catalog row. Unknown names fall back to Szolgáltatás.
func resolveEntryTypeID(typeName string) (int, string, error) {
	name := utils.CanonicalEntryType(typeName)
	if name == "" {
		name = utils.DefaultEntryType()
	}
	var id int
	err := db.DB.QueryRow(`SELECT id FROM entry_types WHERE name = $1`, name).Scan(&id)
	if err == nil {
		return id, name, nil
	}
	err = db.DB.QueryRow(`SELECT id FROM entry_types WHERE name = $1`, utils.DefaultEntryType()).Scan(&id)
	if err != nil {
		return 0, "", err
	}
	return id, utils.DefaultEntryType(), nil
}
