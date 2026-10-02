package handlers

import (
	"backend/internal/db"
	"backend/internal/utils"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// MigrateEntryCategories is a legacy migrator. It no-ops once directory_catalog_v2 is set.
func MigrateEntryCategories() {
	var flag string
	err := db.DB.QueryRow(`SELECT value FROM site_settings WHERE key = $1`, directoryCatalogV2Key).Scan(&flag)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("MigrateEntryCategories (flag check): %v", err)
		return
	}
	if err == nil && flag == "1" {
		log.Println("skipped: directory_catalog_v2")
		return
	}
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
	if canon != "" {
		var cid int
		err := db.DB.QueryRow(`SELECT id FROM entry_categories WHERE name = $1`, canon).Scan(&cid)
		if err == nil {
			return cid, canon, nil
		}
	}
	trimmed := strings.TrimSpace(name)
	if trimmed != "" {
		var cid int
		err := db.DB.QueryRow(`SELECT id FROM entry_categories WHERE name = $1`, trimmed).Scan(&cid)
		if err == nil {
			return cid, trimmed, nil
		}
	}
	return 0, "", fmt.Errorf("unknown entry category: %s", strings.TrimSpace(name))
}
