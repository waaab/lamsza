package weather

import (
	"backend/internal/clock"
	"backend/internal/config"
	"backend/internal/db"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// These tests write the weather tables of the scratch database that
// `npm run test:go` builds (TEST_DATABASE_URL, a *_test name; R8). The
// providers are a local httptest server: no test calls the internet.

func init() {
	config.Load()
	testURL, err := db.TestDatabaseURL()
	if err != nil {
		log.Fatal(err)
	}
	config.AppConfig.DatabaseURL = testURL
	db.InitDB()
	MigrateWeatherTranslations()
	MigrateWeather()
	workerSpacing = 0
}

// fakeProviders serves recorded answers and counts the calls.
type fakeProviders struct {
	mu        sync.Mutex
	calls     map[string]int
	status    map[string]int // per provider: 0 means 200
	lastIMS   string
	lastUA    string
	lastQuery string
}

func (f *fakeProviders) hit(name string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls[name]++
	return f.status[name]
}

func (f *fakeProviders) count(name string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[name]
}

func startFakeProviders(t *testing.T) *fakeProviders {
	t.Helper()
	f := &fakeProviders{calls: map[string]int{}, status: map[string]int{}}
	serve := func(name, file string, metHeaders bool) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			code := f.hit(name)
			if metHeaders {
				f.mu.Lock()
				f.lastIMS, f.lastUA, f.lastQuery = r.Header.Get("If-Modified-Since"), r.Header.Get("User-Agent"), r.URL.RawQuery
				f.mu.Unlock()
				w.Header().Set("Last-Modified", "Thu, 08 Oct 2026 21:11:26 GMT")
				w.Header().Set("Expires", clock.Now().Add(40*time.Minute).UTC().Format(http.TimeFormat))
				if code == 0 && r.Header.Get("If-Modified-Since") == "Thu, 08 Oct 2026 21:11:26 GMT" {
					code = http.StatusNotModified
				}
			}
			if code != 0 && code != http.StatusOK {
				w.WriteHeader(code)
				return
			}
			b, err := os.ReadFile("testdata/" + file)
			if err != nil {
				t.Error(err)
			}
			w.Write(b)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/weatherapi/locationforecast/2.0/complete", serve("metno", "metno_complete.json", true))
	mux.HandleFunc("/weatherapi/sunrise/3.0/sun", serve("sun", "metno_sun.json", false))
	mux.HandleFunc("/weatherapi/sunrise/3.0/moon", serve("moon", "metno_moon.json", false))
	mux.HandleFunc("/v1/forecast.json", serve("weatherapi", "weatherapi_forecast.json", false))
	mux.HandleFunc("/data/2.5/forecast", serve("owm", "owm_forecast.json", false))
	srv := httptest.NewServer(mux)

	oldMet, oldWA, oldOWM := metnoBaseURL, weatherAPIBaseURL, owmBaseURL
	oldKeys := [2]string{config.AppConfig.WeatherAPIComKey, config.AppConfig.WeatherAPIKey}
	metnoBaseURL, weatherAPIBaseURL, owmBaseURL = srv.URL+"/weatherapi", srv.URL+"/v1", srv.URL
	config.AppConfig.WeatherAPIComKey, config.AppConfig.WeatherAPIKey = "test-key", "test-key"
	t.Cleanup(func() {
		srv.Close()
		metnoBaseURL, weatherAPIBaseURL, owmBaseURL = oldMet, oldWA, oldOWM
		config.AppConfig.WeatherAPIComKey, config.AppConfig.WeatherAPIKey = oldKeys[0], oldKeys[1]
	})
	return f
}

// atTime stands the clock inside the recorded forecast.
func atTime(t *testing.T, s string) time.Time {
	t.Helper()
	ts := utc(t, s)
	prev := clock.Now
	clock.Now = func() time.Time { return ts }
	t.Cleanup(func() { clock.Now = prev })
	return ts
}

// testPlace is Csíkszereda from the reference data, with coordinates.
func testPlace(t *testing.T) Place {
	t.Helper()
	p, err := settlementBySlug("csikszereda")
	if err != nil {
		t.Fatalf("reference settlement csikszereda: %v", err)
	}
	lat, lon := 46.3593, 25.8017
	p.Latitude, p.Longitude, p.HasLocation = &lat, &lon, true
	clearWeather(t, p)
	t.Cleanup(func() { clearWeather(t, p) })
	return p
}

func clearWeather(t *testing.T, p Place) {
	t.Helper()
	for _, q := range []string{
		`DELETE FROM weather_forecast_cache WHERE place_kind = $1 AND place_id = $2`,
		`DELETE FROM weather_astro_daily WHERE place_kind = $1 AND place_id = $2`,
	} {
		if _, err := db.DB.Exec(q, p.Kind, p.ID); err != nil {
			t.Fatal(err)
		}
	}
	for _, q := range []string{
		`DELETE FROM weather_obs_hourly WHERE settlement_id = $1`,
		`DELETE FROM weather_daily_archive WHERE settlement_id = $1`,
	} {
		if _, err := db.DB.Exec(q, p.ID); err != nil {
			t.Fatal(err)
		}
	}
}

func allProviders() []string {
	return []string{providerMETNo, providerWeatherAPICom, providerOpenWeatherMap}
}

func mustCache(t *testing.T, p Place) *cacheRow {
	t.Helper()
	row, err := readCache(p.Kind, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	return row
}

func newWorker() *worker { return &worker{geocodeFailed: map[string]time.Time{}} }

func TestWorkerStoresMETForecastObsAndAstro(t *testing.T) {
	f := startFakeProviders(t)
	atTime(t, "2026-10-08T21:10:00Z")
	p := testPlace(t)

	newWorker().refresh(p, allProviders())

	row := mustCache(t, p)
	if row.Source != providerMETNo || row.Forecast == nil || row.CoordSource != coordsStored {
		t.Fatalf("cache %+v", row)
	}
	if row.ExpiresAt == nil || row.ExpiresAt.Sub(clock.Now()) < 30*time.Minute {
		t.Fatalf("expires %v: MET's Expires header was not used", row.ExpiresAt)
	}
	if f.count("weatherapi")+f.count("owm") != 0 {
		t.Fatal("a fallback was called although MET answered")
	}
	if !strings.Contains(f.lastUA, "lamsza") {
		t.Fatalf("User-Agent %q does not identify us", f.lastUA)
	}
	if !strings.Contains(f.lastQuery, "lat=46.3593") || !strings.Contains(f.lastQuery, "lon=25.8017") {
		t.Fatalf("query %q", f.lastQuery)
	}
	obs, err := readObs(p.ID, utc(t, "2026-10-08T21:00:00Z"), utc(t, "2026-10-08T22:00:00Z"))
	if err != nil || len(obs) != 1 || *obs[0].Temp != 10.3 {
		t.Fatalf("obs %+v %v", obs, err)
	}
	if _, err := readAstro(p.Kind, p.ID, "2026-10-09"); err != nil {
		t.Fatalf("astro for the Bucharest day: %v", err)
	}
}

func TestWorkerAsksMETWithIfModifiedSince(t *testing.T) {
	f := startFakeProviders(t)
	now := atTime(t, "2026-10-08T21:10:00Z")
	p := testPlace(t)
	w := newWorker()
	w.refresh(p, allProviders())
	before := mustCache(t, p)

	clock.Now = func() time.Time { return now.Add(2 * time.Hour) }
	w.refresh(p, allProviders())

	if f.lastIMS != "Thu, 08 Oct 2026 21:11:26 GMT" {
		t.Fatalf("If-Modified-Since %q", f.lastIMS)
	}
	after := mustCache(t, p)
	if !after.FetchedAt.Equal(*before.FetchedAt) {
		t.Fatal("a 304 rewrote the forecast")
	}
	if !after.ExpiresAt.After(*before.ExpiresAt) {
		t.Fatal("a 304 did not move the next refresh")
	}
}

func TestWorkerKeepsARecentForecastWhenMETFails(t *testing.T) {
	f := startFakeProviders(t)
	now := atTime(t, "2026-10-08T21:10:00Z")
	p := testPlace(t)
	w := newWorker()
	w.refresh(p, allProviders())

	f.status["metno"] = http.StatusInternalServerError
	clock.Now = func() time.Time { return now.Add(time.Hour) } // expired, within the grace
	w.refresh(p, allProviders())

	row := mustCache(t, p)
	if row.Source != providerMETNo {
		t.Fatalf("source %q: a short MET outage replaced its forecast", row.Source)
	}
	if f.count("weatherapi")+f.count("owm") != 0 {
		t.Fatal("a fallback was called during the grace period")
	}
	if row.ExpiresAt.Sub(clock.Now()) != retryAfterFailure {
		t.Fatalf("next try in %v", row.ExpiresAt.Sub(clock.Now()))
	}
}

func TestWorkerFallsBackInOrderWhenMETIsDownTooLong(t *testing.T) {
	f := startFakeProviders(t)
	now := atTime(t, "2026-10-08T21:10:00Z")
	p := testPlace(t)
	w := newWorker()
	w.refresh(p, allProviders())

	f.status["metno"] = http.StatusServiceUnavailable
	clock.Now = func() time.Time { return now.Add(5 * time.Hour) }
	w.refresh(p, allProviders())
	if row := mustCache(t, p); row.Source != providerWeatherAPICom {
		t.Fatalf("source %q, want WeatherAPI.com", row.Source)
	}
	if f.count("owm") != 0 {
		t.Fatal("OpenWeatherMap was called although WeatherAPI.com answered")
	}

	f.status["weatherapi"] = http.StatusInternalServerError
	clock.Now = func() time.Time { return now.Add(10 * time.Hour) }
	w.refresh(p, allProviders())
	if row := mustCache(t, p); row.Source != providerOpenWeatherMap {
		t.Fatalf("source %q, want OpenWeatherMap", row.Source)
	}
}

// The archive is MET only: a fallback forecast never becomes a reading.
func TestFallbackDataNeverReachesTheArchive(t *testing.T) {
	f := startFakeProviders(t)
	atTime(t, "2026-10-09T00:10:00Z")
	p := testPlace(t)
	f.status["metno"] = http.StatusInternalServerError

	newWorker().refresh(p, allProviders())

	if row := mustCache(t, p); row.Source != providerWeatherAPICom {
		t.Fatalf("source %q", row.Source)
	}
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM weather_obs_hourly WHERE settlement_id = $1`, p.ID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("%d hourly rows from a fallback", n)
	}
}

func TestArchiveYesterdayFoldsTheHours(t *testing.T) {
	atTime(t, "2026-10-10T08:00:00Z")
	p := testPlace(t)
	start := utc(t, "2026-10-08T21:00:00Z") // 2026-10-09 00:00 in Bucharest
	for h := 0; h < 24; h++ {
		s := Step{Time: start.Add(time.Duration(h) * time.Hour), Hours: 1, Symbol: "cloudy",
			Temp: fp(float64(h)), PrecipMm: fp(0.5)}
		if err := saveObs(p.ID, s); err != nil {
			t.Fatal(err)
		}
	}
	newWorker().archiveYesterday([]Place{p})

	days, err := readDailyArchive(p.ID, "2026-10-09", "2026-10-09")
	if err != nil || len(days) != 1 {
		t.Fatalf("days %+v %v", days, err)
	}
	d := days[0]
	if *d.TMin != 0 || *d.TMax != 23 || !near(d.PrecipMm, 12) || d.Hours != 24 || d.Symbol != "cloudy" {
		t.Fatalf("day %+v", d)
	}
}

func get(t *testing.T, h http.HandlerFunc, path string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest(http.MethodGet, path, nil))
	var body map[string]any
	if strings.HasPrefix(strings.TrimSpace(rr.Body.String()), "{") {
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
	}
	return rr, body
}

func TestHandlersReadOnlyTheCache(t *testing.T) {
	f := startFakeProviders(t)
	atTime(t, "2026-10-09T06:30:00Z")
	p := testPlace(t)

	rr, _ := get(t, HandleWeather, "/api/weather?slug=csikszereda")
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("no cache yet: got %d", rr.Code)
	}
	rr, _ = get(t, HandleForecast, "/api/weather/forecast?slug=csikszereda")
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("forecast with no cache: got %d", rr.Code)
	}
	if f.count("metno") != 0 {
		t.Fatal("a GET called a provider for a place the worker keeps")
	}

	newWorker().refresh(p, allProviders())
	calls := f.count("metno")

	rr, w := get(t, HandleWeather, "/api/weather?slug=csikszereda")
	if rr.Code != http.StatusOK {
		t.Fatalf("weather: %d %s", rr.Code, rr.Body)
	}
	if w["source"] != "MET Norway" || w["symbol"] == "" || w["temp_min"] == nil || w["temp_max"] == nil {
		t.Fatalf("weather %v", w)
	}

	rr, fc := get(t, HandleForecast, "/api/weather/forecast?slug=csikszereda")
	if rr.Code != http.StatusOK {
		t.Fatalf("forecast: %d %s", rr.Code, rr.Body)
	}
	daily, _ := fc["daily"].([]any)
	hourly, _ := fc["hourly"].([]any)
	if len(daily) < 7 || len(hourly) < 40 || fc["astro"] == nil || fc["source_name"] != "MET Norway" {
		t.Fatalf("forecast: %d days, %d hours, astro %v", len(daily), len(hourly), fc["astro"])
	}
	for _, d := range daily[1:] {
		if h := d.(map[string]any)["hours"].(float64); h < minDayHours {
			t.Fatalf("a %v-hour day is listed: %v", h, d)
		}
	}
	first := daily[0].(map[string]any)
	if first["date"] != "2026-10-09" || first["desc"] == "" {
		t.Fatalf("first day %v", first)
	}

	rr = httptest.NewRecorder()
	HandlePlaces(rr, httptest.NewRequest(http.MethodGet, "/api/weather/places", nil))
	var places []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &places); err != nil || len(places) == 0 {
		t.Fatalf("places: %v %s", err, rr.Body)
	}
	var withNow int
	for _, pl := range places {
		if pl["now"] != nil {
			withNow++
			if pl["slug"] != "csikszereda" {
				t.Fatalf("%v has weather although only csikszereda was refreshed", pl["slug"])
			}
		}
	}
	if withNow != 1 {
		t.Fatalf("%d places with weather, want 1", withNow)
	}

	rr, _ = get(t, HandleArchive, "/api/weather/archive?slug=csikszereda&days=7")
	if rr.Code != http.StatusOK {
		t.Fatalf("archive: %d", rr.Code)
	}
	rr, _ = get(t, HandleArchive, "/api/weather/archive?slug=csikszereda&date=2026-13-01")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("bad date: %d", rr.Code)
	}
	rr, _ = get(t, HandleForecast, "/api/weather/forecast?slug=no-such-place")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("unknown place: %d", rr.Code)
	}
	if f.count("metno") != calls {
		t.Fatal("a GET called MET")
	}
}
