package account

import (
	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

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

type updateListingBody struct {
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
	Photos         json.RawMessage `json:"photos"`
	RatingsEnabled bool            `json:"ratings_enabled"`
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

func scanMembershipDB(entryID, userID int) (membershipResponse, bool, error) {
	var resp membershipResponse
	err := db.DB.QueryRow(`
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

func canEditListing(role, status string) bool {
	return status == "active" && (role == "owner" || role == "member")
}

func isActiveOwner(role, status string) bool {
	return role == "owner" && status == "active"
}

const maxListingPhotos = 24
const maxListingPhotoTextRunes = 500
const listingImageURLPrefix = "/api/media/entry-images/"
const defaultListingPhotoWidth = 1600
const defaultListingPhotoHeight = 1200

type listingPhoto struct {
	URL         string `json:"url"`
	Alt         string `json:"alt"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

func photosOrEmpty(raw json.RawMessage) string {
	return string(sanitizeListingPhotos(raw))
}

func sanitizeListingPhotos(raw json.RawMessage) json.RawMessage {
	raw = photosArrayOrEmpty(raw)
	var in []listingPhoto
	if err := json.Unmarshal(raw, &in); err != nil {
		return json.RawMessage(`[]`)
	}
	out := make([]listingPhoto, 0, len(in))
	for _, p := range in {
		if len(out) >= maxListingPhotos {
			break
		}
		url := strings.TrimSpace(p.URL)
		if !validListingPhotoURL(url) {
			continue
		}
		w, h := p.Width, p.Height
		if w < 1 || w > 10000 {
			w = defaultListingPhotoWidth
		}
		if h < 1 || h > 10000 {
			h = defaultListingPhotoHeight
		}
		out = append(out, listingPhoto{
			URL:         url,
			Alt:         clipListingPhotoRunes(strings.TrimSpace(p.Alt), maxListingPhotoTextRunes),
			Title:       clipListingPhotoRunes(strings.TrimSpace(p.Title), maxListingPhotoTextRunes),
			Description: clipListingPhotoRunes(strings.TrimSpace(p.Description), maxListingPhotoTextRunes),
			Width:       w,
			Height:      h,
		})
	}
	b, err := json.Marshal(out)
	if err != nil {
		return json.RawMessage(`[]`)
	}
	return json.RawMessage(b)
}

func photosArrayOrEmpty(b []byte) json.RawMessage {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" {
		return json.RawMessage(`[]`)
	}
	if !json.Valid(b) || !strings.HasPrefix(s, "[") {
		return json.RawMessage(`[]`)
	}
	return json.RawMessage(b)
}

func validListingURL(u string) bool {
	u = strings.TrimSpace(u)
	if u == "" {
		return true
	}
	lower := strings.ToLower(u)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func nextUniqueEntrySlug(tx *sql.Tx, base string) (string, error) {
	base = strings.TrimSpace(base)
	if base == "" {
		return "", fmt.Errorf("empty slug")
	}
	for i := 0; i < 100; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i+1)
		}
		var existingID int
		err := tx.QueryRow(`SELECT id FROM entries WHERE slug = $1`, candidate).Scan(&existingID)
		if err == sql.ErrNoRows {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("could not allocate unique slug")
}

func validListingPhotoURL(u string) bool {
	if u == "" {
		return false
	}
	lower := strings.ToLower(u)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") || strings.Contains(u, "://") && !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return false
	}
	if strings.HasPrefix(u, listingImageURLPrefix) {
		base := u[len(listingImageURLPrefix):]
		return base != "" && !strings.Contains(base, "/") && !strings.Contains(base, "..")
	}
	return strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "http://")
}

func clipListingPhotoRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max])
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
		if idStr := strings.TrimSpace(r.URL.Query().Get("id")); idStr != "" {
			handleGetListingDetail(w, r, u.ID, idStr)
		} else {
			handleListListings(w, u.ID)
		}
	case http.MethodPost:
		handleCreateListing(w, r, u.ID)
	case http.MethodPatch:
		handleUpdateListing(w, r, u.ID)
	case http.MethodDelete:
		handleDeleteListing(w, r, u.ID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type catalogItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type listingCatalogResponse struct {
	Categories []catalogItem `json:"categories"`
	Types      []catalogItem `json:"types"`
}

func HandleListingCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u, err := auth.UserFromRequest(r)
	if err != nil || u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp := listingCatalogResponse{
		Categories: []catalogItem{},
		Types:      []catalogItem{},
	}

	rows, err := db.DB.Query(`SELECT id, name FROM entry_categories ORDER BY name ASC, id ASC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		var item catalogItem
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			rows.Close()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Categories = append(resp.Categories, item)
	}
	rows.Close()

	rows, err = db.DB.Query(`SELECT id, name FROM entry_types ORDER BY name ASC, id ASC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		var item catalogItem
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			rows.Close()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Types = append(resp.Types, item)
	}
	rows.Close()

	json.NewEncoder(w).Encode(resp)
}

type listingDetailResponse struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	Slug           string          `json:"slug"`
	LocationID     int             `json:"location_id"`
	CategoryID     int             `json:"category_id"`
	TypeID         int             `json:"type_id"`
	URL            string          `json:"url"`
	Phone          string          `json:"phone"`
	Address         string          `json:"address"`
	Notes          string          `json:"notes"`
	Languages      []string        `json:"languages"`
	Hours          json.RawMessage `json:"hours"`
	DeliveryHours  json.RawMessage `json:"delivery_hours"`
	Photos         json.RawMessage `json:"photos"`
	Published      bool            `json:"published"`
	RatingsEnabled bool            `json:"ratings_enabled"`
}

func handleGetListingDetail(w http.ResponseWriter, r *http.Request, userID int, idStr string) {
	entryID, err := strconv.Atoi(idStr)
	if err != nil || entryID <= 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	membership, found, err := scanMembershipDB(entryID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found || !canEditListing(membership.Role, membership.Status) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var detail listingDetailResponse
	var pqLanguages []string
	var hours, delivery, photos []byte
	err = db.DB.QueryRow(`
		SELECT e.id, e.name, COALESCE(e.slug, ''), e.location_id, e.category_id, e.type_id,
			COALESCE(e.url, ''), COALESCE(e.phone, ''), COALESCE(e.address, ''), COALESCE(e.notes, ''),
			e.languages, COALESCE(e.hours, '{}'::jsonb), COALESCE(e.delivery_hours, '{}'::jsonb),
			COALESCE(e.photos, '[]'::jsonb), e.published, COALESCE(e.ratings_enabled, false)
		FROM entries e
		WHERE e.id = $1
	`, entryID).Scan(&detail.ID, &detail.Name, &detail.Slug, &detail.LocationID, &detail.CategoryID, &detail.TypeID,
		&detail.URL, &detail.Phone, &detail.Address, &detail.Notes, pq.Array(&pqLanguages),
		&hours, &delivery, &photos, &detail.Published, &detail.RatingsEnabled)
	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(pqLanguages) == 0 {
		detail.Languages = []string{"HU"}
	} else {
		detail.Languages = pqLanguages
	}
	detail.Hours = json.RawMessage(jsonObjectOrEmptyListing(hours))
	detail.DeliveryHours = json.RawMessage(jsonObjectOrEmptyListing(delivery))
	detail.Photos = sanitizeListingPhotos(photos)
	json.NewEncoder(w).Encode(detail)
}

type listingMemberRow struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

func HandleListingMembers(w http.ResponseWriter, r *http.Request) {
	u, err := auth.UserFromRequest(r)
	if err != nil || u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	switch r.Method {
	case http.MethodGet:
		handleListListingMembers(w, r, u.ID)
	case http.MethodDelete:
		handleDeleteListingMember(w, r, u.ID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleListListingMembers(w http.ResponseWriter, r *http.Request, ownerUserID int) {
	entryID, err := strconv.Atoi(r.URL.Query().Get("entry_id"))
	if err != nil || entryID <= 0 {
		http.Error(w, "entry_id required", http.StatusBadRequest)
		return
	}

	owner, found, err := scanMembershipDB(entryID, ownerUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found || !isActiveOwner(owner.Role, owner.Status) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	rows, err := db.DB.Query(`
		SELECT m.user_id, COALESCE(u.email, ''), m.role, m.status
		FROM entry_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.entry_id = $1 AND m.role = 'member' AND m.status = 'active'
		ORDER BY u.email ASC, m.user_id ASC
	`, entryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []listingMemberRow{}
	for rows.Next() {
		var row listingMemberRow
		if err := rows.Scan(&row.UserID, &row.Email, &row.Role, &row.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out = append(out, row)
	}
	json.NewEncoder(w).Encode(out)
}

func handleDeleteListingMember(w http.ResponseWriter, r *http.Request, ownerUserID int) {
	entryID, err := strconv.Atoi(r.URL.Query().Get("entry_id"))
	if err != nil || entryID <= 0 {
		http.Error(w, "entry_id required", http.StatusBadRequest)
		return
	}
	targetUserID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil || targetUserID <= 0 {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	owner, found, err := scanMembershipDB(entryID, ownerUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found || !isActiveOwner(owner.Role, owner.Status) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var targetRole string
	err = db.DB.QueryRow(`
		SELECT role FROM entry_members WHERE entry_id = $1 AND user_id = $2
	`, entryID, targetUserID).Scan(&targetRole)
	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if targetRole == "owner" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	_, err = db.DB.Exec(`DELETE FROM entry_members WHERE entry_id = $1 AND user_id = $2`, entryID, targetUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
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
	if !validListingURL(body.URL) {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
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

	baseSlug := utils.Slugify(body.Name)
	if baseSlug == "" {
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
	}
	hours := jsonObjectOrEmptyListing(body.Hours)
	delivery := jsonObjectOrEmptyListing(body.DeliveryHours)

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	slug, err := nextUniqueEntrySlug(tx, baseSlug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var entryID int
	err = tx.QueryRow(`
		INSERT INTO entries (type_id, location_id, category_id, cat_name, name, slug, url, phone, address, notes, languages, verified, published, hours, delivery_hours, photos, ratings_enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, false, false, $12::jsonb, $13::jsonb, '[]'::jsonb, false)
		RETURNING id
	`, body.TypeID, body.LocationID, body.CategoryID, catName, body.Name, slug, strings.TrimSpace(body.URL), body.Phone, body.Address, body.Notes, pq.Array(body.Languages), hours, delivery).Scan(&entryID)
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

func handleUpdateListing(w http.ResponseWriter, r *http.Request, userID int) {
	entryID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || entryID <= 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	membership, found, err := scanMembershipDB(entryID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found || !canEditListing(membership.Role, membership.Status) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var body updateListingBody
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
	if !validListingURL(body.URL) {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	var catName string
	err = db.DB.QueryRow(`SELECT name FROM entry_categories WHERE id = $1`, body.CategoryID).Scan(&catName)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid category_id", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hours := jsonObjectOrEmptyListing(body.Hours)
	delivery := jsonObjectOrEmptyListing(body.DeliveryHours)
	photos := photosOrEmpty(body.Photos)

	_, err = db.DB.Exec(`
		UPDATE entries SET
			type_id = $1, location_id = $2, category_id = $3, cat_name = $4, name = $5,
			url = $6, phone = $7, address = $8, notes = $9, languages = $10,
			hours = $11::jsonb, delivery_hours = $12::jsonb, photos = $13::jsonb,
			ratings_enabled = $14
		WHERE id = $15
	`, body.TypeID, body.LocationID, body.CategoryID, catName, body.Name,
		strings.TrimSpace(body.URL), body.Phone, body.Address, body.Notes, pq.Array(body.Languages),
		hours, delivery, photos, body.RatingsEnabled, entryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleDeleteListing(w http.ResponseWriter, r *http.Request, userID int) {
	entryID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || entryID <= 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	membership, found, err := scanMembershipDB(entryID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found || !isActiveOwner(membership.Role, membership.Status) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	_, err = db.DB.Exec(`DELETE FROM entries WHERE id = $1`, entryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
