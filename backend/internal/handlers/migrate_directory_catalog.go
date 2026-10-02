package handlers

import (
	"backend/internal/db"
	"backend/internal/utils"
	"database/sql"
	"fmt"
	"log"
)

const directoryCatalogV2Key = "directory_catalog_v2"

// Tables wiped by the catalog migration or emptied via CASCADE. entry_tags and
// entry_members use composite primary keys and have no serial sequence.
var directorySerialTables = []string{
	"entries",
	"tags",
	"entry_reviews",
	"entry_suggestions",
	"entry_types",
	"entry_categories",
}

func alignDirectorySequences(exec interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}) error {
	for _, table := range directorySerialTables {
		stmt := fmt.Sprintf(`
DO $migrate$
BEGIN
  IF pg_get_serial_sequence('%s', 'id') IS NOT NULL THEN
    IF (SELECT COUNT(*) FROM %s) = 0 THEN
      PERFORM setval(pg_get_serial_sequence('%s', 'id'), 1, false);
    ELSE
      PERFORM setval(pg_get_serial_sequence('%s', 'id'), (SELECT MAX(id) FROM %s), true);
    END IF;
  END IF;
END
$migrate$`, table, table, table, table, table)
		if _, err := exec.Exec(stmt); err != nil {
			return fmt.Errorf("%s: %w", table, err)
		}
	}
	return nil
}

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
	if err != nil && err != sql.ErrNoRows {
		log.Printf("MigrateDirectoryCatalog (flag check): %v", err)
		return
	}
	alreadyDone := err == nil && flag == "1"

	if !alreadyDone {
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

	if err := alignDirectorySequences(db.DB); err != nil {
		log.Printf("MigrateDirectoryCatalog (align sequences): %v", err)
		return
	}
}
