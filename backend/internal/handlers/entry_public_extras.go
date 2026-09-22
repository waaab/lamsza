package handlers

import (
	"backend/internal/db"
	"backend/internal/models"
	"encoding/json"
	"strconv"
)

// ApplyPublicEntryExtras enforces public listing vs claimed-extras on an Entry.
// Unclaimed public listings must not expose photos, hours, or ratings even if admin data exists.
// Claimed listings keep stored extras and gain ratings_enabled.
func ApplyPublicEntryExtras(e *models.Entry, viewerUserID int) {
	if !e.Claimed {
		// Unclaimed: strip photos, hours, delivery_hours, ratings
		e.Photos = json.RawMessage("[]")
		e.Hours = json.RawMessage("{}")
		e.DeliveryHours = json.RawMessage("{}")
		e.RatingsEnabled = false
		return
	}

	// Claimed: scan ratings_enabled from DB
	var ratingsEnabled bool
	entryIDInt, _ := strconv.Atoi(e.ID)
	err := db.DB.QueryRow(`SELECT COALESCE(ratings_enabled, false) FROM entries WHERE id = $1`, entryIDInt).Scan(&ratingsEnabled)
	if err != nil {
		e.RatingsEnabled = false
		return
	}
	e.RatingsEnabled = ratingsEnabled

	if !ratingsEnabled {
		return
	}

	// Load rating and review count
	var avgRating *float64
	var reviewCount int
	err = db.DB.QueryRow(`
		SELECT 
			ROUND(AVG(score)::numeric, 1),
			COUNT(*)
		FROM entry_reviews
		WHERE entry_id = $1
	`, entryIDInt).Scan(&avgRating, &reviewCount)
	if err == nil {
		e.Rating = avgRating
		e.ReviewCount = reviewCount
	}

	// Load up to 50 reviews, newest first
	rows, err := db.DB.Query(`
		SELECT 
			r.id,
			COALESCE(NULLIF(u.name, ''), 'Felhasználó'),
			COALESCE(u.photo, ''),
			r.score,
			r.body,
			r.created_at,
			r.updated_at
		FROM entry_reviews r
		JOIN users u ON u.id = r.user_id
		WHERE r.entry_id = $1
		ORDER BY r.created_at DESC
		LIMIT 50
	`, entryIDInt)
	if err != nil {
		return
	}
	defer rows.Close()

	var reviews []models.PublicReview
	for rows.Next() {
		var r models.PublicReview
		var id int
		if err := rows.Scan(&id, &r.AuthorName, &r.AuthorPhoto, &r.Score, &r.Text, &r.CreatedAt, &r.UpdatedAt); err == nil {
			r.ID = strconv.Itoa(id)
			reviews = append(reviews, r)
		}
	}
	e.Reviews = reviews

	// Load my_review if viewerUserID > 0
	if viewerUserID > 0 {
		var r models.PublicReview
		var id int
		err = db.DB.QueryRow(`
			SELECT 
				r.id,
				COALESCE(NULLIF(u.name, ''), 'Felhasználó'),
				COALESCE(u.photo, ''),
				r.score,
				r.body,
				r.created_at,
				r.updated_at
			FROM entry_reviews r
			JOIN users u ON u.id = r.user_id
			WHERE r.entry_id = $1 AND r.user_id = $2
		`, entryIDInt, viewerUserID).Scan(&id, &r.AuthorName, &r.AuthorPhoto, &r.Score, &r.Text, &r.CreatedAt, &r.UpdatedAt)
		if err == nil {
			r.ID = strconv.Itoa(id)
			e.MyReview = &r
		}
	}
}
