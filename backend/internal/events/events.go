package events

import (
	"backend/internal/db"
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
)

func HandleEvents(w http.ResponseWriter, r *http.Request) {
	locationSlug := r.URL.Query().Get("location_slug")
	countySlug := r.URL.Query().Get("county_slug")

	var rows *sql.Rows
	var err error

	if locationSlug != "" {
		rows, err = db.DB.Query(`
			SELECT e.id, e.location_id, l.name, l.slug, l.county, l.county_slug, e.title, e.description, 
			       e.start_date::text, e.start_time::text, e.end_date::text, e.end_time::text, e.event_type, e.organizer
			FROM events e
			JOIN locations l ON e.location_id = l.id
			WHERE l.slug = $1
			ORDER BY e.start_date ASC`, locationSlug)
	} else if countySlug != "" {
		rows, err = db.DB.Query(`
			SELECT e.id, e.location_id, l.name, l.slug, l.county, l.county_slug, e.title, e.description, 
			       e.start_date::text, e.start_time::text, e.end_date::text, e.end_time::text, e.event_type, e.organizer
			FROM events e
			JOIN locations l ON e.location_id = l.id
			WHERE l.county_slug = $1
			ORDER BY e.start_date ASC`, countySlug)
	} else {
		rows, err = db.DB.Query(`
			SELECT e.id, e.location_id, l.name, l.slug, l.county, l.county_slug, e.title, e.description, 
			       e.start_date::text, e.start_time::text, e.end_date::text, e.end_time::text, e.event_type, e.organizer
			FROM events e
			JOIN locations l ON e.location_id = l.id
			ORDER BY e.start_date ASC`)
	}

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	events := []models.Event{}
	for rows.Next() {
		var ev models.Event
		if err := rows.Scan(&ev.ID, &ev.LocationID, &ev.LocationName, &ev.LocationSlug, &ev.County, &ev.CountySlug, &ev.Title, &ev.Description, &ev.StartDate, &ev.StartTime, &ev.EndDate, &ev.EndTime, &ev.EventType, &ev.Organizer); err == nil {
			events = append(events, ev)
		}
	}
	json.NewEncoder(w).Encode(events)
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
