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
	ID               int    `json:"id"`
	LocationID       int    `json:"location_id"`
	LocationName     string `json:"location_name"`
	LocationSlug     string `json:"location_slug"`
	County           string `json:"county"`
	CountySlug       string `json:"county_slug"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	StartDate        string `json:"start_date"`
	StartTime        string `json:"start_time"`
	EndDate          string `json:"end_date"`
	EndTime          string `json:"end_time"`
	EventType        string `json:"event_type"`
	Organizer        string `json:"organizer"`
	LocationType     string `json:"location_type"`
	DefaultVenueID     *int   `json:"default_venue_id,omitempty"`
	DefaultVenueName   string `json:"default_venue_name,omitempty"`
	DefaultVenueSlug   string `json:"default_venue_slug,omitempty"`
	// HasSchedule is true when the event has at least one napi program day (same idea as the detail #program block).
	HasSchedule bool `json:"has_schedule"`
}

type AdminEvent struct {
	ID               int    `json:"id"`
	LocationID       *int   `json:"location_id"`
	DefaultVenueID   *int   `json:"default_venue_id"`
	DefaultVenueName string `json:"default_venue_name,omitempty"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	StartDate        string `json:"start_date"`
	StartTime        string `json:"start_time"`
	EndDate          string `json:"end_date"`
	EndTime          string `json:"end_time"`
	EventType        string `json:"event_type"`
	Organizer        string `json:"organizer"`
}

// VenueType is a configurable label for venues.kind (slug identifies the row).
type VenueType struct {
	ID      int    `json:"id"`
	Slug    string `json:"slug"`
	LabelHu string `json:"label_hu"`
}

// Venue is a named site within a settlement (arena, market square, etc.).
// Name is the Hungarian (primary) label; name_ro / name_de are optional translations.
type Venue struct {
	ID              int      `json:"id"`
	SettlementID    int      `json:"settlement_id"`
	Name            string   `json:"name"`
	NameRO          string   `json:"name_ro,omitempty"`
	NameDE          string   `json:"name_de,omitempty"`
	Slug            string   `json:"slug"`
	Kind            string   `json:"kind"`
	KindLabel       string   `json:"kind_label,omitempty"`
	Address         string   `json:"address,omitempty"`
	Notes           string   `json:"notes,omitempty"`
	Latitude        *float64 `json:"latitude,omitempty"`
	Longitude       *float64 `json:"longitude,omitempty"`
	SeatingCapacity *int     `json:"seating_capacity,omitempty"`
	Description     string   `json:"description,omitempty"`
	SettlementName  string   `json:"settlement_name,omitempty"`
	SettlementSlug  string   `json:"settlement_slug,omitempty"`
	CountyName      string   `json:"county_name,omitempty"`
	CountySlug      string   `json:"county_slug,omitempty"`
}

// EventScheduleActivity is one line in the per-day program (optional schedule).
// ActivityType: opening | match | closing | other (sport blocks, ceremonies, etc.).
// StartsAt / EndsAt may be empty; ends_at often omitted for matches with unknown finish time.
type EventScheduleActivity struct {
	ID           int    `json:"id"`
	ActivityType string `json:"activity_type"`
	StartsAt     string `json:"starts_at"`
	EndsAt       string `json:"ends_at"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	SortOrder    int    `json:"sort_order"`
	VenueID      *int   `json:"venue_id,omitempty"`
	VenueName    string `json:"venue_name,omitempty"`
	VenueSlug    string `json:"venue_slug,omitempty"`
}

// EventScheduleDay is one calendar day in the optional program.
type EventScheduleDay struct {
	ID           int                     `json:"id"`
	ScheduleDate string                  `json:"schedule_date"`
	Notes        string                  `json:"notes"`
	SortOrder    int                     `json:"sort_order"`
	Activities   []EventScheduleActivity `json:"activities"`
}

// EventWithSchedule is the public detail payload (event + optional schedule).
type EventWithSchedule struct {
	Event
	Schedule []EventScheduleDay `json:"schedule"`
}

type Mondas struct {
	ID          int    `json:"id"`
	Text        string `json:"text"`
	DisplayDate string `json:"display_date"` // YYYY-MM-DD — day the quote is shown on the homepage
	CreatedAt   string `json:"created_at"`
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
	LocationID *int     `json:"location_id"`
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
