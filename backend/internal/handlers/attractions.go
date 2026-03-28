package handlers

import (
	"backend/internal/db"
	"backend/internal/utils"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

type Attraction struct {
	ID            int      `json:"id"`
	CountyID      int      `json:"county_id"`
	CountySlug    string   `json:"county_slug"`
	CountyName    string   `json:"county_name"`
	Name          string   `json:"name"`
	NameRo        string   `json:"name_ro"`
	NameDe        string   `json:"name_de"`
	Slug          string   `json:"slug"`
	Description   string   `json:"description"`
	Latitude      float64  `json:"latitude,omitempty"`
	Longitude     float64  `json:"longitude,omitempty"`
	FeaturedImage string   `json:"featured_image,omitempty"`
	Content       string   `json:"content,omitempty"`
	Images        []string `json:"images,omitempty"`
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
	ID         int    `json:"id"`
	Name       string `json:"name"`
	NameRo     string `json:"name_ro"`
	NameDe     string `json:"name_de"`
	Slug       string `json:"slug"`
	Content    string `json:"content"`
}

func HandleAttractions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	countySlug := strings.TrimSpace(strings.ToLower(q.Get("county_slug")))
	slug := strings.TrimSpace(strings.ToLower(q.Get("slug")))

	if slug != "" && countySlug != "" {
		// Single attraction detail
		var a Attraction
		err := db.DB.QueryRow(`
			SELECT a.id, a.county_id, c.slug, c.name, a.name, COALESCE(a.name_ro,''), COALESCE(a.name_de,''),
				a.slug, COALESCE(a.description,''), COALESCE(a.featured_image,''), COALESCE(a.content,''),
				gl.latitude, gl.longitude
			FROM attractions a
			JOIN counties c ON a.county_id = c.id
			LEFT JOIN geo_locations gl ON a.location_id = gl.id
			WHERE LOWER(a.slug) = $1 AND LOWER(c.slug) = $2
		`, slug, countySlug).Scan(&a.ID, &a.CountyID, &a.CountySlug, &a.CountyName, &a.Name, &a.NameRo, &a.NameDe,
			&a.Slug, &a.Description, &a.FeaturedImage, &a.Content, &a.Latitude, &a.Longitude)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		rows, _ := db.DB.Query("SELECT url FROM attraction_images WHERE attraction_id = $1 ORDER BY sort_order, id", a.ID)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var url string
				if rows.Scan(&url) == nil {
					a.Images = append(a.Images, url)
				}
			}
		}
		json.NewEncoder(w).Encode(a)
		return
	}

	// List attractions (optionally filtered by county)
	query := `
		SELECT a.id, a.county_id, c.slug, c.name, a.name, COALESCE(a.name_ro,''), COALESCE(a.name_de,''),
			a.slug, COALESCE(a.description,''), COALESCE(a.featured_image,''), COALESCE(a.content,''),
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
		if err := rows.Scan(&a.ID, &a.CountyID, &a.CountySlug, &a.CountyName, &a.Name, &a.NameRo, &a.NameDe,
			&a.Slug, &a.Description, &a.FeaturedImage, &a.Content, &a.Latitude, &a.Longitude); err == nil {
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

// HandleAdminCounties updates county markdown content (PUT JSON: id, content).
func HandleAdminCounties(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		ID      int    `json:"id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if body.ID == 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	res, err := db.DB.Exec("UPDATE counties SET content = $1 WHERE id = $2", body.Content, body.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// HandleAdminHistoricalSeats updates historical seat markdown content (PUT JSON: id, content).
func HandleAdminHistoricalSeats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		ID      int    `json:"id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if body.ID == 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	res, err := db.DB.Exec("UPDATE historical_seats SET content = $1 WHERE id = $2", body.Content, body.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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

// HandleAdminAttractions: GET list, POST create, PUT update, DELETE
func HandleAdminAttractions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		q := r.URL.Query()
		countySlug := strings.TrimSpace(strings.ToLower(q.Get("county_slug")))
		query := `
			SELECT a.id, a.county_id, c.slug, c.name, a.name, COALESCE(a.name_ro,''), COALESCE(a.name_de,''),
				a.slug, COALESCE(a.description,''), COALESCE(a.featured_image,''), COALESCE(a.content,''),
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
			if err := rows.Scan(&a.ID, &a.CountyID, &a.CountySlug, &a.CountyName, &a.Name, &a.NameRo, &a.NameDe,
				&a.Slug, &a.Description, &a.FeaturedImage, &a.Content, &a.Latitude, &a.Longitude); err == nil {
				rows2, _ := db.DB.Query("SELECT url FROM attraction_images WHERE attraction_id = $1 ORDER BY sort_order, id", a.ID)
				if rows2 != nil {
					for rows2.Next() {
						var url string
						if rows2.Scan(&url) == nil {
							a.Images = append(a.Images, url)
						}
					}
					rows2.Close()
				}
				list = append(list, a)
			}
		}
		if list == nil {
			list = []Attraction{}
		}
		json.NewEncoder(w).Encode(list)

	case "POST":
		var a struct {
			CountySlug    string   `json:"county_slug"`
			Name          string   `json:"name"`
			NameRo        string   `json:"name_ro"`
			NameDe        string   `json:"name_de"`
			Slug          string   `json:"slug"`
			Description   string   `json:"description"`
			Latitude      float64  `json:"latitude"`
			Longitude     float64  `json:"longitude"`
			FeaturedImage string   `json:"featured_image"`
			Content       string   `json:"content"`
			Images        []string `json:"images"`
		}
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slug := utils.Slugify(a.Name)
		if a.Slug != "" {
			slug = utils.Slugify(a.Slug)
		}
		var countyID int
		if err := db.DB.QueryRow("SELECT id FROM counties WHERE slug = $1", strings.ToLower(a.CountySlug)).Scan(&countyID); err != nil {
			http.Error(w, "County not found", http.StatusBadRequest)
			return
		}
		var glID *int
		if a.Latitude != 0 || a.Longitude != 0 {
			var id int
			if err := db.DB.QueryRow("INSERT INTO geo_locations (latitude, longitude) VALUES ($1, $2) RETURNING id", a.Latitude, a.Longitude).Scan(&id); err == nil {
				glID = &id
			}
		}
		var attID int
		err := db.DB.QueryRow(`
			INSERT INTO attractions (county_id, name, name_ro, name_de, slug, description, location_id, featured_image, content)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id
		`, countyID, a.Name, a.NameRo, a.NameDe, slug, a.Description, glID, a.FeaturedImage, a.Content).Scan(&attID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for i, url := range a.Images {
			if url != "" {
				db.DB.Exec("INSERT INTO attraction_images (attraction_id, url, sort_order) VALUES ($1, $2, $3)", attID, url, i)
			}
		}
		json.NewEncoder(w).Encode(map[string]int{"id": attID})

	case "PUT":
		var a struct {
			ID            int      `json:"id"`
			CountySlug    string   `json:"county_slug"`
			Name          string   `json:"name"`
			NameRo        string   `json:"name_ro"`
			NameDe        string   `json:"name_de"`
			Slug          string   `json:"slug"`
			Description   string   `json:"description"`
			Latitude      float64  `json:"latitude"`
			Longitude     float64  `json:"longitude"`
			FeaturedImage string   `json:"featured_image"`
			Content       string   `json:"content"`
			Images        []string `json:"images"`
		}
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slug := utils.Slugify(a.Name)
		if a.Slug != "" {
			slug = utils.Slugify(a.Slug)
		}
		var countyID int
		if err := db.DB.QueryRow("SELECT id FROM counties WHERE slug = $1", strings.ToLower(a.CountySlug)).Scan(&countyID); err != nil {
			http.Error(w, "County not found", http.StatusBadRequest)
			return
		}
		var glID *int
		if a.Latitude != 0 || a.Longitude != 0 {
			var oldID *int
			db.DB.QueryRow("SELECT location_id FROM attractions WHERE id = $1", a.ID).Scan(&oldID)
			if oldID != nil {
				db.DB.Exec("UPDATE geo_locations SET latitude=$1, longitude=$2 WHERE id=$3", a.Latitude, a.Longitude, *oldID)
				glID = oldID
			} else {
				var id int
				if err := db.DB.QueryRow("INSERT INTO geo_locations (latitude, longitude) VALUES ($1, $2) RETURNING id", a.Latitude, a.Longitude).Scan(&id); err == nil {
					glID = &id
				}
			}
		}
		_, err := db.DB.Exec(`
			UPDATE attractions SET county_id=$1, name=$2, name_ro=$3, name_de=$4, slug=$5, description=$6, location_id=$7, featured_image=$8, content=$9
			WHERE id=$10
		`, countyID, a.Name, a.NameRo, a.NameDe, slug, a.Description, glID, a.FeaturedImage, a.Content, a.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		db.DB.Exec("DELETE FROM attraction_images WHERE attraction_id = $1", a.ID)
		for i, url := range a.Images {
			if url != "" {
				db.DB.Exec("INSERT INTO attraction_images (attraction_id, url, sort_order) VALUES ($1, $2, $3)", a.ID, url, i)
			}
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}
		var locID *int
		db.DB.QueryRow("SELECT location_id FROM attractions WHERE id = $1", id).Scan(&locID)
		db.DB.Exec("DELETE FROM attraction_images WHERE attraction_id = $1", id)
		db.DB.Exec("DELETE FROM attractions WHERE id = $1", id)
		if locID != nil {
			db.DB.Exec("DELETE FROM geo_locations WHERE id = $1", *locID)
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
