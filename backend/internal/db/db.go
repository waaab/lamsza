package db

import (
	"backend/internal/config"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	connStr := config.AppConfig.DatabaseURL
	if err := guardTestConnection(connStr); err != nil {
		log.Fatal(err)
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Pool limits. database/sql defaults to unlimited open connections, so a
	// burst of /api/search requests (several connections each) can ask for
	// more than Postgres allows and every query then fails at once.
	// SetMaxOpenConns makes the extra requests wait for a connection instead.
	maxOpen := config.AppConfig.DBMaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 25
	}
	maxIdle := config.AppConfig.DBMaxIdleConns
	if maxIdle <= 0 || maxIdle > maxOpen {
		maxIdle = maxOpen
	}
	DB.SetMaxOpenConns(maxOpen)
	DB.SetMaxIdleConns(maxIdle)
	DB.SetConnMaxLifetime(30 * time.Minute)
	DB.SetConnMaxIdleTime(5 * time.Minute)

	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Database connection established (max open %d, max idle %d)", maxOpen, maxIdle)
	dropLocationsLegacy()
}

func dropLocationsLegacy() {
	var exists bool
	if err := DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'locations_legacy'
		)`).Scan(&exists); err != nil {
		log.Printf("locations_legacy check: %v", err)
		return
	}
	if !exists {
		return
	}
	if _, err := DB.Exec(`DROP TABLE locations_legacy`); err != nil {
		log.Printf("drop locations_legacy: %v", err)
		return
	}
	log.Println("Dropped leftover table locations_legacy")
}
