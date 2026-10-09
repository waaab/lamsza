package settings

import (
	"backend/internal/config"
	"backend/internal/db"
	"testing"
)

// The site_settings keys lamsza-admin's Beállítások tab writes
// (lamsza-admin frontend/src/lib/settingsSections.js). Lamsza seeds and reads
// each of them; renaming one here breaks the admin setting silently, so this
// list changes only together with the admin's (whose own test,
// TestSiteSettingsKeysAreReadByLamsza, checks the other direction).
var adminEditableKeys = []string{
	"social_facebook_url", "social_twitter_url", "social_instagram_url",
	"my_location_slug",
	"weather_provider_metno_enabled", "weather_provider_weatherapi_enabled", "weather_provider_openweathermap_enabled",
	"weather_icon_style", "weather_cache_ttl_minutes",
}

// The boot migration seeds every admin-editable key, plus the two counters
// the admin backend bumps (weather_cache_version, quick_links_version).
func TestSiteSettingsSeedsTheAdminKeys(t *testing.T) {
	config.Load()
	testURL, err := db.TestDatabaseURL()
	if err != nil {
		t.Fatal(err)
	}
	config.AppConfig.DatabaseURL = testURL
	db.InitDB()
	MigrateSiteSettings()

	for _, key := range append(adminEditableKeys, "weather_cache_version", "quick_links_version") {
		var n int
		if err := db.DB.QueryRow(`SELECT COUNT(*) FROM site_settings WHERE key = $1`, key).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Errorf("site_settings has no %q row after the boot seed", key)
		}
	}
}
