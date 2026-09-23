package handlers

import (
	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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
				e.id, COALESCE(typ.name, ''), COALESCE(ec.name, ''), e.name, e.slug, 
				s.name, s.slug, c.name, c.slug, s.type, 
				COALESCE(s.name_ro, ''), COALESCE(s.name_de, ''),
				COALESCE(e.phone, ''), COALESCE(e.address, ''), COALESCE(e.notes, ''), 
				e.languages, COALESCE(e.url, ''),
				EXISTS (SELECT 1 FROM entry_members m WHERE m.entry_id = e.id AND m.role = 'owner' AND m.status = 'active'), COALESCE(e.verified, false), COALESCE(e.hours, '{}'::jsonb), COALESCE(e.delivery_hours, '{}'::jsonb),
				COALESCE(e.photos, '[]'::jsonb),
				CASE WHEN unaccent(LOWER(e.name)) = unaccent(LOWER($1)) THEN true ELSE false END as is_direct_match,
				ts_rank_cd(e.search_vector, plainto_tsquery('simple', $2)) as rank,
				COALESCE(array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL), ARRAY[]::text[]),
				COALESCE(e.ratings_enabled, false)
			FROM entries e
			JOIN entry_types typ ON typ.id = e.type_id
			JOIN settlements s ON e.location_id = s.id
			JOIN counties c ON s.county_id = c.id
			LEFT JOIN entry_categories ec ON e.category_id = ec.id
			LEFT JOIN entry_tags et ON e.id = et.entry_id
			LEFT JOIN tags t ON et.tag_id = t.id
			WHERE e.search_vector @@ plainto_tsquery('simple', $2)
				AND e.published = true
		`
		params = append(params, q, normalizedQ)
		paramIdx = 3
	} else {
		sqlQuery = `
			SELECT 
				e.id, COALESCE(typ.name, ''), COALESCE(ec.name, ''), e.name, e.slug, 
				s.name, s.slug, c.name, c.slug, s.type, 
				COALESCE(s.name_ro, ''), COALESCE(s.name_de, ''),
				COALESCE(e.phone, ''), COALESCE(e.address, ''), COALESCE(e.notes, ''), 
				e.languages, COALESCE(e.url, ''),
				EXISTS (SELECT 1 FROM entry_members m WHERE m.entry_id = e.id AND m.role = 'owner' AND m.status = 'active'), COALESCE(e.verified, false), COALESCE(e.hours, '{}'::jsonb), COALESCE(e.delivery_hours, '{}'::jsonb),
				COALESCE(e.photos, '[]'::jsonb),
				CASE WHEN unaccent(LOWER(e.name)) = unaccent(LOWER($1)) THEN true ELSE false END as is_direct_match,
				0 as rank,
				COALESCE(array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL), ARRAY[]::text[]),
				COALESCE(e.ratings_enabled, false)
			FROM entries e
			JOIN entry_types typ ON typ.id = e.type_id
			JOIN settlements s ON e.location_id = s.id
			JOIN counties c ON s.county_id = c.id
			LEFT JOIN entry_categories ec ON e.category_id = ec.id
			LEFT JOIN entry_tags et ON e.id = et.entry_id
			LEFT JOIN tags t ON et.tag_id = t.id
			WHERE e.published = true
		`
		params = append(params, q)
		paramIdx = 2
	}

	if category != "" {
		sqlQuery += " AND (unaccent(ec.name) ILIKE unaccent($" + fmt.Sprintf("%d", paramIdx) + ") OR pg_slugify(ec.name) = pg_slugify($" + fmt.Sprintf("%d", paramIdx) + "))"
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
		sqlQuery += " GROUP BY e.id, typ.name, ec.name, s.name, s.slug, c.name, c.slug, s.type, s.name_ro, s.name_de, e.verified, e.hours, e.delivery_hours, e.photos ORDER BY is_direct_match DESC, rank DESC, btrim(e.name) ASC, e.id ASC"
	} else {
		sqlQuery += " GROUP BY e.id, typ.name, ec.name, s.name, s.slug, c.name, c.slug, s.type, s.name_ro, s.name_de, e.verified, e.hours, e.delivery_hours, e.photos ORDER BY is_direct_match DESC, btrim(e.name) ASC, e.id ASC"
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
		var pqTags []string
		var rank float64
		var hours, delivery, photos []byte
		if err := rows.Scan(&e.ID, &e.Type, &e.Category, &e.Name, &e.Slug, &e.Location, &e.LocationSlug, &e.LocationCounty, &e.CountySlug, &e.LocationType, &e.LocationRo, &e.LocationDe, &e.Phone, &e.Address, &e.Notes, pq.Array(&pqLanguages), &e.URL, &e.Claimed, &e.Verified, &hours, &delivery, &photos, &e.IsDirectMatch, &rank, pq.Array(&pqTags), &e.RatingsEnabled); err != nil {
			log.Printf("EntriesHandler scan error: %v", err)
			continue
		}
		e.Name = strings.TrimSpace(e.Name)
		e.Languages = pqLanguages
		e.Type = utils.CanonicalEntryType(e.Type)
		e.Hours = jsonObjectOrEmpty(hours)
		e.DeliveryHours = jsonObjectOrEmpty(delivery)
		e.Photos = sanitizePhotos(photos)
		if pqTags != nil {
			e.Tags = pqTags
		} else {
			e.Tags = []string{}
		}
		entries = append(entries, e)
	}
	// encoding/json encodes nil []string as null; clients expect [] for "no tags".
	viewerUserID := 0
	if user, err := auth.UserFromRequest(r); err == nil && user != nil {
		viewerUserID = user.ID
	}
	for i := range entries {
		if entries[i].Tags == nil {
			entries[i].Tags = []string{}
		}
		ApplyPublicEntryExtrasMode(&entries[i], viewerUserID, true)
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
	var hours, delivery, photos []byte
	err := db.DB.QueryRow(`
		SELECT 
			e.id, COALESCE(typ.name, ''), COALESCE(ec.name, ''), e.name, e.slug, 
			s.name, s.slug, c.name, c.slug, s.type, 
			COALESCE(s.name_ro, ''), COALESCE(s.name_de, ''),
			COALESCE(e.phone, ''), COALESCE(e.address, ''), COALESCE(e.notes, ''), 
			e.languages, COALESCE(e.url, ''),
			EXISTS (SELECT 1 FROM entry_members m WHERE m.entry_id = e.id AND m.role = 'owner' AND m.status = 'active'), COALESCE(e.verified, false), COALESCE(e.hours, '{}'::jsonb), COALESCE(e.delivery_hours, '{}'::jsonb),
			COALESCE(e.photos, '[]'::jsonb),
			COALESCE(e.ratings_enabled, false)
		FROM entries e
		JOIN entry_types typ ON typ.id = e.type_id
		JOIN settlements s ON e.location_id = s.id
		JOIN counties c ON s.county_id = c.id
		LEFT JOIN entry_categories ec ON e.category_id = ec.id
		WHERE e.slug = $1 AND e.published = true`, slug).Scan(&e.ID, &e.Type, &e.Category, &e.Name, &e.Slug, &e.Location, &e.LocationSlug, &e.LocationCounty, &e.CountySlug, &e.LocationType, &e.LocationRo, &e.LocationDe, &e.Phone, &e.Address, &e.Notes, pq.Array(&pqLanguages), &e.URL, &e.Claimed, &e.Verified, &hours, &delivery, &photos, &e.RatingsEnabled)

	if err != nil {
		http.Error(w, "Entry not found", 404)
		return
	}
	e.Name = strings.TrimSpace(e.Name)
	e.Languages = pqLanguages
	e.Type = utils.CanonicalEntryType(e.Type)
	e.Hours = jsonObjectOrEmpty(hours)
	e.DeliveryHours = jsonObjectOrEmpty(delivery)
	e.Photos = sanitizePhotos(photos)

	rows, _ := db.DB.Query("SELECT t.name FROM tags t JOIN entry_tags et ON t.id = et.tag_id WHERE et.entry_id = $1", e.ID)
	defer rows.Close()
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err == nil {
			e.Tags = append(e.Tags, tag)
		}
	}

	// Apply public entry extras based on claimed status
	viewerUserID := 0
	if user, err := auth.UserFromRequest(r); err == nil && user != nil {
		viewerUserID = user.ID
	}
	ApplyPublicEntryExtras(&e, viewerUserID)

	json.NewEncoder(w).Encode(e)
}
