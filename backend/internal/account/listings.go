package account

import (
	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/utils"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/lib/pq"
)

func migrateEntryMembers() {
	_, err := db.DB.Exec(`
		CREATE TABLE IF NOT EXISTS entry_members (
			entry_id INT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
			user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(16) NOT NULL,
			status VARCHAR(16) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (entry_id, user_id)
		)
	`)
	if err != nil {
		log.Printf("entry_members table: %v", err)
		return
	}
	_, err = db.DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS entry_members_one_active_owner
		ON entry_members (entry_id)
		WHERE role = 'owner' AND status = 'active'
	`)
	if err != nil {
		log.Printf("entry_members unique index: %v", err)
	}
}

type membershipResponse struct {
	EntryID int    `json:"entry_id"`
	Role    string `json:"role"`
	Status  string `json:"status"`
}

type claimBody struct {
	EntryID int `json:"entry_id"`
}

type createListingBody struct {
	Name           string          `json:"name"`
	LocationID     int             `json:"location_id"`
	CategoryID     int             `json:"category_id"`
	TypeID         int             `json:"type_id"`
	URL            string          `json:"url"`
	Phone          string          `json:"phone"`
	Address        string          `json:"address"`
	Notes          string          `json:"notes"`
	Languages      []string        `json:"languages"`
	Hours          json.RawMessage `json:"hours"`
	DeliveryHours  json.RawMessage `json:"delivery_hours"`
}

type listingItem struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Role       string `json:"role"`
	Status     string `json:"status"`
	Published  bool   `json:"published"`
	Verified   bool   `json:"verified"`
	LocationID int    `json:"location_id"`
	CategoryID int    `json:"category_id"`
	TypeID     int    `json:"type_id"`
}

type listingsResponse struct {
	Owned       []listingItem `json:"owned"`
	Member      []listingItem `json:"member"`
	Pending     []listingItem `json:"pending"`
	Unpublished []listingItem `json:"unpublished"`
}

func scanMembership(tx *sql.Tx, entryID, userID int) (membershipResponse, bool, error) {
	var resp membershipResponse
	err := tx.QueryRow(`
		SELECT entry_id, role, status
		FROM entry_members
		WHERE entry_id = $1 AND user_id = $2
	`, entryID, userID).Scan(&resp.EntryID, &resp.Role, &resp.Status)
	if err == sql.ErrNoRows {
		return resp, false, nil
	}
	if err != nil {
		return resp, false, err
	}
	return resp, true, nil
}

func hasActiveOwner(tx *sql.Tx, entryID int) (bool, error) {
	var exists bool
	err := tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM entry_members
			WHERE entry_id = $1 AND role = 'owner' AND status = 'active'
		)
	`, entryID).Scan(&exists)
	return exists, err
}

func insertMembership(tx *sql.Tx, entryID, userID int, role, status string) (membershipResponse, error) {
	resp := membershipResponse{EntryID: entryID, Role: role, Status: status}
	_, err := tx.Exec(`
		INSERT INTO entry_members (entry_id, user_id, role, status)
		VALUES ($1, $2, $3, $4)
	`, entryID, userID, role, status)
	return resp, err
}

func writeMembership(w http.ResponseWriter, resp membershipResponse) {
	json.NewEncoder(w).Encode(resp)
}

func HandleClaimListing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u, err := auth.UserFromRequest(r)
	if err != nil || u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body claimBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.EntryID <= 0 {
		http.Error(w, "entry_id required", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var published bool
	err = tx.QueryRow(`
		SELECT published FROM entries WHERE id = $1 FOR UPDATE
	`, body.EntryID).Scan(&published)
	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !published {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	existing, found, err := scanMembership(tx, body.EntryID, u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if found {
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeMembership(w, existing)
		return
	}

	ownerExists, err := hasActiveOwner(tx, body.EntryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var resp membershipResponse
	if !ownerExists {
		resp, err = insertMembership(tx, body.EntryID, u.ID, "owner", "active")
	} else {
		resp, err = insertMembership(tx, body.EntryID, u.ID, "member", "pending")
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeMembership(w, resp)
}

func HandleListings(w http.ResponseWriter, r *http.Request) {
	u, err := auth.UserFromRequest(r)
	if err != nil || u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	switch r.Method {
	case http.MethodGet:
		handleListListings(w, u.ID)
	case http.MethodPost:
		handleCreateListing(w, r, u.ID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleListListings(w http.ResponseWriter, userID int) {
	resp := listingsResponse{
		Owned:       []listingItem{},
		Member:      []listingItem{},
		Pending:     []listingItem{},
		Unpublished: []listingItem{},
	}
	rows, err := db.DB.Query(`
		SELECT e.id, e.name, COALESCE(e.slug, ''), m.role, m.status, e.published, COALESCE(e.verified, false),
			e.location_id, e.category_id, e.type_id
		FROM entry_members m
		JOIN entries e ON e.id = m.entry_id
		WHERE m.user_id = $1
		ORDER BY e.name ASC, e.id ASC
	`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item listingItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Role, &item.Status, &item.Published, &item.Verified,
			&item.LocationID, &item.CategoryID, &item.TypeID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		switch {
		case item.Role == "owner" && item.Status == "active" && !item.Published:
			resp.Unpublished = append(resp.Unpublished, item)
		case item.Role == "owner" && item.Status == "active":
			resp.Owned = append(resp.Owned, item)
		case item.Status == "pending":
			resp.Pending = append(resp.Pending, item)
		case item.Role == "member" && item.Status == "active":
			resp.Member = append(resp.Member, item)
		}
	}
	json.NewEncoder(w).Encode(resp)
}

func jsonObjectOrEmptyListing(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "{}"
	}
	return string(raw)
}

func handleCreateListing(w http.ResponseWriter, r *http.Request, userID int) {
	var body createListingBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Name == "" || body.LocationID <= 0 || body.CategoryID <= 0 || body.TypeID <= 0 {
		http.Error(w, "name, location_id, category_id, and type_id required", http.StatusBadRequest)
		return
	}
	if len(body.Languages) == 0 {
		body.Languages = []string{"HU"}
	}

	var catName string
	err := db.DB.QueryRow(`SELECT name FROM entry_categories WHERE id = $1`, body.CategoryID).Scan(&catName)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid category_id", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slug := utils.Slugify(body.Name)
	hours := jsonObjectOrEmptyListing(body.Hours)
	delivery := jsonObjectOrEmptyListing(body.DeliveryHours)

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var entryID int
	err = tx.QueryRow(`
		INSERT INTO entries (type_id, location_id, category_id, cat_name, name, slug, url, phone, address, notes, languages, verified, published, hours, delivery_hours, photos)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, false, false, $12::jsonb, $13::jsonb, '[]'::jsonb)
		RETURNING id
	`, body.TypeID, body.LocationID, body.CategoryID, catName, body.Name, slug, body.URL, body.Phone, body.Address, body.Notes, pq.Array(body.Languages), hours, delivery).Scan(&entryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`
		INSERT INTO entry_members (entry_id, user_id, role, status)
		VALUES ($1, $2, 'owner', 'active')
	`, entryID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          entryID,
		"name":        body.Name,
		"slug":        slug,
		"location_id": body.LocationID,
		"category_id": body.CategoryID,
		"type_id":     body.TypeID,
		"published":   false,
		"verified":    false,
		"role":        "owner",
		"status":      "active",
	})
}
