package settings

import (
	"backend/internal/db"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// MigrateSiteSettings creates site_settings table and seeds default values if missing.
// Call once after db.InitDB() so the user does not need to run SQL manually.
func MigrateSiteSettings() {
	_, err := db.DB.Exec(`
		CREATE TABLE IF NOT EXISTS site_settings (
			key VARCHAR(100) PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Printf("site_settings create: %v", err)
		return
	}
	_, err = db.DB.Exec(`
		INSERT INTO site_settings (key, value) VALUES
		  ('weather_cache_ttl_minutes', '15'),
		  ('weather_cache_version', '1'),
		  ('quick_links_version', '1'),
		  ('weather_icon_style', 'emoji'),
		  ('weather_active_users_estimate', '10000'),
		  ('weather_provider_default', 'open_meteo'),
		  ('weather_provider_open_meteo_enabled', 'true'),
		  ('weather_provider_weatherapi_enabled', 'true'),
		  ('weather_provider_openweathermap_enabled', 'true'),
		  ('my_location_slug', 'csikszereda')
		ON CONFLICT (key) DO NOTHING
	`)
	if err != nil {
		log.Printf("site_settings seed: %v", err)
		return
	}
	log.Println("Site settings table ready")
}

// PublicConfig is returned by GET /api/config/public (no auth)
type PublicConfig struct {
	WeatherCacheTTLMinutes int    `json:"weather_cache_ttl_minutes"`
	WeatherCacheVersion    string `json:"weather_cache_version"`
	QuickLinksVersion      string `json:"quick_links_version"`
	QuickLinksCount        int    `json:"quick_links_count"`
	WeatherIconStyle       string `json:"weather_icon_style"`
	MyLocationSlug         string `json:"my_location_slug"`
	MyLocationName         string `json:"my_location_name"`
	MyLocationCounty       string `json:"my_location_county"`
	MyLocationCountySlug   string `json:"my_location_county_slug"`
	MyLocationType         string `json:"my_location_type"`
}

// HandlePublicConfig returns weather cache config for frontend
func HandlePublicConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ttl, _ := getSetting("weather_cache_ttl_minutes", "15")
	version, _ := getSetting("weather_cache_version", "1")
	qlVersion, _ := getSetting("quick_links_version", "1")
	iconStyle, _ := getSetting("weather_icon_style", "emoji")
	myLocSlug, _ := getSetting("my_location_slug", "csikszereda")
	ttlInt, _ := strconv.Atoi(ttl)
	qlCount := 0
	_ = db.DB.QueryRow("SELECT COUNT(*) FROM quick_links").Scan(&qlCount)

	var myLocName, myLocCounty, myLocCountySlug, myLocType string
	_ = db.DB.QueryRow(
		"SELECT COALESCE(s.name,''), COALESCE(c.name,''), COALESCE(c.slug,''), COALESCE(s.type,'') FROM settlements s JOIN counties c ON s.county_id = c.id WHERE s.slug = $1",
		myLocSlug,
	).Scan(&myLocName, &myLocCounty, &myLocCountySlug, &myLocType)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PublicConfig{
		WeatherCacheTTLMinutes: ttlInt,
		WeatherCacheVersion:    version,
		QuickLinksVersion:      qlVersion,
		QuickLinksCount:        qlCount,
		WeatherIconStyle:       iconStyle,
		MyLocationSlug:         myLocSlug,
		MyLocationName:         myLocName,
		MyLocationCounty:       myLocCounty,
		MyLocationCountySlug:   myLocCountySlug,
		MyLocationType:         myLocType,
	})
}

// GetSetting returns a site_settings value or default if missing/error (e.g. table not yet migrated)
func GetSetting(key, defaultVal string) (string, error) {
	var val string
	err := db.DB.QueryRow("SELECT value FROM site_settings WHERE key = $1", key).Scan(&val)
	if err != nil {
		return defaultVal, err
	}
	return val, nil
}

func getSetting(key, defaultVal string) (string, error) {
	return GetSetting(key, defaultVal)
}

func getSettingInt(key string, defaultVal int) (int, error) {
	s, err := getSetting(key, strconv.Itoa(defaultVal))
	if err != nil {
		return defaultVal, err
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal, err
	}
	return v, nil
}

// HandleAdminSettings: GET returns all settings, PUT accepts JSON body { "key": "value", ... }
func HandleAdminSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := db.DB.Query("SELECT key, value FROM site_settings ORDER BY LOWER(key) ASC")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		m := make(map[string]string)
		for rows.Next() {
			var k, v string
			if err := rows.Scan(&k, &v); err != nil {
				continue
			}
			m[k] = v
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
		return
	case http.MethodPut:
		var m map[string]string
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		for k, v := range m {
			_, err := db.DB.Exec(
				`INSERT INTO site_settings (key, value, updated_at) VALUES ($1, $2, CURRENT_TIMESTAMP)
				 ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = CURRENT_TIMESTAMP`,
				k, v,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// ClearWeatherCache increments weather_cache_version (admin)
func ClearWeatherCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var val string
	_ = db.DB.QueryRow("SELECT value FROM site_settings WHERE key = 'weather_cache_version'").Scan(&val)
	next := 1
	if n, err := strconv.Atoi(val); err == nil {
		next = n + 1
	}
	_, _ = db.DB.Exec(
		`INSERT INTO site_settings (key, value, updated_at) VALUES ('weather_cache_version', $1, CURRENT_TIMESTAMP)
		 ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = CURRENT_TIMESTAMP`,
		strconv.Itoa(next),
	)
	w.WriteHeader(http.StatusOK)
}

// IncrementQuickLinksVersion bumps quick_links_version so frontend cache is invalidated
func IncrementQuickLinksVersion() {
	var val string
	_ = db.DB.QueryRow("SELECT value FROM site_settings WHERE key = 'quick_links_version'").Scan(&val)
	next := 1
	if n, err := strconv.Atoi(val); err == nil {
		next = n + 1
	}
	_, _ = db.DB.Exec(
		`INSERT INTO site_settings (key, value, updated_at) VALUES ('quick_links_version', $1, CURRENT_TIMESTAMP)
		 ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = CURRENT_TIMESTAMP`,
		strconv.Itoa(next),
	)
}
