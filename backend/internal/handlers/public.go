package handlers

import (
	"backend/internal/db"
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"backend/internal/utils"

	"github.com/lib/pq"
)

func EntriesHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	tag := r.URL.Query().Get("tag")
	locationSlug := r.URL.Query().Get("location_slug")
	countySlug := r.URL.Query().Get("county_slug")

	var rows *sql.Rows
	var err error

	params := []interface{}{}
	paramIdx := 1
	sqlQuery := ""

	normalizedQ := utils.Slugify(q)
	log.Printf("EntriesHandler: q=%q, normalizedQ=%q", q, normalizedQ)
	if q != "" && normalizedQ != "" {
		sqlQuery = `
			SELECT 
				e.id, COALESCE(e.type, ''), COALESCE(ec.name, ''), e.name, e.slug, 
				s.name, s.slug, c.name, c.slug, s.type, 
				COALESCE(s.name_ro, ''), COALESCE(s.name_de, ''),
				COALESCE(e.phone, ''), COALESCE(e.address, ''), COALESCE(e.notes, ''), 
				e.languages, COALESCE(e.url, ''),
				CASE WHEN unaccent(LOWER(e.name)) = unaccent(LOWER($1)) THEN true ELSE false END as is_direct_match,
				ts_rank_cd(e.search_vector, plainto_tsquery('simple', $2)) as rank
			FROM entries e
			JOIN settlements s ON e.location_id = s.id
			JOIN counties c ON s.county_id = c.id
			LEFT JOIN entry_categories ec ON e.category_id = ec.id
			LEFT JOIN entry_tags et ON e.id = et.entry_id
			LEFT JOIN tags t ON et.tag_id = t.id
			WHERE e.search_vector @@ plainto_tsquery('simple', $2)
		`
		params = append(params, q, normalizedQ)
		paramIdx = 3
	} else {
		sqlQuery = `
			SELECT 
				e.id, COALESCE(e.type, ''), COALESCE(ec.name, ''), e.name, e.slug, 
				s.name, s.slug, c.name, c.slug, s.type, 
				COALESCE(s.name_ro, ''), COALESCE(s.name_de, ''),
				COALESCE(e.phone, ''), COALESCE(e.address, ''), COALESCE(e.notes, ''), 
				e.languages, COALESCE(e.url, ''),
				CASE WHEN unaccent(LOWER(e.name)) = unaccent(LOWER($1)) THEN true ELSE false END as is_direct_match,
				0 as rank
			FROM entries e
			JOIN settlements s ON e.location_id = s.id
			JOIN counties c ON s.county_id = c.id
			LEFT JOIN entry_categories ec ON e.category_id = ec.id
			LEFT JOIN entry_tags et ON e.id = et.entry_id
			LEFT JOIN tags t ON et.tag_id = t.id
			WHERE 1=1
		`
		params = append(params, q)
		paramIdx = 2
	}

	if category != "" {
		sqlQuery += " AND (unaccent(ec.name) ILIKE unaccent($" + fmt.Sprintf("%d", paramIdx) + ") OR e.category ILIKE $" + fmt.Sprintf("%d", paramIdx) + ")"
		params = append(params, category)
		paramIdx++
	}
	if tag != "" {
		sqlQuery += " AND unaccent(t.name) ILIKE unaccent($" + fmt.Sprintf("%d", paramIdx) + ")"
		params = append(params, tag)
		paramIdx++
	}
	if locationSlug != "" {
		sqlQuery += " AND s.slug = $" + fmt.Sprintf("%d", paramIdx)
		params = append(params, locationSlug)
		paramIdx++
	}
	if countySlug != "" {
		sqlQuery += " AND c.slug = $" + fmt.Sprintf("%d", paramIdx)
		params = append(params, countySlug)
		paramIdx++
	}

	if q != "" && normalizedQ != "" {
		sqlQuery += " GROUP BY e.id, ec.name, s.name, s.slug, c.name, c.slug, s.type, s.name_ro, s.name_de ORDER BY is_direct_match DESC, rank DESC, e.name ASC"
	} else {
		sqlQuery += " GROUP BY e.id, ec.name, s.name, s.slug, c.name, c.slug, s.type, s.name_ro, s.name_de ORDER BY is_direct_match DESC, e.name ASC"
	}

	log.Printf("EntriesHandler query: %s", sqlQuery)
	log.Printf("EntriesHandler params: %v", params)

	rows, err = db.DB.Query(sqlQuery, params...)
	if err != nil {
		log.Printf("EntriesHandler query error: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	entries := []models.Entry{}
	for rows.Next() {
		var e models.Entry
		var pqLanguages []string
		var rank float64
		if err := rows.Scan(&e.ID, &e.Type, &e.Category, &e.Name, &e.Slug, &e.Location, &e.LocationSlug, &e.LocationCounty, &e.CountySlug, &e.LocationType, &e.LocationRo, &e.LocationDe, &e.Phone, &e.Address, &e.Notes, pq.Array(&pqLanguages), &e.URL, &e.IsDirectMatch, &rank); err != nil {
			log.Printf("EntriesHandler scan error: %v", err)
			continue
		}
		e.Languages = pqLanguages
		entries = append(entries, e)
	}
	json.NewEncoder(w).Encode(entries)
	log.Printf("EntriesHandler found %d entries", len(entries))
}

func EntryDetailHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Missing slug", 400)
		return
	}

	var e models.Entry
	var pqLanguages []string
	err := db.DB.QueryRow(`
		SELECT 
			e.id, COALESCE(e.type, ''), COALESCE(ec.name, ''), e.name, e.slug, 
			s.name, s.slug, c.name, c.slug, s.type, 
			COALESCE(s.name_ro, ''), COALESCE(s.name_de, ''),
			COALESCE(e.phone, ''), COALESCE(e.address, ''), COALESCE(e.notes, ''), 
			e.languages, COALESCE(e.url, '')
		FROM entries e
		JOIN settlements s ON e.location_id = s.id
		JOIN counties c ON s.county_id = c.id
		LEFT JOIN entry_categories ec ON e.category_id = ec.id
		WHERE e.slug = $1`, slug).Scan(&e.ID, &e.Type, &e.Category, &e.Name, &e.Slug, &e.Location, &e.LocationSlug, &e.LocationCounty, &e.CountySlug, &e.LocationType, &e.LocationRo, &e.LocationDe, &e.Phone, &e.Address, &e.Notes, pq.Array(&pqLanguages), &e.URL)

	if err != nil {
		http.Error(w, "Entry not found", 404)
		return
	}
	e.Languages = pqLanguages

	rows, _ := db.DB.Query("SELECT t.name FROM tags t JOIN entry_tags et ON t.id = et.tag_id WHERE et.entry_id = $1", e.ID)
	defer rows.Close()
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err == nil {
			e.Tags = append(e.Tags, tag)
		}
	}

	json.NewEncoder(w).Encode(e)
}
