package handlers

// dashboardCountTables maps dashboard card id → physical table name (fixed whitelist only).
var dashboardCountTables = []struct {
	Key   string
	Table string
}{
	{"mondasok", "mondasok"},
	{"quicklinks", "quick_links"},
	{"newsfeeds", "news_feeds"},
	{"locations", "locations"},
	{"counties", "counties"},
	{"venues", "venues"},
	{"attractions", "attractions"},
	{"events", "events"},
	{"websites", "websites"},
	{"entries", "entries"},
	{"entry_categories", "entry_categories"},
	{"entry_types", "entry_types"},
	{"tags", "tags"},
	{"pages", "pages"},
	{"page_faq", "page_faq_sections"},
	{"weather_translations", "weather_desc_translations"},
	{"users", "users"},
	{"settings", "site_settings"},
}
