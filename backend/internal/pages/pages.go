package pages

import (
	"backend/internal/db"
	"encoding/json"
	"log"
	"net/http"
)

type Page struct {
	ID        int    `json:"id"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Greeting  string `json:"greeting"`
	Content   string `json:"content"`
	UpdatedAt string `json:"updated_at"`
}

func MigratePages() {
	_, err := db.DB.Exec(`
		CREATE TABLE IF NOT EXISTS pages (
			id SERIAL PRIMARY KEY,
			slug VARCHAR(255) NOT NULL UNIQUE,
			title VARCHAR(255) NOT NULL DEFAULT '',
			content TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Printf("pages create: %v", err)
		return
	}
	_, _ = db.DB.Exec(`ALTER TABLE pages ADD COLUMN IF NOT EXISTS greeting TEXT NOT NULL DEFAULT ''`)

	seeds := []struct{ slug, title, greeting string }{
		{"home", "Na Lámsza!", "Erdélyi magyar startlap és kereső. Az internet székely kapuja."},
		{"esemenyek", "Székelyföldi események", "Válogass a legfrissebb székelyföldi események közül."},
		{"hirek", "Friss hírek erdélyi forrásból", "Helyi hírcsatornák legfrissebb hírei időrendben."},
		{"megyek", "Székelyföldi Megyék", "Válassz megyét a települések és tartalmak böngészéséhez."},
		{"varosok", "Székelyföldi Városok", "Székelyföldi városok listája megyénként."},
		{"falvak", "Székelyföldi Falvak", "Falvak és községek listája megyénként."},
		{"szekek", "Székelyföld történelmi székei", "A székely székek és a hozzájuk kapcsolódó megyék."},
		{"index", "Index", "Minden, ami helyi, egy helyen: szakemberek, intézmények, szolgáltatások."},
		{"index/szolgaltatasok", "Szolgáltatások", "Az index szolgáltatástípusú bejegyzései - nem a teljes címtár."},
		{"index/weboldalak", "Weboldalak", "Ellenőrzött, hiteles, előnyben részesített helyi weboldalak az indexen."},
		{"terkep", "Székelyföld Térkép", "Hamarosan érkezik az interaktív térképünk helyi adatokkal!"},
		{"valtozasnaplo", "Változásnapló", "Újítások, javítások - emberi nyelven."},
		{"iranyelvek", "Irányelvek", "Adatvédelem, sütik és felhasználási feltételek - összefoglaló."},
		{"iranyelvek/adatvedelem", "Adatvédelem", "Milyen adatot kezelünk, miért, mennyi ideig - és milyen jogai vannak."},
		{"iranyelvek/sutik", "Sütik", "Hogyan használjuk a sütiket és mire valók."},
		{"iranyelvek/feltetelek", "Feltételek", "A szolgáltatás igénybevételének feltételei."},
	}
	for _, s := range seeds {
		_, _ = db.DB.Exec(
			`INSERT INTO pages (slug, title, greeting) VALUES ($1, $2, $3) ON CONFLICT (slug) DO NOTHING`,
			s.slug, s.title, s.greeting,
		)
	}
	// Meglévő sorok: üres bevezető feltöltése (egyszeri jelleggel üres greeting esetén).
	for _, s := range seeds {
		_, _ = db.DB.Exec(
			`UPDATE pages SET greeting = $2 WHERE slug = $1 AND COALESCE(TRIM(greeting), '') = ''`,
			s.slug, s.greeting,
		)
	}
	log.Println("Pages table ready")
}

// HandlePublicPage serves a single page by slug (GET /api/pages?slug=...)
func HandlePublicPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Missing slug parameter", http.StatusBadRequest)
		return
	}
	var p Page
	err := db.DB.QueryRow(
		"SELECT id, slug, title, greeting, content, updated_at::text FROM pages WHERE slug = $1", slug,
	).Scan(&p.ID, &p.Slug, &p.Title, &p.Greeting, &p.Content, &p.UpdatedAt)
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}
