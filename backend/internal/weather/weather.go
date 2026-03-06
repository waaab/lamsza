package weather

import (
	"backend/internal/db"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func HandleWeather(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	apiKey := r.URL.Query().Get("appid")

	if slug == "" || apiKey == "" {
		http.Error(w, "Missing slug or appid", http.StatusBadRequest)
		return
	}

	var name, nameRo string
	err := db.DB.QueryRow("SELECT name, COALESCE(name_ro, '') FROM locations WHERE slug = $1", slug).Scan(&name, &nameRo)
	if err != nil {
		log.Println("Weather slug resolution error:", err)
		err = db.DB.QueryRow("SELECT name, COALESCE(name_ro, '') FROM locations WHERE unaccent(name) ILIKE $1 OR unaccent(name_ro) ILIKE $1 LIMIT 1", slug).Scan(&name, &nameRo)
		if err != nil {
			name = strings.Title(strings.ReplaceAll(slug, "-", " "))
		}
	}

	searchName := name
	if nameRo != "" {
		searchName = nameRo
	}

	weatherURL := "https://api.openweathermap.org/data/2.5/weather?q=" + url.QueryEscape(searchName) + "&appid=" + apiKey + "&units=metric&lang=hu"
	resp, err := http.Get(weatherURL)
	if err != nil {
		http.Error(w, "Weather provider error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
