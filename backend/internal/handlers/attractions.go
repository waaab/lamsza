package handlers

import (
	"backend/internal/db"
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type AttractionImage struct {
	URL       string `json:"url"`
	Copyright string `json:"copyright,omitempty"`
}

type Attraction struct {
	ID                     int      `json:"id"`
	CountyID               int      `json:"county_id"`
	CountySlug             string   `json:"county_slug"`
	CountyName             string   `json:"county_name"`
	Name                   string   `json:"name"`
	NameRo                 string   `json:"name_ro"`
	NameDe                 string   `json:"name_de"`
	Slug                   string   `json:"slug"`
	Description            string   `json:"description"`
	Latitude               float64  `json:"latitude,omitempty"`
	Longitude              float64  `json:"longitude,omitempty"`
	FeaturedImage          string   `json:"featured_image,omitempty"`
	FeaturedImageCopyright string   `json:"featured_image_copyright,omitempty"`
	Content                string   `json:"content,omitempty"`
	Activities             []string `json:"activities,omitempty"`
	Prohibitions           []string `json:"prohibitions,omitempty"`
	// The attraction's facts, shown as tiles; nil (absent in JSON) when not set.
	ElevationM        *float64                `json:"elevation_m,omitempty"`
	AreaKm2           *float64                `json:"area_km2,omitempty"`
	DepthM            *float64                `json:"depth_m,omitempty"`
	Images            []AttractionImage       `json:"images,omitempty"`
	SuggestionPending bool                    `json:"suggestion_pending"`
	Contributors      []AttractionContributor `json:"contributors"`
}

type HistoricalSeat struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	NameRo  string `json:"name_ro"`
	NameDe  string `json:"name_de"`
	Slug    string `json:"slug"`
	Content string `json:"content"`
}

type County struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	NameRo  string `json:"name_ro"`
	NameDe  string `json:"name_de"`
	Slug    string `json:"slug"`
	Content string `json:"content"`
}

func HandleAttractions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	countySlug := strings.TrimSpace(strings.ToLower(q.Get("county_slug")))
	slug := strings.TrimSpace(strings.ToLower(q.Get("slug")))
	if nearID := strings.TrimSpace(q.Get("near_id")); nearID != "" {
		writeNearbyAttractions(w, nearID)
		return
	}

	if slug != "" && countySlug != "" {
		// Single attraction detail
		var a Attraction
		var activitiesText, prohibitionsText string
		// No geo_locations row (an attraction without coordinates) gives NULLs.
		var lat, lon sql.NullFloat64
		err := db.DB.QueryRow(`
			SELECT a.id, a.county_id, c.slug, c.name, a.name, COALESCE(a.name_ro,''), COALESCE(a.name_de,''),
				a.slug, COALESCE(a.description,''), COALESCE(a.featured_image,''), COALESCE(a.featured_image_copyright,''),
				COALESCE(a.content,''), COALESCE(a.activities,''), COALESCE(a.prohibitions,''),
				gl.latitude, gl.longitude, a.elevation_m, a.area_km2, a.depth_m
			FROM attractions a
			JOIN counties c ON a.county_id = c.id
			LEFT JOIN geo_locations gl ON a.location_id = gl.id
			WHERE LOWER(a.slug) = $1 AND LOWER(c.slug) = $2
		`, slug, countySlug).Scan(&a.ID, &a.CountyID, &a.CountySlug, &a.CountyName, &a.Name, &a.NameRo, &a.NameDe,
			&a.Slug, &a.Description, &a.FeaturedImage, &a.FeaturedImageCopyright, &a.Content, &activitiesText, &prohibitionsText, &lat, &lon,
			&a.ElevationM, &a.AreaKm2, &a.DepthM)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		a.Latitude, a.Longitude = lat.Float64, lon.Float64
		a.Activities = splitActivities(activitiesText)
		a.Prohibitions = splitActivities(prohibitionsText)
		a.Images = loadAttractionImages(a.ID)
		attachAttractionPublicExtras(&a)
		json.NewEncoder(w).Encode(a)
		return
	}

	// List attractions (optionally filtered by county)
	query := `
		SELECT a.id, a.county_id, c.slug, c.name, a.name, COALESCE(a.name_ro,''), COALESCE(a.name_de,''),
			a.slug, COALESCE(a.description,''), COALESCE(a.featured_image,''), COALESCE(a.featured_image_copyright,''),
			COALESCE(a.content,''), COALESCE(a.activities,''), COALESCE(a.prohibitions,''),
			gl.latitude, gl.longitude
		FROM attractions a
		JOIN counties c ON a.county_id = c.id
		LEFT JOIN geo_locations gl ON a.location_id = gl.id
	`
	args := []interface{}{}
	if countySlug != "" {
		query += " WHERE LOWER(c.slug) = $1"
		args = append(args, countySlug)
	}
	query += " ORDER BY a.name ASC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var list []Attraction
	for rows.Next() {
		var a Attraction
		var activitiesText, prohibitionsText string
		var lat, lon sql.NullFloat64
		if err := rows.Scan(&a.ID, &a.CountyID, &a.CountySlug, &a.CountyName, &a.Name, &a.NameRo, &a.NameDe,
			&a.Slug, &a.Description, &a.FeaturedImage, &a.FeaturedImageCopyright, &a.Content, &activitiesText, &prohibitionsText, &lat, &lon); err == nil {
			a.Latitude, a.Longitude = lat.Float64, lon.Float64
			a.Activities = splitActivities(activitiesText)
			a.Prohibitions = splitActivities(prohibitionsText)
			list = append(list, a)
		}
	}
	if list == nil {
		list = []Attraction{}
	}
	json.NewEncoder(w).Encode(list)
}

func HandleHistoricalSeats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	slug := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("slug")))
	if slug != "" {
		var h HistoricalSeat
		err := db.DB.QueryRow(
			`SELECT id, name, COALESCE(name_ro,''), COALESCE(name_de,''), slug, COALESCE(content,'') FROM historical_seats WHERE LOWER(slug) = $1`,
			slug,
		).Scan(&h.ID, &h.Name, &h.NameRo, &h.NameDe, &h.Slug, &h.Content)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h)
		return
	}
	rows, err := db.DB.Query("SELECT id, name, COALESCE(name_ro,''), COALESCE(name_de,''), slug, COALESCE(content,'') FROM historical_seats ORDER BY name ASC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var list []HistoricalSeat
	for rows.Next() {
		var h HistoricalSeat
		if err := rows.Scan(&h.ID, &h.Name, &h.NameRo, &h.NameDe, &h.Slug, &h.Content); err == nil {
			list = append(list, h)
		}
	}
	if list == nil {
		list = []HistoricalSeat{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func HandleCounties(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rows, err := db.DB.Query("SELECT id, name, COALESCE(name_ro,''), COALESCE(name_de,''), slug, COALESCE(content,'') FROM counties ORDER BY name ASC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var list []County
	for rows.Next() {
		var c County
		if err := rows.Scan(&c.ID, &c.Name, &c.NameRo, &c.NameDe, &c.Slug, &c.Content); err == nil {
			list = append(list, c)
		}
	}
	if list == nil {
		list = []County{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func MigrateAttractions() {
	statements := []string{
		`ALTER TABLE attractions ADD COLUMN IF NOT EXISTS activities TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE attractions ADD COLUMN IF NOT EXISTS prohibitions TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE attractions ADD COLUMN IF NOT EXISTS featured_image_copyright TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE attraction_images ADD COLUMN IF NOT EXISTS copyright TEXT NOT NULL DEFAULT ''`,
		// The attraction's facts for the page's tiles (UI_BASELINE "szf-pages"); NULL = not shown.
		`ALTER TABLE attractions ADD COLUMN IF NOT EXISTS elevation_m NUMERIC(7,1)`,
		`ALTER TABLE attractions ADD COLUMN IF NOT EXISTS area_km2 NUMERIC(10,3) CHECK (area_km2 >= 0)`,
		`ALTER TABLE attractions ADD COLUMN IF NOT EXISTS depth_m NUMERIC(7,1) CHECK (depth_m >= 0)`,
		`CREATE TABLE IF NOT EXISTS attraction_suggestions (
			id SERIAL PRIMARY KEY,
			attraction_id INT NOT NULL REFERENCES attractions(id) ON DELETE CASCADE,
			user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			changes JSONB NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'accepted', 'denied')),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS attraction_suggestions_one_open
			ON attraction_suggestions (attraction_id)
			WHERE status = 'open'`,
		`CREATE TABLE IF NOT EXISTS attraction_contributors (
			attraction_id INT NOT NULL REFERENCES attractions(id) ON DELETE CASCADE,
			user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (attraction_id, user_id)
		)`,
	}
	for _, q := range statements {
		if _, err := db.DB.Exec(q); err != nil {
			log.Printf("MigrateAttractions: %v", err)
		}
	}
}

func splitActivities(raw string) []string {
	out := []string{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// Nearby rings widen until at least one other attraction is found.
// Settlements have no coordinates, so villages, towns, and cities are distance steps, then the county.
var nearbyAttractionTiers = []struct {
	km    float64
	scope string
}{
	{30, "near"},
	{60, "vicinity"},
	{100, "area"},
}

type nearbyAttraction struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description string  `json:"description,omitempty"`
	CountySlug  string  `json:"county_slug"`
	CountyName  string  `json:"county_name"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
}

func writeNearbyAttractions(w http.ResponseWriter, idRaw string) {
	id, err := strconv.Atoi(idRaw)
	if err != nil || id <= 0 {
		http.Error(w, "near_id required", http.StatusBadRequest)
		return
	}
	var originCounty int
	var originLat, originLon sql.NullFloat64
	err = db.DB.QueryRow(`
		SELECT a.county_id, gl.latitude, gl.longitude
		FROM attractions a
		LEFT JOIN geo_locations gl ON a.location_id = gl.id
		WHERE a.id = $1
	`, id).Scan(&originCounty, &originLat, &originLon)
	if err == sql.ErrNoRows {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := db.DB.Query(`
		SELECT a.id, a.county_id, c.slug, c.name, a.name, a.slug, COALESCE(a.description, ''),
			gl.latitude, gl.longitude
		FROM attractions a
		JOIN counties c ON a.county_id = c.id
		LEFT JOIN geo_locations gl ON a.location_id = gl.id
		WHERE a.id <> $1
		ORDER BY a.name ASC
	`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var all []candidateNearby
	originHasCoords := originLat.Valid && originLon.Valid
	for rows.Next() {
		var item nearbyAttraction
		var county int
		var lat, lon sql.NullFloat64
		if err := rows.Scan(&item.ID, &county, &item.CountySlug, &item.CountyName, &item.Name, &item.Slug, &item.Description, &lat, &lon); err != nil {
			continue
		}
		c := candidateNearby{item: item, county: county}
		if originHasCoords && lat.Valid && lon.Valid {
			item.Latitude = lat.Float64
			item.Longitude = lon.Float64
			c.item = item
			c.hasDist = true
			c.distKm = haversineKm(originLat.Float64, originLon.Float64, lat.Float64, lon.Float64)
		}
		all = append(all, c)
	}

	scope := "county"
	chosen := []candidateNearby{}
	if originHasCoords {
		for _, tier := range nearbyAttractionTiers {
			var matched []candidateNearby
			for _, c := range all {
				if c.hasDist && c.distKm <= tier.km {
					matched = append(matched, c)
				}
			}
			if len(matched) > 0 {
				scope = tier.scope
				chosen = matched
				break
			}
		}
	}
	if len(chosen) == 0 {
		scope = "county"
		for _, c := range all {
			if c.county == originCounty {
				chosen = append(chosen, c)
			}
		}
	}
	sortNearby(chosen)

	out := make([]nearbyAttraction, 0, len(chosen))
	for _, c := range chosen {
		if len(out) >= 12 {
			break
		}
		out = append(out, c.item)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"scope":       scope,
		"attractions": out,
	})
}

func sortNearby(items []candidateNearby) {
	for i := 1; i < len(items); i++ {
		j := i
		for j > 0 && nearbyLess(items[j], items[j-1]) {
			items[j], items[j-1] = items[j-1], items[j]
			j--
		}
	}
}

type candidateNearby struct {
	item    nearbyAttraction
	county  int
	hasDist bool
	distKm  float64
}

func nearbyLess(a, b candidateNearby) bool {
	if a.hasDist != b.hasDist {
		return a.hasDist
	}
	if a.hasDist && b.hasDist && a.distKm != b.distKm {
		return a.distKm < b.distKm
	}
	return a.item.Name < b.item.Name
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthKm = 6371.0
	r1 := lat1 * math.Pi / 180
	r2 := lat2 * math.Pi / 180
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	h := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(r1)*math.Cos(r2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return earthKm * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}

func loadAttractionImages(id int) []AttractionImage {
	out := []AttractionImage{}
	rows, err := db.DB.Query(`SELECT url, COALESCE(copyright, '') FROM attraction_images WHERE attraction_id = $1 ORDER BY sort_order, id`, id)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var img AttractionImage
		if rows.Scan(&img.URL, &img.Copyright) == nil && strings.TrimSpace(img.URL) != "" {
			out = append(out, img)
		}
	}
	return out
}
