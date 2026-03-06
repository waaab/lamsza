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
	"backend/internal/search"
	"backend/internal/weather"
)

func main() {
	config.Load()
	db.InitDB()

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

	// Optional Modules
	if config.AppConfig.Features.Weather {
		mux.HandleFunc("/api/weather", middleware.ApplyCORS(weather.HandleWeather))
		log.Println("Module [Weather] enabled")
	}

	if config.AppConfig.Features.Events {
		mux.HandleFunc("/api/events", middleware.ApplyCORS(events.HandleEvents))
		mux.HandleFunc("/api/admin/events", middleware.ApplyCORS(events.HandleAdminEvents))
		log.Println("Module [Events] enabled")
	}

	if config.AppConfig.Features.News {
		mux.HandleFunc("/api/county_news", middleware.ApplyCORS(news.HandleCountyNews))
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
