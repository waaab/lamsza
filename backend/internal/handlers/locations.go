package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"backend/internal/db"
	"backend/internal/models"
)

// HandlePublicLocations serves the place list used by public pages. The writes
// live in lamsza-admin, so this is read-only.
func HandlePublicLocations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	typeFilter := q.Get("type")
	countySlugFilter := q.Get("county_slug")

	baseQuery := "SELECT id, name, COALESCE(name_ro, ''), COALESCE(name_de, ''), COALESCE(county, ''), COALESCE(county_slug, ''), COALESCE(type, ''), COALESCE(slug, ''), COALESCE(post_code, ''), COALESCE(coordinates, ''), COALESCE(population, ''), COALESCE(area, ''), COALESCE(crest, ''), parent_id, COALESCE(is_county_seat, false) FROM locations"
	args := []interface{}{}
	argIdx := 1
	var conditions []string

	if typeFilter != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(type) = LOWER($%d)", argIdx))
		args = append(args, typeFilter)
		argIdx++
	}
	if countySlugFilter != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(county_slug) = LOWER($%d)", argIdx))
		args = append(args, countySlugFilter)
		argIdx++
	}
	if len(conditions) > 0 {
		baseQuery += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			baseQuery += " AND " + conditions[i]
		}
	}
	baseQuery += " ORDER BY LOWER(name) ASC, id ASC"

	rows, err := db.DB.Query(baseQuery, args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var res []models.Location
	for rows.Next() {
		var loc models.Location
		if err := rows.Scan(&loc.ID, &loc.Name, &loc.NameRo, &loc.NameDe, &loc.County, &loc.CountySlug, &loc.Type, &loc.Slug, &loc.PostCode, &loc.Coordinates, &loc.Population, &loc.Area, &loc.Crest, &loc.ParentID, &loc.IsCountySeat); err == nil {
			res = append(res, loc)
		}
	}
	if res == nil {
		res = []models.Location{}
	}
	json.NewEncoder(w).Encode(res)
}
