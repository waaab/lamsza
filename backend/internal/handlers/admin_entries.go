package handlers

import (
	"backend/internal/db"
	"backend/internal/models"
	"backend/internal/utils"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/lib/pq"
)

func HandleAdminEntries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		sqlQuery := `
			SELECT 
				e.id, e.type, e.location_id, e.category_id, COALESCE(e.category, ''), e.name, COALESCE(e.slug, ''),
				COALESCE(e.url, ''), COALESCE(e.phone, ''), COALESCE(e.address, ''), COALESCE(e.notes, ''), 
				e.languages,
				COALESCE(array_agg(t.name) FILTER (WHERE t.name IS NOT NULL), ARRAY[]::VARCHAR[]) as tags
			FROM entries e
			LEFT JOIN entry_tags et ON e.id = et.entry_id
			LEFT JOIN tags t ON et.tag_id = t.id
			GROUP BY e.id, e.type, e.location_id, e.category_id, e.category, e.name, e.url, e.phone, e.address, e.notes, e.languages
			ORDER BY LOWER(e.name) ASC, e.id ASC
		`
		rows, err := db.DB.Query(sqlQuery)
		if err != nil {
			log.Println("handleAdminEntries GET error:", err)
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []models.AdminEntry
		for rows.Next() {
			var s models.AdminEntry
			var pqLanguages []string
			var pqTags []string
			var locID sql.NullInt64
			if err := rows.Scan(&s.ID, &s.Type, &locID, &s.CategoryID, &s.Category, &s.Name, &s.Slug, &s.URL, &s.Phone, &s.Address, &s.Notes, pq.Array(&pqLanguages), pq.Array(&pqTags)); err == nil {
				if locID.Valid {
					v := int(locID.Int64)
					s.LocationID = &v
				}
				s.Languages = pqLanguages
				s.Tags = pqTags
				res = append(res, s)
			} else {
				log.Println("handleAdminEntries rows.Scan error:", err)
			}
		}
		if res == nil {
			res = []models.AdminEntry{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var s models.AdminEntry
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if s.Type == "" {
			s.Type = "service"
		}
		if len(s.Languages) == 0 {
			s.Languages = []string{"HU"}
		}

		var catID interface{} = s.CategoryID
		if s.CategoryID == nil {
			catID = nil
		}

		s.Slug = utils.Slugify(s.Name)
		err := db.DB.QueryRow("INSERT INTO entries (type, location_id, category_id, category, name, slug, url, phone, address, notes, languages) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id",
			s.Type, s.LocationID, catID, s.Category, s.Name, s.Slug, s.URL, s.Phone, s.Address, s.Notes, pq.Array(s.Languages)).Scan(&s.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		for _, tagName := range s.Tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}
			var tagID int
			err = db.DB.QueryRow("INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id", tagName).Scan(&tagID)
			if err == nil {
				db.DB.Exec("INSERT INTO entry_tags (entry_id, tag_id) VALUES ($1, $2)", s.ID, tagID)
			}
		}

		json.NewEncoder(w).Encode(s)

	case "PUT":
		var s models.AdminEntry
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if s.Type == "" {
			s.Type = "service"
		}

		var catID interface{} = s.CategoryID
		if s.CategoryID == nil {
			catID = nil
		}

		s.Slug = utils.Slugify(s.Name)
		_, err := db.DB.Exec(`UPDATE entries SET 
            type=$1, location_id=$2, category_id=$3, category=$4, name=$5, slug=$6, url=$7, 
            phone=$8, address=$9, notes=$10, languages=$11 
            WHERE id=$12`,
			s.Type, s.LocationID, catID, s.Category, s.Name, s.Slug, s.URL, s.Phone, s.Address, s.Notes, pq.Array(s.Languages), s.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		db.DB.Exec("DELETE FROM entry_tags WHERE entry_id = $1", s.ID)
		for _, tagName := range s.Tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}
			var tagID int
			err = db.DB.QueryRow("INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id", tagName).Scan(&tagID)
			if err == nil {
				db.DB.Exec("INSERT INTO entry_tags (entry_id, tag_id) VALUES ($1, $2)", s.ID, tagID)
			}
		}

		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.DB.Exec("DELETE FROM entries WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func HandleAdminEntryCategories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, err := db.DB.Query("SELECT id, name FROM entry_categories ORDER BY name ASC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []models.EntryCategory
		for rows.Next() {
			var sc models.EntryCategory
			if err := rows.Scan(&sc.ID, &sc.Name); err == nil {
				res = append(res, sc)
			}
		}
		if res == nil {
			res = []models.EntryCategory{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var sc models.EntryCategory
		if err := json.NewDecoder(r.Body).Decode(&sc); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		slug := utils.Slugify(sc.Name)
		err := db.DB.QueryRow("INSERT INTO entry_categories (name, slug) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET slug=EXCLUDED.slug RETURNING id", sc.Name, slug).Scan(&sc.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(sc)

	case "PUT":
		var sc models.EntryCategory
		if err := json.NewDecoder(r.Body).Decode(&sc); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		slug := utils.Slugify(sc.Name)
		_, err := db.DB.Exec("UPDATE entry_categories SET name=$1, slug=$2 WHERE id=$3", sc.Name, slug, sc.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.DB.Exec("DELETE FROM entry_categories WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func HandleAdminEntryTypes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, err := db.DB.Query("SELECT id, name FROM entry_types ORDER BY name ASC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []models.EntryType
		for rows.Next() {
			var et models.EntryType
			if err := rows.Scan(&et.ID, &et.Name); err == nil {
				res = append(res, et)
			}
		}
		if res == nil {
			res = []models.EntryType{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var et models.EntryType
		if err := json.NewDecoder(r.Body).Decode(&et); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err := db.DB.QueryRow("INSERT INTO entry_types (name) VALUES ($1) RETURNING id", et.Name).Scan(&et.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(et)

	case "PUT":
		var et models.EntryType
		if err := json.NewDecoder(r.Body).Decode(&et); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, err := db.DB.Exec("UPDATE entry_types SET name=$1 WHERE id=$2", et.Name, et.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.DB.Exec("DELETE FROM entry_types WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}
