package events

import (
	"backend/internal/db"
	"encoding/json"
	"net/http"
)

// filterOptionsResponse drives the /esemenyek sidebar (types + locations).
type filterOptionsResponse struct {
	EventTypes []string          `json:"event_types"`
	Locations  []locationOption  `json:"locations"`
}

type locationOption struct {
	Name         string `json:"name"`
	LocationSlug string `json:"location_slug"`
	CountySlug   string `json:"county_slug"`
	CountyName   string `json:"county_name"`
}

// HandleEventFilterOptions GET — distinct types and locations among upcoming events.
func HandleEventFilterOptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var out filterOptionsResponse

	rows, err := db.DB.Query(`
		SELECT DISTINCT e.event_type
		FROM events e
		JOIN settlements s ON e.location_id = s.id
		JOIN counties c ON s.county_id = c.id
		WHERE e.end_date >= CURRENT_DATE
		ORDER BY e.event_type`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			rows.Close()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if t != "" {
			out.EventTypes = append(out.EventTypes, t)
		}
	}
	rows.Close()

	rows2, err := db.DB.Query(`
		SELECT DISTINCT s.name, s.slug, c.slug, c.name
		FROM events e
		JOIN settlements s ON e.location_id = s.id
		JOIN counties c ON s.county_id = c.id
		WHERE e.end_date >= CURRENT_DATE
		ORDER BY c.name, s.name`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		var loc locationOption
		if err := rows2.Scan(&loc.Name, &loc.LocationSlug, &loc.CountySlug, &loc.CountyName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out.Locations = append(out.Locations, loc)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
