# Weather Multi-Provider, Caching & Admin Controls – Implementation Plan

## Overview

**Main goal:** Keep weather data as fresh as possible for all visitors on lamsza.com.

**Scope:**
1. Multi-provider weather (Open-Meteo, WeatherAPI.com, OpenWeatherMap)
2. Caching mechanisms + admin controls
3. Source and timestamp on weather card
4. Current date/time widget on homepage

---

## Part 1: Multi-Provider Weather (3 APIs)

### 1.1 Provider Summary

| Provider        | API Key | Free Tier              | Notes                          |
|----------------|---------|------------------------|--------------------------------|
| Open-Meteo     | No      | ~10k req/day           | No key, geocoding by name      |
| WeatherAPI.com | Yes     | 1M calls/month (~33k/d)| Key from weatherapi.com        |
| OpenWeatherMap | Yes     | 60/min, 1k/day         | Already integrated             |

### 1.2 Backend Changes

**Config** (`backend/internal/config/config.go`):
- Add `WeatherAPIKey` (OpenWeatherMap) – already exists
- Add `WeatherAPIComKey` (WeatherAPI.com)
- Open-Meteo: no key

**Weather package** (`backend/internal/env` or `config`):
- Add `WEATHER_API_COM_KEY` to `.env`

**Strategy:** Try providers in order; admin sets default (tried first) and can enable/disable each.
- Default provider is tried first
- Then other enabled providers as fallbacks
- Disabled providers are skipped

**Unified response shape** (for frontend):
```json
{
  "temp": 5,
  "temp_min": 2,
  "desc": "részben felhős",
  "icon": "02d",
  "source": "Open-Meteo",
  "fetched_at": 1709827200
}
```

**Icon mapping:** Each provider uses different icon codes; backend normalizes to OpenWeatherMap-style (`01d`, `02d`, etc.) for frontend compatibility.

**Provider selection (admin-controlled):** Backend reads `site_settings` on each request (or caches briefly):
1. Get `weather_provider_default` and enabled flags
2. Build try order: default first, then other enabled providers
3. Try each until one returns valid data
4. If default is disabled, skip it and use first enabled provider

### 1.3 What You Need To Do

1. **WeatherAPI.com:** Sign up at https://www.weatherapi.com/ and add to `.env`:
   ```
   WEATHER_API_COM_KEY=your_key_here
   ```
2. **OpenWeatherMap:** Already configured via `WEATHER_API_KEY`
3. **Open-Meteo:** No action needed

---

## Part 2: Caching & Admin Controls

### 2.1 Architecture

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ WeatherWidget│────▶│ GET /api/config  │     │ site_settings   │
│ (frontend)   │     │ (public, no auth)│────▶│ (DB table)      │
└─────────────┘     └──────────────────┘     └─────────────────┘
       │                        │
       │                        │ Returns: weather_cache_ttl_minutes,
       │                        │          weather_cache_version
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│ localStorage: weather_cache_{slug}                            │
│ { temp, desc, icon, source, timestamp, cache_version }       │
│ Valid if: (now - timestamp) < TTL AND cache_version matches  │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Database

**New table:** `site_settings`

```sql
CREATE TABLE site_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed
INSERT INTO site_settings (key, value) VALUES
  ('weather_cache_ttl_minutes', '15'),
  ('weather_cache_version', '1'),
  ('weather_active_users_estimate', '10000'),
  ('weather_provider_default', 'open_meteo'),
  ('weather_provider_open_meteo_enabled', 'true'),
  ('weather_provider_weatherapi_enabled', 'true'),
  ('weather_provider_openweathermap_enabled', 'true')
ON CONFLICT (key) DO NOTHING;
```

### 2.3 Backend Endpoints

| Endpoint | Auth | Purpose |
|----------|------|---------|
| `GET /api/config/public` | No | Returns `{ weather_cache_ttl_minutes, weather_cache_version }` for frontend |
| `GET /api/admin/settings` | Admin | All settings |
| `PUT /api/admin/settings` | Admin | Update settings |
| `POST /api/admin/settings/clear-weather-cache` | Admin | Increment `weather_cache_version` |

### 2.4 Smart TTL (Optional Phase 2)

- Track weather API requests per minute in memory
- `TTL = active_users_estimate / rate_limit` (e.g. 10000/60 ≈ 167 min for free tier)
- Or: if `requests_this_minute > 0.8 * limit`, auto-increase TTL
- Admin can override with manual TTL

### 2.5 Admin UI

**New tab or section: "Beállítások" (Settings)**

**Időjárás (Weather):**
- **Alapértelmezett szolgáltató:** dropdown – Open-Meteo | WeatherAPI.com | OpenWeatherMap (which to try first)
- **Szolgáltatók engedélyezése:** checkboxes for each provider (Open-Meteo, WeatherAPI.com, OpenWeatherMap)
- **Időjárás cache TTL (perc):** number input, default 15
- **Aktív felhasználók becslése:** number (for smart TTL), default 10000
- **"Időjárás cache törlése"** button → increments `weather_cache_version` → all clients refetch on next load

**Provider logic:** Backend tries the default provider first; if it fails or is disabled, falls back to other enabled providers in order (Open-Meteo → WeatherAPI.com → OpenWeatherMap).

---

## Part 3: Weather Card – Source & Timestamp

### 3.1 Display

**Current:**
```
Utoljára frissítve: 14:32
OpenWeatherMap
```

**New:**
```
Utoljára frissítve: 14:32
Forrás: Open-Meteo
```

Or with multiple providers (if we ever merge data):
```
Utoljára frissítve: 14:32
Forrás: Open-Meteo, WeatherAPI.com
```

- `source` comes from API response
- `fetched_at` / `timestamp` shown as "Utoljára frissítve"

---

## Part 4: Current Date & Time Widget (Homepage)

### 4.1 Placement

- New widget on homepage, e.g. near WeatherWidget or in the same column
- Compact: "2025. március 6., csütörtök · 14:32"

### 4.2 Implementation

- **Client-side only** – `new Date()` in Svelte
- Update every minute (or every second if desired) via `setInterval`
- Format: Hungarian locale (`hu-HU`)
- No API, no cache

### 4.3 Styling

- Match existing widget style (e.g. `.weather-card` or similar)
- Small, unobtrusive

---

## Implementation Order

| Phase | Task | Depends On |
|-------|------|------------|
| **1** | Multi-provider weather backend | Config keys |
| **2** | WeatherWidget: use new API shape, show source + timestamp | Phase 1 |
| **3** | `site_settings` table + migration | - |
| **4** | `GET /api/config/public` | Phase 3 |
| **5** | WeatherWidget: fetch config, use TTL + cache_version | Phase 4 |
| **6** | Admin settings endpoints | Phase 3 |
| **7** | Admin UI: Settings tab | Phase 6 |
| **8** | Date/time widget on homepage | - |
| **9** | (Optional) Smart TTL, request tracking | Phase 5 |

---

## File Changes Summary

| File | Changes |
|------|---------|
| `backend/internal/config/config.go` | Add `WeatherAPIComKey` |
| `backend/internal/weather/weather.go` | Multi-provider fetch, unified response, read provider settings from DB |
| `backend/migrations/site_settings.sql` | New migration |
| `backend/main.go` | Register `/api/config/public`, admin settings routes |
| `src/lib/components/WeatherWidget.svelte` | Config fetch, source/timestamp, cache_version |
| `src/routes/(public)/+page.svelte` | Add DateTimeWidget |
| `src/lib/components/DateTimeWidget.svelte` | New component |
| `src/routes/admin/+page.svelte` | Settings tab, clear cache button |
| `src/styles/global.css` | DateTimeWidget styles |

---

## Your Action Items

1. **WeatherAPI.com:** Get free API key, add `WEATHER_API_COM_KEY` to `.env`
2. **OpenWeatherMap:** Ensure `WEATHER_API_KEY` is set in `.env`
3. **Run migration** after `site_settings` migration is added
