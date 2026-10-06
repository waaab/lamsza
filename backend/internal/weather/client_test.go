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

// /api/weather/county runs in waves of maxWeatherFanOut. Budget plus the worst
// case for one city must stay under the server WriteTimeout (30s in main.go),
// or the last wave answers into a closed connection.
func TestWeatherFanOutFitsInTheServerWriteTimeout(t *testing.T) {
	const serverWriteTimeout = 30 * time.Second

	worstCase := weatherFanOutBudget + 3*weatherClientTimeout
	if worstCase >= serverWriteTimeout {
		t.Fatalf("worst case %v is not under the %v WriteTimeout", worstCase, serverWriteTimeout)
	}
	if maxWeatherFanOut <= 0 {
		t.Fatal("maxWeatherFanOut must cap the goroutine-per-city fan-out")
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
