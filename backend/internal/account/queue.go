package account

import (
	"backend/internal/db"
	"encoding/json"
	"net/http"
)

type queueUnpublished struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	OwnerEmail string `json:"owner_email"`
}

type queueMember struct {
	EntryID   int    `json:"entry_id"`
	EntryName string `json:"entry_name"`
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
}

type queueResponse struct {
	Unpublished []queueUnpublished `json:"unpublished"`
	Members     []queueMember      `json:"members"`
}

func QueueCount() (int, error) {
	var count int
	err := db.DB.QueryRow(`
		SELECT
		  (SELECT COUNT(*) FROM entries WHERE published = false) +
		  (SELECT COUNT(*) FROM entry_members WHERE status = 'pending')
	`).Scan(&count)
	return count, err
}

func HandleListingQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := queueResponse{
		Unpublished: []queueUnpublished{},
		Members:     []queueMember{},
	}

	rows, err := db.DB.Query(`
		SELECT e.id, e.name, COALESCE(e.slug, ''), COALESCE(u.email, '')
		FROM entries e
		LEFT JOIN entry_members m ON m.entry_id = e.id AND m.role = 'owner' AND m.status = 'active'
		LEFT JOIN users u ON u.id = m.user_id
		WHERE e.published = false
		ORDER BY e.name ASC, e.id ASC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item queueUnpublished
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.OwnerEmail); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Unpublished = append(resp.Unpublished, item)
	}

	rows, err = db.DB.Query(`
		SELECT m.entry_id, e.name, m.user_id, u.email
		FROM entry_members m
		JOIN entries e ON e.id = m.entry_id
		JOIN users u ON u.id = m.user_id
		WHERE m.status = 'pending'
		ORDER BY e.name ASC, m.user_id ASC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item queueMember
		if err := rows.Scan(&item.EntryID, &item.EntryName, &item.UserID, &item.Email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Members = append(resp.Members, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type publishBody struct {
	EntryID int `json:"entry_id"`
}

func HandleListingQueuePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body publishBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.EntryID <= 0 {
		http.Error(w, "entry_id required", http.StatusBadRequest)
		return
	}

	res, err := db.DB.Exec(`UPDATE entries SET published = true WHERE id = $1 AND published = false`, body.EntryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		var exists bool
		err = db.DB.QueryRow(`SELECT EXISTS (SELECT 1 FROM entries WHERE id = $1)`, body.EntryID).Scan(&exists)
		if err != nil || !exists {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

type memberActionBody struct {
	EntryID int    `json:"entry_id"`
	UserID  int    `json:"user_id"`
	Action  string `json:"action"`
}

func HandleListingQueueMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body memberActionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.EntryID <= 0 || body.UserID <= 0 {
		http.Error(w, "entry_id and user_id required", http.StatusBadRequest)
		return
	}

	switch body.Action {
	case "approve":
		res, err := db.DB.Exec(`
			UPDATE entry_members SET status = 'active'
			WHERE entry_id = $1 AND user_id = $2 AND status = 'pending'
		`, body.EntryID, body.UserID)
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
			DELETE FROM entry_members
			WHERE entry_id = $1 AND user_id = $2 AND status = 'pending'
		`, body.EntryID, body.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	default:
		http.Error(w, "action must be approve or reject", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
