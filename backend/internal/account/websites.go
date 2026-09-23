package account

import (
	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/webdomain"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

type websiteListItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	URL         string `json:"url"`
}

type websitesListResponse struct {
	Websites []websiteListItem `json:"websites"`
}

type submitWebsiteBody struct {
	Domain      string `json:"domain"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type submitWebsiteResponse struct {
	ID          int    `json:"id"`
	Domain      string `json:"domain"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func HandleWebsites(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleWebsitesList(w, r)
	case http.MethodPost:
		handleWebsiteSubmit(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleWebsitesList(w http.ResponseWriter, r *http.Request) {
	resp := websitesListResponse{Websites: []websiteListItem{}}
	rows, err := db.DB.Query(`
		SELECT id, title, description, domain_key
		FROM websites
		WHERE status = 'approved'
		ORDER BY title ASC, id ASC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item websiteListItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Domain); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		item.URL = "https://" + item.Domain
		resp.Websites = append(resp.Websites, item)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleWebsiteSubmit(w http.ResponseWriter, r *http.Request) {
	u, err := auth.UserFromRequest(r)
	if err != nil || u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var banned bool
	err = db.DB.QueryRow(`SELECT website_banned FROM users WHERE id = $1`, u.ID).Scan(&banned)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if banned {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "website_banned"})
		return
	}

	var body submitWebsiteBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	domainKey, err := webdomain.CanonicalDomain(body.Domain)
	if err != nil {
		writeWebsiteFieldError(w, "domain")
		return
	}
	title, err := webdomain.Plain(body.Title, 120)
	if err != nil {
		writeWebsiteFieldError(w, "title")
		return
	}
	description, err := webdomain.Plain(body.Description, 300)
	if err != nil {
		writeWebsiteFieldError(w, "description")
		return
	}
	submittedHost, err := submittedHost(body.Domain)
	if err != nil {
		writeWebsiteFieldError(w, "domain")
		return
	}

	if conflict := checkWebsiteDomainConflict(domainKey); conflict != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(conflict)
		return
	}

	var id int
	err = db.DB.QueryRow(`
		INSERT INTO websites (domain_key, submitted_host, title, description, status, user_id)
		VALUES ($1, $2, $3, $4, 'pending', $5)
		RETURNING id
	`, domainKey, submittedHost, title, description, u.ID).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(submitWebsiteResponse{
		ID:          id,
		Domain:      domainKey,
		Title:       title,
		Description: description,
		Status:      "pending",
	})
}

func writeWebsiteFieldError(w http.ResponseWriter, field string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"error": "invalid", "field": field})
}

func submittedHost(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", errors.New("empty domain")
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return "", errors.New("invalid domain")
	}
	return strings.ToLower(u.Hostname()), nil
}

func checkWebsiteDomainConflict(domainKey string) map[string]interface{} {
	var (
		status    string
		title     string
		entryID   sql.NullInt64
		entryName sql.NullString
		entrySlug sql.NullString
		published sql.NullBool
	)
	err := db.DB.QueryRow(`
		SELECT w.status, w.title, w.entry_id, e.name, e.slug, e.published
		FROM websites w
		LEFT JOIN entries e ON e.id = w.entry_id
		WHERE w.domain_key = $1
	`, domainKey).Scan(&status, &title, &entryID, &entryName, &entrySlug, &published)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return map[string]interface{}{"error": "internal"}
	}

	switch status {
	case "pending":
		return map[string]interface{}{"error": "domain_pending"}
	case "approved":
		if entryID.Valid && published.Valid && published.Bool {
			return map[string]interface{}{
				"error": "domain_taken",
				"listing": map[string]string{
					"name": entryName.String,
					"slug": entrySlug.String,
				},
			}
		}
		return map[string]interface{}{
			"error": "domain_taken",
			"website": map[string]string{
				"title":  title,
				"domain": domainKey,
			},
		}
	default:
		return nil
	}
}
