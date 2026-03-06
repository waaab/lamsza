package news

import (
	"backend/internal/db"
	"backend/internal/models"
	"encoding/json"
	"net/http"
)

func HandleAdminNewsFeeds(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, err := db.DB.Query("SELECT id, title, feed_url, COALESCE(bg_color, '#ffebd6') FROM news_feeds ORDER BY id ASC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var res []models.NewsFeed
		for rows.Next() {
			var nf models.NewsFeed
			if err := rows.Scan(&nf.ID, &nf.Title, &nf.FeedURL, &nf.BgColor); err == nil {
				res = append(res, nf)
			}
		}
		if res == nil {
			res = []models.NewsFeed{}
		}
		json.NewEncoder(w).Encode(res)

	case "POST":
		var nf models.NewsFeed
		if err := json.NewDecoder(r.Body).Decode(&nf); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		bgColor := nf.BgColor
		if bgColor == "" {
			bgColor = "#ffebd6"
		}
		err := db.DB.QueryRow("INSERT INTO news_feeds (title, feed_url, bg_color) VALUES ($1, $2, $3) RETURNING id", nf.Title, nf.FeedURL, bgColor).Scan(&nf.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		nf.BgColor = bgColor
		json.NewEncoder(w).Encode(nf)

	case "PUT":
		var nf models.NewsFeed
		if err := json.NewDecoder(r.Body).Decode(&nf); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, err := db.DB.Exec("UPDATE news_feeds SET title=$1, feed_url=$2, bg_color=$3 WHERE id=$4",
			nf.Title, nf.FeedURL, nf.BgColor, nf.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id != "" {
			db.DB.Exec("DELETE FROM news_feeds WHERE id = $1", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func HandleCountyNews(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Missing slug", 400)
		return
	}

	rows, err := db.DB.Query("SELECT id, title, feed_url, bg_color FROM news_feeds WHERE county_slug = $1 OR county_slug IS NULL", slug)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var feeds []models.NewsFeed
	for rows.Next() {
		var f models.NewsFeed
		if err := rows.Scan(&f.ID, &f.Title, &f.FeedURL, &f.BgColor); err == nil {
			feeds = append(feeds, f)
		}
	}
	json.NewEncoder(w).Encode(feeds)
}
