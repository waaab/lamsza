package handlers

import (
	"backend/internal/db"
	"backend/internal/models"
	"encoding/json"
	"log"
	"net/http"
)

// MigrateSettlementLocationTypes creates catalog + seeds default settlement types (idempotent).
func MigrateSettlementLocationTypes() {
	_, err := db.DB.Exec(`
CREATE TABLE IF NOT EXISTS settlement_location_types (
	id SERIAL PRIMARY KEY,
	slug VARCHAR(64) NOT NULL UNIQUE,
	label_hu TEXT NOT NULL,
	sort_order INT NOT NULL DEFAULT 0
)`)
	if err != nil {
		log.Printf("MigrateSettlementLocationTypes (create): %v", err)
		return
	}
	_, err = db.DB.Exec(`
INSERT INTO settlement_location_types (slug, label_hu, sort_order) VALUES
	('municípium', 'Municípium', 0),
	('város', 'Város', 1),
	('község', 'Község', 2),
	('falu', 'Falu', 3),
	('megye', 'Megye', 4)
ON CONFLICT (slug) DO NOTHING`)
	if err != nil {
		log.Printf("MigrateSettlementLocationTypes (seed): %v", err)
	}
}

func fetchAllSettlementLocationTypes() ([]models.SettlementLocationType, error) {
	rows, err := db.DB.Query(`
		SELECT id, slug, label_hu, sort_order FROM settlement_location_types
		ORDER BY sort_order ASC, LOWER(label_hu) ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.SettlementLocationType
	for rows.Next() {
		var t models.SettlementLocationType
		if err := rows.Scan(&t.ID, &t.Slug, &t.LabelHu, &t.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if out == nil {
		out = []models.SettlementLocationType{}
	}
	return out, nil
}

// HandlePublicSettlementLocationTypes GET /api/settlement_location_types
func HandlePublicSettlementLocationTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	list, err := fetchAllSettlementLocationTypes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
