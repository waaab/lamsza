package main

import (
	"log"
	"net/http"
	"time"

	"backend/internal/account"
	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/events"
	"backend/internal/handlers"
	"backend/internal/health"
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
	handlers.MigrateEntrySuggestions()
	handlers.MigrateEntryProfileView()
	handlers.MigrateEntryLocationSearch()
	handlers.MigrateSettlementLocationTypes()
	events.Migrate()
	handlers.MigrateAttractions()
	auth.Migrate()
	account.Migrate()
	account.MigrateWebsites()
	// Schema and seed only. The directory wipe is now a hand-run one-shot:
	// migrations/0001_directory_catalog_v2.sql.
	handlers.MigrateDirectoryCatalog()

	mux := http.DefaultServeMux

	mux.HandleFunc("/api/health", middleware.ApplyCORS(health.HandleHealth))

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
	mux.HandleFunc("/api/account/websites/lookup", middleware.ApplyCORS(account.HandleWebsiteLookup))
	mux.HandleFunc("/api/account/websites", middleware.ApplyCORS(account.HandleAccountWebsites))
	mux.HandleFunc("/api/websites", middleware.ApplyCORS(account.HandleWebsites))

	// Core Module (Always Enabled)
	mux.HandleFunc("/api/entries", middleware.ApplyCORS(handlers.EntriesHandler))
	mux.HandleFunc("/api/directory", middleware.ApplyCORS(handlers.EntriesHandler))
	mux.HandleFunc("/api/entry-categories", middleware.ApplyCORS(handlers.HandlePublicEntryCategories))
	mux.HandleFunc("/api/entry", middleware.ApplyCORS(handlers.EntryDetailHandler))
	mux.HandleFunc("/api/entry/related", middleware.ApplyCORS(handlers.HandleEntryRelated))
	mux.HandleFunc("/api/entry/reviews", middleware.ApplyCORS(handlers.HandleEntryReviews))
	mux.HandleFunc("/api/entry/suggestion-form", middleware.ApplyCORS(account.HandleSuggestionForm))
	mux.HandleFunc("/api/entry/suggestions", middleware.ApplyCORS(account.HandleEntrySuggestions))
	mux.HandleFunc("/api/locations", middleware.ApplyCORS(handlers.HandlePublicLocations))
	mux.HandleFunc("/api/settlement_location_types", middleware.ApplyCORS(handlers.HandlePublicSettlementLocationTypes))
	// Media: public read from shared dirs; admin uploads live in lamsza-admin
	mux.Handle("/api/media/entry-images/", middleware.ApplyCORS(http.StripPrefix("/api/media/entry-images/", middleware.NoDirListing(http.FileServer(http.Dir(handlers.EntryImagesDir())))).ServeHTTP))
	mux.Handle("/api/media/event-images/", middleware.ApplyCORS(http.StripPrefix("/api/media/event-images/", middleware.NoDirListing(http.FileServer(http.Dir(handlers.EventImagesDir())))).ServeHTTP))
	mux.HandleFunc("/api/attractions", middleware.ApplyCORS(handlers.HandleAttractions))
	mux.HandleFunc("/api/attraction-suggestions", middleware.ApplyCORS(handlers.HandleAttractionSuggestions))
	mux.HandleFunc("/api/historical_seats", middleware.ApplyCORS(handlers.HandleHistoricalSeats))
	mux.HandleFunc("/api/counties", middleware.ApplyCORS(handlers.HandleCounties))

	// Public config (weather cache TTL, version)
	mux.HandleFunc("/api/config/public", middleware.ApplyCORS(settings.HandlePublicConfig))

	// Pages (public)
	mux.HandleFunc("/api/pages", middleware.ApplyCORS(pages.HandlePublicPage))
	mux.HandleFunc("/api/page_faq", middleware.ApplyCORS(pagefaq.HandlePublic))

	// Optional Modules
	if config.AppConfig.Features.Weather {
		mux.HandleFunc("/api/weather", middleware.ApplyCORS(weather.HandleWeather))
		mux.HandleFunc("/api/weather/county", middleware.ApplyCORS(weather.HandleCountyWeather))
		log.Println("Module [Weather] enabled")
	}

	if config.AppConfig.Features.Events {
		mux.HandleFunc("/api/events", middleware.ApplyCORS(events.HandleEvents))
		mux.HandleFunc("/api/events/filter-options", middleware.ApplyCORS(events.HandleEventFilterOptions))
		mux.HandleFunc("/api/events/detail", middleware.ApplyCORS(events.HandleEventDetail))
		mux.HandleFunc("/api/venues", middleware.ApplyCORS(venues.HandlePublic))
		mux.HandleFunc("/api/venue_types", middleware.ApplyCORS(venues.HandlePublicVenueTypes))
		log.Println("Module [Events] enabled")
	}

	if config.AppConfig.Features.News {
		mux.HandleFunc("/api/news", middleware.ApplyCORS(news.HandleNews))
		mux.HandleFunc("/api/news/feeds", middleware.ApplyCORS(news.HandlePublicNewsFeeds))
		news.WarmNews()
		log.Println("Module [News] enabled")
	}

	if config.AppConfig.Features.Mondasok {
		mux.HandleFunc("/api/mondasok", middleware.ApplyCORS(mondasok.HandlePublicMondasok))
		log.Println("Module [Mondasok] enabled")
	}

	if config.AppConfig.Features.QuickLinks {
		mux.HandleFunc("/api/quick_links", middleware.ApplyCORS(links.HandlePublicQuickLinks))
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
	srv := newServer(":"+port, middleware.LimitBody(handler))
	log.Printf("Backend API active on port %s\n", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// newServer builds the API server with its timeouts.
//
// A bare http.ListenAndServe has no timeouts at all, so a half-open or slow
// connection is never dropped and a few hundred of them hold the server
// forever (slowloris).
//
// WriteTimeout starts when the request headers are read, so it has to be
// longer than the slowest handler. The slowest are /api/proxy (15s client
// timeout) and /api/weather/county (fan-out budget plus one city).
func newServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
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
