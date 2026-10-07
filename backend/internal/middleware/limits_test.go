package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func readAllHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 4096)
		for {
			_, err := r.Body.Read(buf)
			if err != nil {
				if err.Error() == "EOF" {
					w.WriteHeader(http.StatusOK)
					return
				}
				http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
				return
			}
		}
	})
}

func TestMaxBodyBytesForDefault(t *testing.T) {
	if got := MaxBodyBytesFor("/api/auth/google"); got != DefaultMaxBodyBytes {
		t.Fatalf("default limit = %d, want %d", got, DefaultMaxBodyBytes)
	}
}

func TestMaxBodyBytesForOverrides(t *testing.T) {
	cases := map[string]int64{
		"/api/account/import": ImportMaxBodyBytes,
		"/api/search":         DefaultMaxBodyBytes,
		// Uploads moved to lamsza-admin; a stray request here gets the default.
		"/api/admin/entry-images": DefaultMaxBodyBytes,
	}
	for path, want := range cases {
		if got := MaxBodyBytesFor(path); got != want {
			t.Errorf("%s limit = %d, want %d", path, got, want)
		}
	}
}

func TestLimitBodyRejectsOversizedBody(t *testing.T) {
	body := strings.Repeat("a", DefaultMaxBodyBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/google", strings.NewReader(body))
	rec := httptest.NewRecorder()

	LimitBody(readAllHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestLimitBodyAcceptsBodyAtTheLimit(t *testing.T) {
	body := strings.Repeat("a", DefaultMaxBodyBytes)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/google", strings.NewReader(body))
	rec := httptest.NewRecorder()

	LimitBody(readAllHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLimitBodyUsesThePerPathOverride(t *testing.T) {
	body := strings.Repeat("a", DefaultMaxBodyBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/api/account/import", strings.NewReader(body))
	rec := httptest.NewRecorder()

	LimitBody(readAllHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (override should allow it)", rec.Code, http.StatusOK)
	}
}
