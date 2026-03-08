package search

import (
	"backend/internal/db"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

func HandleAutosuggest(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	if len(query) < 3 {
		w.Write([]byte(`[]`))
		return
	}

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

	rows, err := db.DB.Query(sqlQuery, query)
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

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "Missing url parameter", http.StatusBadRequest)
		return
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for name, values := range r.Header {
		if strings.ToLower(name) == "host" || strings.ToLower(name) == "origin" {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		lower := strings.ToLower(name)
		if lower == "access-control-allow-origin" ||
			lower == "access-control-allow-methods" ||
			lower == "access-control-allow-headers" {
			continue
		}
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
