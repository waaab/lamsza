package main

import (
	"backend/internal/auth"
	"backend/internal/db"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"
)

// The public site and the admin app (lamsza-admin, :3000) share this database.
// Until BOG-45 they also shared one cookie name and one `sessions` table, so a
// token minted by either one was accepted by the other. They are two stores
// now: `sessions` here, `admin_sessions` for admin. These tests hold the public
// half of that boundary - the admin half lives in
// lamsza-admin/backend/internal/auth/session_boundary_test.go.
//
// This repo owns the DDL for both tables because it owns this schema; the admin
// process runs no DDL. So an accidental revert here is what would silently
// re-merge the two stores, and that is what these tests catch.

func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func adminTestUserID(t *testing.T) int {
	t.Helper()
	var id int
	if err := db.DB.QueryRow(`SELECT id FROM users WHERE email = $1`, "admin@test.lamsza").Scan(&id); err != nil {
		t.Fatalf("admin test user: %v", err)
	}
	return id
}

// mintAdminSession writes a row the way the admin app does: into
// admin_sessions, never into sessions.
func mintAdminSession(t *testing.T) string {
	t.Helper()
	token := "bog45-admin-token-" + time.Now().Format("20060102150405.000000000")
	hash := hashSessionToken(token)
	_, err := db.DB.Exec(
		`INSERT INTO admin_sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		hash, adminTestUserID(t), time.Now().Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("insert admin session: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.DB.Exec(`DELETE FROM admin_sessions WHERE token_hash = $1`, hash)
	})
	return token
}

// TestAdminSessionIsRejectedByPublicAPI is the cross-host rejection, seen from
// the public side: a token the admin app minted buys nothing here, even though
// it belongs to a real allowlisted admin in the shared users table.
func TestAdminSessionIsRejectedByPublicAPI(t *testing.T) {
	token := mintAdminSession(t)
	cookie := &http.Cookie{Name: auth.SessionCookieName, Value: token}

	// This app routes no /api/admin/* any more (BOG-42), so the paths worth
	// probing are the signed-in surface it does still own.
	for _, path := range []string{"/api/auth/me", "/api/account/favorites", "/api/account/listings", "/api/account/websites"} {
		rr := doRequestWithCookie(t, http.MethodGet, path, nil, cookie)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("GET %s with an admin-minted token: got %d, want 401 (body %q)",
				path, rr.Code, rr.Body.String())
		}
	}
}

// TestPublicSessionNeverLandsInAdminSessions is the other direction, at the
// store level: a sign-in here must not write a row the admin API would accept.
func TestPublicSessionNeverLandsInAdminSessions(t *testing.T) {
	if testAllowlistedSession == nil {
		t.Fatal("no public session cookie from the test login")
	}
	hash := hashSessionToken(testAllowlistedSession.Value)

	var inPublic int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM sessions WHERE token_hash = $1`, hash).Scan(&inPublic); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if inPublic != 1 {
		t.Fatalf("public token in sessions: got %d rows, want 1", inPublic)
	}

	var inAdmin int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM admin_sessions WHERE token_hash = $1`, hash).Scan(&inAdmin); err != nil {
		t.Fatalf("count admin_sessions: %v", err)
	}
	if inAdmin != 0 {
		t.Errorf("a public sign-in wrote %d row(s) into admin_sessions; the admin API would accept them", inAdmin)
	}
}

// TestPublicSessionCookieNameIsUnchanged pins the name the admin app has to
// differ from. Renaming this constant to the admin one would merge the two
// cookies back together without any test in this repo failing otherwise.
func TestPublicSessionCookieNameIsUnchanged(t *testing.T) {
	if auth.SessionCookieName != "lamsza_session" {
		t.Fatalf("public session cookie = %q, want \"lamsza_session\" (admin uses \"lamsza_admin_session\")",
			auth.SessionCookieName)
	}
}

// TestAdminSessionsTableExists proves auth.Migrate() still creates the table
// the admin app depends on. Admin runs no DDL, so if this boot path loses the
// statement, admin sign-in breaks at runtime with nothing failing here.
func TestAdminSessionsTableExists(t *testing.T) {
	want := map[string]bool{"token_hash": false, "user_id": false, "expires_at": false, "created_at": false}
	rows, err := db.DB.Query(
		`SELECT column_name FROM information_schema.columns WHERE table_name = 'admin_sessions'`)
	if err != nil {
		t.Fatalf("read admin_sessions columns: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		if _, ok := want[name]; ok {
			want[name] = true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	for name, found := range want {
		if !found {
			t.Errorf("admin_sessions is missing column %q", name)
		}
	}
}
