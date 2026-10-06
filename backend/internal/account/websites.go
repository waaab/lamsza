package account

import (
	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/directory"
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
	Claimed     bool   `json:"claimed"`
}

type websiteListItem struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Domain      string   `json:"domain"`
	URL         string   `json:"url"`
	Claimed     bool     `json:"claimed"`
	CategoryID  int      `json:"category_id"`
	Category    string   `json:"category"`
	Categories  []string `json:"categories"`
	EntryID     int      `json:"entry_id"`
	EntrySlug   string   `json:"entry_slug,omitempty"`
}

type websitesListResponse struct {
	Websites []websiteListItem `json:"websites"`
}

type submitWebsiteBody struct {
	Domain      string `json:"domain"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CategoryID  int    `json:"category_id"`
	CategoryIDs []int  `json:"category_ids"`
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
			claimed                 bool
		)
		err = db.DB.QueryRow(`
			SELECT w.id, w.title, w.description, w.domain_key, w.entry_id, e.published,
				EXISTS (
					SELECT 1 FROM entry_members m
					WHERE m.entry_id = w.entry_id AND m.role = 'owner' AND m.status = 'active'
				)
			FROM websites w
			LEFT JOIN entries e ON e.id = w.entry_id
			WHERE w.domain_key = $1 AND w.status = 'approved'
		`, domainKey).Scan(&id, &title, &description, &key, &entryID, &published, &claimed)
		if err == nil {
			domainQuery = true
			hits = append(hits, WebsiteHit{
				ID:          id,
				Title:       title,
				Description: description,
				Domain:      key,
				URL:         "https://" + key,
				Claimed:     claimed,
			})
			if entryID.Valid && published.Valid && published.Bool {
				publishedEntryID = int(entryID.Int64)
			}
			return hits, domainQuery, publishedEntryID
		}
	}

	pattern := "%" + strings.ToLower(q) + "%"
	rows, err := db.DB.Query(`
		SELECT w.id, w.title, w.description, w.domain_key,
			EXISTS (
				SELECT 1 FROM entry_members m
				WHERE m.entry_id = w.entry_id AND m.role = 'owner' AND m.status = 'active'
			)
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
		if err := rows.Scan(&hit.ID, &hit.Title, &hit.Description, &hit.Domain, &hit.Claimed); err != nil {
			continue
		}
		hit.URL = "https://" + hit.Domain
		hits = append(hits, hit)
	}
	return hits, false, 0
}

type accountWebsiteItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	Claimed     bool   `json:"claimed"`
}

func HandleAccountWebsites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u, err := auth.UserFromRequest(r)
	if err != nil || u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	rows, err := db.DB.Query(`
		SELECT w.id, w.title, w.description, w.domain_key, w.status,
			EXISTS (
				SELECT 1 FROM entry_members m
				WHERE m.entry_id = w.entry_id AND m.role = 'owner' AND m.status = 'active'
			)
		FROM websites w
		WHERE w.user_id = $1
		ORDER BY w.created_at DESC, w.id DESC
	`, u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	items := []accountWebsiteItem{}
	for rows.Next() {
		var item accountWebsiteItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Domain, &item.Status, &item.Claimed); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		item.URL = "https://" + item.Domain
		items = append(items, item)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"websites": items})
}

type websiteLookup struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	Claimed     bool   `json:"claimed"`
	EntryID     int    `json:"entry_id"`
	Published   bool   `json:"published"`
	Membership  string `json:"membership"`
}

func HandleWebsiteLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u, err := auth.UserFromRequest(r)
	if err != nil || u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	domainKey, err := webdomain.CanonicalDomain(r.URL.Query().Get("url"))
	if err != nil || domainKey == "" {
		json.NewEncoder(w).Encode(map[string]any{"website": nil})
		return
	}

	var item websiteLookup
	var entryID sql.NullInt64
	var published sql.NullBool
	var membership sql.NullString
	err = db.DB.QueryRow(`
		SELECT w.id, w.title, w.description, w.domain_key, w.status, w.entry_id, e.published,
			EXISTS (
				SELECT 1 FROM entry_members m
				WHERE m.entry_id = w.entry_id AND m.role = 'owner' AND m.status = 'active'
			),
			(
				SELECT m.role || ':' || m.status
				FROM entry_members m
				WHERE m.entry_id = w.entry_id AND m.user_id = $2
				LIMIT 1
			)
		FROM websites w
		LEFT JOIN entries e ON e.id = w.entry_id
		WHERE w.domain_key = $1
	`, domainKey, u.ID).Scan(
		&item.ID, &item.Title, &item.Description, &item.Domain, &item.Status,
		&entryID, &published, &item.Claimed, &membership,
	)
	if err == sql.ErrNoRows {
		json.NewEncoder(w).Encode(map[string]any{"website": nil})
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if entryID.Valid {
		item.EntryID = int(entryID.Int64)
	}
	item.Published = published.Valid && published.Bool
	item.URL = "https://" + item.Domain
	if membership.Valid {
		item.Membership = membership.String
	}
	json.NewEncoder(w).Encode(map[string]any{"website": item})
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
		SELECT w.id, w.title, w.description, w.domain_key, COALESCE(w.category_id, 0), COALESCE(c.name, ''),
			COALESCE(w.entry_id, 0), COALESCE(e.slug, ''),
			EXISTS (
				SELECT 1 FROM entry_members m
				WHERE m.entry_id = w.entry_id AND m.role = 'owner' AND m.status = 'active'
			)
		FROM websites w
		LEFT JOIN entry_categories c ON c.id = w.category_id
		LEFT JOIN entries e ON e.id = w.entry_id AND e.published = true
		WHERE w.status = 'approved'
		ORDER BY w.title ASC, w.id ASC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item websiteListItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Domain, &item.CategoryID, &item.Category, &item.EntryID, &item.EntrySlug, &item.Claimed); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		item.URL = "https://" + item.Domain
		item.Categories = []string{}
		resp.Websites = append(resp.Websites, item)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ids := make([]int, 0, len(resp.Websites))
	for _, item := range resp.Websites {
		ids = append(ids, item.ID)
	}
	names, err := directory.NamesForWebsites(db.DB, ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for i := range resp.Websites {
		got := names[resp.Websites[i].ID]
		if len(got) == 0 && resp.Websites[i].Category != "" {
			got = []string{resp.Websites[i].Category}
		}
		resp.Websites[i].Categories = got
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
	categoryIDs, err := directory.NormalizeCategoryIDs(body.CategoryID, body.CategoryIDs)
	if err != nil {
		writeWebsiteFieldError(w, "category_id")
		return
	}
	if _, err = directory.ValidateLeaves(db.DB, categoryIDs); err != nil {
		writeWebsiteFieldError(w, "category_id")
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
		INSERT INTO websites (domain_key, submitted_host, title, description, status, user_id, category_id)
		VALUES ($1, $2, $3, $4, 'pending', $5, $6)
		RETURNING id
	`, domainKey, submittedHost, title, description, u.ID, body.CategoryID).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = directory.ReplaceWebsiteCategories(db.DB, id, categoryIDs); err != nil {
		writeWebsiteFieldError(w, "category_id")
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
	ID          int    `json:"id"`
	Action      string `json:"action"`
	CategoryID  int    `json:"category_id"`
	CategoryIDs []int  `json:"category_ids"`
}
