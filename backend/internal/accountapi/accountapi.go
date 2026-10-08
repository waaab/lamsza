// Package accountapi is the internal account API (/internal/account/): how
// Szótár's and Játszótér's backends read what a user chose on Lámsza, matched
// by Google ID (UI_BASELINE "acc-page"). Today that is the display name
// ("Megjelenített név"), which Lámsza owns and the other apps copy at sign-in
// and when their Fiók page opens.
//
// It follows the network's internal-API rules (WAYS_OF_WORKING R18): a path
// outside /api/, so production nginx never proxies it; loopback only, with no
// proxy header; a Bearer token per calling app, compared in constant time; no
// token configured means the API is off. The guard is Szótár's
// internal/adminapi guard with one token per caller.
package accountapi

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	"backend/internal/config"
	"backend/internal/db"
)

// Prefix is where the internal account API lives.
const Prefix = "/internal/account/"

// Guard lets a request through only when all of these hold:
//
//   - at least one caller token is configured (ACCOUNT_SZOTAR_TOKEN,
//     ACCOUNT_JATSZOTER_TOKEN); none means the API is off (404);
//   - the path is under Prefix (404 otherwise);
//   - it comes from loopback and carries no proxy header (403);
//   - it carries "Authorization: Bearer <token>" matching one of the
//     configured tokens, compared in constant time (401).
func Guard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokens := callerTokens()
		if len(tokens) == 0 || !strings.HasPrefix(r.URL.Path, Prefix) {
			http.NotFound(w, r)
			return
		}
		if forwarded(r) || !fromLoopback(r) {
			writeError(w, http.StatusForbidden, "Csak a gépen belülről érhető el.")
			return
		}
		got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || !matchesAny(strings.TrimSpace(got), tokens) {
			writeError(w, http.StatusUnauthorized, "Érvénytelen szolgáltatási token.")
			return
		}
		next(w, r)
	}
}

// HandleProfile answers GET /internal/account/profile?google_sub=… with what
// the other apps copy from Lámsza: {"display_name": "…"}. An empty name means
// the user has not chosen one (the apps then show Google's name). 404: no
// Lámsza account with that Google ID.
func HandleProfile(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != Prefix+"profile" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Csak GET.")
		return
	}
	sub := strings.TrimSpace(r.URL.Query().Get("google_sub"))
	if sub == "" {
		writeError(w, http.StatusBadRequest, "Hiányzik a google_sub.")
		return
	}
	var displayName string
	err := db.DB.QueryRow(`SELECT display_name FROM users WHERE google_sub = $1`, sub).Scan(&displayName)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "Nincs ilyen Lámsza-fiók.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Adatbázis hiba.")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{"display_name": displayName})
}

func callerTokens() []string {
	var tokens []string
	for _, t := range []string{config.AppConfig.AccountSzotarToken, config.AppConfig.AccountJatszoterToken} {
		if t != "" {
			tokens = append(tokens, t)
		}
	}
	return tokens
}

func matchesAny(got string, tokens []string) bool {
	ok := false
	for _, t := range tokens {
		if subtle.ConstantTimeCompare([]byte(got), []byte(t)) == 1 {
			ok = true
		}
	}
	return ok
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func forwarded(r *http.Request) bool {
	for _, h := range []string{"X-Forwarded-For", "X-Real-IP", "Forwarded"} {
		if r.Header.Get(h) != "" {
			return true
		}
	}
	return false
}

func fromLoopback(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
