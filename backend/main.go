package main

import (
	"log"
	"net/http"

	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/events"
	"backend/internal/handlers"
	"backend/internal/links"
	"backend/internal/middleware"
	"backend/internal/mondasok"
	"backend/internal/news"
	"backend/internal/pagefaq"
	"backend/internal/pages"
	"backend/internal/search"
	"backend/internal/settings"
	"backend/internal/weather"
)

func main() {
	config.Load()
	db.InitDB()
	db.SeedHistoricalSeatsContent()
	settings.MigrateSiteSettings()
	weather.MigrateWeatherTranslations()
	pages.MigratePages()
	pagefaq.Migrate()

	mux := http.DefaultServeMux

	// Core Module (Always Enabled)
	mux.HandleFunc("/api/entries", middleware.ApplyCORS(handlers.EntriesHandler))
	mux.HandleFunc("/api/directory", middleware.ApplyCORS(handlers.EntriesHandler))
	mux.HandleFunc("/api/entry", middleware.ApplyCORS(handlers.EntryDetailHandler))
	mux.HandleFunc("/api/locations", middleware.ApplyCORS(handlers.HandleAdminLocations))
	mux.HandleFunc("/api/admin/entries", middleware.ApplyCORS(handlers.HandleAdminEntries))
	mux.HandleFunc("/api/admin/entry_categories", middleware.ApplyCORS(handlers.HandleAdminEntryCategories))
	mux.HandleFunc("/api/admin/entry_types", middleware.ApplyCORS(handlers.HandleAdminEntryTypes))
	mux.HandleFunc("/api/admin/locations", middleware.ApplyCORS(handlers.HandleAdminLocations))
	mux.HandleFunc("/api/admin/county_seat", middleware.ApplyCORS(handlers.HandleSetCountySeat))
	mux.HandleFunc("/api/attractions", middleware.ApplyCORS(handlers.HandleAttractions))
	mux.HandleFunc("/api/historical_seats", middleware.ApplyCORS(handlers.HandleHistoricalSeats))
	mux.HandleFunc("/api/counties", middleware.ApplyCORS(handlers.HandleCounties))
	mux.HandleFunc("/api/admin/counties", middleware.ApplyCORS(handlers.HandleAdminCounties))
	mux.HandleFunc("/api/admin/historical_seats", middleware.ApplyCORS(handlers.HandleAdminHistoricalSeats))
	mux.HandleFunc("/api/admin/attractions", middleware.ApplyCORS(handlers.HandleAdminAttractions))

	// Public config (weather cache TTL, version) + admin settings
	mux.HandleFunc("/api/config/public", middleware.ApplyCORS(settings.HandlePublicConfig))
	mux.HandleFunc("/api/admin/settings", middleware.ApplyCORS(settings.HandleAdminSettings))
	mux.HandleFunc("/api/admin/settings/clear-weather-cache", middleware.ApplyCORS(settings.ClearWeatherCache))

	// Pages (public + admin)
	mux.HandleFunc("/api/pages", middleware.ApplyCORS(pages.HandlePublicPage))
	mux.HandleFunc("/api/admin/pages", middleware.ApplyCORS(pages.HandleAdminPages))
	mux.HandleFunc("/api/page_faq", middleware.ApplyCORS(pagefaq.HandlePublic))
	mux.HandleFunc("/api/admin/page_faq", middleware.ApplyCORS(pagefaq.HandleAdmin))

	// Optional Modules
	if config.AppConfig.Features.Weather {
		mux.HandleFunc("/api/weather", middleware.ApplyCORS(weather.HandleWeather))
		mux.HandleFunc("/api/weather/county", middleware.ApplyCORS(weather.HandleCountyWeather))
		mux.HandleFunc("/api/admin/weather_translations", middleware.ApplyCORS(weather.HandleAdminWeatherTranslations))
		log.Println("Module [Weather] enabled")
	}

	if config.AppConfig.Features.Events {
		mux.HandleFunc("/api/events", middleware.ApplyCORS(events.HandleEvents))
		mux.HandleFunc("/api/events/detail", middleware.ApplyCORS(events.HandleEventDetail))
		mux.HandleFunc("/api/admin/events", middleware.ApplyCORS(events.HandleAdminEvents))
		log.Println("Module [Events] enabled")
	}

	if config.AppConfig.Features.News {
		mux.HandleFunc("/api/news", middleware.ApplyCORS(news.HandleNews))
		mux.HandleFunc("/api/admin/news_feeds", middleware.ApplyCORS(news.HandleAdminNewsFeeds))
		log.Println("Module [News] enabled")
	}

	if config.AppConfig.Features.Mondasok {
		mux.HandleFunc("/api/admin/mondasok", middleware.ApplyCORS(mondasok.HandleAdminMondasok))
		log.Println("Module [Mondasok] enabled")
	}

	if config.AppConfig.Features.QuickLinks {
		mux.HandleFunc("/api/admin/quick_links", middleware.ApplyCORS(links.HandleAdminQuickLinks))
		log.Println("Module [QuickLinks] enabled")
	}

	if config.AppConfig.Features.Search {
		mux.HandleFunc("/api/search", middleware.ApplyCORS(search.HandleUnifiedSearch))
		mux.HandleFunc("/api/proxy", middleware.ApplyCORS(search.ProxyHandler))
		mux.HandleFunc("/api/autosuggest", middleware.ApplyCORS(search.HandleAutosuggest))
		log.Println("Module [Search] enabled")
	}

	port := config.AppConfig.Port
	log.Printf("Backend API active on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
