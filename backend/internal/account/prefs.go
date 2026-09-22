package account

import (
	"backend/internal/auth"
	"backend/internal/db"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func Migrate() {
	log.Println("Account preferences ready")
}

func NormalizeTheme(v string) (string, error) {
	switch v {
	case "light", "dark", "system":
		return v, nil
	default:
		return "", fmt.Errorf("invalid theme")
	}
}

func ClampSlots(n int) int {
	if n < 7 {
		return 7
	}
	if n > 14 {
		return 14
	}
	return n
}

type preferencesBody struct {
	Theme          *string `json:"theme"`
	QuicklinkSlots *int    `json:"quicklink_slots"`
}

type importBody struct {
	Theme          string          `json:"theme"`
	QuicklinkSlots *int            `json:"quicklink_slots"`
	Links          json.RawMessage `json:"links"`
	History        json.RawMessage `json:"history"`
}

func HandlePreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u, err := auth.UserFromRequest(r)
	if err != nil || u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body preferencesBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	setClauses := make([]string, 0, 2)
	args := make([]interface{}, 0, 3)
	argN := 1

	if body.Theme != nil {
		theme, err := NormalizeTheme(*body.Theme)
		if err != nil {
			http.Error(w, "invalid theme", http.StatusBadRequest)
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("theme = $%d", argN))
		args = append(args, theme)
		argN++
	}
	if body.QuicklinkSlots != nil {
		slots := *body.QuicklinkSlots
		if slots < 7 || slots > 14 {
			http.Error(w, "invalid quicklink_slots", http.StatusBadRequest)
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("quicklink_slots = $%d", argN))
		args = append(args, slots)
		argN++
	}

	if len(setClauses) == 0 {
		if err := auth.WriteMe(w, u.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	args = append(args, u.ID)
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", joinSetClauses(setClauses), argN)
	if _, err := db.DB.Exec(query, args...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := auth.WriteMe(w, u.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u, err := auth.UserFromRequest(r)
	if err != nil || u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var (
		theme           sql.NullString
		quicklinkSlots  sql.NullInt64
		prefsImportedAt sql.NullTime
	)
	err = db.DB.QueryRow(`
		SELECT theme, quicklink_slots, prefs_imported_at
		FROM users WHERE id = $1
	`, u.ID).Scan(&theme, &quicklinkSlots, &prefsImportedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if prefsImportedAt.Valid {
		if err := auth.WriteMe(w, u.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var body importBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	setClauses := make([]string, 0, 3)
	args := make([]interface{}, 0, 4)
	argN := 1

	if !theme.Valid {
		if normalized, err := NormalizeTheme(body.Theme); err == nil {
			setClauses = append(setClauses, fmt.Sprintf("theme = $%d", argN))
			args = append(args, normalized)
			argN++
		}
	}
	if !quicklinkSlots.Valid && body.QuicklinkSlots != nil {
		setClauses = append(setClauses, fmt.Sprintf("quicklink_slots = $%d", argN))
		args = append(args, ClampSlots(*body.QuicklinkSlots))
		argN++
	}

	setClauses = append(setClauses, "prefs_imported_at = NOW()")
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", joinSetClauses(setClauses), argN)
	args = append(args, u.ID)

	if _, err := db.DB.Exec(query, args...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := auth.WriteMe(w, u.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func joinSetClauses(clauses []string) string {
	out := clauses[0]
	for i := 1; i < len(clauses); i++ {
		out += ", " + clauses[i]
	}
	return out
}
