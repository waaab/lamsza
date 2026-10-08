package weather

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWeatherClientHasATimeout(t *testing.T) {
	if weatherClient.Timeout <= 0 {
		t.Fatal("weatherClient has no timeout: a hanging provider would leak a goroutine and a socket per city")
	}
}

func TestWeatherHTTPGetGivesUpOnAHangingProvider(t *testing.T) {
	if testing.Short() {
		t.Skip("waits for the client timeout")
	}

	hang := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-hang
	}))
	defer func() {
		close(hang)
		srv.Close()
	}()

	start := time.Now()
	if _, err := weatherHTTPGet(srv.URL); err == nil {
		t.Fatal("expected a timeout error from a provider that never answers")
	}
	if elapsed := time.Since(start); elapsed > weatherClientTimeout+2*time.Second {
		t.Fatalf("gave up after %v, expected about %v", elapsed, weatherClientTimeout)
	}
}
