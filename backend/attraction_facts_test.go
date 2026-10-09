package main

import (
	"backend/internal/db"
	"encoding/json"
	"testing"
)

// The attraction's facts (elevation, area, depth) reach the public detail
// API, are absent when not set, and the database refuses a negative area or
// depth (UI_BASELINE "szf-pages"). The test attractions have no coordinates,
// which the detail and the list must take (NULL latitude and longitude).
func TestAttractionFacts(t *testing.T) {
	var countyID int
	if err := db.DB.QueryRow(`SELECT id FROM counties WHERE slug = 'hargita'`).Scan(&countyID); err != nil {
		t.Fatalf("county: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.DB.Exec(`DELETE FROM attractions WHERE slug IN ('teszt-to', 'teszt-to-ures')`)
	})
	if _, err := db.DB.Exec(`
		INSERT INTO attractions (county_id, name, slug, description, elevation_m, area_km2, depth_m)
		VALUES ($1, 'Teszt-tó', 'teszt-to', 'leírás', 946, 0.22, 7),
		       ($1, 'Üres tó', 'teszt-to-ures', 'leírás', NULL, NULL, NULL)`, countyID); err != nil {
		t.Fatal(err)
	}

	rr := doRequest(t, "GET", "/api/attractions?county_slug=hargita&slug=teszt-to", nil)
	if rr.Code != 200 {
		t.Fatalf("detail: %d %s", rr.Code, rr.Body.String())
	}
	var got struct {
		ElevationM *float64 `json:"elevation_m"`
		AreaKm2    *float64 `json:"area_km2"`
		DepthM     *float64 `json:"depth_m"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ElevationM == nil || *got.ElevationM != 946 || got.AreaKm2 == nil || *got.AreaKm2 != 0.22 || got.DepthM == nil || *got.DepthM != 7 {
		t.Fatalf("facts: %+v (%s)", got, rr.Body.String())
	}

	// An attraction without coordinates is still listed (it used to be dropped).
	rr = doRequest(t, "GET", "/api/attractions?county_slug=hargita", nil)
	var list []struct {
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	listed := false
	for _, a := range list {
		listed = listed || a.Slug == "teszt-to-ures"
	}
	if !listed {
		t.Errorf("an attraction without coordinates is missing from the list: %s", rr.Body.String())
	}

	rr = doRequest(t, "GET", "/api/attractions?county_slug=hargita&slug=teszt-to-ures", nil)
	var raw map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"elevation_m", "area_km2", "depth_m"} {
		if _, ok := raw[k]; ok {
			t.Errorf("%s present for an attraction without it: %s", k, rr.Body.String())
		}
	}

	for _, col := range []string{"area_km2", "depth_m"} {
		if _, err := db.DB.Exec(`UPDATE attractions SET ` + col + ` = -1 WHERE slug = 'teszt-to'`); err == nil {
			t.Errorf("negative %s accepted", col)
		}
	}
}
