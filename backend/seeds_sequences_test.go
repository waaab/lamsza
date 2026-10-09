package main

import (
	"backend/internal/account"
	"backend/internal/db"
	"backend/internal/events"
	"backend/internal/handlers"
	"backend/internal/pagefaq"
	"backend/internal/pages"
	"testing"
)

// The boot seeds insert only missing rows, so a backend start that finds them
// all in place takes no sequence value (lamsza OPEN_ITEMS: catalog seeds burn
// sequence numbers). With INSERT ... ON CONFLICT DO NOTHING every start moved
// pages, page_faq_sections, historical_seats, catalog_event_types/_subtypes,
// settlement_location_types and websites ahead with no row change.
func TestBootSeedsDoNotBurnSequences(t *testing.T) {
	seed := func() {
		db.SeedHistoricalSeatsContent()
		pages.MigratePages()
		pagefaq.Migrate()
		handlers.MigrateSettlementLocationTypes()
		events.Migrate()
		account.MigrateWebsites()
	}
	tables := []string{
		"historical_seats", "pages", "page_faq_sections", "settlement_location_types",
		"catalog_event_types", "catalog_event_subtypes", "websites",
	}
	type state struct {
		last   int64
		called bool
		rows   int
	}
	read := func() map[string]state {
		out := map[string]state{}
		for _, table := range tables {
			var seq string
			if err := db.DB.QueryRow(`SELECT pg_get_serial_sequence($1, 'id')`, table).Scan(&seq); err != nil {
				t.Fatalf("%s sequence: %v", table, err)
			}
			var st state
			if err := db.DB.QueryRow(`SELECT last_value, is_called FROM `+seq).Scan(&st.last, &st.called); err != nil {
				t.Fatalf("%s: %v", seq, err)
			}
			if err := db.DB.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&st.rows); err != nil {
				t.Fatalf("%s rows: %v", table, err)
			}
			out[table] = st
		}
		return out
	}

	seed() // the first boot may insert what is missing
	before := read()
	seed() // a second boot finds everything in place
	after := read()
	for _, table := range tables {
		if before[table] != after[table] {
			t.Errorf("%s: a second boot moved it from %+v to %+v", table, before[table], after[table])
		}
	}
}
