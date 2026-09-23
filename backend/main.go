package main

import (
	"log"
	"net/http"

	"backend/internal/account"
	"backend/internal/auth"
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
	"backend/internal/venues"
	"backend/internal/weather"
)

func main() {
	config.Load()
	db.InitDB()
	db.SeedHistoricalSeatsContent()
	mondasok.Migrate()
	settings.MigrateSiteSettings()
	weather.MigrateWeatherTranslations()
	pages.MigratePages()
	pagefaq.Migrate()
	handlers.MigrateEntryVerified()
	handlers.MigrateEntryReviews()
	handlers.MigrateEntryLocationSearch()
	handlers.MigrateSettlementLocationTypes()
	events.Migrate()
	auth.Migrate()
	account.Migrate()
	account.MigrateWebsites()

	mux := http.DefaultServeMux
	admin := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.ApplyCORS(auth.RequireAdmin(h))
	}

	mux.HandleFunc("/api/auth/google", middleware.ApplyCORS(auth.HandleGoogleLogin))
	mux.HandleFunc("/api/auth/me", middleware.ApplyCORS(auth.HandleMe))
	mux.HandleFunc("/api/auth/logout", middleware.ApplyCORS(auth.HandleLogout))
	mux.HandleFunc("/api/account/preferences", middleware.ApplyCORS(account.HandlePreferences))
	mux.HandleFunc("/api/account/import", middleware.ApplyCORS(account.HandleImport))
	mux.HandleFunc("/api/account/links", middleware.ApplyCORS(account.HandleLinks))
	mux.HandleFunc("/api/account/history", middleware.ApplyCORS(account.HandleHistory))
	mux.HandleFunc("/api/account/favorites", middleware.ApplyCORS(account.HandleFavorites))
	mux.HandleFunc("/api/account/listings/catalog", middleware.ApplyCORS(account.HandleListingCatalog))
	mux.HandleFunc("/api/account/listings/claim", middleware.ApplyCORS(account.HandleClaimListing))
	mux.HandleFunc("/api/account/listings/members", middleware.ApplyCORS(account.HandleListingMembers))
	mux.HandleFunc("/api/account/listings", middleware.ApplyCORS(account.HandleListings))
	mux.HandleFunc("/api/websites", middleware.ApplyCORS(account.HandleWebsites))

	// Core Module (Always Enabled)
	mux.HandleFunc("/api/entries", middleware.ApplyCORS(handlers.EntriesHandler))
	mux.HandleFunc("/api/directory", middleware.ApplyCORS(handlers.EntriesHandler))
	mux.HandleFunc("/api/entry", middleware.ApplyCORS(handlers.EntryDetailHandler))
	mux.HandleFunc("/api/entry/related", middleware.ApplyCORS(handlers.HandleEntryRelated))
	mux.HandleFunc("/api/entry/reviews", middleware.ApplyCORS(handlers.HandleEntryReviews))
	mux.HandleFunc("/api/locations", middleware.ApplyCORS(handlers.HandleAdminLocations))
	mux.HandleFunc("/api/settlement_location_types", middleware.ApplyCORS(handlers.HandlePublicSettlementLocationTypes))
	mux.HandleFunc("/api/admin/listing-queue", admin(account.HandleListingQueue))
	mux.HandleFunc("/api/admin/listing-queue/publish", admin(account.HandleListingQueuePublish))
	mux.HandleFunc("/api/admin/listing-queue/member", admin(account.HandleListingQueueMember))
	mux.HandleFunc("/api/admin/entries", middleware.ApplyCORS(handlers.HandleAdminEntries))
	mux.HandleFunc("/api/admin/entry_categories", middleware.ApplyCORS(handlers.HandleAdminEntryCategories))
	mux.HandleFunc("/api/admin/entry_types", middleware.ApplyCORS(handlers.HandleAdminEntryTypes))
	mux.HandleFunc("/api/admin/locations", middleware.ApplyCORS(handlers.HandleAdminLocations))
	mux.HandleFunc("/api/admin/settlement_location_types", middleware.ApplyCORS(handlers.HandleAdminSettlementLocationTypes))
	mux.HandleFunc("/api/admin/county_seat", middleware.ApplyCORS(handlers.HandleSetCountySeat))
	mux.HandleFunc("/api/admin/dashboard_stats", middleware.ApplyCORS(handlers.HandleAdminDashboardStats))
	mux.HandleFunc("/api/admin/entry-images", middleware.ApplyCORS(handlers.HandleEntryImageUpload))
	mux.Handle("/api/media/entry-images/", middleware.ApplyCORS(http.StripPrefix("/api/media/entry-images/", http.FileServer(http.Dir(handlers.EntryImagesDir()))).ServeHTTP))
	mux.Handle("/api/media/event-images/", middleware.ApplyCORS(http.StripPrefix("/api/media/event-images/", http.FileServer(http.Dir(handlers.EventImagesDir()))).ServeHTTP))
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
		mux.HandleFunc("/api/events/filter-options", middleware.ApplyCORS(events.HandleEventFilterOptions))
		mux.HandleFunc("/api/events/detail", middleware.ApplyCORS(events.HandleEventDetail))
		mux.HandleFunc("/api/venues", middleware.ApplyCORS(venues.HandlePublic))
		mux.HandleFunc("/api/venue_types", middleware.ApplyCORS(venues.HandlePublicVenueTypes))
		mux.HandleFunc("/api/admin/events", middleware.ApplyCORS(events.HandleAdminEvents))
		mux.HandleFunc("/api/admin/events/schedule", middleware.ApplyCORS(events.HandleAdminEventSchedule))
		mux.HandleFunc("/api/admin/catalog_event_types", middleware.ApplyCORS(events.HandleAdminCatalogEventTypes))
		mux.HandleFunc("/api/admin/catalog_event_subtypes", middleware.ApplyCORS(events.HandleAdminCatalogEventSubtypes))
		mux.HandleFunc("/api/admin/event-images", middleware.ApplyCORS(handlers.HandleEventImageUpload))
		mux.HandleFunc("/api/admin/venues", middleware.ApplyCORS(venues.HandleAdmin))
		mux.HandleFunc("/api/admin/venue_types", middleware.ApplyCORS(venues.HandleAdminVenueTypes))
		log.Println("Module [Events] enabled")
	}

	if config.AppConfig.Features.News {
		mux.HandleFunc("/api/news", middleware.ApplyCORS(news.HandleNews))
		mux.HandleFunc("/api/news/feeds", middleware.ApplyCORS(news.HandlePublicNewsFeeds))
		mux.HandleFunc("/api/admin/news_feeds", middleware.ApplyCORS(news.HandleAdminNewsFeeds))
		log.Println("Module [News] enabled")
	}

	if config.AppConfig.Features.Mondasok {
		mux.HandleFunc("/api/mondasok", middleware.ApplyCORS(mondasok.HandlePublicMondasok))
		mux.HandleFunc("/api/admin/mondasok", middleware.ApplyCORS(mondasok.HandleAdminMondasok))
		log.Println("Module [Mondasok] enabled")
	}

	if config.AppConfig.Features.QuickLinks {
		mux.HandleFunc("/api/quick_links", middleware.ApplyCORS(links.HandlePublicQuickLinks))
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
	var handler http.Handler = mux
	if !config.AppConfig.DataAPI {
		log.Println("Data API stopped. Google sign-in and /api/config/public stay up.")
		handler = dataAPIGate(mux)
	}
	log.Printf("Backend API active on port %s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func dataAPIGate(next http.Handler) http.Handler {
	open := map[string]bool{
		"/api/auth/google":   true,
		"/api/auth/me":       true,
		"/api/auth/logout":   true,
		"/api/config/public": true,
	}
	stopped := middleware.ApplyCORS(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "data api stopped", http.StatusServiceUnavailable)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if open[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}
		stopped(w, r)
	})
}
