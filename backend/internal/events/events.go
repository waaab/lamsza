package events

import (
	"backend/internal/db"
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type paginatedResponse struct {
	Events []models.Event `json:"events"`
	Total  int            `json:"total"`
}

func HandleEvents(w http.ResponseWriter, r *http.Request) {
	locationSlug := r.URL.Query().Get("location_slug")
	countySlug := r.URL.Query().Get("county_slug")
	organizer := r.URL.Query().Get("organizer")
	locationType := r.URL.Query().Get("location_type")

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 0
	offset := 0
	if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
		limit = v
	}
	if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
		offset = v
	}

	var total int
	var rows *sql.Rows
	var err error

	baseSelect := `SELECT e.id, e.location_id, s.name, s.slug, c.name, c.slug, e.title, e.description,
		e.start_date::text, e.start_time::text, e.end_date::text, e.end_time::text, e.event_type, e.organizer, COALESCE(s.type, '')
		FROM events e
		JOIN settlements s ON e.location_id = s.id
		JOIN counties c ON s.county_id = c.id`

	baseCount := `SELECT COUNT(*) FROM events e JOIN settlements s ON e.location_id = s.id JOIN counties c ON s.county_id = c.id`

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "e.end_date >= CURRENT_DATE")

	if organizer != "" {
		conditions = append(conditions, fmt.Sprintf("e.organizer = $%d", argIdx))
		args = append(args, organizer)
		argIdx++
	} else if locationSlug != "" {
		conditions = append(conditions, fmt.Sprintf("(s.slug = $%d OR s.parent_id = (SELECT id FROM settlements WHERE slug = $%d))", argIdx, argIdx))
		args = append(args, locationSlug)
		argIdx++
	} else if countySlug != "" {
		conditions = append(conditions, fmt.Sprintf("c.slug = $%d", argIdx))
		args = append(args, countySlug)
		argIdx++
	}

	if locationType != "" {
		conditions = append(conditions, fmt.Sprintf("s.type = $%d", argIdx))
		args = append(args, locationType)
		argIdx++
	}

	where := " WHERE " + strings.Join(conditions, " AND ")

	err = db.DB.QueryRow(baseCount+where, args...).Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	query := baseSelect + where + " ORDER BY e.start_date ASC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, offset)
	}
	rows, err = db.DB.Query(query, args...)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	events := []models.Event{}
	for rows.Next() {
		var ev models.Event
		if err := rows.Scan(&ev.ID, &ev.LocationID, &ev.LocationName, &ev.LocationSlug, &ev.County, &ev.CountySlug, &ev.Title, &ev.Description, &ev.StartDate, &ev.StartTime, &ev.EndDate, &ev.EndTime, &ev.EventType, &ev.Organizer, &ev.LocationType); err == nil {
			events = append(events, ev)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paginatedResponse{Events: events, Total: total})
}

func HandleEventDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var ev models.Event
	err = db.DB.QueryRow(`
		SELECT e.id, e.location_id, l.name, l.slug, l.county, l.county_slug, e.title, e.description,
		       e.start_date::text, e.start_time::text, e.end_date::text, e.end_time::text, e.event_type, e.organizer, COALESCE(l.type, '')
		FROM events e
		JOIN locations l ON e.location_id = l.id
		WHERE e.id = $1`, id).Scan(&ev.ID, &ev.LocationID, &ev.LocationName, &ev.LocationSlug, &ev.County, &ev.CountySlug, &ev.Title, &ev.Description, &ev.StartDate, &ev.StartTime, &ev.EndDate, &ev.EndTime, &ev.EventType, &ev.Organizer, &ev.LocationType)
	if err == sql.ErrNoRows {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(ev)
}

func HandleAdminEvents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, err := db.DB.Query(`SELECT id, location_id, title, description, start_date::text, start_time::text, end_date::text, end_time::text, event_type, organizer FROM events ORDER BY start_date DESC`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		events := []models.AdminEvent{}
		for rows.Next() {
			var ev models.AdminEvent
			if err := rows.Scan(&ev.ID, &ev.LocationID, &ev.Title, &ev.Description, &ev.StartDate, &ev.StartTime, &ev.EndDate, &ev.EndTime, &ev.EventType, &ev.Organizer); err == nil {
				events = append(events, ev)
			}
		}
		json.NewEncoder(w).Encode(events)

	case "POST":
		var ev models.AdminEvent
		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err := db.DB.QueryRow(`INSERT INTO events (location_id, title, description, start_date, start_time, end_date, end_time, event_type, organizer) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
			ev.LocationID, ev.Title, ev.Description, ev.StartDate, ev.StartTime, ev.EndDate, ev.EndTime, ev.EventType, ev.Organizer).Scan(&ev.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(ev)

	case "PUT":
		var ev models.AdminEvent
		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, err := db.DB.Exec(`UPDATE events SET location_id=$1, title=$2, description=$3, start_date=$4, start_time=$5, end_date=$6, end_time=$7, event_type=$8, organizer=$9 WHERE id=$10`,
			ev.LocationID, ev.Title, ev.Description, ev.StartDate, ev.StartTime, ev.EndDate, ev.EndTime, ev.EventType, ev.Organizer, ev.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.DB.Exec("DELETE FROM events WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}
