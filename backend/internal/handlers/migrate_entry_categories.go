package handlers

import (
	"backend/internal/db"
	"backend/internal/utils"
	"fmt"
	"log"
	"strings"
)

// MigrateEntryCategories ensures the short catalog exists, remaps entries.category_id,
// drops the leftover varchar category column, and deletes unused junk labels.
// Existing catalog display names are left alone so an admin rename (e.g. Mesterember)
// is not overwritten on every backend start.
func MigrateEntryCategories() {
	rows, err := db.DB.Query(`SELECT id, name, COALESCE(slug, '') FROM entry_categories`)
	if err != nil {
		log.Printf("MigrateEntryCategories (list catalog): %v", err)
	} else {
		type catRow struct {
			id         int
			name, slug string
		}
		var existing []catRow
		for rows.Next() {
			var r catRow
			if scanErr := rows.Scan(&r.id, &r.name, &r.slug); scanErr == nil {
				existing = append(existing, r)
			}
		}
		rows.Close()

		for _, name := range utils.SeedEntryCategories() {
			slug := utils.Slugify(name)
			var id int
			var existingSlug string
			found := false
			for _, r := range existing {
				if r.name == name || r.slug == slug || utils.CanonicalEntryCategory(r.name) == name {
					id = r.id
					existingSlug = r.slug
					found = true
					break
				}
			}
			if found {
				// Keep admin display names. Only backfill a missing slug.
				if strings.TrimSpace(existingSlug) == "" {
					if _, uerr := db.DB.Exec(`UPDATE entry_categories SET slug = $1 WHERE id = $2`, slug, id); uerr != nil {
						log.Printf("MigrateEntryCategories (slug %s): %v", name, uerr)
					}
				}
			} else if _, ierr := db.DB.Exec(`INSERT INTO entry_categories (name, slug) VALUES ($1, $2)`, name, slug); ierr != nil {
				log.Printf("MigrateEntryCategories (seed %s): %v", name, ierr)
			}
		}
	}

	var hasCategoryCol bool
	if err := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = 'entries' AND column_name = 'category'
		)`).Scan(&hasCategoryCol); err != nil {
		log.Printf("MigrateEntryCategories (column check): %v", err)
	}

	query := `
		SELECT e.id, e.category_id, COALESCE(c.name, '')
		FROM entries e
		LEFT JOIN entry_categories c ON c.id = e.category_id`
	if hasCategoryCol {
		query = `
			SELECT e.id, e.category_id, COALESCE(c.name, ''), COALESCE(e.category, '')
			FROM entries e
			LEFT JOIN entry_categories c ON c.id = e.category_id`
	}

	rows, err = db.DB.Query(query)
	if err != nil {
		log.Printf("MigrateEntryCategories (list): %v", err)
		return
	}
	defer rows.Close()

	type row struct {
		id  int
		raw string
	}
	var pending []row
	for rows.Next() {
		var id int
		var catID *int
		var joined string
		var varchar string
		var scanErr error
		if hasCategoryCol {
			scanErr = rows.Scan(&id, &catID, &joined, &varchar)
		} else {
			scanErr = rows.Scan(&id, &catID, &joined)
		}
		if scanErr != nil {
			log.Printf("MigrateEntryCategories (scan): %v", scanErr)
			continue
		}
		raw := joined
		if strings.TrimSpace(raw) == "" {
			raw = varchar
		}
		pending = append(pending, row{id: id, raw: raw})
	}

	for _, p := range pending {
		catID, _, resErr := resolveEntryCategoryID(nil, p.raw)
		if resErr != nil {
			log.Printf("MigrateEntryCategories (resolve %d): %v", p.id, resErr)
			continue
		}
		if _, err := db.DB.Exec(`UPDATE entries SET category_id = $1 WHERE id = $2`, catID, p.id); err != nil {
			log.Printf("MigrateEntryCategories (update %d): %v", p.id, err)
		}
	}

	if _, err := db.DB.Exec(`
		UPDATE entries e SET cat_name = c.name
		FROM entry_categories c
		WHERE e.category_id = c.id`); err != nil {
		log.Printf("MigrateEntryCategories (cat_name): %v", err)
	}

	if hasCategoryCol {
		if _, err := db.DB.Exec(`ALTER TABLE entries DROP COLUMN IF EXISTS category`); err != nil {
			log.Printf("MigrateEntryCategories (drop category): %v", err)
		}
	}

	if _, err := db.DB.Exec(`
		UPDATE entries
		SET category_id = (SELECT id FROM entry_categories WHERE name = $1 LIMIT 1)
		WHERE category_id IS NULL`, utils.DefaultEntryCategory()); err != nil {
		log.Printf("MigrateEntryCategories (null category_id): %v", err)
	}

	if _, err := db.DB.Exec(`ALTER TABLE entries ALTER COLUMN category_id SET NOT NULL`); err != nil {
		log.Printf("MigrateEntryCategories (category_id NOT NULL): %v", err)
	}

	seed := utils.SeedEntryCategories()
	placeholders := make([]string, 0, len(seed)*2)
	args := make([]interface{}, 0, len(seed)*2)
	for _, name := range seed {
		args = append(args, name)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
	}
	slugPlaceholders := make([]string, 0, len(seed))
	for _, name := range seed {
		args = append(args, utils.Slugify(name))
		slugPlaceholders = append(slugPlaceholders, fmt.Sprintf("$%d", len(args)))
	}
	del := fmt.Sprintf(`
		DELETE FROM entry_categories c
		WHERE NOT EXISTS (SELECT 1 FROM entries e WHERE e.category_id = c.id)
		  AND c.name NOT IN (%s)
		  AND COALESCE(c.slug, '') NOT IN (%s)`, strings.Join(placeholders, ", "), strings.Join(slugPlaceholders, ", "))
	if _, err := db.DB.Exec(del, args...); err != nil {
		log.Printf("MigrateEntryCategories (prune unused): %v", err)
	}

	log.Println("Entry categories migrated to short catalog FK")
}

func resolveEntryCategoryID(id *int, name string) (int, string, error) {
	if id != nil && *id > 0 {
		var n string
		err := db.DB.QueryRow(`SELECT name FROM entry_categories WHERE id = $1`, *id).Scan(&n)
		if err == nil {
			return *id, n, nil
		}
	}
	canon := utils.CanonicalEntryCategory(name)
	if canon == "" {
		canon = utils.DefaultEntryCategory()
	}
	var cid int
	err := db.DB.QueryRow(`SELECT id FROM entry_categories WHERE name = $1`, canon).Scan(&cid)
	if err == nil {
		return cid, canon, nil
	}
	trimmed := strings.TrimSpace(name)
	if trimmed != "" && trimmed != canon {
		err = db.DB.QueryRow(`SELECT id FROM entry_categories WHERE name = $1`, trimmed).Scan(&cid)
		if err == nil {
			return cid, trimmed, nil
		}
	}
	err = db.DB.QueryRow(`SELECT id FROM entry_categories WHERE name = $1`, utils.DefaultEntryCategory()).Scan(&cid)
	if err != nil {
		return 0, "", err
	}
	return cid, utils.DefaultEntryCategory(), nil
}
