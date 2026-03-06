package mondasok

import (
	"backend/internal/db"
	"backend/internal/models"
	"encoding/json"
	"net/http"
)

func HandleAdminMondasok(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, err := db.DB.Query("SELECT id, text, created_at FROM mondasok ORDER BY id DESC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []models.Mondas
		for rows.Next() {
			var m models.Mondas
			if err := rows.Scan(&m.ID, &m.Text, &m.CreatedAt); err == nil {
				res = append(res, m)
			}
		}
		if res == nil {
			res = []models.Mondas{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var m models.Mondas
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err := db.DB.QueryRow("INSERT INTO mondasok (text) VALUES ($1) RETURNING id, created_at", m.Text).Scan(&m.ID, &m.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(m)

	case "PUT":
		var m models.Mondas
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, err := db.DB.Exec("UPDATE mondasok SET text=$1 WHERE id=$2", m.Text, m.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.DB.Exec("DELETE FROM mondasok WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}
