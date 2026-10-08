package weather

import (
	"backend/internal/db"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"
)

// Places the weather is kept for. A settlement is archived; an attraction
// only gets the forecast its page shows.
const (
	placeSettlement = "settlement"
	placeAttraction = "attraction"
)

// Where a place's coordinates came from: entered in the admin app, or found
// by name because nobody has entered them yet.
const (
	coordsStored   = "stored"
	coordsGeocoded = "geocoded"
)

// MigrateWeather creates the weather cache and archive tables.
//
// weather_forecast_cache holds the last forecast per place; the worker
// rewrites it, a GET only reads it (WAYS_OF_WORKING R14).
// weather_obs_hourly and weather_daily_archive start on launch day and
// hold MET Norway values only, so the series stays one source: the hourly
// row is MET's analysis-time step for that hour, not a station reading.
func MigrateWeather() {
	_, err := db.DB.Exec(`
		CREATE TABLE IF NOT EXISTS weather_forecast_cache (
			place_kind    VARCHAR(20) NOT NULL,
			place_id      INTEGER NOT NULL,
			latitude      DOUBLE PRECISION NOT NULL,
			longitude     DOUBLE PRECISION NOT NULL,
			coord_source  VARCHAR(20) NOT NULL,
			source        VARCHAR(30),
			fetched_at    TIMESTAMPTZ,
			expires_at    TIMESTAMPTZ,
			last_modified TEXT,
			payload       JSONB,
			PRIMARY KEY (place_kind, place_id)
		);
		CREATE TABLE IF NOT EXISTS weather_astro_daily (
			place_kind VARCHAR(20) NOT NULL,
			place_id   INTEGER NOT NULL,
			day        DATE NOT NULL,
			sunrise    TIMESTAMPTZ,
			sunset     TIMESTAMPTZ,
			moonrise   TIMESTAMPTZ,
			moonset    TIMESTAMPTZ,
			moon_phase DOUBLE PRECISION,
			PRIMARY KEY (place_kind, place_id, day)
		);
		CREATE TABLE IF NOT EXISTS weather_obs_hourly (
			settlement_id INTEGER NOT NULL REFERENCES settlements(id) ON DELETE CASCADE,
			observed_at   TIMESTAMPTZ NOT NULL,
			symbol        VARCHAR(60) NOT NULL,
			temp          DOUBLE PRECISION,
			feels         DOUBLE PRECISION,
			precip_mm     DOUBLE PRECISION,
			wind_kph      DOUBLE PRECISION,
			wind_dir      DOUBLE PRECISION,
			humidity      DOUBLE PRECISION,
			pressure_hpa  DOUBLE PRECISION,
			cloud_pct     DOUBLE PRECISION,
			PRIMARY KEY (settlement_id, observed_at)
		);
		CREATE TABLE IF NOT EXISTS weather_daily_archive (
			settlement_id INTEGER NOT NULL REFERENCES settlements(id) ON DELETE CASCADE,
			day           DATE NOT NULL,
			symbol        VARCHAR(60) NOT NULL,
			tmin          DOUBLE PRECISION,
			tmax          DOUBLE PRECISION,
			precip_mm     DOUBLE PRECISION NOT NULL DEFAULT 0,
			wind_max_kph  DOUBLE PRECISION,
			hours         INTEGER NOT NULL,
			PRIMARY KEY (settlement_id, day)
		);
	`)
	if err != nil {
		log.Printf("weather tables create: %v", err)
		return
	}
	log.Println("Weather cache and archive tables ready")
}

// Place is a settlement or an attraction the weather is kept for.
type Place struct {
	Kind        string
	ID          int
	Slug        string
	Name        string
	NameRo      string
	County      string
	CountySlug  string
	CountyRo    string
	Type        string
	CountySeat  bool
	Latitude    *float64
	Longitude   *float64
	Elevation   *int
	HasLocation bool
}

const placeSettlementSelect = `
	SELECT s.id, s.slug, s.name, COALESCE(s.name_ro, ''), c.name, c.slug, COALESCE(c.name_ro, ''),
	       s.type, COALESCE(s.is_county_seat, false), gl.latitude, gl.longitude, gl.elevation
	FROM settlements s
	JOIN counties c ON c.id = s.county_id
	LEFT JOIN geo_locations gl ON gl.id = s.location_id`

func scanSettlement(sc interface{ Scan(...any) error }) (Place, error) {
	p := Place{Kind: placeSettlement}
	var elev sql.NullInt64
	err := sc.Scan(&p.ID, &p.Slug, &p.Name, &p.NameRo, &p.County, &p.CountySlug, &p.CountyRo,
		&p.Type, &p.CountySeat, &p.Latitude, &p.Longitude, &elev)
	if elev.Valid {
		e := int(elev.Int64)
		p.Elevation = &e
	}
	p.HasLocation = p.Latitude != nil && p.Longitude != nil
	return p, err
}

// loadPlaces returns every settlement and every attraction with a location.
func loadPlaces() ([]Place, error) {
	rows, err := db.DB.Query(placeSettlementSelect + ` ORDER BY s.is_county_seat DESC, s.id`)
	if err != nil {
		return nil, err
	}
	var out []Place
	for rows.Next() {
		p, err := scanSettlement(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	arows, err := db.DB.Query(`
		SELECT a.id, a.slug, a.name, gl.latitude, gl.longitude, gl.elevation
		FROM attractions a JOIN geo_locations gl ON gl.id = a.location_id
		ORDER BY a.id`)
	if err != nil {
		return nil, err
	}
	defer arows.Close()
	for arows.Next() {
		p := Place{Kind: placeAttraction, HasLocation: true}
		var lat, lon float64
		var elev sql.NullInt64
		if err := arows.Scan(&p.ID, &p.Slug, &p.Name, &lat, &lon, &elev); err != nil {
			return nil, err
		}
		p.Latitude, p.Longitude = &lat, &lon
		if elev.Valid {
			e := int(elev.Int64)
			p.Elevation = &e
		}
		out = append(out, p)
	}
	return out, arows.Err()
}

// settlementBySlug finds a settlement. Slugs are unique per county only;
// with no county the oldest one wins, as everywhere else on the site.
func settlementBySlug(slug string) (Place, error) {
	row := db.DB.QueryRow(placeSettlementSelect+` WHERE s.slug = $1 ORDER BY s.id LIMIT 1`, slug)
	return scanSettlement(row)
}

func attractionBySlug(slug string) (Place, error) {
	p := Place{Kind: placeAttraction, HasLocation: true}
	var lat, lon float64
	var elev sql.NullInt64
	err := db.DB.QueryRow(`
		SELECT a.id, a.slug, a.name, gl.latitude, gl.longitude, gl.elevation
		FROM attractions a JOIN geo_locations gl ON gl.id = a.location_id
		WHERE a.slug = $1 ORDER BY a.id LIMIT 1`, slug).Scan(&p.ID, &p.Slug, &p.Name, &lat, &lon, &elev)
	p.Latitude, p.Longitude = &lat, &lon
	if elev.Valid {
		e := int(elev.Int64)
		p.Elevation = &e
	}
	return p, err
}

// cacheRow is one place's row in weather_forecast_cache.
type cacheRow struct {
	Latitude     float64
	Longitude    float64
	CoordSource  string
	Source       string
	FetchedAt    *time.Time
	ExpiresAt    *time.Time
	LastModified string
	Forecast     *Forecast
}

var errNoCache = errors.New("no cached weather")

func readCache(kind string, id int) (*cacheRow, error) {
	var r cacheRow
	var source, lastMod sql.NullString
	var fetched, expires sql.NullTime
	var payload []byte
	err := db.DB.QueryRow(`
		SELECT latitude, longitude, coord_source, source, fetched_at, expires_at, last_modified, payload
		FROM weather_forecast_cache WHERE place_kind = $1 AND place_id = $2`, kind, id).
		Scan(&r.Latitude, &r.Longitude, &r.CoordSource, &source, &fetched, &expires, &lastMod, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNoCache
	}
	if err != nil {
		return nil, err
	}
	r.Source, r.LastModified = source.String, lastMod.String
	if fetched.Valid {
		r.FetchedAt = &fetched.Time
	}
	if expires.Valid {
		r.ExpiresAt = &expires.Time
	}
	if len(payload) > 0 {
		var fc Forecast
		if err := json.Unmarshal(payload, &fc); err == nil {
			r.Forecast = &fc
		}
	}
	return &r, nil
}

// saveCoords records where a place is. Changed coordinates drop the cached
// forecast, which belonged to the old spot.
func saveCoords(kind string, id int, lat, lon float64, src string) error {
	_, err := db.DB.Exec(`
		INSERT INTO weather_forecast_cache (place_kind, place_id, latitude, longitude, coord_source)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (place_kind, place_id) DO UPDATE SET
			latitude = EXCLUDED.latitude, longitude = EXCLUDED.longitude, coord_source = EXCLUDED.coord_source,
			source = NULL, fetched_at = NULL, expires_at = NULL, last_modified = NULL, payload = NULL
		WHERE weather_forecast_cache.latitude <> EXCLUDED.latitude
		   OR weather_forecast_cache.longitude <> EXCLUDED.longitude
		   OR weather_forecast_cache.coord_source <> EXCLUDED.coord_source`,
		kind, id, lat, lon, src)
	return err
}

func saveForecast(kind string, id int, fc *Forecast, fetched, expires time.Time, lastModified string) error {
	payload, err := json.Marshal(fc)
	if err != nil {
		return err
	}
	_, err = db.DB.Exec(`
		UPDATE weather_forecast_cache
		SET source = $3, fetched_at = $4, expires_at = $5, last_modified = NULLIF($6, ''), payload = $7
		WHERE place_kind = $1 AND place_id = $2`,
		kind, id, fc.Source, fetched, expires, lastModified, payload)
	return err
}

// setExpires moves only the next refresh time (MET said 304, or every
// provider failed and the old forecast stays).
func setExpires(kind string, id int, expires time.Time) error {
	_, err := db.DB.Exec(`UPDATE weather_forecast_cache SET expires_at = $3 WHERE place_kind = $1 AND place_id = $2`,
		kind, id, expires)
	return err
}

func hasAstro(kind string, id int, day string) bool {
	var one int
	err := db.DB.QueryRow(`SELECT 1 FROM weather_astro_daily WHERE place_kind = $1 AND place_id = $2 AND day = $3`,
		kind, id, day).Scan(&one)
	return err == nil
}

func saveAstro(kind string, id int, a *Astro) error {
	_, err := db.DB.Exec(`
		INSERT INTO weather_astro_daily (place_kind, place_id, day, sunrise, sunset, moonrise, moonset, moon_phase)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (place_kind, place_id, day) DO UPDATE SET
			sunrise = EXCLUDED.sunrise, sunset = EXCLUDED.sunset, moonrise = EXCLUDED.moonrise,
			moonset = EXCLUDED.moonset, moon_phase = EXCLUDED.moon_phase`,
		kind, id, a.Date, a.Sunrise, a.Sunset, a.Moonrise, a.Moonset, a.MoonPhase)
	return err
}

func readAstro(kind string, id int, day string) (*Astro, error) {
	a := &Astro{Date: day}
	var rise, set, mrise, mset sql.NullTime
	var phase sql.NullFloat64
	err := db.DB.QueryRow(`
		SELECT sunrise, sunset, moonrise, moonset, moon_phase FROM weather_astro_daily
		WHERE place_kind = $1 AND place_id = $2 AND day = $3`, kind, id, day).
		Scan(&rise, &set, &mrise, &mset, &phase)
	if err != nil {
		return nil, err
	}
	for _, p := range []struct {
		src *sql.NullTime
		dst **time.Time
	}{{&rise, &a.Sunrise}, {&set, &a.Sunset}, {&mrise, &a.Moonrise}, {&mset, &a.Moonset}} {
		if p.src.Valid {
			t := p.src.Time.UTC()
			*p.dst = &t
		}
	}
	if phase.Valid {
		a.MoonPhase = &phase.Float64
	}
	return a, nil
}

// saveObs stores a MET step as the reading for its hour. A later refresh in
// the same hour overwrites it with the newer analysis.
func saveObs(settlementID int, s Step) error {
	_, err := db.DB.Exec(`
		INSERT INTO weather_obs_hourly (settlement_id, observed_at, symbol, temp, feels, precip_mm,
			wind_kph, wind_dir, humidity, pressure_hpa, cloud_pct)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (settlement_id, observed_at) DO UPDATE SET
			symbol = EXCLUDED.symbol, temp = EXCLUDED.temp, feels = EXCLUDED.feels,
			precip_mm = EXCLUDED.precip_mm, wind_kph = EXCLUDED.wind_kph, wind_dir = EXCLUDED.wind_dir,
			humidity = EXCLUDED.humidity, pressure_hpa = EXCLUDED.pressure_hpa, cloud_pct = EXCLUDED.cloud_pct`,
		settlementID, s.Time, s.Symbol, s.Temp, s.Feels, s.PrecipMm, s.WindKph, s.WindDir,
		s.Humidity, s.PressureHpa, s.CloudPct)
	return err
}

// readObs returns a settlement's stored hours in [from, to).
func readObs(settlementID int, from, to time.Time) ([]Step, error) {
	rows, err := db.DB.Query(`
		SELECT observed_at, symbol, temp, feels, precip_mm, wind_kph, wind_dir, humidity, pressure_hpa, cloud_pct
		FROM weather_obs_hourly
		WHERE settlement_id = $1 AND observed_at >= $2 AND observed_at < $3
		ORDER BY observed_at`, settlementID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Step
	for rows.Next() {
		s := Step{Hours: 1}
		if err := rows.Scan(&s.Time, &s.Symbol, &s.Temp, &s.Feels, &s.PrecipMm, &s.WindKph, &s.WindDir,
			&s.Humidity, &s.PressureHpa, &s.CloudPct); err != nil {
			return nil, err
		}
		s.Time = s.Time.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

func saveDailyArchive(settlementID int, d Day) error {
	_, err := db.DB.Exec(`
		INSERT INTO weather_daily_archive (settlement_id, day, symbol, tmin, tmax, precip_mm, wind_max_kph, hours)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (settlement_id, day) DO UPDATE SET
			symbol = EXCLUDED.symbol, tmin = EXCLUDED.tmin, tmax = EXCLUDED.tmax,
			precip_mm = EXCLUDED.precip_mm, wind_max_kph = EXCLUDED.wind_max_kph, hours = EXCLUDED.hours`,
		settlementID, d.Date, d.Symbol, d.TMin, d.TMax, d.PrecipMm, d.WindMaxKph, d.Hours)
	return err
}

// readDailyArchive returns a settlement's archived days in [from, to],
// newest first.
func readDailyArchive(settlementID int, from, to string) ([]Day, error) {
	rows, err := db.DB.Query(`
		SELECT to_char(day, 'YYYY-MM-DD'), symbol, tmin, tmax, precip_mm, wind_max_kph, hours
		FROM weather_daily_archive
		WHERE settlement_id = $1 AND day BETWEEN $2 AND $3
		ORDER BY day DESC`, settlementID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Day{}
	for rows.Next() {
		var d Day
		if err := rows.Scan(&d.Date, &d.Symbol, &d.TMin, &d.TMax, &d.PrecipMm, &d.WindMaxKph, &d.Hours); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// loadDescOverrides reads the admin's Hungarian texts, keyed by base symbol
// code, once per request instead of once per step.
func loadDescOverrides(lang string) map[string]string {
	out := map[string]string{}
	rows, err := db.DB.Query(`SELECT source_text, translated_text FROM weather_desc_translations WHERE lang = $1`, lang)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var k, v string
		if rows.Scan(&k, &v) == nil && v != "" {
			if knownBase[k] {
				out[k] = v
			}
		}
	}
	return out
}

// describe is a symbol's Hungarian text, the admin's own wording first.
func describe(code string, overrides map[string]string) string {
	base, _ := splitSymbol(code)
	if v, ok := overrides[base]; ok {
		return v
	}
	return symbolDescHU(code)
}
