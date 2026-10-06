package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"backend/internal/db"
	"backend/internal/directory"
	"backend/internal/utils"

	"github.com/lib/pq"
)

// The directory write API lives in lamsza-admin now (BOG-42), so these tests
// cannot POST /api/admin/entries to build their fixtures any more. They insert
// the row themselves instead. Keep this file as the only place that writes the
// directory tables from a test: a public handler under test must never need it.

type entryFixture struct {
	Name            string
	LocationID      interface{}
	Type            string // defaults to "Vállalkozás"
	Category        string // defaults to "Bútor"
	URL             string
	Phone           string
	Address         string
	Notes           string
	Languages       []string // defaults to {"HU"}
	Verified        bool
	Claimed         bool
	Published       *bool // defaults to the column default, true
	RatingsEnabled  bool
	Hours           interface{}
	HoursEnabled    bool
	DeliveryHours   interface{}
	DeliveryEnabled bool
	Photos          interface{}
	Tags            []string
}

func lookupID(t *testing.T, table, name string) int {
	t.Helper()
	var id int
	if err := db.DB.QueryRow("SELECT id FROM "+table+" WHERE name = $1", name).Scan(&id); err != nil {
		t.Fatalf("%s %q not found: %v", table, name, err)
	}
	return id
}

func jsonOrDefault(t *testing.T, v interface{}, fallback string) string {
	t.Helper()
	if v == nil {
		return fallback
	}
	if s, ok := v.(string); ok {
		return s
	}
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal fixture json: %v", err)
	}
	return string(raw)
}

// createEntryFixture inserts one directory entry and returns its id and slug.
func createEntryFixture(t *testing.T, f entryFixture) (int, string) {
	t.Helper()
	if f.Type == "" {
		f.Type = utils.EntryTypeVallalkozas
	}
	if f.Category == "" {
		f.Category = "Bútor"
	}
	if len(f.Languages) == 0 {
		f.Languages = []string{"HU"}
	}
	published := true
	if f.Published != nil {
		published = *f.Published
	}
	typeID := lookupID(t, "entry_types", utils.CanonicalEntryType(f.Type))
	catID := lookupID(t, "entry_categories", utils.CanonicalEntryCategory(f.Category))
	slug := utils.Slugify(f.Name)

	var id int
	err := db.DB.QueryRow(`INSERT INTO entries
		(type_id, location_id, category_id, cat_name, name, slug, url, phone, address, notes,
		 languages, verified, claimed, published, ratings_enabled,
		 hours, hours_enabled, delivery_hours, delivery_enabled, photos)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16::jsonb,$17,$18::jsonb,$19,$20::jsonb)
		RETURNING id`,
		typeID, f.LocationID, catID, utils.CanonicalEntryCategory(f.Category), f.Name, slug,
		f.URL, f.Phone, f.Address, f.Notes, pq.Array(f.Languages),
		f.Verified, f.Claimed, published, f.RatingsEnabled,
		jsonOrDefault(t, f.Hours, "{}"), f.HoursEnabled,
		jsonOrDefault(t, f.DeliveryHours, "{}"), f.DeliveryEnabled,
		jsonOrDefault(t, f.Photos, "[]"),
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert entry fixture %q: %v", f.Name, err)
	}
	if err := directory.ReplaceEntryCategories(db.DB, id, []int{catID}); err != nil {
		t.Fatalf("link entry fixture categories: %v", err)
	}
	for _, tag := range f.Tags {
		var tagID int
		if err := db.DB.QueryRow("INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id", tag).Scan(&tagID); err != nil {
			t.Fatalf("insert tag fixture %q: %v", tag, err)
		}
		if _, err := db.DB.Exec("INSERT INTO entry_tags (entry_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", id, tagID); err != nil {
			t.Fatalf("link tag fixture %q: %v", tag, err)
		}
	}
	return id, slug
}

// createEntry is the short form most tests want: a plain published entry.
func createEntry(t *testing.T, name string, locID interface{}) (int, string) {
	t.Helper()
	return createEntryFixture(t, entryFixture{Name: name, LocationID: locID})
}

func deleteEntryFixture(t *testing.T, id interface{}) {
	t.Helper()
	db.DB.Exec("DELETE FROM websites WHERE entry_id = $1", id)
	db.DB.Exec("DELETE FROM entries WHERE id = $1", id)
}

// setEntryColumns is the fixture stand-in for an admin PUT.
func setEntryColumns(t *testing.T, id interface{}, assignments map[string]interface{}) {
	t.Helper()
	set := ""
	args := []interface{}{}
	i := 1
	for col, v := range assignments {
		if i > 1 {
			set += ", "
		}
		set += fmt.Sprintf("%s = $%d", col, i)
		args = append(args, v)
		i++
	}
	args = append(args, id)
	if _, err := db.DB.Exec(fmt.Sprintf("UPDATE entries SET %s WHERE id = $%d", set, i), args...); err != nil {
		t.Fatalf("update entry fixture: %v", err)
	}
}

// approveWebsiteFixture is what the admin "approve" action does to a pending
// website submission. The action itself lives in lamsza-admin.
func approveWebsiteFixture(t *testing.T, id interface{}) {
	t.Helper()
	res, err := db.DB.Exec(`UPDATE websites
		SET status = 'approved', approved_at = NOW(),
		    approved_by = (SELECT id FROM users WHERE LOWER(email) = 'admin@test.lamsza')
		WHERE id = $1 AND status = 'pending'`, id)
	if err != nil {
		t.Fatalf("approve website fixture: %v", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		t.Fatalf("approve website fixture: no pending website %v", id)
	}
}

// publishEntryFixture is the admin listing-queue publish action.
func publishEntryFixture(t *testing.T, id interface{}) {
	t.Helper()
	if _, err := db.DB.Exec(`UPDATE entries SET published = true WHERE id = $1`, id); err != nil {
		t.Fatalf("publish entry fixture: %v", err)
	}
}

// acceptPendingClaim promotes a pending claim to an active owner. The admin
// listing queue that used to do this now lives in lamsza-admin.
func acceptPendingClaim(t *testing.T, entryID interface{}, email string) {
	t.Helper()
	res, err := db.DB.Exec(`UPDATE entry_members SET status = 'active'
		WHERE entry_id = $1
		  AND user_id = (SELECT id FROM users WHERE LOWER(email) = LOWER($2))
		  AND role = 'owner' AND status = 'pending'`, entryID, email)
	if err != nil {
		t.Fatalf("accept pending claim: %v", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		t.Fatalf("accept pending claim: no pending owner row for entry %v / %s", entryID, email)
	}
}
