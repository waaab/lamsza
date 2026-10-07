package main

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"backend/internal/clock"
	"backend/internal/db"
)

// standAt fixes the backend's clock at a UTC instant for one test (R19).
func standAt(t *testing.T, utc string) {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, utc)
	if err != nil {
		t.Fatal(err)
	}
	prev := clock.Now
	clock.Now = func() time.Time { return ts }
	t.Cleanup(func() { clock.Now = prev })
}

func eventListed(t *testing.T, title string) bool {
	t.Helper()
	rr := doRequest(t, "GET", "/api/events?limit=500", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/events: %d %s", rr.Code, rr.Body.String())
	}
	var out struct {
		Events []struct {
			Title string `json:"title"`
		} `json:"events"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	for _, e := range out.Events {
		if e.Title == title {
			return true
		}
	}
	return false
}

// An event stays listed through its last day in Bucharest and goes at
// Bucharest midnight, not at UTC midnight and not three hours later (R19).
func TestEventsEndAtBucharestMidnight(t *testing.T) {
	var locID, typeID int
	if err := db.DB.QueryRow(`SELECT id FROM settlements ORDER BY id LIMIT 1`).Scan(&locID); err != nil {
		t.Fatal(err)
	}
	if err := db.DB.QueryRow(`SELECT id FROM catalog_event_types ORDER BY id LIMIT 1`).Scan(&typeID); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		day, before, after string
	}{
		{"2026-07-15", "2026-07-15T20:59:00Z", "2026-07-15T22:30:00Z"}, // summer: 23:59 vs 01:30 next day (UTC still the 15th)
		{"2026-01-15", "2026-01-15T21:59:00Z", "2026-01-15T22:00:00Z"}, // winter: 23:59 vs 00:00 next day
		{"2026-03-29", "2026-03-29T20:59:00Z", "2026-03-29T21:00:00Z"}, // spring clock change: the short day
		{"2026-10-25", "2026-10-25T21:59:00Z", "2026-10-25T22:00:00Z"}, // autumn clock change: the long day
	} {
		title := "r19-teszt-" + c.day
		var id int
		if err := db.DB.QueryRow(`INSERT INTO events (title, location_id, event_type_id, start_date, end_date)
			VALUES ($1, $2, $3, $4, $4) RETURNING id`, title, locID, typeID, c.day).Scan(&id); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { db.DB.Exec(`DELETE FROM events WHERE id = $1`, id) })

		standAt(t, c.before)
		if !eventListed(t, title) {
			t.Errorf("event on %s hidden at %s (its last minute in Bucharest)", c.day, c.before)
		}
		standAt(t, c.after)
		if eventListed(t, title) {
			t.Errorf("event on %s still listed at %s (already the next day in Bucharest)", c.day, c.after)
		}
	}
}
