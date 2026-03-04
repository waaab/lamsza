package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"unicode"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func slugify(s string) string {
	s = strings.ToLower(s)
	// Simple Hungarian replacement
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ö", "o", "ő", "o", "ú", "u", "ü", "u", "ű", "u",
	)
	s = replacer.Replace(s)

	var res strings.Builder
	lastDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			res.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			res.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(res.String(), "-")
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers directly
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Missing url param"}`))
		return
	}

	// Fix: Encode the URL components to prevent Go 400 Bad Request on spaces like "Miercurea Ciuc"
	parsedURL, err := url.Parse(targetURL)
	if err == nil {
		parsedURL.RawQuery = parsedURL.Query().Encode()
		targetURL = parsedURL.String()
	}

	log.Println("Proxy request for:", targetURL)

	// Fetch the URL with a custom User-Agent to prevent 500 blocks, utilizing a client that handles redirects
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml, */*")

	// Create a Cookie Jar to handle redirect cookies from Incapsula/Cloudflare
	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: jar,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Proxy error:", err)
		http.Error(w, "Error fetching target URL", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy specific necessary headers from the response
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	// Stream the body out to the client
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Error streaming response for %s: %v", targetURL, err)
	}
}

type Entry struct {
	ID             string   `json:"id"`
	Type           string   `json:"type"`
	Category       string   `json:"category"`
	Name           string   `json:"name"`
	Slug           string   `json:"slug"`
	Location       string   `json:"location"`
	LocationSlug   string   `json:"location_slug"`
	LocationCounty string   `json:"location_county"`
	CountySlug     string   `json:"county_slug"`
	LocationType   string   `json:"location_type"`
	LocationRo     string   `json:"location_ro"`
	LocationDe     string   `json:"location_de"`
	Phone          string   `json:"phone"`
	Address        string   `json:"address"`
	Notes          string   `json:"notes"`
	Tags           []string `json:"tags"`
	Languages      []string `json:"languages"`
	URL            string   `json:"url"`
	IsDirectMatch  bool     `json:"is_direct_match"`
}

var db *sql.DB

func initDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Use default local postgres connection string
		connStr = "postgres://lamsza_user:lamsza_password@localhost:5433/lamsza?sslmode=disable"
	}
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Println("DB Open error:", err)
	} else {
		if err = db.Ping(); err != nil {
			log.Println("DB Ping error:", err)
		} else {
			log.Println("Connected to PostgreSQL")
			// Auto-create entry_types table if not exists
			db.Exec(`CREATE TABLE IF NOT EXISTS entry_types (
				id SERIAL PRIMARY KEY,
				name VARCHAR(50) NOT NULL UNIQUE
			)`)
			db.Exec(`INSERT INTO entry_types (name) VALUES ('service'), ('business'), ('other') ON CONFLICT (name) DO NOTHING`)
		}
	}
}

func entriesHandler(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	query := strings.ToLower(r.URL.Query().Get("q"))
	category := strings.ToLower(r.URL.Query().Get("category"))
	county := strings.ToLower(r.URL.Query().Get("county"))
	locationSlug := strings.ToLower(r.URL.Query().Get("location_slug"))
	// Fallback for generic 'slug' param
	if locationSlug == "" {
		locationSlug = strings.ToLower(r.URL.Query().Get("slug"))
	}
	entryType := strings.ToLower(r.URL.Query().Get("type"))

	if db == nil {
		// Fallback if DB is down
		w.Write([]byte(`[]`))
		return
	}

	sqlQuery := `
		SELECT 
			e.id, COALESCE(e.type, ''), COALESCE(ec.name, e.category, ''), e.name, COALESCE(e.slug, ''),
			l.name as location, COALESCE(l.slug, ''), COALESCE(l.county, ''), COALESCE(l.county_slug, ''), COALESCE(l.type, ''), 
			COALESCE(l.name_ro, ''), COALESCE(l.name_de, ''),
			COALESCE(e.phone, ''), COALESCE(e.address, ''), COALESCE(e.notes, ''), 
			COALESCE(e.url, ''), e.languages,
			COALESCE(array_agg(t.name) FILTER (WHERE t.name IS NOT NULL), ARRAY[]::VARCHAR[]) as tags
		FROM entries e
		JOIN locations l ON e.location_id = l.id
		LEFT JOIN entry_categories ec ON e.category_id = ec.id
		LEFT JOIN entry_tags et ON e.id = et.entry_id
		LEFT JOIN tags t ON et.tag_id = t.id
		WHERE 1=1
	`
	args := []interface{}{}
	argId := 1

	if category != "" && category != "osszes" {
		sqlQuery += ` AND e.category = $` + javaToPg(argId)
		args = append(args, category)
		argId++
	}

	if county != "" {
		sqlQuery += ` AND LOWER(l.county) = $` + javaToPg(argId)
		args = append(args, county)
		argId++
	}

	if locationSlug != "" {
		sqlQuery += ` AND L.slug = $` + javaToPg(argId)
		args = append(args, locationSlug)
		argId++
	}

	if entryType != "" {
		sqlQuery += ` AND LOWER(e.type) = $` + javaToPg(argId)
		args = append(args, entryType)
		argId++
	}

	if query != "" {
		// Use unaccent and trigram-powered LIKE to support diacritic-insensitive searches (e.g., 'korhaz' matches 'kórház')
		// while still utilizing pg_trgm indexes.
		// Added multilingual search: also check Romanian (name_ro) and German (name_de) location names.
		likeQuery := "%" + query + "%"
		sqlQuery += ` AND (
			unaccent(LOWER(e.name)) LIKE unaccent(LOWER($` + javaToPg(argId) + `)) 
			OR unaccent(LOWER(l.name)) LIKE unaccent(LOWER($` + javaToPg(argId) + `)) 
			OR unaccent(LOWER(l.name_ro)) LIKE unaccent(LOWER($` + javaToPg(argId) + `)) 
			OR unaccent(LOWER(l.name_de)) LIKE unaccent(LOWER($` + javaToPg(argId) + `)) 
			OR unaccent(LOWER(e.notes)) LIKE unaccent(LOWER($` + javaToPg(argId) + `)) 
			OR unaccent(LOWER(e.category)) LIKE unaccent(LOWER($` + javaToPg(argId) + `)) 
			OR EXISTS(SELECT 1 FROM entry_tags et2 JOIN tags t2 ON et2.tag_id = t2.id WHERE et2.entry_id = e.id AND unaccent(LOWER(t2.name)) LIKE unaccent(LOWER($` + javaToPg(argId) + `)))
		)`
		args = append(args, likeQuery)
		argId++
	}

	// Group by e.id so array_agg resolves appropriately, and order
	sqlQuery += ` GROUP BY e.id, e.type, ec.name, e.category, e.name, e.slug, l.name, l.slug, l.county, l.county_slug, l.type, l.name_ro, l.name_de, e.phone, e.address, e.notes, e.url, e.languages ORDER BY e.name ASC`
	rows, err := db.Query(sqlQuery, args...)
	if err != nil {
		log.Println("Query error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []Entry
	for rows.Next() {
		var s Entry
		// Postgres arrays map to pq.StringArray
		var pqLanguages []string
		var pqTags []string

		if err := rows.Scan(
			&s.ID, &s.Type, &s.Category, &s.Name, &s.Slug,
			&s.Location, &s.LocationSlug, &s.LocationCounty, &s.CountySlug, &s.LocationType,
			&s.LocationRo, &s.LocationDe,
			&s.Phone, &s.Address, &s.Notes,
			&s.URL, pq.Array(&pqLanguages), pq.Array(&pqTags),
		); err == nil {
			s.Languages = pqLanguages
			s.Tags = pqTags
			// Simple direct match logic: lower name == lower query
			if query != "" && strings.Contains(strings.ToLower(s.Name), query) {
				s.IsDirectMatch = true
			}
			results = append(results, s)
		} else {
			log.Println("Row scan error:", err)
		}
	}

	if results == nil {
		results = []Entry{} // ensure we output [] instead of null
	}
	json.NewEncoder(w).Encode(results)
}

// Helper to convert index to postgres placeholder (just returns the index as a string)
func javaToPg(i int) string {
	importStr := "0"
	if i == 1 {
		importStr = "1"
	}
	if i == 2 {
		importStr = "2"
	}
	if i == 3 {
		importStr = "3"
	}
	if i == 4 {
		importStr = "4"
	}
	if i == 5 {
		importStr = "5"
	}
	if i == 6 {
		importStr = "6"
	}
	return importStr
}

type Mondas struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

func handleAdminMondasok(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if db == nil {
		w.Write([]byte(`[]`))
		return
	}

	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, text, created_at FROM mondasok ORDER BY id DESC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []Mondas
		for rows.Next() {
			var m Mondas
			if err := rows.Scan(&m.ID, &m.Text, &m.CreatedAt); err == nil {
				res = append(res, m)
			}
		}
		if res == nil {
			res = []Mondas{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var m Mondas
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err := db.QueryRow("INSERT INTO mondasok (text) VALUES ($1) RETURNING id, created_at", m.Text).Scan(&m.ID, &m.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(m)

	case "PUT":
		var m Mondas
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, err := db.Exec("UPDATE mondasok SET text=$1 WHERE id=$2", m.Text, m.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.Exec("DELETE FROM mondasok WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

type QuickLink struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	BgColor string `json:"bg_color"`
}

func handleAdminQuickLinks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if db == nil {
		w.Write([]byte(`[]`))
		return
	}

	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, title, url, COALESCE(bg_color, '#ffffff') FROM quick_links ORDER BY id ASC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []QuickLink
		for rows.Next() {
			var ql QuickLink
			if err := rows.Scan(&ql.ID, &ql.Title, &ql.URL, &ql.BgColor); err == nil {
				res = append(res, ql)
			}
		}
		if res == nil {
			res = []QuickLink{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var ql QuickLink
		if err := json.NewDecoder(r.Body).Decode(&ql); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err := db.QueryRow("INSERT INTO quick_links (title, url, bg_color) VALUES ($1, $2, $3) RETURNING id", ql.Title, ql.URL, ql.BgColor).Scan(&ql.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(ql)

	case "PUT":
		var ql QuickLink
		if err := json.NewDecoder(r.Body).Decode(&ql); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, err := db.Exec("UPDATE quick_links SET title=$1, url=$2, bg_color=$3 WHERE id=$4",
			ql.Title, ql.URL, ql.BgColor, ql.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.Exec("DELETE FROM quick_links WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

type NewsFeed struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	FeedURL string `json:"feed_url"`
	BgColor string `json:"bg_color"`
}

func handleAdminNewsFeeds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if db == nil {
		w.Write([]byte(`[]`))
		return
	}

	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, title, feed_url, COALESCE(bg_color, '#ffebd6') FROM news_feeds ORDER BY id ASC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []NewsFeed
		for rows.Next() {
			var nf NewsFeed
			if err := rows.Scan(&nf.ID, &nf.Title, &nf.FeedURL, &nf.BgColor); err == nil {
				res = append(res, nf)
			}
		}
		if res == nil {
			res = []NewsFeed{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var nf NewsFeed
		if err := json.NewDecoder(r.Body).Decode(&nf); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		bgColor := nf.BgColor
		if bgColor == "" {
			bgColor = "#ffebd6"
		}
		err := db.QueryRow("INSERT INTO news_feeds (title, feed_url, bg_color) VALUES ($1, $2, $3) RETURNING id", nf.Title, nf.FeedURL, bgColor).Scan(&nf.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		nf.BgColor = bgColor
		json.NewEncoder(w).Encode(nf)

	case "PUT":
		var nf NewsFeed
		if err := json.NewDecoder(r.Body).Decode(&nf); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, err := db.Exec("UPDATE news_feeds SET title=$1, feed_url=$2, bg_color=$3 WHERE id=$4",
			nf.Title, nf.FeedURL, nf.BgColor, nf.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.Exec("DELETE FROM news_feeds WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

type Location struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	NameRo     string `json:"name_ro"`
	NameDe     string `json:"name_de"`
	County     string `json:"county"`
	CountySlug string `json:"county_slug"`
	Type       string `json:"type"`
	Slug       string `json:"slug"`
}

func handleAdminLocations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if db == nil {
		w.Write([]byte(`[]`))
		return
	}

	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, name, COALESCE(name_ro, ''), COALESCE(name_de, ''), COALESCE(county, ''), COALESCE(county_slug, ''), COALESCE(type, ''), COALESCE(slug, '') FROM locations ORDER BY name ASC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []Location
		for rows.Next() {
			var loc Location
			if err := rows.Scan(&loc.ID, &loc.Name, &loc.NameRo, &loc.NameDe, &loc.County, &loc.CountySlug, &loc.Type, &loc.Slug); err == nil {
				res = append(res, loc)
			}
		}
		if res == nil {
			res = []Location{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var l Location
		if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		l.Slug = slugify(l.Name)
		l.CountySlug = slugify(l.County)
		err := db.QueryRow("INSERT INTO locations (name, name_ro, name_de, county, county_slug, type, slug) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
			l.Name, l.NameRo, l.NameDe, l.County, l.CountySlug, l.Type, l.Slug).Scan(&l.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(l)

	case "PUT":
		var l Location
		if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		l.Slug = slugify(l.Name)
		l.CountySlug = slugify(l.County)
		_, err := db.Exec("UPDATE locations SET name=$1, name_ro=$2, name_de=$3, county=$4, county_slug=$5, type=$6, slug=$7 WHERE id=$8",
			l.Name, l.NameRo, l.NameDe, l.County, l.CountySlug, l.Type, l.Slug, l.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.Exec("DELETE FROM locations WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

type AdminEntry struct {
	ID         int      `json:"id"`
	Type       string   `json:"type"`
	LocationID int      `json:"location_id"`
	CategoryID *int     `json:"category_id"`
	Category   string   `json:"category"`
	Name       string   `json:"name"`
	Slug       string   `json:"slug"`
	URL        string   `json:"url"`
	Phone      string   `json:"phone"`
	Address    string   `json:"address"`
	Notes      string   `json:"notes"`
	Languages  []string `json:"languages"`
	Tags       []string `json:"tags"`
}

func handleAdminEntries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if db == nil {
		w.Write([]byte(`[]`))
		return
	}

	switch r.Method {
	case "GET":
		// Use ARRAY_AGG to collect individual tags from the relational table for each entry
		sqlQuery := `
			SELECT 
				e.id, e.type, e.location_id, e.category_id, COALESCE(e.category, ''), e.name, COALESCE(e.slug, ''),
				COALESCE(e.url, ''), COALESCE(e.phone, ''), COALESCE(e.address, ''), COALESCE(e.notes, ''), 
				e.languages,
				COALESCE(array_agg(t.name) FILTER (WHERE t.name IS NOT NULL), ARRAY[]::VARCHAR[]) as tags
			FROM entries e
			LEFT JOIN entry_tags et ON e.id = et.entry_id
			LEFT JOIN tags t ON et.tag_id = t.id
			GROUP BY e.id, e.type, e.location_id, e.category_id, e.category, e.name, e.url, e.phone, e.address, e.notes, e.languages
			ORDER BY e.id DESC
		`
		rows, err := db.Query(sqlQuery)
		if err != nil {
			log.Println("handleAdminEntries GET error:", err)
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []AdminEntry
		for rows.Next() {
			var s AdminEntry
			var pqLanguages []string
			var pqTags []string
			if err := rows.Scan(&s.ID, &s.Type, &s.LocationID, &s.CategoryID, &s.Category, &s.Name, &s.Slug, &s.URL, &s.Phone, &s.Address, &s.Notes, pq.Array(&pqLanguages), pq.Array(&pqTags)); err == nil {
				s.Languages = pqLanguages
				s.Tags = pqTags
				res = append(res, s)
			} else {
				log.Println("handleAdminEntries rows.Scan error:", err)
			}
		}
		if res == nil {
			res = []AdminEntry{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var s AdminEntry
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if s.Type == "" {
			s.Type = "service"
		}
		if len(s.Languages) == 0 {
			s.Languages = []string{"HU"}
		}

		// Handle potential nil value for category_id
		var catID interface{} = s.CategoryID
		if s.CategoryID == nil {
			catID = nil
		}

		s.Slug = slugify(s.Name)
		err := db.QueryRow("INSERT INTO entries (type, location_id, category_id, category, name, slug, url, phone, address, notes, languages) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id",
			s.Type, s.LocationID, catID, s.Category, s.Name, s.Slug, s.URL, s.Phone, s.Address, s.Notes, pq.Array(s.Languages)).Scan(&s.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// 2. Establish Tags
		for _, tagName := range s.Tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}
			var tagID int
			err = db.QueryRow("INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id", tagName).Scan(&tagID)
			if err == nil {
				db.Exec("INSERT INTO entry_tags (entry_id, tag_id) VALUES ($1, $2)", s.ID, tagID)
			}
		}

		json.NewEncoder(w).Encode(s)

	case "PUT":
		var s AdminEntry
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if s.Type == "" {
			s.Type = "service"
		}

		// Handle potential nil value for category_id
		var catID interface{} = s.CategoryID
		if s.CategoryID == nil {
			catID = nil
		}

		s.Slug = slugify(s.Name)
		_, err := db.Exec(`UPDATE entries SET 
            type=$1, location_id=$2, category_id=$3, category=$4, name=$5, slug=$6, url=$7, 
            phone=$8, address=$9, notes=$10, languages=$11 
            WHERE id=$12`,
			s.Type, s.LocationID, catID, s.Category, s.Name, s.Slug, s.URL, s.Phone, s.Address, s.Notes, pq.Array(s.Languages), s.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// 2. Clear previous tags and establish new ones
		db.Exec("DELETE FROM entry_tags WHERE entry_id = $1", s.ID)
		for _, tagName := range s.Tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}
			var tagID int
			err = db.QueryRow("INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id", tagName).Scan(&tagID)
			if err == nil {
				db.Exec("INSERT INTO entry_tags (entry_id, tag_id) VALUES ($1, $2)", s.ID, tagID)
			}
		}

		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.Exec("DELETE FROM entries WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

type EntryCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func handleAdminEntryCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if db == nil {
		w.Write([]byte(`[]`))
		return
	}

	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, name FROM entry_categories ORDER BY name ASC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []EntryCategory
		for rows.Next() {
			var sc EntryCategory
			if err := rows.Scan(&sc.ID, &sc.Name); err == nil {
				res = append(res, sc)
			}
		}
		if res == nil {
			res = []EntryCategory{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var sc EntryCategory
		if err := json.NewDecoder(r.Body).Decode(&sc); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		slug := slugify(sc.Name)
		err := db.QueryRow("INSERT INTO entry_categories (name, slug) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET slug=EXCLUDED.slug RETURNING id", sc.Name, slug).Scan(&sc.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(sc)

	case "PUT":
		var sc EntryCategory
		if err := json.NewDecoder(r.Body).Decode(&sc); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		slug := slugify(sc.Name)
		_, err := db.Exec("UPDATE entry_categories SET name=$1, slug=$2 WHERE id=$3", sc.Name, slug, sc.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.Exec("DELETE FROM entry_categories WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

type EntryType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func handleAdminEntryTypes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if db == nil {
		w.Write([]byte(`[]`))
		return
	}

	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, name FROM entry_types ORDER BY name ASC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []EntryType
		for rows.Next() {
			var et EntryType
			if err := rows.Scan(&et.ID, &et.Name); err == nil {
				res = append(res, et)
			}
		}
		if res == nil {
			res = []EntryType{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var et EntryType
		if err := json.NewDecoder(r.Body).Decode(&et); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err := db.QueryRow("INSERT INTO entry_types (name) VALUES ($1) RETURNING id", et.Name).Scan(&et.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(et)

	case "PUT":
		var et EntryType
		if err := json.NewDecoder(r.Body).Decode(&et); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, err := db.Exec("UPDATE entry_types SET name=$1 WHERE id=$2", et.Name, et.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.Exec("DELETE FROM entry_types WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func handleService(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Missing slug", 400)
		return
	}

	sqlQuery := `
		SELECT 
			e.id, COALESCE(e.type, ''), COALESCE(ec.name, e.category, ''), e.name, COALESCE(e.slug, ''),
			l.name as location, COALESCE(l.slug, ''), COALESCE(l.county, ''), COALESCE(l.county_slug, ''), COALESCE(l.type, ''), 
			COALESCE(l.name_ro, ''), COALESCE(l.name_de, ''),
			COALESCE(e.phone, ''), COALESCE(e.address, ''), COALESCE(e.notes, ''), 
			COALESCE(e.url, ''), e.languages,
			COALESCE(array_agg(t.name) FILTER (WHERE t.name IS NOT NULL), ARRAY[]::VARCHAR[]) as tags
		FROM entries e
		JOIN locations l ON e.location_id = l.id
		LEFT JOIN entry_categories ec ON e.category_id = ec.id
		LEFT JOIN entry_tags et ON e.id = et.entry_id
		LEFT JOIN tags t ON et.tag_id = t.id
		WHERE e.slug = $1
		GROUP BY e.id, e.type, ec.name, e.category, e.name, e.slug, l.name, l.slug, l.county, l.county_slug, l.type, l.name_ro, l.name_de, e.phone, e.address, e.notes, e.url, e.languages
	`
	var s Entry
	var pqLanguages []string
	var pqTags []string
	err := db.QueryRow(sqlQuery, slug).Scan(
		&s.ID, &s.Type, &s.Category, &s.Name, &s.Slug,
		&s.Location, &s.LocationSlug, &s.LocationCounty, &s.CountySlug, &s.LocationType,
		&s.LocationRo, &s.LocationDe,
		&s.Phone, &s.Address, &s.Notes,
		&s.URL, pq.Array(&pqLanguages), pq.Array(&pqTags),
	)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Not found"}`))
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}
	s.Languages = pqLanguages
	s.Tags = pqTags
	json.NewEncoder(w).Encode(s)
}

func handleCountyNews(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	slug := r.URL.Query().Get("slug")
	if slug == "" {
		w.Write([]byte(`[]`))
		return
	}

	// 1. Resolve slug to location name and county
	var name, nameRo, county string
	err := db.QueryRow("SELECT name, COALESCE(name_ro, ''), COALESCE(county, '') FROM locations WHERE slug = $1", slug).Scan(&name, &nameRo, &county)
	if err != nil {
		log.Println("News slug resolution error:", err)
		w.Write([]byte(`[]`))
		return
	}

	// 2. Fetch all news feeds URLs
	rows, err := db.Query("SELECT title, feed_url FROM news_feeds")
	if err != nil {
		log.Println("Error fetching news feeds:", err)
		w.Write([]byte(`[]`))
		return
	}
	defer rows.Close()

	// 3. Implement context-aware mock filtering
	// In a full implementation, we would fetch from 'rows' (feed URLs) and filter the content.
	type NewsItem struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Source  string `json:"source"`
		Date    string `json:"date"`
		Summary string `json:"summary"`
		BgColor string `json:"bg_color"`
	}

	filteredNews := []NewsItem{
		{
			Title:   fmt.Sprintf("Friss hírek: %s és környéke", name),
			Link:    "#",
			Source:  "Lamsza Regionális",
			Date:    "Ma",
			Summary: fmt.Sprintf("A legfrissebb események %s területéről és a %s megyei régióból.", name, county),
			BgColor: "#e3f2fd",
		},
	}

	if nameRo != "" {
		filteredNews = append(filteredNews, NewsItem{
			Title:   fmt.Sprintf("Noutăți din %s", nameRo),
			Link:    "#",
			Source:  "Lamsza Info",
			Date:    "Azi",
			Summary: fmt.Sprintf("Cele mai importante știri din regiunea %s.", nameRo),
			BgColor: "#f1f8e9",
		})
	}

	log.Printf("Filtered %d regional news items for: %s\n", len(filteredNews), name)
	json.NewEncoder(w).Encode(filteredNews)
}

func handleWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	slug := r.URL.Query().Get("slug")
	apiKey := r.URL.Query().Get("appid")

	if slug == "" || apiKey == "" {
		http.Error(w, "Missing slug or appid", http.StatusBadRequest)
		return
	}

	// 1. Resolve slug to a searchable location name (prefer name or name_ro)
	var name, nameRo string
	// Explicit check for known Szekely city names if slug is just the slugified version
	err := db.QueryRow("SELECT name, COALESCE(name_ro, '') FROM locations WHERE slug = $1", slug).Scan(&name, &nameRo)
	if err != nil {
		log.Println("Weather slug resolution error:", err)
		// Try fuzzy name match if slug doesn't match
		err = db.QueryRow("SELECT name, COALESCE(name_ro, '') FROM locations WHERE unaccent(name) ILIKE $1 OR unaccent(name_ro) ILIKE $1 LIMIT 1", slug).Scan(&name, &nameRo)
		if err != nil {
			name = strings.Title(strings.ReplaceAll(slug, "-", " "))
		}
	}

	searchName := name
	if nameRo != "" {
		// OpenWeatherMap often works better with the Romanian name for official lookups
		searchName = nameRo
	}

	// 2. Fetch from OpenWeatherMap
	weatherURL := "https://api.openweathermap.org/data/2.5/weather?q=" + url.QueryEscape(searchName) + "&appid=" + apiKey + "&units=metric&lang=hu"
	resp, err := http.Get(weatherURL)
	if err != nil {
		http.Error(w, "Weather provider error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func handleAutosuggest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	query := strings.ToLower(r.URL.Query().Get("q"))
	if len(query) < 3 {
		w.Write([]byte(`[]`))
		return
	}

	if db == nil {
		w.Write([]byte(`[]`))
		return
	}

	// SQL query to prioritize exact matches, prefix matches, and then fuzzy similarity
	// using pg_trgm for intelligent typo-tolerance.
	sqlQuery := `
		SELECT word FROM (
			SELECT word, priority, 
			       similarity(LOWER(word), LOWER($1)) as sim 
			FROM (
				SELECT name as word, 1 as priority FROM entry_categories
				UNION ALL
				SELECT name as word, 2 as priority FROM tags
				UNION ALL
				SELECT name as word, 3 as priority FROM locations
				UNION ALL
				SELECT name_ro as word, 3 as priority FROM locations WHERE name_ro IS NOT NULL
				UNION ALL
				SELECT name_de as word, 3 as priority FROM locations WHERE name_de IS NOT NULL
			) raw
			WHERE word != '' 
			  AND (
			      unaccent(LOWER(word)) LIKE '%' || unaccent(LOWER($1)) || '%' 
			      OR similarity(unaccent(LOWER(word)), unaccent(LOWER($1))) > 0.2
			  )
			GROUP BY word, priority
		) s
		ORDER BY 
			CASE 
				WHEN unaccent(LOWER(word)) = unaccent(LOWER($1)) THEN 1
				WHEN unaccent(LOWER(word)) LIKE unaccent(LOWER($1)) || '%' THEN 2
				ELSE 3
			END,
			sim DESC,
			priority ASC,
			word ASC
		LIMIT 10
	`

	rows, err := db.Query(sqlQuery, query)
	if err != nil {
		log.Println("Autosuggest query error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := []string{}
	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err == nil {
			results = append(results, word)
		}
	}

	if results == nil {
		results = []string{}
	}
	json.NewEncoder(w).Encode(results)
}

func main() {
	initDB()
	http.HandleFunc("/api/proxy", proxyHandler)
	http.HandleFunc("/api/entries", entriesHandler)
	http.HandleFunc("/api/directory", entriesHandler) // Alias for public frontend
	http.HandleFunc("/api/service", handleService)
	http.HandleFunc("/api/county_news", handleCountyNews)
	http.HandleFunc("/api/weather", handleWeather)
	http.HandleFunc("/api/autosuggest", handleAutosuggest)
	http.HandleFunc("/api/locations", handleAdminLocations) // Public GET access for locations list

	http.HandleFunc("/api/admin/mondasok", handleAdminMondasok)
	http.HandleFunc("/api/admin/quick_links", handleAdminQuickLinks)
	http.HandleFunc("/api/admin/news_feeds", handleAdminNewsFeeds)

	http.HandleFunc("/api/admin/entries", handleAdminEntries)
	http.HandleFunc("/api/admin/entry_categories", handleAdminEntryCategories)
	http.HandleFunc("/api/admin/entry_types", handleAdminEntryTypes)
	http.HandleFunc("/api/admin/locations", handleAdminLocations)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Backend API active on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
