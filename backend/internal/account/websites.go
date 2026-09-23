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

// WebsiteHit is a public website search result.
type WebsiteHit struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	URL         string `json:"url"`
}

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

func SearchWebsites(q string) (hits []WebsiteHit, domainQuery bool, publishedEntryID int) {
	hits = []WebsiteHit{}
	domainKey, err := webdomain.CanonicalDomain(q)
	if err == nil {
		var (
			id                      int
			title, description, key string
			entryID                 sql.NullInt64
			published               sql.NullBool
		)
		err = db.DB.QueryRow(`
			SELECT w.id, w.title, w.description, w.domain_key, w.entry_id, e.published
			FROM websites w
			LEFT JOIN entries e ON e.id = w.entry_id
			WHERE w.domain_key = $1 AND w.status = 'approved'
		`, domainKey).Scan(&id, &title, &description, &key, &entryID, &published)
		if err == nil {
			domainQuery = true
			hits = append(hits, WebsiteHit{
				ID:          id,
				Title:       title,
				Description: description,
				Domain:      key,
				URL:         "https://" + key,
			})
			if entryID.Valid && published.Valid && published.Bool {
				publishedEntryID = int(entryID.Int64)
			}
			return hits, domainQuery, publishedEntryID
		}
	}

	pattern := "%" + strings.ToLower(q) + "%"
	rows, err := db.DB.Query(`
		SELECT w.id, w.title, w.description, w.domain_key
		FROM websites w
		LEFT JOIN entries e ON e.id = w.entry_id
		WHERE w.status = 'approved'
		  AND (unaccent(lower(w.title)) ILIKE unaccent($1) OR unaccent(lower(w.description)) ILIKE unaccent($1))
		  AND (w.entry_id IS NULL OR e.published = false)
		ORDER BY w.title ASC, w.id ASC
		LIMIT 8
	`, pattern)
	if err != nil {
		return hits, false, 0
	}
	defer rows.Close()
	for rows.Next() {
		var hit WebsiteHit
		if err := rows.Scan(&hit.ID, &hit.Title, &hit.Description, &hit.Domain); err != nil {
			continue
		}
		hit.URL = "https://" + hit.Domain
		hits = append(hits, hit)
	}
	return hits, false, 0
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

	conflict, err := checkWebsiteDomainConflict(domainKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if conflict != nil {
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

func checkWebsiteDomainConflict(domainKey string) (map[string]interface{}, error) {
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
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	switch status {
	case "pending":
		return map[string]interface{}{"error": "domain_pending"}, nil
	case "approved":
		if entryID.Valid && published.Valid && published.Bool {
			return map[string]interface{}{
				"error": "domain_taken",
				"listing": map[string]string{
					"name": entryName.String,
					"slug": entrySlug.String,
				},
			}, nil
		}
		return map[string]interface{}{
			"error": "domain_taken",
			"website": map[string]string{
				"title":  title,
				"domain": domainKey,
			},
		}, nil
	default:
		return nil, nil
	}
}

type adminWebsiteBody struct {
	ID     int    `json:"id"`
	Action string `json:"action"`
}

func HandleAdminWebsite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body adminWebsiteBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.ID <= 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	switch body.Action {
	case "approve":
		res, err := db.DB.Exec(`
			UPDATE websites SET status = 'approved' WHERE id = $1 AND status = 'pending'
		`, body.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	case "reject":
		res, err := db.DB.Exec(`
			DELETE FROM websites WHERE id = $1 AND status = 'pending'
		`, body.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	case "ban":
		tx, err := db.DB.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var userID sql.NullInt64
		err = tx.QueryRow(`
			SELECT user_id FROM websites WHERE id = $1 AND status = 'pending'
		`, body.ID).Scan(&userID)
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !userID.Valid {
			http.Error(w, "user required", http.StatusBadRequest)
			return
		}

		res, err := tx.Exec(`DELETE FROM websites WHERE id = $1 AND status = 'pending'`, body.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		_, err = tx.Exec(`UPDATE users SET website_banned = true WHERE id = $1`, userID.Int64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "action must be approve, reject, or ban", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
