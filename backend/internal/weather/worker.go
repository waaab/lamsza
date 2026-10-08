package weather

import (
	"backend/internal/clock"
	"backend/internal/db"
	"context"
	"errors"
	"log"
	"strings"
	"time"
)

// The worker is the only code that calls the weather providers on a
// schedule and the only code that writes the weather tables. A GET reads
// what it left (WAYS_OF_WORKING R14).
const (
	// workerTick is how often the worker looks for places that are due.
	workerTick = time.Minute
	// defaultTTL is used when MET sends no Expires, and for fallback data,
	// so MET is asked again soon.
	defaultTTL = 30 * time.Minute
	// minTTL keeps a bad Expires header from making us poll.
	minTTL = 5 * time.Minute
	// retryAfterFailure is the wait after every provider failed.
	retryAfterFailure = 10 * time.Minute
	// staleGrace is how long past its expiry a MET forecast is still shown
	// before the fallbacks are asked. A short MET outage changes nothing.
	staleGrace = 3 * time.Hour
	// geocodeRetry is the wait after a name could not be geocoded.
	geocodeRetry = 6 * time.Hour
)

// providerOrder returns the providers to try, in the fixed order MET
// Norway, WeatherAPI.com, OpenWeatherMap, minus any the admin switched off.
// MET is the only one that reaches a week ahead and the archive's only
// source; the other two are there for when it is down.
func providerOrder() []string {
	enabled := map[string]bool{
		providerMETNo:          getSettingBool("weather_provider_metno_enabled", true),
		providerWeatherAPICom:  getSettingBool("weather_provider_weatherapi_enabled", true),
		providerOpenWeatherMap: getSettingBool("weather_provider_openweathermap_enabled", true),
	}
	var order []string
	for _, p := range []string{providerMETNo, providerWeatherAPICom, providerOpenWeatherMap} {
		if enabled[p] {
			order = append(order, p)
		}
	}
	if len(order) == 0 {
		order = []string{providerMETNo}
	}
	return order
}

func getSetting(key, defaultVal string) (string, error) {
	var val string
	err := db.DB.QueryRow("SELECT value FROM site_settings WHERE key = $1", key).Scan(&val)
	if err != nil {
		return defaultVal, err
	}
	return val, nil
}

func getSettingBool(key string, defaultVal bool) bool {
	s, err := getSetting(key, "")
	if err != nil || s == "" {
		return defaultVal
	}
	return strings.ToLower(s) == "true" || s == "1"
}

// workerSpacing spreads requests out: MET asks for no bursts. A variable so
// tests can run without the pauses.
var workerSpacing = 500 * time.Millisecond

// StartWorker runs the refresh loop until ctx ends.
func StartWorker(ctx context.Context) {
	go func() {
		log.Println("Weather worker started")
		w := &worker{geocodeFailed: map[string]time.Time{}}
		for {
			w.runOnce()
			select {
			case <-ctx.Done():
				return
			case <-time.After(workerTick):
			}
		}
	}()
}

type worker struct {
	geocodeFailed map[string]time.Time
	archivedDay   string
}

func (w *worker) runOnce() {
	places, err := loadPlaces()
	if err != nil {
		log.Printf("weather worker: load places: %v", err)
		return
	}
	order := providerOrder()
	for _, p := range places {
		w.refresh(p, order)
	}
	w.archiveYesterday(places)
}

func placeKey(p Place) string { return p.Kind + ":" + p.Slug }

// coordsFor returns where to ask about a place: the admin's coordinates, a
// name we geocoded earlier, or a fresh geocode. ok is false if none.
func (w *worker) coordsFor(p Place, cached *cacheRow) (lat, lon float64, ok bool) {
	if p.HasLocation {
		if err := saveCoords(p.Kind, p.ID, *p.Latitude, *p.Longitude, coordsStored); err != nil {
			log.Printf("weather worker: save coords %s: %v", placeKey(p), err)
			return 0, 0, false
		}
		return *p.Latitude, *p.Longitude, true
	}
	if cached != nil && cached.CoordSource == coordsGeocoded {
		return cached.Latitude, cached.Longitude, true
	}
	if p.Kind != placeSettlement {
		return 0, 0, false
	}
	if t, failed := w.geocodeFailed[placeKey(p)]; failed && clock.Now().Before(t.Add(geocodeRetry)) {
		return 0, 0, false
	}
	name := p.NameRo
	if name == "" {
		name = p.Name
	}
	lat, lon, err := geocodeOWM(name, p.CountyRo)
	if err != nil {
		w.geocodeFailed[placeKey(p)] = clock.Now()
		log.Printf("weather worker: geocode %s (%q): %v; enter its coordinates in the admin app", placeKey(p), name, err)
		return 0, 0, false
	}
	time.Sleep(workerSpacing)
	if err := saveCoords(p.Kind, p.ID, lat, lon, coordsGeocoded); err != nil {
		log.Printf("weather worker: save coords %s: %v", placeKey(p), err)
		return 0, 0, false
	}
	return lat, lon, true
}

func (w *worker) refresh(p Place, order []string) {
	cached, err := readCache(p.Kind, p.ID)
	if err != nil && !errors.Is(err, errNoCache) {
		log.Printf("weather worker: read cache %s: %v", placeKey(p), err)
		return
	}
	if errors.Is(err, errNoCache) {
		cached = nil
	}
	lat, lon, ok := w.coordsFor(p, cached)
	if !ok {
		return
	}
	// Coordinates may have changed (saveCoords then clears the forecast).
	if cached, err = readCache(p.Kind, p.ID); err != nil {
		return
	}

	now := clock.Now()
	if cached.ExpiresAt == nil || !now.Before(*cached.ExpiresAt) {
		w.refreshForecast(p, lat, lon, cached, order, now)
		cached, _ = readCache(p.Kind, p.ID)
	}
	if cached != nil && cached.Source == providerMETNo && cached.Forecast != nil && p.Kind == placeSettlement {
		w.recordObs(p, cached.Forecast, now)
	}
	w.refreshAstro(p, lat, lon, now)
}

func (w *worker) refreshForecast(p Place, lat, lon float64, cached *cacheRow, order []string, now time.Time) {
	stillGood := cached.Forecast != nil && cached.ExpiresAt != nil && now.Before(cached.ExpiresAt.Add(staleGrace))
	metTried := false
	for _, prov := range order {
		// After a failed MET call, a fallback only replaces a forecast that
		// is too old to show: a short MET outage changes nothing.
		if prov != providerMETNo && metTried && stillGood {
			break
		}
		switch prov {
		case providerMETNo:
			metTried = true
			ims := ""
			if cached.Source == providerMETNo {
				ims = cached.LastModified
			}
			res, err := fetchMETNo(lat, lon, p.Elevation, ims)
			time.Sleep(workerSpacing)
			if err != nil {
				log.Printf("weather worker: met.no %s: %v", placeKey(p), err)
				continue
			}
			expires := clampExpires(res.Expires, now)
			if res.NotModified {
				if err := setExpires(p.Kind, p.ID, expires); err != nil {
					log.Printf("weather worker: %s: %v", placeKey(p), err)
				}
				return
			}
			if err := saveForecast(p.Kind, p.ID, res.Forecast, now, expires, res.LastModified); err != nil {
				log.Printf("weather worker: save %s: %v", placeKey(p), err)
			}
			return
		case providerWeatherAPICom, providerOpenWeatherMap:
			fetch := fetchWeatherAPICom
			if prov == providerOpenWeatherMap {
				fetch = fetchOpenWeatherMap
			}
			fc, err := fetch(lat, lon)
			if errors.Is(err, errNoKey) {
				continue
			}
			time.Sleep(workerSpacing)
			if err != nil {
				log.Printf("weather worker: %s %s: %v", prov, placeKey(p), err)
				continue
			}
			if err := saveForecast(p.Kind, p.ID, fc, now, now.Add(defaultTTL), ""); err != nil {
				log.Printf("weather worker: save %s: %v", placeKey(p), err)
			}
			log.Printf("weather worker: %s served by fallback %s", placeKey(p), prov)
			return
		}
	}
	// Nothing new: keep what we have and try again a little later.
	if err := setExpires(p.Kind, p.ID, now.Add(retryAfterFailure)); err != nil {
		log.Printf("weather worker: %s: %v", placeKey(p), err)
	}
}

// clampExpires keeps MET's Expires within sane bounds.
func clampExpires(exp, now time.Time) time.Time {
	if exp.IsZero() {
		return now.Add(defaultTTL)
	}
	if exp.Before(now.Add(minTTL)) {
		return now.Add(minTTL)
	}
	if exp.After(now.Add(6 * time.Hour)) {
		return now.Add(6 * time.Hour)
	}
	return exp
}

// recordObs stores the MET step for the current hour as that hour's
// reading. Only steps of one hour count; a 6-hour step is not a reading.
func (w *worker) recordObs(p Place, fc *Forecast, now time.Time) {
	hour := now.UTC().Truncate(time.Hour)
	for _, s := range fc.Steps {
		if s.Time.Equal(hour) && s.Hours == 1 {
			if err := saveObs(p.ID, s); err != nil {
				log.Printf("weather worker: obs %s: %v", placeKey(p), err)
			}
			return
		}
	}
}

// refreshAstro fetches today's sun and moon once per place per day.
func (w *worker) refreshAstro(p Place, lat, lon float64, now time.Time) {
	local := now.In(clock.Zone)
	day := local.Format("2006-01-02")
	if hasAstro(p.Kind, p.ID, day) {
		return
	}
	a, err := fetchMETNoAstro(lat, lon, day, local.Format("-07:00"))
	time.Sleep(workerSpacing)
	if err != nil {
		log.Printf("weather worker: astro %s: %v", placeKey(p), err)
		return
	}
	if err := saveAstro(p.Kind, p.ID, a); err != nil {
		log.Printf("weather worker: save astro %s: %v", placeKey(p), err)
	}
}

// archiveYesterday folds yesterday's hourly readings into one archive row
// per settlement, once a day. Rerunning it is harmless (upsert).
func (w *worker) archiveYesterday(places []Place) {
	today := clock.Now().In(clock.Zone)
	yesterday := today.AddDate(0, 0, -1).Format("2006-01-02")
	if w.archivedDay == yesterday {
		return
	}
	from, to, err := dayBounds(yesterday)
	if err != nil {
		return
	}
	for _, p := range places {
		if p.Kind != placeSettlement {
			continue
		}
		obs, err := readObs(p.ID, from, to)
		if err != nil {
			log.Printf("weather worker: archive %s: %v", placeKey(p), err)
			return
		}
		if len(obs) == 0 {
			continue
		}
		for _, d := range BuildDays(obs, clock.Zone) {
			if d.Date == yesterday {
				if err := saveDailyArchive(p.ID, d); err != nil {
					log.Printf("weather worker: archive %s: %v", placeKey(p), err)
				}
			}
		}
	}
	w.archivedDay = yesterday
}

// dayBounds is a Bucharest calendar day as a UTC half-open interval; the
// clock-change days are 23 or 25 hours long.
func dayBounds(day string) (time.Time, time.Time, error) {
	start, err := time.ParseInLocation("2006-01-02", day, clock.Zone)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start.UTC(), start.AddDate(0, 0, 1).UTC(), nil
}
