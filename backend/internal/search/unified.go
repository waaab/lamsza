package search

import (
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/models"
	"backend/internal/news"
	"backend/internal/utils"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/lib/pq"
)

// AttractionSearchHit is a lightweight shape for unified search JSON.
type AttractionSearchHit struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	CountySlug  string `json:"county_slug"`
	CountyName  string `json:"county_name"`
	Description string `json:"description,omitempty"`
}

// HistoricalSeatSearchHit is a lightweight shape for unified search JSON.
type HistoricalSeatSearchHit struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type UnifiedSearchResult struct {
	Locations         []models.Location         `json:"locations"`
	Entries           []models.Entry            `json:"entries"`
	Events            []models.Event            `json:"events"`
	News              []newsSearchItem          `json:"news"`
	Attractions       []AttractionSearchHit     `json:"attractions"`
	HistoricalSeats   []HistoricalSeatSearchHit `json:"historical_seats"`
}

type newsSearchItem struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	PubDate int64  `json:"pubDate"`
	Source  string `json:"source"`
	BgColor string `json:"bgColor"`
	Image   string `json:"image,omitempty"`
}

func HandleUnifiedSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		json.NewEncoder(w).Encode(UnifiedSearchResult{
			Locations:       []models.Location{},
			Entries:         []models.Entry{},
			Events:          []models.Event{},
			News:            []newsSearchItem{},
			Attractions:     []AttractionSearchHit{},
			HistoricalSeats: []HistoricalSeatSearchHit{},
		})
		return
	}

	normalizedQ := utils.Slugify(q)
	pattern := "%" + strings.ToLower(q) + "%"

	var wg sync.WaitGroup
	var locations []models.Location
	var entries []models.Entry
	var events []models.Event
	var newsItems []newsSearchItem
	var attractionHits []AttractionSearchHit
	var seatHits []HistoricalSeatSearchHit

	// Search locations (ILIKE on name, name_ro, name_de, county)
	wg.Add(1)
	go func() {
		defer wg.Done()
		rows, err := db.DB.Query(`
			SELECT id, name, COALESCE(name_ro,''), COALESCE(name_de,''), COALESCE(county,''),
				COALESCE(county_slug,''), COALESCE(type,''), COALESCE(slug,''),
				COALESCE(post_code,''), COALESCE(coordinates,''), COALESCE(population,''),
				COALESCE(area,''), COALESCE(crest,''), parent_id, COALESCE(is_county_seat, false)
			FROM locations
			WHERE unaccent(LOWER(name)) ILIKE unaccent($1)
			   OR unaccent(LOWER(COALESCE(name_ro,''))) ILIKE unaccent($1)
			   OR unaccent(LOWER(COALESCE(name_de,''))) ILIKE unaccent($1)
			   OR unaccent(LOWER(COALESCE(county,''))) ILIKE unaccent($1)
			ORDER BY name ASC
			LIMIT 15
		`, pattern)
		if err != nil {
			log.Printf("UnifiedSearch locations error: %v", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var loc models.Location
			if err := rows.Scan(&loc.ID, &loc.Name, &loc.NameRo, &loc.NameDe, &loc.County, &loc.CountySlug, &loc.Type, &loc.Slug, &loc.PostCode, &loc.Coordinates, &loc.Population, &loc.Area, &loc.Crest, &loc.ParentID, &loc.IsCountySeat); err == nil {
				locations = append(locations, loc)
			}
		}
	}()

	// Search attractions (látnivalók)
	wg.Add(1)
	go func() {
		defer wg.Done()
		rows, err := db.DB.Query(`
			SELECT a.id, a.name, a.slug, c.slug, c.name, COALESCE(a.description,'')
			FROM attractions a
			JOIN counties c ON a.county_id = c.id
			WHERE unaccent(LOWER(a.name)) ILIKE unaccent($1)
			   OR unaccent(LOWER(COALESCE(a.name_ro,''))) ILIKE unaccent($1)
			   OR unaccent(LOWER(COALESCE(a.name_de,''))) ILIKE unaccent($1)
			   OR unaccent(LOWER(COALESCE(a.description,''))) ILIKE unaccent($1)
			ORDER BY a.name ASC
			LIMIT 12
		`, pattern)
		if err != nil {
			log.Printf("UnifiedSearch attractions error: %v", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var h AttractionSearchHit
			if err := rows.Scan(&h.ID, &h.Name, &h.Slug, &h.CountySlug, &h.CountyName, &h.Description); err == nil {
				attractionHits = append(attractionHits, h)
			}
		}
	}()

	// Search historical seats (székek)
	wg.Add(1)
	go func() {
		defer wg.Done()
		rows, err := db.DB.Query(`
			SELECT id, name, slug
			FROM historical_seats
			WHERE unaccent(LOWER(name)) ILIKE unaccent($1)
			   OR unaccent(LOWER(COALESCE(name_ro,''))) ILIKE unaccent($1)
			   OR unaccent(LOWER(COALESCE(name_de,''))) ILIKE unaccent($1)
			ORDER BY name ASC
			LIMIT 8
		`, pattern)
		if err != nil {
			log.Printf("UnifiedSearch historical_seats error: %v", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var h HistoricalSeatSearchHit
			if err := rows.Scan(&h.ID, &h.Name, &h.Slug); err == nil {
				seatHits = append(seatHits, h)
			}
		}
	}()

	// Search entries (reuse FTS logic from EntriesHandler)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if normalizedQ == "" {
			return
		}
		rows, err := db.DB.Query(`
			SELECT e.id, COALESCE(e.type,''), COALESCE(ec.name,''), e.name, e.slug,
				s.name, s.slug, c.name, c.slug, s.type,
				COALESCE(s.name_ro,''), COALESCE(s.name_de,''),
				COALESCE(e.phone,''), COALESCE(e.address,''), COALESCE(e.notes,''),
				e.languages, COALESCE(e.url,''),
				CASE WHEN unaccent(LOWER(e.name)) = unaccent(LOWER($1)) THEN true ELSE false END as is_direct_match,
				ts_rank_cd(e.search_vector, plainto_tsquery('simple', $2)) as rank
			FROM entries e
			JOIN settlements s ON e.location_id = s.id
			JOIN counties c ON s.county_id = c.id
			LEFT JOIN entry_categories ec ON e.category_id = ec.id
			LEFT JOIN entry_tags et ON e.id = et.entry_id
			LEFT JOIN tags t ON et.tag_id = t.id
			WHERE e.search_vector @@ plainto_tsquery('simple', $2)
			GROUP BY e.id, ec.name, s.name, s.slug, c.name, c.slug, s.type, s.name_ro, s.name_de
			ORDER BY is_direct_match DESC, rank DESC, e.name ASC
			LIMIT 20
		`, q, normalizedQ)
		if err != nil {
			log.Printf("UnifiedSearch entries error: %v", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var e models.Entry
			var pqLanguages []string
			var rank float64
			if err := rows.Scan(&e.ID, &e.Type, &e.Category, &e.Name, &e.Slug, &e.Location, &e.LocationSlug, &e.LocationCounty, &e.CountySlug, &e.LocationType, &e.LocationRo, &e.LocationDe, &e.Phone, &e.Address, &e.Notes, pq.Array(&pqLanguages), &e.URL, &e.IsDirectMatch, &rank); err == nil {
				e.Languages = pqLanguages
				entries = append(entries, e)
			}
		}
	}()

	// Search events (if feature enabled)
	if config.AppConfig.Features.Events {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rows, err := db.DB.Query(`
				SELECT e.id, e.location_id, s.name, s.slug, c.name, c.slug,
					e.title, COALESCE(e.description, ''), e.start_date::text, COALESCE(e.start_time::text, ''),
					e.end_date::text, COALESCE(e.end_time::text, ''), e.event_type, COALESCE(e.organizer, '')
				FROM events e
				JOIN settlements s ON e.location_id = s.id
				JOIN counties c ON s.county_id = c.id
				WHERE e.end_date >= CURRENT_DATE
				  AND (
					unaccent(LOWER(e.title)) ILIKE unaccent($1)
					OR unaccent(LOWER(COALESCE(e.description,''))) ILIKE unaccent($1)
					OR unaccent(LOWER(COALESCE(e.organizer,''))) ILIKE unaccent($1)
					OR unaccent(LOWER(s.name)) ILIKE unaccent($1)
				  )
				ORDER BY e.start_date ASC
				LIMIT 15
			`, pattern)
			if err != nil {
				log.Printf("UnifiedSearch events error: %v", err)
				return
			}
			defer rows.Close()
			for rows.Next() {
				var ev models.Event
				if err := rows.Scan(&ev.ID, &ev.LocationID, &ev.LocationName, &ev.LocationSlug, &ev.County, &ev.CountySlug, &ev.Title, &ev.Description, &ev.StartDate, &ev.StartTime, &ev.EndDate, &ev.EndTime, &ev.EventType, &ev.Organizer); err == nil {
					events = append(events, ev)
				}
			}
		}()
	}

	// Search news (if feature enabled) - fetch and filter
	if config.AppConfig.Features.News {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items := news.FetchNewsItems(50)
			qLower := strings.ToLower(q)
			for _, item := range items {
				if strings.Contains(strings.ToLower(item.Title), qLower) {
					newsItems = append(newsItems, newsSearchItem{
						Title:   item.Title,
						Link:    item.Link,
						PubDate: item.PubDate,
						Source:  item.Source,
						BgColor: item.BgColor,
						Image:   item.Image,
					})
					if len(newsItems) >= 15 {
						break
					}
				}
			}
		}()
	}

	wg.Wait()

	if locations == nil {
		locations = []models.Location{}
	}
	if entries == nil {
		entries = []models.Entry{}
	}
	if events == nil {
		events = []models.Event{}
	}
	if newsItems == nil {
		newsItems = []newsSearchItem{}
	}
	if attractionHits == nil {
		attractionHits = []AttractionSearchHit{}
	}
	if seatHits == nil {
		seatHits = []HistoricalSeatSearchHit{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UnifiedSearchResult{
		Locations:       locations,
		Entries:         entries,
		Events:          events,
		News:            newsItems,
		Attractions:     attractionHits,
		HistoricalSeats: seatHits,
	})
}
