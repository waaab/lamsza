// Package weather serves the site's weather: the widgets, the /idojaras
// pages and the archive.
//
// Sources are free for commercial use, because the network will carry ads:
// MET Norway first (CC BY 4.0), WeatherAPI.com and OpenWeatherMap as
// fallbacks. Open-Meteo is not used: its free API is non-commercial only.
// A background worker (worker.go) fills the cache and the archive; every
// handler here only reads them.
package weather

import (
	"backend/internal/clock"
	"backend/internal/db"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

// forecastDays is how many days the forecast route returns at most. MET
// reaches about nine and a half days ahead.
const forecastDays = 10

// hourlyAhead is how far the hour-by-hour strip reaches.
const hourlyAhead = 48 * time.Hour

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// placeFromRequest finds the place a request names: ?slug= is a settlement,
// else an attraction.
func placeFromRequest(r *http.Request) (Place, int, string) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		return Place{}, http.StatusBadRequest, "missing slug"
	}
	p, err := settlementBySlug(slug)
	if err == nil {
		return p, 0, ""
	}
	if !errors.Is(err, sql.ErrNoRows) {
		log.Printf("weather: settlement %q: %v", slug, err)
		return Place{}, http.StatusInternalServerError, "lookup failed"
	}
	if p, err = attractionBySlug(slug); err == nil {
		return p, 0, ""
	}
	return Place{}, http.StatusNotFound, "unknown place"
}

// cachedForecast returns a place's cached forecast, if it has one.
func cachedForecast(p Place) (*cacheRow, bool) {
	row, err := readCache(p.Kind, p.ID)
	if err != nil || row.Forecast == nil || len(row.Forecast.Steps) == 0 {
		return nil, false
	}
	return row, true
}

func r1(v *float64) *float64 {
	if v == nil {
		return nil
	}
	return fp(round1(*v))
}

func roundInt(v *float64) *int {
	if v == nil {
		return nil
	}
	i := int(math.Round(*v))
	return &i
}

// ---- GET /api/weather (the widgets) ---------------------------------------

// UnifiedWeatherResponse is what the home, county and attraction widgets and
// the search answer read. Icon is the old OpenWeatherMap-style code the emoji
// style uses; Symbol is the MET code the animated icons use.
type UnifiedWeatherResponse struct {
	Slug      string   `json:"slug,omitempty"`
	Temp      int      `json:"temp"`
	TempMin   *int     `json:"temp_min,omitempty"`
	TempMax   *int     `json:"temp_max,omitempty"`
	Desc      string   `json:"desc"`
	Icon      string   `json:"icon"`
	Symbol    string   `json:"symbol"`
	Source    string   `json:"source"`
	FetchedAt int64    `json:"fetched_at"`
	Humidity  *int     `json:"humidity,omitempty"`
	WindKph   *float64 `json:"wind_kph,omitempty"`
	PrecipMm  *float64 `json:"precip_mm,omitempty"`
}

// summarize turns a forecast into the widget's "now" plus today's range.
func summarize(fc *Forecast, fetched time.Time, now time.Time, overrides map[string]string) (*UnifiedWeatherResponse, bool) {
	cur, ok := CurrentStep(fc.Steps, now)
	if !ok || cur.Temp == nil {
		return nil, false
	}
	out := &UnifiedWeatherResponse{
		Temp:      int(math.Round(*cur.Temp)),
		Desc:      describe(cur.Symbol, overrides),
		Icon:      legacyIcon(cur.Symbol),
		Symbol:    cur.Symbol,
		Source:    providerDisplayName(fc.Source),
		FetchedAt: fetched.Unix(),
		Humidity:  roundInt(cur.Humidity),
		WindKph:   r1(cur.WindKph),
		PrecipMm:  r1(cur.PrecipMm),
	}
	today := now.In(clock.Zone).Format("2006-01-02")
	for _, d := range BuildDays(fc.Steps, clock.Zone) {
		if d.Date == today {
			out.TempMin, out.TempMax = roundInt(d.TMin), roundInt(d.TMax)
			break
		}
	}
	return out, true
}

// HandleWeather answers the widgets: ?slug= (a settlement or an attraction)
// or ?lat=&lon= (any point; served live, never stored).
func HandleWeather(w http.ResponseWriter, r *http.Request) {
	now := clock.Now()
	overrides := loadDescOverrides("hu")
	q := r.URL.Query()

	if q.Get("slug") == "" && q.Get("lat") != "" && q.Get("lon") != "" {
		lat, err1 := strconv.ParseFloat(q.Get("lat"), 64)
		lon, err2 := strconv.ParseFloat(q.Get("lon"), 64)
		if err1 != nil || err2 != nil || math.Abs(lat) > 90 || math.Abs(lon) > 180 {
			writeError(w, http.StatusBadRequest, "bad coordinates")
			return
		}
		fc, err := liveForecast(lat, lon)
		if err != nil {
			log.Printf("weather: live %v,%v: %v", lat, lon, err)
			writeError(w, http.StatusBadGateway, "weather provider error")
			return
		}
		if out, ok := summarize(fc, now, now, overrides); ok {
			writeJSON(w, http.StatusOK, out)
			return
		}
		writeError(w, http.StatusBadGateway, "weather provider error")
		return
	}

	p, status, msg := placeFromRequest(r)
	if status != 0 {
		writeError(w, status, msg)
		return
	}
	row, ok := cachedForecast(p)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "weather not ready yet")
		return
	}
	out, ok := summarize(row.Forecast, *row.FetchedAt, now, overrides)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "weather not ready yet")
		return
	}
	out.Slug = p.Slug
	writeJSON(w, http.StatusOK, out)
}

// liveForecast asks the providers directly, for a point the worker does not
// keep. Nothing is stored.
func liveForecast(lat, lon float64) (*Forecast, error) {
	var lastErr error
	for _, prov := range providerOrder() {
		var fc *Forecast
		var err error
		switch prov {
		case providerMETNo:
			var res *metnoResult
			if res, err = fetchMETNo(lat, lon, nil, ""); err == nil {
				fc = res.Forecast
			}
		case providerWeatherAPICom:
			fc, err = fetchWeatherAPICom(lat, lon)
		case providerOpenWeatherMap:
			fc, err = fetchOpenWeatherMap(lat, lon)
		}
		if err == nil && fc != nil {
			return fc, nil
		}
		if err != nil && !errors.Is(err, errNoKey) {
			lastErr = err
		}
	}
	if lastErr == nil {
		lastErr = errors.New("no provider answered")
	}
	return nil, lastErr
}

// ---- GET /api/weather/county ----------------------------------------------

type cityWeather struct {
	City         string  `json:"city"`
	Slug         string  `json:"slug"`
	Temp         float64 `json:"temp"`
	TempMin      *int    `json:"temp_min,omitempty"`
	TempMax      *int    `json:"temp_max,omitempty"`
	Desc         string  `json:"desc"`
	Icon         string  `json:"icon"`
	Symbol       string  `json:"symbol"`
	IsCountySeat bool    `json:"is_county_seat"`
	Source       string  `json:"source,omitempty"`
}

// HandleCountyWeather returns the cached weather of a county's towns.
func HandleCountyWeather(w http.ResponseWriter, r *http.Request) {
	countySlug := r.URL.Query().Get("slug")
	if countySlug == "" {
		writeError(w, http.StatusBadRequest, "missing slug")
		return
	}
	rows, err := db.DB.Query(placeSettlementSelect+`
		WHERE c.slug = $1 AND s.type IN ('város', 'municípium')
		ORDER BY s.name`, countySlug)
	if err != nil {
		log.Printf("weather county %q: %v", countySlug, err)
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	var places []Place
	for rows.Next() {
		if p, err := scanSettlement(rows); err == nil {
			places = append(places, p)
		}
	}
	rows.Close()

	now := clock.Now()
	overrides := loadDescOverrides("hu")
	results := []cityWeather{}
	for _, p := range places {
		row, ok := cachedForecast(p)
		if !ok {
			continue
		}
		s, ok := summarize(row.Forecast, *row.FetchedAt, now, overrides)
		if !ok {
			continue
		}
		results = append(results, cityWeather{
			City: p.Name, Slug: p.Slug, Temp: float64(s.Temp), TempMin: s.TempMin, TempMax: s.TempMax,
			Desc: s.Desc, Icon: s.Icon, Symbol: s.Symbol, IsCountySeat: p.CountySeat, Source: s.Source,
		})
	}
	writeJSON(w, http.StatusOK, results)
}

// ---- GET /api/weather/forecast --------------------------------------------

type placeOut struct {
	Kind        string   `json:"kind"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	NameRo      string   `json:"name_ro,omitempty"`
	County      string   `json:"county,omitempty"`
	CountySlug  string   `json:"county_slug,omitempty"`
	Type        string   `json:"type,omitempty"`
	Latitude    *float64 `json:"lat,omitempty"`
	Longitude   *float64 `json:"lon,omitempty"`
	Elevation   *int     `json:"elevation,omitempty"`
	CoordSource string   `json:"coord_source,omitempty"`
}

func toPlaceOut(p Place, row *cacheRow) placeOut {
	out := placeOut{Kind: p.Kind, Slug: p.Slug, Name: p.Name, NameRo: p.NameRo, County: p.County,
		CountySlug: p.CountySlug, Type: p.Type, Elevation: p.Elevation}
	if row != nil {
		lat, lon := row.Latitude, row.Longitude
		out.Latitude, out.Longitude, out.CoordSource = &lat, &lon, row.CoordSource
	}
	return out
}

type stepOut struct {
	Step
	Desc string `json:"desc"`
}

type dayOut struct {
	Day
	Desc string `json:"desc"`
}

type forecastOut struct {
	Place      placeOut  `json:"place"`
	Source     string    `json:"source"`
	SourceName string    `json:"source_name"`
	FetchedAt  time.Time `json:"fetched_at"`
	// Stale: the providers have not answered for a while; the page says so.
	Stale   bool      `json:"stale"`
	Current stepOut   `json:"current"`
	Hourly  []stepOut `json:"hourly"`
	Daily   []dayOut  `json:"daily"`
	Astro   *Astro    `json:"astro"`
}

// HandleForecast returns a place's forecast: now, the next 48 hours and
// the days ahead, with today's sun and moon.
func HandleForecast(w http.ResponseWriter, r *http.Request) {
	p, status, msg := placeFromRequest(r)
	if status != 0 {
		writeError(w, status, msg)
		return
	}
	row, ok := cachedForecast(p)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "weather not ready yet")
		return
	}
	now := clock.Now()
	overrides := loadDescOverrides("hu")
	fc := row.Forecast

	cur, _ := CurrentStep(fc.Steps, now)
	out := forecastOut{
		Place:      toPlaceOut(p, row),
		Source:     fc.Source,
		SourceName: providerDisplayName(fc.Source),
		FetchedAt:  row.FetchedAt.UTC(),
		Stale:      row.ExpiresAt != nil && now.After(row.ExpiresAt.Add(staleGrace)),
		Current:    stepOut{Step: cur, Desc: describe(cur.Symbol, overrides)},
		Hourly:     []stepOut{},
		Daily:      []dayOut{},
	}
	from := now.UTC().Truncate(time.Hour)
	for _, s := range StepsBetween(fc.Steps, from, from.Add(hourlyAhead)) {
		out.Hourly = append(out.Hourly, stepOut{Step: s, Desc: describe(s.Symbol, overrides)})
	}
	today := now.In(clock.Zone).Format("2006-01-02")
	for _, d := range BuildDays(fc.Steps, clock.Zone) {
		if d.Date < today || len(out.Daily) >= forecastDays {
			continue
		}
		out.Daily = append(out.Daily, dayOut{Day: d, Desc: describe(d.Symbol, overrides)})
	}
	if a, err := readAstro(p.Kind, p.ID, today); err == nil {
		out.Astro = a
	}
	writeJSON(w, http.StatusOK, out)
}

// ---- GET /api/weather/archive ---------------------------------------------

type archiveOut struct {
	Place placeOut `json:"place"`
	// Since is the first archived day: the archive starts on launch day.
	Since *string   `json:"since"`
	From  string    `json:"from"`
	To    string    `json:"to"`
	Days  []dayOut  `json:"days"`
	Hours []stepOut `json:"hours,omitempty"`
}

// maxArchiveDays caps one request.
const maxArchiveDays = 366

// HandleArchive returns a settlement's archived days: ?slug=&days=30 (the
// last N days, today excluded), or ?slug=&date=YYYY-MM-DD for that day's
// hourly readings.
func HandleArchive(w http.ResponseWriter, r *http.Request) {
	p, status, msg := placeFromRequest(r)
	if status != 0 {
		writeError(w, status, msg)
		return
	}
	if p.Kind != placeSettlement {
		writeError(w, http.StatusNotFound, "only settlements have an archive")
		return
	}
	overrides := loadDescOverrides("hu")
	row, _ := readCache(p.Kind, p.ID)
	out := archiveOut{Place: toPlaceOut(p, row), Days: []dayOut{}}

	var since sql.NullString
	if err := db.DB.QueryRow(`SELECT to_char(MIN(day), 'YYYY-MM-DD') FROM weather_daily_archive WHERE settlement_id = $1`,
		p.ID).Scan(&since); err == nil && since.Valid {
		out.Since = &since.String
	}

	q := r.URL.Query()
	if date := q.Get("date"); date != "" {
		from, to, err := dayBounds(date)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad date")
			return
		}
		obs, err := readObs(p.ID, from, to)
		if err != nil {
			log.Printf("weather archive %s %s: %v", p.Slug, date, err)
			writeError(w, http.StatusInternalServerError, "lookup failed")
			return
		}
		out.From, out.To = date, date
		for _, s := range obs {
			out.Hours = append(out.Hours, stepOut{Step: s, Desc: describe(s.Symbol, overrides)})
		}
		days, err := readDailyArchive(p.ID, date, date)
		if err == nil {
			for _, d := range days {
				out.Days = append(out.Days, dayOut{Day: d, Desc: describe(d.Symbol, overrides)})
			}
		}
		writeJSON(w, http.StatusOK, out)
		return
	}

	n := 30
	if s := q.Get("days"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 {
			writeError(w, http.StatusBadRequest, "bad days")
			return
		}
		n = min(v, maxArchiveDays)
	}
	today := clock.Now().In(clock.Zone)
	out.To = today.AddDate(0, 0, -1).Format("2006-01-02")
	out.From = today.AddDate(0, 0, -n).Format("2006-01-02")
	days, err := readDailyArchive(p.ID, out.From, out.To)
	if err != nil {
		log.Printf("weather archive %s: %v", p.Slug, err)
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	for _, d := range days {
		out.Days = append(out.Days, dayOut{Day: d, Desc: describe(d.Symbol, overrides)})
	}
	writeJSON(w, http.StatusOK, out)
}

// ---- GET /api/weather/places ----------------------------------------------

type placeListItem struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	County     string `json:"county"`
	CountySlug string `json:"county_slug"`
	Type       string `json:"type"`
	CountySeat bool   `json:"is_county_seat"`
	// Now is the widget summary, nil until the worker has a forecast.
	Now *UnifiedWeatherResponse `json:"now"`
}

// HandlePlaces lists every settlement for the /idojaras overview and picker,
// each with its current weather from the cache.
func HandlePlaces(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(placeSettlementSelect + ` ORDER BY c.name, s.name`)
	if err != nil {
		log.Printf("weather places: %v", err)
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	var places []Place
	for rows.Next() {
		if p, err := scanSettlement(rows); err == nil {
			places = append(places, p)
		}
	}
	rows.Close()

	now := clock.Now()
	overrides := loadDescOverrides("hu")
	out := []placeListItem{}
	for _, p := range places {
		it := placeListItem{Slug: p.Slug, Name: p.Name, County: p.County, CountySlug: p.CountySlug,
			Type: p.Type, CountySeat: p.CountySeat}
		if row, ok := cachedForecast(p); ok {
			if s, ok := summarize(row.Forecast, *row.FetchedAt, now, overrides); ok {
				s.Slug = p.Slug
				it.Now = s
			}
		}
		out = append(out, it)
	}
	writeJSON(w, http.StatusOK, out)
}

// ---- admin translations table ---------------------------------------------

// MigrateWeatherTranslations creates the weather_desc_translations table.
// Since the move to MET Norway its source_text is a base symbol code
// ("partlycloudy", "lightrainshowers"): the admin's own Hungarian wording
// for that weather. Older rows keyed by an English provider phrase are
// ignored.
func MigrateWeatherTranslations() {
	_, err := db.DB.Exec(`
		CREATE TABLE IF NOT EXISTS weather_desc_translations (
			id SERIAL PRIMARY KEY,
			source_text VARCHAR(255) NOT NULL,
			lang VARCHAR(10) NOT NULL,
			translated_text VARCHAR(255) NOT NULL,
			UNIQUE(source_text, lang)
		)
	`)
	if err != nil {
		log.Printf("weather_desc_translations create: %v", err)
		return
	}
	log.Println("Weather translations table ready")
}
