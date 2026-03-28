package events

import (
	"backend/internal/db"
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
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
		SELECT id, COALESCE(activity_type, 'other'), COALESCE(starts_at::text, ''), COALESCE(ends_at::text, ''),
		       title, COALESCE(description, ''), sort_order
		FROM event_schedule_activities
		WHERE event_day_id = $1
		ORDER BY sort_order ASC, id ASC`, dayID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.EventScheduleActivity
	for rows.Next() {
		var a models.EventScheduleActivity
		var st, et string
		if err := rows.Scan(&a.ID, &a.ActivityType, &st, &et, &a.Title, &a.Description, &a.SortOrder); err != nil {
			return nil, err
		}
		a.StartsAt = formatTimeForJSON(st)
		a.EndsAt = formatTimeForJSON(et)
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

// HandleAdminEventSchedule GET ?event_id= / PUT full replace body.
func HandleAdminEventSchedule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		idStr := r.URL.Query().Get("event_id")
		if idStr == "" {
			http.Error(w, "event_id kötelező", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			http.Error(w, "érvénytelen event_id", http.StatusBadRequest)
			return
		}
		days, err := fetchScheduleForEvent(id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"event_id": id,
			"days":     days,
		})

	case http.MethodPut:
		var body struct {
			EventID int `json:"event_id"`
			Days    []struct {
				ScheduleDate string `json:"schedule_date"`
				Notes        string `json:"notes"`
				Activities   []struct {
					ActivityType string `json:"activity_type"`
					StartsAt     string `json:"starts_at"`
					EndsAt       string `json:"ends_at"`
					Title        string `json:"title"`
					Description  string `json:"description"`
					SortOrder    int    `json:"sort_order"`
				} `json:"activities"`
			} `json:"days"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if body.EventID < 1 {
			http.Error(w, "event_id kötelező", http.StatusBadRequest)
			return
		}
		var exists int
		err := db.DB.QueryRow(`SELECT 1 FROM events WHERE id = $1`, body.EventID).Scan(&exists)
		if err == sql.ErrNoRows {
			http.Error(w, "esemény nem található", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		tx, err := db.DB.Begin()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer tx.Rollback()

		if _, err := tx.Exec(`DELETE FROM event_schedule_days WHERE event_id = $1`, body.EventID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		for di, d := range body.Days {
			dateStr := strings.TrimSpace(d.ScheduleDate)
			if dateStr == "" {
				continue
			}
			var dayID int
			qerr := tx.QueryRow(`
				INSERT INTO event_schedule_days (event_id, schedule_date, notes, sort_order)
				VALUES ($1, $2::date, $3, $4)
				RETURNING id`,
				body.EventID, dateStr, strings.TrimSpace(d.Notes), di).Scan(&dayID)
			if qerr != nil {
				http.Error(w, qerr.Error(), 500)
				return
			}
			for ai, a := range d.Activities {
				title := strings.TrimSpace(a.Title)
				if title == "" {
					continue
				}
				st := parseTimeOrNil(a.StartsAt)
				et := parseTimeOrNil(a.EndsAt)
				at := normalizeActivityType(a.ActivityType)
				_, execErr := tx.Exec(`
					INSERT INTO event_schedule_activities (event_day_id, activity_type, starts_at, ends_at, title, description, sort_order)
					VALUES ($1, $2, $3, $4, $5, $6, $7)`,
					dayID, at, st, et, title, strings.TrimSpace(a.Description), ai)
				if execErr != nil {
					http.Error(w, execErr.Error(), 500)
					return
				}
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		days, err := fetchScheduleForEvent(body.EventID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"event_id": body.EventID,
			"days":     days,
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func parseTimeOrNil(s string) interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if len(s) == 5 && s[2] == ':' {
		s = s + ":00"
	}
	t, err := time.Parse("15:04:05", s)
	if err != nil {
		return nil
	}
	return t.Format("15:04:05")
}

func normalizeActivityType(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "opening", "match", "closing", "other":
		return s
	default:
		return "other"
	}
}
