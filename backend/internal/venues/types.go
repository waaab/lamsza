package venues

import (
	"backend/internal/db"
	"backend/internal/models"
	"encoding/json"
	"net/http"
)

// HandlePublicVenueTypes GET /api/venue_types - ordered list for public UIs
func HandlePublicVenueTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	list, err := fetchAllVenueTypes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func fetchAllVenueTypes() ([]models.VenueType, error) {
	rows, err := db.DB.Query(`SELECT id, slug, label_hu FROM venue_types ORDER BY LOWER(label_hu) ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.VenueType
	for rows.Next() {
		var t models.VenueType
		if err := rows.Scan(&t.ID, &t.Slug, &t.LabelHu); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if out == nil {
		out = []models.VenueType{}
	}
	return out, nil
}
