# Frontend Manual Test Checklist

Run these tests against the live development server (`npm run dev` on port 5173) with the backend running (port 3131).

**Legend:** `[ ]` = not tested, `[x]` = pass, `[!]` = fail (add notes)

---

## 1. Navigation & Routing

### 1.1 Homepage (`/`)

- [ ] Page loads without console errors
- [ ] Search bar is visible and focused
- [ ] Quick-link buttons render with correct colors
- [ ] Clicking a quick link opens the URL in a new tab
- [ ] Weather widget loads (or shows graceful fallback)
- [ ] News widget loads headlines
- [ ] "Mondás" quote widget shows a random quote
- [ ] Events widget shows upcoming events (if any exist)

### 1.2 News Page (`/hirek`)

- [ ] Page loads and displays news feeds
- [ ] Each feed card shows title and colored background
- [ ] RSS items display with title, date, and link
- [ ] Clicking a news item opens source in new tab

### 1.3 Directory Index (`/index`)

- [ ] Page loads and shows all directory entries
- [ ] Category tabs/links are visible
- [ ] Entry count is displayed
- [ ] Grid/list view toggle works
- [ ] Sort dropdown works (A→Z, Newest)
- [ ] "Load more" button appears and loads more entries

### 1.4 Directory Category (`/index/[category]`)

- [ ] Navigating to `/index/szolgaltatasok` shows filtered entries
- [ ] URL updates correctly
- [ ] Browser back button returns to previous category/page
- [ ] Switching between categories re-fetches and updates the list
- [ ] Breadcrumbs show correct path

### 1.5 Counties (`/megyek`)

- [ ] Page lists all counties
- [ ] Each county links to its detail page
- [ ] Settlement counts are displayed (if applicable)

### 1.6 County Detail (`/[county]-megye`)

- [ ] Page loads and shows county info
- [ ] Settlement list renders for the county
- [ ] Entries for the county are listed
- [ ] Weather widget loads for the county

### 1.7 Towns (`/varosok`)

- [ ] Page lists all towns
- [ ] Each town links to its detail page

### 1.8 Villages (`/falvak`)

- [ ] Page lists all villages
- [ ] Each village links to its detail page

### 1.9 Settlement Detail (`/[county]-megye/[slug]`) — **NAVIGATION BUG FIX**

- [ ] Page loads with correct settlement data
- [ ] Entries for the settlement are displayed
- [ ] Weather widget loads for the settlement
- [ ] News widget loads
- [ ] Events widget shows events for the settlement
- [ ] **Parent city link navigates AND the page re-renders with new data** (critical fix)
- [ ] Navigating between two settlements under the same county updates content
- [ ] Browser back button works correctly after parent-city navigation
- [ ] Breadcrumbs update on navigation

### 1.10 Entry Detail (`/bejegyzes/[slug]`) — **NAVIGATION BUG FIX**

- [ ] Page loads with correct entry data (name, phone, address, etc.)
- [ ] Breadcrumbs show correct path
- [ ] **Navigating from one entry detail to another updates the page content** (critical fix)
- [ ] Browser back button works correctly

### 1.11 Events Page (`/esemenyek`)

- [ ] Page loads and shows events
- [ ] Event cards display title, date, time, organizer
- [ ] Events are sorted by date

### 1.12 Changelog (`/valtozasnaplo`)

- [ ] Page loads and displays changelog entries

### 1.13 Map (`/terkep`)

- [ ] Page loads (placeholder or map renders)

---

## 2. Search Engine

### 2.1 Autosuggest

- [ ] Typing fewer than 3 characters shows no suggestions
- [ ] Typing 3+ characters shows a dropdown with suggestions
- [ ] Suggestions include categories, tags, and location names
- [ ] Clicking a suggestion fills the search input
- [ ] Pressing Enter triggers search

### 2.2 Search Results

- [ ] Search returns directory entries matching the query
- [ ] Results are highlighted by type (green/directory, blue/weather, orange/news)
- [ ] No results shows a fallback message with external search buttons
- [ ] Clearing the search resets the results

---

## 3. Widgets

### 3.1 Weather Widget

- [ ] Shows temperature and description for the current location
- [ ] Handles missing API key gracefully (shows error or placeholder)
- [ ] Updates when navigating to a different location

### 3.2 News Widget

- [ ] Displays RSS headlines
- [ ] Shows feed title and colored card background
- [ ] Links open in new tab

### 3.3 Events Widget

- [ ] Shows upcoming events for the current scope (location/county/all)
- [ ] Displays event date, time, title, organizer

### 3.4 Mondás Widget

- [ ] Shows a random quote from the `mondasok` table
- [ ] Changes on page refresh (or button click if available)

---

## 4. Theme / Dark Mode

- [ ] Light/dark/system toggle is visible in the UI
- [ ] Toggling to dark mode changes background, text, and card colors
- [ ] Toggling to light mode restores original colors
- [ ] System mode follows OS preference
- [ ] Theme choice persists across page reload (localStorage)
- [ ] No flash of unstyled content (FOUC) on initial page load

---

## 5. Admin Panel (`/admin`)

### 5.1 Authentication

- [ ] `/admin` shows login form
- [ ] Entering wrong password shows error
- [ ] Entering correct password grants access to dashboard
- [ ] Logout button clears session and returns to login

### 5.2 Dashboard Tabs

- [ ] All tabs are visible: Entries, Locations, Categories, Types, Quick Links, News Feeds, Mondások, Events
- [ ] Switching tabs loads the correct data

### 5.3 CRUD: Entries

- [ ] List loads all entries with name, location, category
- [ ] Create new entry with all fields (name, location, phone, URL, etc.)
- [ ] Edit existing entry — changes persist after reload
- [ ] Delete entry — removed from list

### 5.4 CRUD: Locations

- [ ] List loads all locations
- [ ] Create new location (name, county, type)
- [ ] Edit existing location
- [ ] Delete location

### 5.5 CRUD: Entry Categories

- [ ] List loads all categories
- [ ] Create new category
- [ ] Edit category name
- [ ] Delete category

### 5.6 CRUD: Entry Types

- [ ] List loads all entry types
- [ ] Create new type
- [ ] Edit type name
- [ ] Delete type

### 5.7 CRUD: Quick Links

- [ ] List loads all quick links with title, URL, color
- [ ] Create new quick link
- [ ] Edit existing quick link
- [ ] Delete quick link

### 5.8 CRUD: News Feeds

- [ ] List loads all news feeds
- [ ] Create new feed (title, feed_url, bg_color)
- [ ] Edit existing feed
- [ ] Delete feed

### 5.9 CRUD: Mondások

- [ ] List loads all quotes
- [ ] Create new quote
- [ ] Edit existing quote
- [ ] Delete quote

### 5.10 CRUD: Events

- [ ] List loads all events
- [ ] Create new event (title, location, dates, organizer)
- [ ] Edit existing event
- [ ] Delete event

---

## 6. Responsive Design

- [ ] Homepage renders correctly on mobile (< 480px)
- [ ] Navigation is accessible on mobile (hamburger or scroll)
- [ ] Directory cards stack vertically on small screens
- [ ] Admin panel is usable on tablet (768px+)
- [ ] No horizontal scroll on any page at standard breakpoints

---

## 7. Error Handling

- [ ] API timeout shows user-friendly error message
- [ ] Invalid route shows 404 page or redirect
- [ ] Network failure shows fallback UI (not blank page)
