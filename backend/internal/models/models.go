package models

type Entry struct {
	ID             string   `json:"id"`
	Type           string   `json:"type"`
	Category       string   `json:"category"`
	Name           string   `json:"name"`
	Slug           string   `json:"slug"`
	Location       string   `json:"location"`
	LocationSlug   string   `json:"location_slug"`
	LocationCounty string   `json:"location_county"`
	CountySlug     string   `json:"county_slug"`
	LocationType   string   `json:"location_type"`
	LocationRo     string   `json:"location_ro"`
	LocationDe     string   `json:"location_de"`
	Phone          string   `json:"phone"`
	Address        string   `json:"address"`
	Notes          string   `json:"notes"`
	Tags           []string `json:"tags"`
	Languages      []string `json:"languages"`
	URL            string   `json:"url"`
	IsDirectMatch  bool     `json:"is_direct_match"`
}

type Event struct {
	ID           int    `json:"id"`
	LocationID   int    `json:"location_id"`
	LocationName string `json:"location_name"`
	LocationSlug string `json:"location_slug"`
	County       string `json:"county"`
	CountySlug   string `json:"county_slug"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	StartDate    string `json:"start_date"`
	StartTime    string `json:"start_time"`
	EndDate      string `json:"end_date"`
	EndTime      string `json:"end_time"`
	EventType    string `json:"event_type"`
	Organizer    string `json:"organizer"`
	LocationType string `json:"location_type"`
}

type AdminEvent struct {
	ID          int    `json:"id"`
	LocationID  int    `json:"location_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
	StartTime   string `json:"start_time"`
	EndDate     string `json:"end_date"`
	EndTime     string `json:"end_time"`
	EventType   string `json:"event_type"`
	Organizer   string `json:"organizer"`
}

type Mondas struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

type QuickLink struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	BgColor string `json:"bg_color"`
}

type NewsFeed struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	FeedURL string `json:"feed_url"`
	BgColor string `json:"bg_color"`
}

type Location struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	NameRo       string `json:"name_ro"`
	NameDe       string `json:"name_de"`
	County       string `json:"county"`
	CountySlug   string `json:"county_slug"`
	Type         string `json:"type"`
	Slug         string `json:"slug"`
	PostCode     string `json:"post_code"`
	Coordinates  string `json:"coordinates"`
	Population   string `json:"population"`
	Area         string `json:"area"`
	Crest        string `json:"crest"`
	ParentID     *int   `json:"parent_id"`
	IsCountySeat bool   `json:"is_county_seat"`
}

type AdminEntry struct {
	ID         int      `json:"id"`
	Type       string   `json:"type"`
	LocationID int      `json:"location_id"`
	CategoryID *int     `json:"category_id"`
	Category   string   `json:"category"`
	Name       string   `json:"name"`
	Slug       string   `json:"slug"`
	URL        string   `json:"url"`
	Phone      string   `json:"phone"`
	Address    string   `json:"address"`
	Notes      string   `json:"notes"`
	Languages  []string `json:"languages"`
	Tags       []string `json:"tags"`
}

type EntryCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type EntryType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
