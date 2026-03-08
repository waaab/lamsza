package handlers

import (
	"backend/internal/db"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"
	"log"
	"net/http"
)

func HandleSetCountySeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		LocationID int `json:"location_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.LocationID == 0 {
		http.Error(w, "Invalid request: location_id required", http.StatusBadRequest)
		return
	}

	var countySlug string
	err := db.DB.QueryRow("SELECT county_slug FROM locations WHERE id = $1", body.LocationID).Scan(&countySlug)
	if err != nil {
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := tx.Exec("UPDATE locations SET is_county_seat = false WHERE county_slug = $1", countySlug); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := tx.Exec("UPDATE locations SET is_county_seat = true WHERE id = $1", body.LocationID); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("County seat set: location_id=%d, county_slug=%s", body.LocationID, countySlug)
	w.WriteHeader(http.StatusOK)
}

func HandleAdminLocations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, err := db.DB.Query("SELECT id, name, COALESCE(name_ro, ''), COALESCE(name_de, ''), COALESCE(county, ''), COALESCE(county_slug, ''), COALESCE(type, ''), COALESCE(slug, ''), COALESCE(post_code, ''), COALESCE(coordinates, ''), COALESCE(population, ''), COALESCE(area, ''), COALESCE(crest, ''), parent_id, COALESCE(is_county_seat, false) FROM locations ORDER BY name ASC")
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

	case "POST":
		var l models.Location
		if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		l.Slug = utils.Slugify(l.Name)
		l.CountySlug = utils.Slugify(l.County)
		err := db.DB.QueryRow("INSERT INTO locations (name, name_ro, name_de, county, county_slug, type, slug, post_code, coordinates, population, area, crest, parent_id, is_county_seat) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id",
			l.Name, l.NameRo, l.NameDe, l.County, l.CountySlug, l.Type, l.Slug, l.PostCode, l.Coordinates, l.Population, l.Area, l.Crest, l.ParentID, l.IsCountySeat).Scan(&l.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(l)

	case "PUT":
		var l models.Location
		if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		l.Slug = utils.Slugify(l.Name)
		l.CountySlug = utils.Slugify(l.County)
		_, err := db.DB.Exec("UPDATE locations SET name=$1, name_ro=$2, name_de=$3, county=$4, county_slug=$5, type=$6, slug=$7, post_code=$8, coordinates=$9, population=$10, area=$11, crest=$12, parent_id=$13, is_county_seat=$14 WHERE id=$15",
			l.Name, l.NameRo, l.NameDe, l.County, l.CountySlug, l.Type, l.Slug, l.PostCode, l.Coordinates, l.Population, l.Area, l.Crest, l.ParentID, l.IsCountySeat, l.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.DB.Exec("DELETE FROM locations WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}
