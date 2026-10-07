package events

import (
	"backend/internal/clock"
	"backend/internal/db"
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type paginatedResponse struct {
	Events []models.Event `json:"events"`
	Total  int            `json:"total"`
}

func parseDateQueryParam(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func HandleEvents(w http.ResponseWriter, r *http.Request) {
	locationSlug := r.URL.Query().Get("location_slug")
	countySlug := r.URL.Query().Get("county_slug")
	organizer := r.URL.Query().Get("organizer")
	locationType := r.URL.Query().Get("location_type")
	eventType := strings.TrimSpace(r.URL.Query().Get("event_type"))
	eventSubtype := strings.TrimSpace(r.URL.Query().Get("event_subtype"))
	dateFrom := parseDateQueryParam(r.URL.Query().Get("date_from"))
	dateTo := parseDateQueryParam(r.URL.Query().Get("date_to"))

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

	baseSelect := `SELECT e.id, e.location_id, s.name, s.slug, c.name, c.slug, e.title, COALESCE(e.description, ''), COALESCE(e.featured_image, ''),
		e.start_date::text, COALESCE(e.start_time::text, ''), e.end_date::text, COALESCE(e.end_time::text, ''),
		COALESCE(et.slug, ''), COALESCE(et.label_hu, ''), COALESCE(es.slug, ''), COALESCE(es.label_hu, ''),
		COALESCE(e.access_type, 'public'), COALESCE(e.organizer, ''), COALESCE(e.entry_price, ''), COALESCE(s.type, ''),
		e.default_venue_id, COALESCE(vdef.name, ''), COALESCE(vdef.slug, ''), COALESCE(vdef.kind, ''), COALESCE(vt.label_hu, ''),
		EXISTS (SELECT 1 FROM event_schedule_days esd WHERE esd.event_id = e.id)
		FROM events e
		JOIN settlements s ON e.location_id = s.id
		JOIN counties c ON s.county_id = c.id
		JOIN catalog_event_types et ON e.event_type_id = et.id
		LEFT JOIN catalog_event_subtypes es ON e.event_subtype_id = es.id
		LEFT JOIN venues vdef ON e.default_venue_id = vdef.id
		LEFT JOIN venue_types vt ON vt.slug = vdef.kind
		LEFT JOIN geo_locations gls ON s.location_id = gls.id`

	baseCount := `SELECT COUNT(*) FROM events e JOIN settlements s ON e.location_id = s.id JOIN counties c ON s.county_id = c.id JOIN catalog_event_types et ON e.event_type_id = et.id LEFT JOIN catalog_event_subtypes es ON e.event_subtype_id = es.id LEFT JOIN geo_locations gls ON s.location_id = gls.id`

	var conditions []string
	var args []interface{}
	argIdx := 1

	// Hide events that ended before Bucharest's today (R19), not the
	// database session's (UTC) date.
	conditions = append(conditions, fmt.Sprintf("e.end_date >= $%d::date", argIdx))
	args = append(args, clock.Today())
	argIdx++

	if organizer != "" {
		conditions = append(conditions, fmt.Sprintf("e.organizer = $%d", argIdx))
		args = append(args, organizer)
		argIdx++
	}
	if locationSlug != "" {
		conditions = append(conditions, fmt.Sprintf("(s.slug = $%d OR s.parent_id = (SELECT id FROM settlements WHERE slug = $%d))", argIdx, argIdx))
		args = append(args, locationSlug)
		argIdx++
	}
	if countySlug != "" {
		conditions = append(conditions, fmt.Sprintf("c.slug = $%d", argIdx))
		args = append(args, countySlug)
		argIdx++
	}
	if attractionID := strings.TrimSpace(r.URL.Query().Get("attraction_id")); attractionID != "" {
		if aid, errConv := strconv.Atoi(attractionID); errConv == nil && aid > 0 {
			conditions = append(conditions, fmt.Sprintf("e.attraction_id = $%d", argIdx))
			args = append(args, aid)
			argIdx++
		}
	}
	if excludeAttraction := strings.TrimSpace(r.URL.Query().Get("exclude_attraction_id")); excludeAttraction != "" {
		if aid, errConv := strconv.Atoi(excludeAttraction); errConv == nil && aid > 0 {
			conditions = append(conditions, fmt.Sprintf("(e.attraction_id IS NULL OR e.attraction_id <> $%d)", argIdx))
			args = append(args, aid)
			argIdx++
		}
	}
	if nearLat, errLat := strconv.ParseFloat(strings.TrimSpace(r.URL.Query().Get("near_lat")), 64); errLat == nil {
		if nearLon, errLon := strconv.ParseFloat(strings.TrimSpace(r.URL.Query().Get("near_lon")), 64); errLon == nil {
			km := 30.0
			if parsed, errKm := strconv.ParseFloat(strings.TrimSpace(r.URL.Query().Get("near_km")), 64); errKm == nil && parsed > 0 && parsed <= 200 {
				km = parsed
			}
			aid := 0
			if parsed, errAid := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("exclude_attraction_id"))); errAid == nil && parsed > 0 {
				aid = parsed
			}
			conditions = append(conditions, fmt.Sprintf(`(
				(gls.latitude IS NOT NULL AND gls.longitude IS NOT NULL AND (
					6371 * acos(LEAST(1.0, GREATEST(-1.0,
						cos(radians($%d)) * cos(radians(gls.latitude)) * cos(radians(gls.longitude) - radians($%d))
						+ sin(radians($%d)) * sin(radians(gls.latitude))
					)))
				) <= $%d)
				OR (
					(gls.latitude IS NULL OR gls.longitude IS NULL)
					AND $%d > 0
					AND c.id = (SELECT county_id FROM attractions WHERE id = $%d)
				)
			)`, argIdx, argIdx+1, argIdx, argIdx+2, argIdx+3, argIdx+4))
			args = append(args, nearLat, nearLon, km, aid, aid)
			argIdx += 5
		}
	}

	if locationType != "" {
		conditions = append(conditions, fmt.Sprintf("s.type = $%d", argIdx))
		args = append(args, locationType)
		argIdx++
	}

	if eventType != "" {
		conditions = append(conditions, fmt.Sprintf("et.slug = $%d", argIdx))
		args = append(args, eventType)
		argIdx++
	}
	if eventSubtype != "" {
		conditions = append(conditions, fmt.Sprintf("es.slug = $%d", argIdx))
		args = append(args, eventSubtype)
		argIdx++
	}
	// Overlap with [dateFrom, dateTo]: includes multi-day events on every day in range (day pick, month range).
	if dateFrom != "" && dateTo != "" {
		conditions = append(conditions, fmt.Sprintf("e.start_date <= $%d::date AND e.end_date >= $%d::date", argIdx, argIdx+1))
		args = append(args, dateTo, dateFrom)
		argIdx += 2
	} else if dateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("e.end_date >= $%d::date", argIdx))
		args = append(args, dateFrom)
		argIdx++
	} else if dateTo != "" {
		conditions = append(conditions, fmt.Sprintf("e.start_date <= $%d::date", argIdx))
		args = append(args, dateTo)
		argIdx++
	}

	// Strict: event must be for the same settlement as the venue, and use this venue as default or on schedule.
	if venueIDStr := strings.TrimSpace(r.URL.Query().Get("venue_id")); venueIDStr != "" {
		vid, errConv := strconv.Atoi(venueIDStr)
		if errConv == nil && vid > 0 {
			conditions = append(conditions, fmt.Sprintf(`EXISTS (
				SELECT 1 FROM venues vq
				WHERE vq.id = $%d
				  AND vq.settlement_id = e.location_id
				  AND (
					e.default_venue_id = vq.id
					OR EXISTS (
						SELECT 1 FROM event_schedule_days esd
						JOIN event_schedule_activities esa ON esa.event_day_id = esd.id
						WHERE esd.event_id = e.id AND esa.venue_id = vq.id
					)
				  )
			)`, argIdx))
			args = append(args, vid)
			argIdx++
		}
	}

	// Same settlement + venues of this kind (for “related” lists); optional exclude_venue_id drops strict-at-this-venue rows.
	venueKindSlug := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("venue_kind")))
	if venueKindSlug != "" && locationSlug != "" && countySlug != "" {
		conditions = append(conditions, fmt.Sprintf(`(
			EXISTS (
				SELECT 1 FROM venues vd
				WHERE vd.id = e.default_venue_id AND vd.kind = $%d AND vd.settlement_id = e.location_id
			)
			OR EXISTS (
				SELECT 1 FROM event_schedule_days esd
				JOIN event_schedule_activities esa ON esa.event_day_id = esd.id
				JOIN venues va ON va.id = esa.venue_id AND va.settlement_id = e.location_id
				WHERE esd.event_id = e.id AND va.kind = $%d
			)
		)`, argIdx, argIdx+1))
		args = append(args, venueKindSlug, venueKindSlug)
		argIdx += 2

		if exStr := strings.TrimSpace(r.URL.Query().Get("exclude_venue_id")); exStr != "" {
			exid, errEx := strconv.Atoi(exStr)
			if errEx == nil && exid > 0 {
				conditions = append(conditions, fmt.Sprintf(`NOT EXISTS (
					SELECT 1 FROM venues vqx
					WHERE vqx.id = $%d
					  AND vqx.settlement_id = e.location_id
					  AND (
						e.default_venue_id = vqx.id
						OR EXISTS (
							SELECT 1 FROM event_schedule_days esdx
							JOIN event_schedule_activities esax ON esax.event_day_id = esdx.id
							WHERE esdx.event_id = e.id AND esax.venue_id = vqx.id
						)
					  )
				)`, argIdx))
				args = append(args, exid)
				argIdx++
			}
		}
	}

	where := " WHERE " + strings.Join(conditions, " AND ")

	err = db.DB.QueryRow(baseCount+where, args...).Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	orderSQL := "e.start_date ASC"
	if strings.TrimSpace(strings.ToLower(r.URL.Query().Get("sort"))) == "title" {
		orderSQL = "e.title ASC, e.start_date ASC"
	}
	query := baseSelect + where + " ORDER BY " + orderSQL
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
		var defVID sql.NullInt64
		var defVName, defVSlug, defVKind, defVKindLabel string
		if err := rows.Scan(&ev.ID, &ev.LocationID, &ev.LocationName, &ev.LocationSlug, &ev.County, &ev.CountySlug, &ev.Title, &ev.Description, &ev.FeaturedImage, &ev.StartDate, &ev.StartTime, &ev.EndDate, &ev.EndTime, &ev.EventType, &ev.EventTypeLabel, &ev.EventSubtype, &ev.EventSubtypeLabel, &ev.AccessType, &ev.Organizer, &ev.EntryPrice, &ev.LocationType, &defVID, &defVName, &defVSlug, &defVKind, &defVKindLabel, &ev.HasSchedule); err != nil {
			log.Printf("HandleEvents rows.Scan: %v", err)
			continue
		}
		if defVID.Valid {
			x := int(defVID.Int64)
			ev.DefaultVenueID = &x
		}
		ev.DefaultVenueName = defVName
		ev.DefaultVenueSlug = defVSlug
		ev.DefaultVenueKind = defVKind
		ev.DefaultVenueKindLabel = defVKindLabel
		events = append(events, ev)
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
	var defVID sql.NullInt64
	var defVName, defVSlug, defVKind, defVKindLabel string
	err = db.DB.QueryRow(`
		SELECT e.id, e.location_id, s.name, s.slug, c.name, c.slug, e.title, COALESCE(e.description, ''), COALESCE(e.featured_image, ''),
		       e.start_date::text, COALESCE(e.start_time::text, ''), e.end_date::text, COALESCE(e.end_time::text, ''),
		       COALESCE(et.slug, ''), COALESCE(et.label_hu, ''), COALESCE(es.slug, ''), COALESCE(es.label_hu, ''),
		       COALESCE(e.access_type, 'public'), COALESCE(e.organizer, ''), COALESCE(e.entry_price, ''), COALESCE(s.type, ''),
		       e.default_venue_id, COALESCE(vdef.name, ''), COALESCE(vdef.slug, ''), COALESCE(vdef.kind, ''), COALESCE(vt.label_hu, ''),
		       EXISTS (SELECT 1 FROM event_schedule_days esd WHERE esd.event_id = e.id),
		       COALESCE(e.featured_image_copyright, '')
		FROM events e
		JOIN settlements s ON e.location_id = s.id
		JOIN counties c ON s.county_id = c.id
		JOIN catalog_event_types et ON e.event_type_id = et.id
		LEFT JOIN catalog_event_subtypes es ON e.event_subtype_id = es.id
		LEFT JOIN venues vdef ON e.default_venue_id = vdef.id
		LEFT JOIN venue_types vt ON vt.slug = vdef.kind
		WHERE e.id = $1`, id).Scan(&ev.ID, &ev.LocationID, &ev.LocationName, &ev.LocationSlug, &ev.County, &ev.CountySlug, &ev.Title, &ev.Description, &ev.FeaturedImage, &ev.StartDate, &ev.StartTime, &ev.EndDate, &ev.EndTime, &ev.EventType, &ev.EventTypeLabel, &ev.EventSubtype, &ev.EventSubtypeLabel, &ev.AccessType, &ev.Organizer, &ev.EntryPrice, &ev.LocationType, &defVID, &defVName, &defVSlug, &defVKind, &defVKindLabel, &ev.HasSchedule, &ev.FeaturedImageCopyright)
	if err == nil && defVID.Valid {
		x := int(defVID.Int64)
		ev.DefaultVenueID = &x
	}
	if err == nil {
		ev.DefaultVenueName = defVName
		ev.DefaultVenueSlug = defVSlug
		ev.DefaultVenueKind = defVKind
		ev.DefaultVenueKindLabel = defVKindLabel
	}
	if err == sql.ErrNoRows {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	schedule, err := fetchScheduleForEvent(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	out := models.EventWithSchedule{Event: ev, Schedule: schedule}
	json.NewEncoder(w).Encode(out)
}
