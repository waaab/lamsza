package accountapi

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/config"
	"backend/internal/db"
)

const (
	szotarToken    = "szotar-test-token"
	jatszoterToken = "jatszoter-test-token"
	testSub        = "accountapi-test-sub"
)

func init() {
	config.Load()
	testURL, err := db.TestDatabaseURL()
	if err != nil {
		log.Fatal(err)
	}
	config.AppConfig.DatabaseURL = testURL
	db.InitDB()
}

func withTokens(t *testing.T, szotar, jatszoter string) {
	t.Helper()
	old1, old2 := config.AppConfig.AccountSzotarToken, config.AppConfig.AccountJatszoterToken
	config.AppConfig.AccountSzotarToken, config.AppConfig.AccountJatszoterToken = szotar, jatszoter
	t.Cleanup(func() { config.AppConfig.AccountSzotarToken, config.AppConfig.AccountJatszoterToken = old1, old2 })
}

func request(path, remote, token string, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r.RemoteAddr = remote
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	Guard(HandleProfile)(w, r)
	return w
}

func TestGuardRefusals(t *testing.T) {
	path := Prefix + "profile?google_sub=" + testSub
	cases := []struct {
		name       string
		szotar     string
		jatszoter  string
		path       string
		remote     string
		token      string
		headers    map[string]string
		wantStatus int
	}{
		{"no token configured: the API is off", "", "", path, "127.0.0.1:5000", szotarToken, nil, http.StatusNotFound},
		{"outside the prefix", szotarToken, "", "/internal/other", "127.0.0.1:5000", szotarToken, nil, http.StatusNotFound},
		{"not from loopback", szotarToken, "", path, "203.0.113.7:5000", szotarToken, nil, http.StatusForbidden},
		{"came through nginx (X-Real-IP)", szotarToken, "", path, "127.0.0.1:5000", szotarToken, map[string]string{"X-Real-IP": "203.0.113.7"}, http.StatusForbidden},
		{"came through a proxy (X-Forwarded-For)", szotarToken, "", path, "127.0.0.1:5000", szotarToken, map[string]string{"X-Forwarded-For": "203.0.113.7"}, http.StatusForbidden},
		{"came through a proxy (Forwarded)", szotarToken, "", path, "127.0.0.1:5000", szotarToken, map[string]string{"Forwarded": "for=203.0.113.7"}, http.StatusForbidden},
		{"no bearer token", szotarToken, "", path, "127.0.0.1:5000", "", nil, http.StatusUnauthorized},
		{"wrong bearer token", szotarToken, jatszoterToken, path, "127.0.0.1:5000", "nope", nil, http.StatusUnauthorized},
		{"the other app's token when only Szótár's is set", szotarToken, "", path, "127.0.0.1:5000", jatszoterToken, nil, http.StatusUnauthorized},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			withTokens(t, c.szotar, c.jatszoter)
			if got := request(c.path, c.remote, c.token, c.headers).Code; got != c.wantStatus {
				t.Fatalf("status %d, want %d", got, c.wantStatus)
			}
		})
	}
}

func TestProfileReturnsTheDisplayNameByGoogleID(t *testing.T) {
	withTokens(t, szotarToken, jatszoterToken)
	if _, err := db.DB.Exec(`DELETE FROM users WHERE google_sub = $1`, testSub); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`INSERT INTO users (google_sub, email, name, display_name) VALUES ($1, 'accountapi@test.lamsza', 'Teszt Elek', 'Elek')`, testSub); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.DB.Exec(`DELETE FROM users WHERE google_sub = $1`, testSub) })

	for _, token := range []string{szotarToken, jatszoterToken} {
		w := request(Prefix+"profile?google_sub="+testSub, "[::1]:5000", token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status %d, body %s", w.Code, w.Body.String())
		}
		var body map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body["display_name"] != "Elek" {
			t.Fatalf("display_name = %q, want Elek", body["display_name"])
		}
		if len(body) != 1 {
			t.Fatalf("answer has %d fields, want only display_name: %v", len(body), body)
		}
	}

	if w := request(Prefix+"profile?google_sub=no-such-sub", "127.0.0.1:5000", szotarToken, nil); w.Code != http.StatusNotFound {
		t.Fatalf("unknown Google ID: status %d, want 404", w.Code)
	}
	if w := request(Prefix+"other?google_sub="+testSub, "127.0.0.1:5000", szotarToken, nil); w.Code != http.StatusNotFound {
		t.Fatalf("another path under the prefix: status %d, want 404", w.Code)
	}
	if w := request(Prefix+"profile", "127.0.0.1:5000", szotarToken, nil); w.Code != http.StatusBadRequest {
		t.Fatalf("no google_sub: status %d, want 400", w.Code)
	}
}
