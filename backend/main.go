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
	if err := handlers.EnsureEventImagesDir(); err != nil {
		log.Printf("event images directory: %v", err)
	}
	if err := handlers.EnsureEntryImagesDir(); err != nil {
		log.Printf("entry images directory: %v", err)
	}
	handlers.MigrateSettlementLocationTypes()
	db.SeedHistoricalSeatsContent()
	mondasok.Migrate()
	settings.MigrateSiteSettings()
	weather.MigrateWeatherTranslations()
	pages.MigratePages()
	pagefaq.Migrate()
	events.Migrate()
	handlers.MigrateEntryTypes()
	handlers.MigrateEntryCategories()
	handlers.MigrateEntryAvailability()
	handlers.MigrateEntryVerified()
	auth.Migrate()
	account.Migrate()

	mux := http.DefaultServeMux
	pub := middleware.ApplyCORS
	admin := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.ApplyCORS(auth.RequireAdmin(h))
	}

	mux.HandleFunc("/api/auth/google", pub(auth.HandleGoogleLogin))
	mux.HandleFunc("/api/auth/me", pub(auth.HandleMe))
	mux.HandleFunc("/api/auth/logout", pub(auth.HandleLogout))
	mux.HandleFunc("/api/account/preferences", pub(account.HandlePreferences))
	mux.HandleFunc("/api/account/import", pub(account.HandleImport))
	mux.HandleFunc("/api/account/links", pub(account.HandleLinks))
	mux.HandleFunc("/api/account/history", pub(account.HandleHistory))
	mux.HandleFunc("/api/account/favorites", pub(account.HandleFavorites))
	mux.HandleFunc("/api/account/listings/claim", pub(account.HandleClaimListing))
	mux.HandleFunc("/api/account/listings/catalog", pub(account.HandleListingCatalog))
	mux.HandleFunc("/api/account/listings/members", pub(account.HandleListingMembers))
	mux.HandleFunc("/api/account/listings", pub(account.HandleListings))

	// Core Module (Always Enabled)
	mux.HandleFunc("/api/entries", pub(handlers.EntriesHandler))
	mux.HandleFunc("/api/directory", pub(handlers.EntriesHandler))
	mux.HandleFunc("/api/entry", pub(handlers.EntryDetailHandler))
	mux.HandleFunc("/api/entry/related", pub(handlers.HandleEntryRelated))
	mux.HandleFunc("/api/locations", pub(handlers.HandleAdminLocations))
	mux.HandleFunc("/api/admin/listing-queue", admin(account.HandleListingQueue))
	mux.HandleFunc("/api/admin/listing-queue/publish", admin(account.HandleListingQueuePublish))
	mux.HandleFunc("/api/admin/listing-queue/member", admin(account.HandleListingQueueMember))
	mux.HandleFunc("/api/admin/entries", admin(handlers.HandleAdminEntries))
	mux.HandleFunc("/api/admin/entry-images", admin(handlers.HandleEntryImageUpload))
	mux.Handle("/api/media/entry-images/", http.StripPrefix("/api/media/entry-images/", http.FileServer(http.Dir(handlers.EntryImagesDir()))))
	mux.HandleFunc("/api/admin/entry_categories", admin(handlers.HandleAdminEntryCategories))
	mux.HandleFunc("/api/admin/entry_types", admin(handlers.HandleAdminEntryTypes))
	mux.HandleFunc("/api/admin/locations", admin(handlers.HandleAdminLocations))
	mux.HandleFunc("/api/admin/dashboard_stats", admin(handlers.HandleAdminDashboardStats))
	mux.HandleFunc("/api/settlement_location_types", pub(handlers.HandlePublicSettlementLocationTypes))
	mux.HandleFunc("/api/admin/settlement_location_types", admin(handlers.HandleAdminSettlementLocationTypes))
	mux.HandleFunc("/api/admin/county_seat", admin(handlers.HandleSetCountySeat))
	mux.HandleFunc("/api/attractions", pub(handlers.HandleAttractions))
	mux.HandleFunc("/api/historical_seats", pub(handlers.HandleHistoricalSeats))
	mux.HandleFunc("/api/counties", pub(handlers.HandleCounties))
	mux.HandleFunc("/api/admin/counties", admin(handlers.HandleAdminCounties))
	mux.HandleFunc("/api/admin/historical_seats", admin(handlers.HandleAdminHistoricalSeats))
	mux.HandleFunc("/api/admin/attractions", admin(handlers.HandleAdminAttractions))

	// Public config (weather cache TTL, version) + admin settings
	mux.HandleFunc("/api/config/public", pub(settings.HandlePublicConfig))
	mux.HandleFunc("/api/admin/settings", admin(settings.HandleAdminSettings))
	mux.HandleFunc("/api/admin/settings/clear-weather-cache", admin(settings.ClearWeatherCache))

	// Pages (public + admin)
	mux.HandleFunc("/api/pages", pub(pages.HandlePublicPage))
	mux.HandleFunc("/api/admin/pages", admin(pages.HandleAdminPages))
	mux.HandleFunc("/api/page_faq", pub(pagefaq.HandlePublic))
	mux.HandleFunc("/api/admin/page_faq", admin(pagefaq.HandleAdmin))

	// Optional Modules
	if config.AppConfig.Features.Weather {
		mux.HandleFunc("/api/weather", pub(weather.HandleWeather))
		mux.HandleFunc("/api/weather/county", pub(weather.HandleCountyWeather))
		mux.HandleFunc("/api/admin/weather_translations", admin(weather.HandleAdminWeatherTranslations))
		log.Println("Module [Weather] enabled")
	}

	if config.AppConfig.Features.Events {
		mux.HandleFunc("/api/events", pub(events.HandleEvents))
		mux.HandleFunc("/api/events/filter-options", pub(events.HandleEventFilterOptions))
		mux.HandleFunc("/api/events/detail", pub(events.HandleEventDetail))
		mux.HandleFunc("/api/venues", pub(venues.HandlePublic))
		mux.HandleFunc("/api/venue_types", pub(venues.HandlePublicVenueTypes))
		mux.HandleFunc("/api/admin/events", admin(events.HandleAdminEvents))
		mux.HandleFunc("/api/admin/catalog_event_types", admin(events.HandleAdminCatalogEventTypes))
		mux.HandleFunc("/api/admin/catalog_event_subtypes", admin(events.HandleAdminCatalogEventSubtypes))
		mux.HandleFunc("/api/admin/event-images", admin(handlers.HandleEventImageUpload))
		mux.Handle("/api/media/event-images/", http.StripPrefix("/api/media/event-images/", http.FileServer(http.Dir(handlers.EventImagesDir()))))
		mux.HandleFunc("/api/admin/events/schedule", admin(events.HandleAdminEventSchedule))
		mux.HandleFunc("/api/admin/venues", admin(venues.HandleAdmin))
		mux.HandleFunc("/api/admin/venue_types", admin(venues.HandleAdminVenueTypes))
		log.Println("Module [Events] enabled")
	}

	if config.AppConfig.Features.News {
		mux.HandleFunc("/api/news", pub(news.HandleNews))
		mux.HandleFunc("/api/news/feeds", pub(news.HandlePublicNewsFeeds))
		mux.HandleFunc("/api/admin/news_feeds", admin(news.HandleAdminNewsFeeds))
		log.Println("Module [News] enabled")
	}

	if config.AppConfig.Features.Mondasok {
		mux.HandleFunc("/api/mondasok", pub(mondasok.HandlePublicMondasok))
		mux.HandleFunc("/api/admin/mondasok", admin(mondasok.HandleAdminMondasok))
		log.Println("Module [Mondasok] enabled")
	}

	if config.AppConfig.Features.QuickLinks {
		mux.HandleFunc("/api/quick_links", pub(links.HandlePublicQuickLinks))
		mux.HandleFunc("/api/admin/quick_links", admin(links.HandleAdminQuickLinks))
		log.Println("Module [QuickLinks] enabled")
	}

	if config.AppConfig.Features.Search {
		mux.HandleFunc("/api/search", pub(search.HandleUnifiedSearch))
		mux.HandleFunc("/api/proxy", pub(search.ProxyHandler))
		mux.HandleFunc("/api/autosuggest", pub(search.HandleAutosuggest))
		log.Println("Module [Search] enabled")
	}

	port := config.AppConfig.Port
	log.Printf("Backend API active on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
