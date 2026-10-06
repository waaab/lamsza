package links

import (
	"backend/internal/db"
	"backend/internal/models"
	"encoding/json"
	"net/http"
)

func HandlePublicQuickLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rows, err := db.DB.Query("SELECT id, title, url, COALESCE(bg_color, '#ffffff') FROM quick_links ORDER BY LOWER(title) ASC, id ASC")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var res []models.QuickLink
	for rows.Next() {
		var ql models.QuickLink
		if err := rows.Scan(&ql.ID, &ql.Title, &ql.URL, &ql.BgColor); err == nil {
			res = append(res, ql)
		}
	}
	if res == nil {
		res = []models.QuickLink{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
