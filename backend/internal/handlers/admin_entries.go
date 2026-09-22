package handlers

import (
	"backend/internal/db"
	"backend/internal/models"
	"backend/internal/utils"
	"database/sql"
	"encoding/json"
	"fmt"
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
				e.id, COALESCE(typ.name, ''), e.location_id, e.category_id, COALESCE(cat.name, ''), e.name, COALESCE(e.slug, ''),
				COALESCE(e.url, ''), COALESCE(e.phone, ''), COALESCE(e.address, ''), COALESCE(e.notes, ''), 
				e.languages,
				COALESCE(e.verified, false), COALESCE(e.hours, '{}'::jsonb), COALESCE(e.delivery_hours, '{}'::jsonb),
				COALESCE(e.photos, '[]'::jsonb),
				COALESCE(array_agg(t.name) FILTER (WHERE t.name IS NOT NULL), ARRAY[]::VARCHAR[]) as tags
			FROM entries e
			JOIN entry_types typ ON typ.id = e.type_id
			LEFT JOIN entry_categories cat ON cat.id = e.category_id
			LEFT JOIN entry_tags et ON e.id = et.entry_id
			LEFT JOIN tags t ON et.tag_id = t.id
			GROUP BY e.id, typ.name, e.location_id, e.category_id, cat.name, e.name, e.url, e.phone, e.address, e.notes, e.languages, e.verified, e.hours, e.delivery_hours, e.photos
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
			var hours, delivery, photos []byte
			if err := rows.Scan(&s.ID, &s.Type, &locID, &s.CategoryID, &s.Category, &s.Name, &s.Slug, &s.URL, &s.Phone, &s.Address, &s.Notes, pq.Array(&pqLanguages), &s.Verified, &hours, &delivery, &photos, pq.Array(&pqTags)); err == nil {
				if locID.Valid {
					v := int(locID.Int64)
					s.LocationID = &v
				}
				s.Languages = pqLanguages
				s.Tags = pqTags
				s.Hours = jsonObjectOrEmpty(hours)
				s.DeliveryHours = jsonObjectOrEmpty(delivery)
				s.Photos = sanitizePhotos(photos)
				s.Type = utils.CanonicalEntryType(s.Type)
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
		typeID, typeName, err := resolveEntryTypeID(s.Type)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		s.Type = typeName
		catID, catName, catErr := resolveEntryCategoryID(s.CategoryID, s.Category)
		if catErr != nil {
			http.Error(w, catErr.Error(), 500)
			return
		}
		s.CategoryID = &catID
		s.Category = catName
		if len(s.Languages) == 0 {
			s.Languages = []string{"HU"}
		}

		s.Slug = utils.Slugify(s.Name)
		s.Hours = jsonObjectOrEmpty(s.Hours)
		s.DeliveryHours = jsonObjectOrEmpty(s.DeliveryHours)
		s.Photos = sanitizePhotos(s.Photos)
		err = db.DB.QueryRow("INSERT INTO entries (type_id, location_id, category_id, cat_name, name, slug, url, phone, address, notes, languages, verified, hours, delivery_hours, photos) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb, $14::jsonb, $15::jsonb) RETURNING id",
			typeID, s.LocationID, catID, catName, s.Name, s.Slug, s.URL, s.Phone, s.Address, s.Notes, pq.Array(s.Languages), s.Verified, string(s.Hours), string(s.DeliveryHours), string(s.Photos)).Scan(&s.ID)
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
		typeID, typeName, err := resolveEntryTypeID(s.Type)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		s.Type = typeName
		catID, catName, catErr := resolveEntryCategoryID(s.CategoryID, s.Category)
		if catErr != nil {
			http.Error(w, catErr.Error(), 500)
			return
		}
		s.CategoryID = &catID
		s.Category = catName

		s.Slug = utils.Slugify(s.Name)
		s.Hours = jsonObjectOrEmpty(s.Hours)
		s.DeliveryHours = jsonObjectOrEmpty(s.DeliveryHours)
		s.Photos = sanitizePhotos(s.Photos)
		_, err = db.DB.Exec(`UPDATE entries SET 
            type_id=$1, location_id=$2, category_id=$3, cat_name=$4, name=$5, slug=$6, url=$7, 
            phone=$8, address=$9, notes=$10, languages=$11, verified=$12, hours=$13::jsonb, delivery_hours=$14::jsonb, photos=$15::jsonb 
            WHERE id=$16`,
			typeID, s.LocationID, catID, catName, s.Name, s.Slug, s.URL, s.Phone, s.Address, s.Notes, pq.Array(s.Languages), s.Verified, string(s.Hours), string(s.DeliveryHours), string(s.Photos), s.ID)
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
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		var n int
		err := db.DB.QueryRow(`SELECT COUNT(*) FROM entries WHERE category_id = $1`, id).Scan(&n)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if n > 0 {
			http.Error(w, fmt.Sprintf("Nem törölhető: %d bejegyzés használja ezt a kategóriát.", n), http.StatusConflict)
			return
		}
		if _, err := db.DB.Exec("DELETE FROM entry_categories WHERE id = $1", id); err != nil {
			http.Error(w, err.Error(), 500)
			return
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
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		var n int
		err := db.DB.QueryRow(`SELECT COUNT(*) FROM entries WHERE type_id = $1`, id).Scan(&n)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if n > 0 {
			http.Error(w, fmt.Sprintf("Nem törölhető: %d bejegyzés használja ezt a típust.", n), http.StatusConflict)
			return
		}
		if _, err := db.DB.Exec("DELETE FROM entry_types WHERE id = $1", id); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
