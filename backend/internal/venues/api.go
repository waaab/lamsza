package venues

import (
	"backend/internal/db"
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

const venueSelectCols = `v.id, v.settlement_id, v.name, COALESCE(v.name_ro, ''), COALESCE(v.name_de, ''), v.slug, v.kind,
		COALESCE(v.address, ''), COALESCE(v.notes, ''),
		v.latitude, v.longitude, v.seating_capacity, COALESCE(v.description, ''),
		COALESCE(vt.label_hu, v.kind),
		s.name, s.slug, c.name, c.slug`

const venueJoins = `FROM venues v
		JOIN settlements s ON v.settlement_id = s.id
		JOIN counties c ON s.county_id = c.id
		LEFT JOIN venue_types vt ON vt.slug = v.kind`

func scanVenueRow(scanner interface {
	Scan(dest ...interface{}) error
}) (models.Venue, error) {
	var v models.Venue
	var lat, lng sql.NullFloat64
	var seat sql.NullInt64
	err := scanner.Scan(
		&v.ID, &v.SettlementID, &v.Name, &v.NameRO, &v.NameDE, &v.Slug, &v.Kind,
		&v.Address, &v.Notes,
		&lat, &lng, &seat, &v.Description,
		&v.KindLabel,
		&v.SettlementName, &v.SettlementSlug, &v.CountyName, &v.CountySlug,
	)
	if err != nil {
		return v, err
	}
	if lat.Valid {
		x := lat.Float64
		v.Latitude = &x
	}
	if lng.Valid {
		x := lng.Float64
		v.Longitude = &x
	}
	if seat.Valid {
		x := int(seat.Int64)
		v.SeatingCapacity = &x
	}
	return v, nil
}

// HandlePublic GET /api/venues?settlement_id= | ?county_slug=&settlement_slug=&venue_slug=
func HandlePublic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	if q.Get("venue_slug") != "" && q.Get("settlement_slug") != "" && q.Get("county_slug") != "" {
		getVenueBySlugs(w,
			strings.TrimSpace(q.Get("county_slug")),
			strings.TrimSpace(q.Get("settlement_slug")),
			strings.TrimSpace(q.Get("venue_slug")),
		)
		return
	}
	listVenues(w, q.Get("settlement_id"))
}

func getVenueBySlugs(w http.ResponseWriter, countySlug, settlementSlug, venueSlug string) {
	countySlug = strings.ToLower(countySlug)
	settlementSlug = strings.ToLower(settlementSlug)
	venueSlug = strings.ToLower(venueSlug)
	if countySlug == "" || settlementSlug == "" || venueSlug == "" {
		http.Error(w, "county_slug, settlement_slug és venue_slug kötelező", http.StatusBadRequest)
		return
	}
	row := db.DB.QueryRow(`
		SELECT `+venueSelectCols+` `+venueJoins+`
		WHERE lower(c.slug) = lower($1) AND lower(s.slug) = lower($2) AND lower(v.slug) = lower($3)`,
		countySlug, settlementSlug, venueSlug)
	v, err := scanVenueRow(row)
	if err == sql.ErrNoRows {
		http.Error(w, "nem található", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func listVenues(w http.ResponseWriter, settlementIDStr string) {
	var rows *sql.Rows
	var err error
	base := `SELECT ` + venueSelectCols + ` ` + venueJoins + ` `
	if strings.TrimSpace(settlementIDStr) != "" {
		sid, err := strconv.Atoi(settlementIDStr)
		if err != nil || sid < 1 {
			http.Error(w, "érvénytelen settlement_id", http.StatusBadRequest)
			return
		}
		rows, err = db.DB.Query(base+`WHERE v.settlement_id = $1 ORDER BY v.name`, sid)
	} else {
		rows, err = db.DB.Query(base + `ORDER BY c.name, s.name, v.name`)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []models.Venue{}
	for rows.Next() {
		v, err := scanVenueRow(rows)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out = append(out, v)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

type venuePayload struct {
	SettlementID    int      `json:"settlement_id"`
	Name            string   `json:"name"`
	NameRO          string   `json:"name_ro"`
	NameDE          string   `json:"name_de"`
	Slug            string   `json:"slug"`
	Kind            string   `json:"kind"`
	Address         string   `json:"address"`
	Notes           string   `json:"notes"`
	Latitude        *float64 `json:"latitude"`
	Longitude       *float64 `json:"longitude"`
	SeatingCapacity *int     `json:"seating_capacity"`
	Description     string   `json:"description"`
}
