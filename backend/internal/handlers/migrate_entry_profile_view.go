package handlers

import (
	"backend/internal/db"
	"log"
	"os"
)

// MigrateEntryProfileView adds hours/delivery switches and social_links (idempotent).
func MigrateEntryProfileView() {
	sqlBytes, err := os.ReadFile("migrations/entry_profile_view.sql")
	if err != nil {
		log.Printf("MigrateEntryProfileView (read sql): %v", err)
		return
	}
	if _, err := db.DB.Exec(string(sqlBytes)); err != nil {
		log.Printf("MigrateEntryProfileView: %v", err)
	}
}
