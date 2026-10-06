package mondasok

import (
	"backend/internal/db"
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

var budapestLoc *time.Location

func init() {
	loc, err := time.LoadLocation("Europe/Budapest")
	if err != nil {
		budapestLoc = time.UTC
		return
	}
	budapestLoc = loc
}

func todayBudapest() string {
	return time.Now().In(budapestLoc).Format("2006-01-02")
}

// Migrate ensures mondasok.display_date exists and is populated (matches migrations/mondasok_display_date.sql).
// Call once after db.InitDB() so existing deployments without a manual migration keep working.
func Migrate() {
	_, err := db.DB.Exec(`ALTER TABLE mondasok ADD COLUMN IF NOT EXISTS display_date DATE`)
	if err != nil {
		log.Printf("mondasok Migrate (add display_date): %v", err)
		return
	}
	_, _ = db.DB.Exec(`
		UPDATE mondasok
		SET display_date = COALESCE((created_at AT TIME ZONE 'UTC')::date, CURRENT_DATE)
		WHERE display_date IS NULL`)
	_, _ = db.DB.Exec(`UPDATE mondasok SET display_date = CURRENT_DATE WHERE display_date IS NULL`)
	_, err = db.DB.Exec(`ALTER TABLE mondasok ALTER COLUMN display_date SET NOT NULL`)
	if err != nil {
		log.Printf("mondasok Migrate (display_date NOT NULL): %v", err)
	}
	_, _ = db.DB.Exec(`ALTER TABLE mondasok ALTER COLUMN display_date SET DEFAULT CURRENT_DATE`)
	_, _ = db.DB.Exec(`CREATE INDEX IF NOT EXISTS idx_mondasok_display_date ON mondasok (display_date)`)
	log.Println("mondasok table ready (display_date)")
}

// normalizeYMD trims ISO date strings from <input type="date"> or JSON (YYYY-MM-DD prefix).
func normalizeYMD(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

// HandlePublicMondasok GET /api/mondasok - quotes for a calendar day.
// Optional ?date=YYYY-MM-DD uses that day (browser should pass local calendar date, same as homepage #datetime).
// If date is missing or invalid, falls back to “today” in Europe/Budapest.
func HandlePublicMondasok(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	today := todayBudapest()
	if q := strings.TrimSpace(r.URL.Query().Get("date")); q != "" {
		n := normalizeYMD(q)
		if n != "" {
			if _, err := time.Parse("2006-01-02", n); err == nil {
				today = n
			}
		}
	}
	rows, err := db.DB.Query(
		`SELECT id, text, display_date::text, created_at::text FROM mondasok WHERE display_date = $1::date ORDER BY id ASC`,
		today,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var res []models.Mondas
	for rows.Next() {
		var m models.Mondas
		var dd sql.NullString
		if err := rows.Scan(&m.ID, &m.Text, &dd, &m.CreatedAt); err != nil {
			continue
		}
		if dd.Valid {
			m.DisplayDate = dd.String
		}
		res = append(res, m)
	}
	if res == nil {
		res = []models.Mondas{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
