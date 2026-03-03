package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

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

type Service struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	Notes    string `json:"notes"`
	URL      string `json:"url"`
	Tags     string `json:"tags"`
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
		}
	}
}

func szolgaltatasokHandler(w http.ResponseWriter, r *http.Request) {
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

	if db == nil {
		// Fallback if DB is down
		w.Write([]byte(`[]`))
		return
	}

	sqlQuery := `
		SELECT s.id, COALESCE(sc.name, s.category, ''), s.name, l.name as location, COALESCE(s.phone, ''), COALESCE(s.address, ''), COALESCE(s.notes, ''), COALESCE(s.url, ''), COALESCE(s.tags, '')
		FROM services s
		JOIN locations l ON s.location_id = l.id
		LEFT JOIN service_categories sc ON s.category_id = sc.id
		WHERE 1=1
	`
	args := []interface{}{}
	argId := 1

	if category != "" && category != "osszes" {
		sqlQuery += ` AND s.category = $` + javaToPg(argId)
		args = append(args, category)
		argId++
	}

	if county != "" {
		sqlQuery += ` AND LOWER(l.county) = $` + javaToPg(argId)
		args = append(args, county)
		argId++
	}

	if query != "" {
		likeQuery := "%" + query + "%"
		sqlQuery += ` AND (LOWER(s.name) LIKE $` + javaToPg(argId) + ` OR LOWER(l.name) LIKE $` + javaToPg(argId) + ` OR LOWER(s.notes) LIKE $` + javaToPg(argId) + ` OR LOWER(s.category) LIKE $` + javaToPg(argId) + ` OR LOWER(s.tags) LIKE $` + javaToPg(argId) + `)`
		args = append(args, likeQuery)
		argId++
	}

	rows, err := db.Query(sqlQuery, args...)
	if err != nil {
		log.Println("Query error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []Service
	for rows.Next() {
		var s Service
		if err := rows.Scan(&s.ID, &s.Category, &s.Name, &s.Location, &s.Phone, &s.Address, &s.Notes, &s.URL, &s.Tags); err == nil {
			results = append(results, s)
		}
	}

	if results == nil {
		results = []Service{} // ensure we output [] instead of null
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

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.Exec("DELETE FROM news_feeds WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

type Location struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	County string `json:"county"`
	Type   string `json:"type"`
}

func handleAdminLocations(w http.ResponseWriter, r *http.Request) {
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
		rows, err := db.Query("SELECT id, name, COALESCE(county, ''), COALESCE(type, '') FROM locations ORDER BY name ASC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []Location
		for rows.Next() {
			var loc Location
			if err := rows.Scan(&loc.ID, &loc.Name, &loc.County, &loc.Type); err == nil {
				res = append(res, loc)
			}
		}
		if res == nil {
			res = []Location{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var loc Location
		if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err := db.QueryRow("INSERT INTO locations (name, county, type) VALUES ($1, $2, $3) RETURNING id", loc.Name, loc.County, loc.Type).Scan(&loc.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(loc)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.Exec("DELETE FROM locations WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

type AdminService struct {
	ID               int    `json:"id"`
	LocationID       int    `json:"location_id"`
	CategoryID       *int   `json:"category_id"`
	Category         string `json:"category"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	Phone            string `json:"phone"`
	Address          string `json:"address"`
	Notes            string `json:"notes"`
	IsMagyarLanguage bool   `json:"is_magyar_language"`
	Tags             string `json:"tags"`
}

func handleAdminServices(w http.ResponseWriter, r *http.Request) {
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
		rows, err := db.Query("SELECT id, location_id, category_id, COALESCE(category, ''), name, COALESCE(url, ''), COALESCE(phone, ''), COALESCE(address, ''), COALESCE(notes, ''), is_magyar_language, COALESCE(tags, '') FROM services ORDER BY id DESC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []AdminService
		for rows.Next() {
			var s AdminService
			if err := rows.Scan(&s.ID, &s.LocationID, &s.CategoryID, &s.Category, &s.Name, &s.URL, &s.Phone, &s.Address, &s.Notes, &s.IsMagyarLanguage, &s.Tags); err == nil {
				res = append(res, s)
			}
		}
		if res == nil {
			res = []AdminService{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var s AdminService
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err := db.QueryRow("INSERT INTO services (location_id, category_id, category, name, url, phone, address, notes, is_magyar_language, tags) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id",
			s.LocationID, s.CategoryID, s.Category, s.Name, s.URL, s.Phone, s.Address, s.Notes, s.IsMagyarLanguage, s.Tags).Scan(&s.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(s)

	case "PUT":
		var s AdminService
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, err := db.Exec(`UPDATE services SET 
            location_id=$1, category_id=$2, category=$3, name=$4, url=$5, 
            phone=$6, address=$7, notes=$8, is_magyar_language=$9, tags=$10 
            WHERE id=$11`,
			s.LocationID, s.CategoryID, s.Category, s.Name, s.URL, s.Phone, s.Address, s.Notes, s.IsMagyarLanguage, s.Tags, s.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.Exec("DELETE FROM services WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

type ServiceCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func handleAdminServiceCategories(w http.ResponseWriter, r *http.Request) {
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
		rows, err := db.Query("SELECT id, name FROM service_categories ORDER BY name ASC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []ServiceCategory
		for rows.Next() {
			var sc ServiceCategory
			if err := rows.Scan(&sc.ID, &sc.Name); err == nil {
				res = append(res, sc)
			}
		}
		if res == nil {
			res = []ServiceCategory{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var sc ServiceCategory
		if err := json.NewDecoder(r.Body).Decode(&sc); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err := db.QueryRow("INSERT INTO service_categories (name) VALUES ($1) RETURNING id", sc.Name).Scan(&sc.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(sc)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.Exec("DELETE FROM service_categories WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
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

	// SQL query to prioritize prefix matches and focus on keywords (categories and tags)
	sqlQuery := `
		SELECT word FROM (
			SELECT word, MIN(priority) as min_priority FROM (
				SELECT name as word, 1 as priority FROM service_categories
				UNION ALL
				SELECT unnest(string_to_array(replace(replace(tags, ' ', ','), ',,', ','), ',')) as word, 2 as priority FROM services
				UNION ALL
				SELECT name as word, 3 as priority FROM locations
			) raw
			WHERE LOWER(word) LIKE '%' || $1 || '%' AND word != ''
			GROUP BY word
		) s
		ORDER BY 
			CASE 
				WHEN LOWER(word) = $1 THEN 1
				WHEN LOWER(word) LIKE $1 || '%' THEN 2
				ELSE 3
			END,
			min_priority,
			word
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
	http.HandleFunc("/api/szolgaltatasok", szolgaltatasokHandler)
	http.HandleFunc("/api/autosuggest", handleAutosuggest)
	http.HandleFunc("/api/admin/mondasok", handleAdminMondasok)
	http.HandleFunc("/api/admin/quick_links", handleAdminQuickLinks)
	http.HandleFunc("/api/admin/news_feeds", handleAdminNewsFeeds)

	http.HandleFunc("/api/admin/services", handleAdminServices)
	http.HandleFunc("/api/admin/service_categories", handleAdminServiceCategories)
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
