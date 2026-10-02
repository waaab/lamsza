package handlers

import (
	"backend/internal/db"
	"backend/internal/utils"
	"database/sql"
	"log"
)

const directoryCatalogV2Key = "directory_catalog_v2"

// MigrateDirectoryCatalog wipes directory rows once and seeds the two-level
// category tree and three entry types. Guarded by site_settings.directory_catalog_v2.
func MigrateDirectoryCatalog() {
	schemaStmts := []string{
		`ALTER TABLE entry_categories ADD COLUMN IF NOT EXISTS parent_id INTEGER REFERENCES entry_categories(id)`,
		`ALTER TABLE entry_categories ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE websites ADD COLUMN IF NOT EXISTS category_id INTEGER REFERENCES entry_categories(id)`,
	}
	for _, stmt := range schemaStmts {
		if _, err := db.DB.Exec(stmt); err != nil {
			log.Printf("MigrateDirectoryCatalog (schema): %v", err)
			return
		}
	}

	var flag string
	err := db.DB.QueryRow(`SELECT value FROM site_settings WHERE key = $1`, directoryCatalogV2Key).Scan(&flag)
	if err == nil && flag == "1" {
		return
	}
	if err != nil && err != sql.ErrNoRows {
		log.Printf("MigrateDirectoryCatalog (flag check): %v", err)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("MigrateDirectoryCatalog (begin): %v", err)
		return
	}
	defer tx.Rollback()

	wipeStmts := []string{
		`DELETE FROM entries`,
		`DELETE FROM tags`,
		`DELETE FROM entry_categories`,
		`DELETE FROM entry_types`,
	}
	for _, stmt := range wipeStmts {
		if _, err := tx.Exec(stmt); err != nil {
			log.Printf("MigrateDirectoryCatalog (wipe): %v", err)
			return
		}
	}

	for _, t := range utils.DirectoryTypes() {
		if _, err := tx.Exec(
			`INSERT INTO entry_types (id, name) VALUES ($1, $2)`,
			t.ID, t.Name,
		); err != nil {
			log.Printf("MigrateDirectoryCatalog (type %s): %v", t.Name, err)
			return
		}
	}

	insertCat := func(node utils.DirectoryNode) error {
		slug := utils.Slugify(node.Name)
		_, err := tx.Exec(
			`INSERT INTO entry_categories (id, name, slug, parent_id, sort_order) VALUES ($1, $2, $3, $4, $5)`,
			node.ID, node.Name, slug, node.ParentID, node.SortOrder,
		)
		return err
	}

	for _, node := range utils.DirectoryParents() {
		if err := insertCat(node); err != nil {
			log.Printf("MigrateDirectoryCatalog (parent %s): %v", node.Name, err)
			return
		}
	}
	for _, node := range utils.DirectoryChildren() {
		if err := insertCat(node); err != nil {
			log.Printf("MigrateDirectoryCatalog (child %s): %v", node.Name, err)
			return
		}
	}

	seqStmts := []string{
		`SELECT setval(pg_get_serial_sequence('entry_types', 'id'), (SELECT COALESCE(MAX(id), 1) FROM entry_types), true)`,
		`SELECT setval(pg_get_serial_sequence('entry_categories', 'id'), (SELECT COALESCE(MAX(id), 1) FROM entry_categories), true)`,
	}
	for _, stmt := range seqStmts {
		if _, err := tx.Exec(stmt); err != nil {
			log.Printf("MigrateDirectoryCatalog (setval): %v", err)
			return
		}
	}

	if _, err := tx.Exec(
		`INSERT INTO site_settings (key, value) VALUES ($1, '1')
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
		directoryCatalogV2Key,
	); err != nil {
		log.Printf("MigrateDirectoryCatalog (flag): %v", err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("MigrateDirectoryCatalog (commit): %v", err)
		return
	}

	log.Println("Directory catalog v2 seeded")
}
