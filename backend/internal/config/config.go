package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL      string
	DBMaxOpenConns   int
	DBMaxIdleConns   int
	Port             string
	WeatherAPIKey    string
	WeatherAPIComKey string
	// METNoUserAgent identifies us to MET Norway, the main weather source.
	// Their terms require an application name and a contact (a site or an
	// email); a request without one may be blocked.
	METNoUserAgent string
	// WeatherWorker runs the background refresh that fills the weather cache
	// and archive. Off for a throwaway server that must not call providers.
	WeatherWorker     bool
	GoogleClientID    string
	AdminGoogleEmails []string
	SzotarOrigin      string
	// AccountSzotarToken and AccountJatszoterToken open the internal account
	// API (/internal/account/) to Szótár's and Játszótér's backends, one
	// secret per app pair (WAYS_OF_WORKING R18). None set: the API is off.
	AccountSzotarToken    string
	AccountJatszoterToken string
	// AllowedOrigins lists the exact browser origins that may read this API
	// cross-origin. Never reflect an unknown Origin back: with
	// Access-Control-Allow-Credentials that hands any site the signed-in reply.
	AllowedOrigins []string
	// DataAPI serves mondások, events, news, and the other content routes.
	// Auth and /api/config/public stay up when this is false.
	DataAPI  bool
	Features struct {
		Weather    bool
		Events     bool
		News       bool
		QuickLinks bool
		Search     bool
	}
}

var AppConfig Config

func Load() {
	// Load optional parent .env first, then cwd .env so the project always wins
	// (Previously ../.env alone could shadow ./.env when the server was started from repo root.)
	// A variable already set in the process environment beats both files, so
	// `DATABASE_URL=... go run .` really uses that URL.
	preset := presetEnv()
	if _, err := os.Stat("../.env"); err == nil {
		loadEnvFile("../.env", preset)
	}
	if _, err := os.Stat(".env"); err == nil {
		loadEnvFile(".env", preset)
	}

	AppConfig.DatabaseURL = getEnv("DATABASE_URL", "postgres://lamsza_user:lamsza_password@localhost:5433/lamsza?sslmode=disable")
	// Pool size. The `lamsza` database is shared with the admin app and
	// Postgres defaults to max_connections=100, so this app must not try to
	// take all of them. /api/search alone opens several connections per
	// request, so an uncapped pool exhausts the server under load.
	AppConfig.DBMaxOpenConns = getIntEnv("DB_MAX_OPEN_CONNS", 25)
	AppConfig.DBMaxIdleConns = getIntEnv("DB_MAX_IDLE_CONNS", 10)
	AppConfig.Port = getEnv("PORT", "3001")
	AppConfig.WeatherAPIKey = getEnv("WEATHER_API_KEY", "")
	AppConfig.WeatherAPIComKey = getEnv("WEATHER_API_COM_KEY", "")
	AppConfig.METNoUserAgent = strings.TrimSpace(getEnv("METNO_USER_AGENT", "lamsza.com weather (+https://lamsza.com)"))
	AppConfig.WeatherWorker = getBoolEnv("WEATHER_WORKER", true)

	AppConfig.GoogleClientID = getEnv("GOOGLE_CLIENT_ID", "")
	AppConfig.AdminGoogleEmails = parseEmailList(getEnv("ADMIN_GOOGLE_EMAILS", "attila.bogozi@gmail.com"))
	AppConfig.SzotarOrigin = strings.TrimRight(getEnv("SZOTAR_ORIGIN", ""), "/")
	AppConfig.AccountSzotarToken = strings.TrimSpace(getEnv("ACCOUNT_SZOTAR_TOKEN", ""))
	AppConfig.AccountJatszoterToken = strings.TrimSpace(getEnv("ACCOUNT_JATSZOTER_TOKEN", ""))
	AppConfig.AllowedOrigins = parseOriginList(getEnv("CORS_ALLOWED_ORIGINS", ""))
	ReloadAllowedOrigins()

	AppConfig.DataAPI = getBoolEnv("DATA_API", true)

	AppConfig.Features.Weather = getBoolEnv("FEATURE_WEATHER", true)
	AppConfig.Features.Events = getBoolEnv("FEATURE_EVENTS", true)
	AppConfig.Features.News = getBoolEnv("FEATURE_NEWS", true)
	AppConfig.Features.QuickLinks = getBoolEnv("FEATURE_QUICKLINKS", true)
	AppConfig.Features.Search = getBoolEnv("FEATURE_SEARCH", true)
}

// defaultAllowedOrigins covers the four production sites plus the local
// development hosts. Set CORS_ALLOWED_ORIGINS to replace the whole list, which
// is what production should do so the development hosts drop out.
var defaultAllowedOrigins = []string{
	"https://lamsza.com",
	"https://www.lamsza.com",
	"https://szotar.lamsza.com",
	"https://jatszoter.lamsza.com",
	"https://admin.lamsza.com",
	"https://lamsza.test",
	"https://szotar.lamsza.test",
	"https://jatszoter.lamsza.test",
	"https://admin.lamsza.test",
	"http://localhost:5173",
	"http://localhost:5174",
	"http://localhost:5175",
	"http://localhost:5176",
	"http://127.0.0.1:5174",
}

var allowedOriginSet map[string]bool

// OriginAllowed reports whether an Origin header value is on the allowlist.
// The match is exact: scheme, host and port must all agree, so a lookalike
// host such as "https://lamsza.com.evil.test" never passes.
func OriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	if allowedOriginSet == nil {
		ReloadAllowedOrigins()
	}
	return allowedOriginSet[origin]
}

// ReloadAllowedOrigins rebuilds the lookup set after AllowedOrigins changes.
// Load calls it; tests call it when they set their own list.
func ReloadAllowedOrigins() {
	origins := AppConfig.AllowedOrigins
	if len(origins) == 0 {
		origins = defaultAllowedOrigins
		AppConfig.AllowedOrigins = origins
	}
	set := make(map[string]bool, len(origins))
	for _, o := range origins {
		set[o] = true
	}
	allowedOriginSet = set
}

func parseOriginList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		origin := strings.TrimRight(strings.TrimSpace(p), "/")
		if origin == "" || seen[origin] {
			continue
		}
		seen[origin] = true
		out = append(out, origin)
	}
	return out
}

func parseEmailList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		email := strings.ToLower(strings.TrimSpace(p))
		if email == "" || seen[email] {
			continue
		}
		seen[email] = true
		out = append(out, email)
	}
	if len(out) == 0 {
		return []string{"attila.bogozi@gmail.com"}
	}
	return out
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		return strings.ToLower(value) == "true"
	}
	return fallback
}

// presetEnv records which keys the process environment set before any .env
// file was read.
func presetEnv() map[string]bool {
	preset := map[string]bool{}
	for _, kv := range os.Environ() {
		if key, _, ok := strings.Cut(kv, "="); ok {
			preset[key] = true
		}
	}
	return preset
}

func loadEnvFile(filename string, preset map[string]bool) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			if preset[key] {
				continue
			}
			os.Setenv(key, strings.TrimSpace(parts[1]))
		}
	}
}
