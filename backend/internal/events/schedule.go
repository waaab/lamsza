package events

import (
	"backend/internal/db"
	"backend/internal/models"
	"database/sql"
	"strings"
	"time"
)

func fetchScheduleForEvent(eventID int) ([]models.EventScheduleDay, error) {
	rows, err := db.DB.Query(`
		SELECT id, schedule_date::text, COALESCE(notes, ''), sort_order
		FROM event_schedule_days
		WHERE event_id = $1
		ORDER BY schedule_date ASC, sort_order ASC, id ASC`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []models.EventScheduleDay
	for rows.Next() {
		var d models.EventScheduleDay
		if err := rows.Scan(&d.ID, &d.ScheduleDate, &d.Notes, &d.SortOrder); err != nil {
			return nil, err
		}
		acts, err := fetchActivitiesForDay(d.ID)
		if err != nil {
			return nil, err
		}
		d.Activities = acts
		days = append(days, d)
	}
	if days == nil {
		days = []models.EventScheduleDay{}
	}
	return days, nil
}

func fetchActivitiesForDay(dayID int) ([]models.EventScheduleActivity, error) {
	rows, err := db.DB.Query(`
		SELECT a.id, COALESCE(a.activity_type, 'other'), COALESCE(a.starts_at::text, ''), COALESCE(a.ends_at::text, ''),
		       a.title, COALESCE(a.description, ''), a.sort_order, a.venue_id, COALESCE(v.name, ''), COALESCE(v.slug, '')
		FROM event_schedule_activities a
		LEFT JOIN venues v ON a.venue_id = v.id
		WHERE a.event_day_id = $1
		ORDER BY a.sort_order ASC, a.id ASC`, dayID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.EventScheduleActivity
	for rows.Next() {
		var a models.EventScheduleActivity
		var st, et string
		var vid sql.NullInt64
		var vname, vslug string
		if err := rows.Scan(&a.ID, &a.ActivityType, &st, &et, &a.Title, &a.Description, &a.SortOrder, &vid, &vname, &vslug); err != nil {
			return nil, err
		}
		a.StartsAt = formatTimeForJSON(st)
		a.EndsAt = formatTimeForJSON(et)
		if vid.Valid {
			x := int(vid.Int64)
			a.VenueID = &x
		}
		a.VenueName = vname
		a.VenueSlug = vslug
		list = append(list, a)
	}
	if list == nil {
		list = []models.EventScheduleActivity{}
	}
	return list, nil
}

func formatTimeForJSON(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	t, err := time.Parse("15:04:05", s)
	if err == nil {
		return t.Format("15:04")
	}
	t2, err := time.Parse("15:04", s)
	if err == nil {
		return t2.Format("15:04")
	}
	return s
}
