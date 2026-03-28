package db

import (
	_ "embed"
	"log"
)

//go:embed historical_seats_seed.sql
var historicalSeatsSeedSQL string

// SeedHistoricalSeatsContent upserts the five default székek with Markdown bodies.
// Skips quietly if historical_seats does not exist (run migration_settlements_attractions.sql first).
func SeedHistoricalSeatsContent() {
	if DB == nil {
		return
	}
	_, err := DB.Exec(historicalSeatsSeedSQL)
	if err != nil {
		log.Printf("db: historical seats seed skipped: %v", err)
		return
	}
	log.Println("db: historical seats default content ensured")
}
