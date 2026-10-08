package weather

import (
	"backend/internal/clock"
	"math"
	"os"
	"testing"
	"time"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func utc(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

// testdata/metno_complete.json is a real Locationforecast answer for
// Csíkszereda, recorded 2026-10-08 21:11 UTC.
func TestParseMETNoFixture(t *testing.T) {
	fc, err := parseMETNo(fixture(t, "metno_complete.json"))
	if err != nil {
		t.Fatal(err)
	}
	if fc.Source != providerMETNo {
		t.Fatalf("source %q", fc.Source)
	}
	first := fc.Steps[0]
	if !first.Time.Equal(utc(t, "2026-10-08T21:00:00Z")) || first.Hours != 1 {
		t.Fatalf("first step %v, %dh", first.Time, first.Hours)
	}
	if first.Symbol != "fair_night" || first.Temp == nil || *first.Temp != 10.3 {
		t.Fatalf("first step symbol %q temp %v", first.Symbol, first.Temp)
	}
	if first.WindKph == nil || first.Feels == nil || first.Humidity == nil || first.PressureHpa == nil {
		t.Fatal("first step lost an instant value")
	}
	var six int
	for _, s := range fc.Steps {
		if s.Hours != 1 && s.Hours != 6 {
			t.Fatalf("step %v has %d hours", s.Time, s.Hours)
		}
		if s.Hours == 6 {
			six++
		}
		if normalizeSymbol(s.Symbol) != s.Symbol {
			t.Fatalf("step %v symbol %q is not normalized", s.Time, s.Symbol)
		}
	}
	if six == 0 {
		t.Fatal("no 6-hour steps; the far days would be missing")
	}
	last := fc.Steps[len(fc.Steps)-1].Time
	if last.Sub(first.Time) < 7*24*time.Hour {
		t.Fatalf("forecast reaches only %v ahead, want a week", last.Sub(first.Time))
	}
}

func TestBuildDaysFromMETNoCoversAWeek(t *testing.T) {
	fc, err := parseMETNo(fixture(t, "metno_complete.json"))
	if err != nil {
		t.Fatal(err)
	}
	days := BuildDays(fc.Steps, clock.Zone)
	if len(days) < 8 {
		t.Fatalf("got %d days", len(days))
	}
	// 21:00 UTC is midnight in Bucharest (EEST): the first day is the 9th.
	if days[0].Date != "2026-10-09" {
		t.Fatalf("first day %s", days[0].Date)
	}
	for _, d := range days {
		if d.TMin == nil || d.TMax == nil || *d.TMin > *d.TMax {
			t.Fatalf("%s: min %v max %v", d.Date, d.TMin, d.TMax)
		}
		if d.Hours > 24 || d.PrecipMm < 0 {
			t.Fatalf("%s: %d hours, %v mm", d.Date, d.Hours, d.PrecipMm)
		}
		if d.Symbol == "" || normalizeSymbol(d.Symbol) != d.Symbol {
			t.Fatalf("%s: symbol %q", d.Date, d.Symbol)
		}
	}
	for _, d := range days[:7] {
		if d.Hours != 24 {
			t.Fatalf("%s covers %d hours, want 24", d.Date, d.Hours)
		}
	}
}

func step(t *testing.T, at string, hours int, symbol string, temp, precip float64) Step {
	return Step{Time: utc(t, at), Hours: hours, Symbol: symbol, Temp: fp(temp), PrecipMm: fp(precip)}
}

// A 6-hour window from 21:00 to 03:00 Bucharest time is split at midnight.
func TestBuildDaysSplitsAWindowAtBucharestMidnight(t *testing.T) {
	days := BuildDays([]Step{step(t, "2026-10-09T18:00:00Z", 6, "rain", 5, 6)}, clock.Zone)
	if len(days) != 2 {
		t.Fatalf("got %d days", len(days))
	}
	if days[0].Date != "2026-10-09" || !near(days[0].PrecipMm, 3) || days[0].Hours != 3 {
		t.Fatalf("first day %+v", days[0])
	}
	if days[1].Date != "2026-10-10" || !near(days[1].PrecipMm, 3) || days[1].Hours != 3 {
		t.Fatalf("second day %+v", days[1])
	}
}

// An hour already inside a longer step is not counted again.
func TestBuildDaysCountsEachHourOnce(t *testing.T) {
	days := BuildDays([]Step{
		step(t, "2026-10-09T06:00:00Z", 6, "rain", 8, 6),
		step(t, "2026-10-09T08:00:00Z", 1, "heavyrain", 9, 5),
	}, clock.Zone)
	if len(days) != 1 || !near(days[0].PrecipMm, 6) || days[0].Hours != 6 {
		t.Fatalf("got %+v", days)
	}
}

// 2026-10-25 is the end of summer time in Romania: that day has 25 hours.
func TestBuildDaysClockChangeDayHas25Hours(t *testing.T) {
	var steps []Step
	start := utc(t, "2026-10-24T21:00:00Z") // midnight, EEST
	for h := 0; h < 25; h++ {
		steps = append(steps, Step{Time: start.Add(time.Duration(h) * time.Hour), Hours: 1, Symbol: "cloudy", Temp: fp(5)})
	}
	days := BuildDays(steps, clock.Zone)
	if len(days) != 1 || days[0].Date != "2026-10-25" {
		t.Fatalf("got %+v", days)
	}
	// The repeated 03:00 hour is one hour of the clock face but two steps.
	if days[0].Hours < 24 {
		t.Fatalf("got %d hours", days[0].Hours)
	}
}

// A clear night and a rainy afternoon make a rainy day.
func TestBuildDaysPicksTheDaytimeSymbol(t *testing.T) {
	var steps []Step
	start := utc(t, "2026-10-08T21:00:00Z")
	for h := 0; h < 24; h++ {
		sym := "clearsky_night"
		local := start.Add(time.Duration(h) * time.Hour).In(clock.Zone).Hour()
		if local >= 7 && local < 19 {
			sym = "rainshowers_day"
		}
		steps = append(steps, Step{Time: start.Add(time.Duration(h) * time.Hour), Hours: 1, Symbol: sym, Temp: fp(float64(h))})
	}
	days := BuildDays(steps, clock.Zone)
	if len(days) != 1 || days[0].Symbol != "rainshowers_day" {
		t.Fatalf("got %+v", days)
	}
	if *days[0].TMin != 0 || *days[0].TMax != 23 {
		t.Fatalf("min %v max %v", *days[0].TMin, *days[0].TMax)
	}
}

func TestCurrentStep(t *testing.T) {
	steps := []Step{
		step(t, "2026-10-09T10:00:00Z", 1, "fair_day", 10, 0),
		step(t, "2026-10-09T11:00:00Z", 1, "cloudy", 11, 0),
	}
	got, _ := CurrentStep(steps, utc(t, "2026-10-09T11:40:00Z"))
	if *got.Temp != 11 {
		t.Fatalf("got %v", *got.Temp)
	}
	got, _ = CurrentStep(steps, utc(t, "2026-10-09T09:00:00Z"))
	if *got.Temp != 10 {
		t.Fatalf("before the forecast: got %v", *got.Temp)
	}
}

// The widgets' second number is today's low, not the current temperature
// again (the old "6°C / 6°C").
func TestSummarizeGivesTodaysRange(t *testing.T) {
	fc, err := parseMETNo(fixture(t, "metno_complete.json"))
	if err != nil {
		t.Fatal(err)
	}
	now := utc(t, "2026-10-09T09:30:00Z")
	out, ok := summarize(fc, now, now, nil)
	if !ok {
		t.Fatal("no summary")
	}
	if out.TempMin == nil || out.TempMax == nil || *out.TempMin >= *out.TempMax {
		t.Fatalf("range %v..%v", out.TempMin, out.TempMax)
	}
	if out.Source != "MET Norway" || out.Symbol == "" || out.Icon == "" || out.Desc == "" {
		t.Fatalf("got %+v", out)
	}
}

func TestParseWeatherAPIComFixture(t *testing.T) {
	fc, err := parseWeatherAPIComForecast(fixture(t, "weatherapi_forecast.json"))
	if err != nil {
		t.Fatal(err)
	}
	if fc.Source != providerWeatherAPICom || len(fc.Steps) != 72 {
		t.Fatalf("source %q, %d steps", fc.Source, len(fc.Steps))
	}
	for _, s := range fc.Steps {
		if s.Hours != 1 || normalizeSymbol(s.Symbol) != s.Symbol || s.Temp == nil {
			t.Fatalf("step %+v", s)
		}
	}
	if days := BuildDays(fc.Steps, clock.Zone); len(days) < 3 {
		t.Fatalf("got %d days", len(days))
	}
}

func TestParseOpenWeatherMapFixture(t *testing.T) {
	fc, err := parseOpenWeatherMapForecast(fixture(t, "owm_forecast.json"))
	if err != nil {
		t.Fatal(err)
	}
	if fc.Source != providerOpenWeatherMap || len(fc.Steps) != 40 {
		t.Fatalf("source %q, %d steps", fc.Source, len(fc.Steps))
	}
	// id 802 (scattered clouds) at night.
	if fc.Steps[0].Symbol != "partlycloudy_night" || fc.Steps[0].Hours != 3 {
		t.Fatalf("first step %+v", fc.Steps[0])
	}
	if fc.Steps[0].PrecipProb == nil || *fc.Steps[0].PrecipProb < 0 || *fc.Steps[0].PrecipProb > 100 {
		t.Fatalf("precip prob %v", fc.Steps[0].PrecipProb)
	}
}

func TestParseMETNoAstroFixtures(t *testing.T) {
	a := &Astro{Date: "2026-10-09"}
	if err := parseMETNoSun(fixture(t, "metno_sun.json"), a); err != nil {
		t.Fatal(err)
	}
	if err := parseMETNoMoon(fixture(t, "metno_moon.json"), a); err != nil {
		t.Fatal(err)
	}
	if a.Sunrise == nil || !a.Sunrise.Equal(utc(t, "2026-10-09T04:25:00Z")) {
		t.Fatalf("sunrise %v", a.Sunrise)
	}
	if a.Sunset == nil || a.Moonrise == nil || a.Moonset == nil {
		t.Fatalf("missing a time: %+v", a)
	}
	if a.MoonPhase == nil || *a.MoonPhase != 338.85 {
		t.Fatalf("moon phase %v", a.MoonPhase)
	}
}

func TestPickGeocodePrefersTheCounty(t *testing.T) {
	body := []byte(`[
		{"lat":45.1,"lon":24.1,"country":"RO","state":"Vâlcea"},
		{"lat":46.36,"lon":25.80,"country":"RO","state":"Harghita"},
		{"lat":47.0,"lon":19.0,"country":"HU","state":"Pest"}]`)
	lat, lon, ok := pickGeocode(body, "Harghita")
	if !ok || lat != 46.36 || lon != 25.80 {
		t.Fatalf("got %v %v %v", lat, lon, ok)
	}
	lat, _, ok = pickGeocode(body, "Covasna")
	if !ok || lat != 45.1 {
		t.Fatalf("no county match should take the first Romanian hit, got %v %v", lat, ok)
	}
	if _, _, ok := pickGeocode([]byte(`[{"lat":1,"lon":1,"country":"HU"}]`), ""); ok {
		t.Fatal("a hit outside Romania was accepted")
	}
	lat, _, ok = pickGeocode(fixture(t, "owm_geocode.json"), "Harghita")
	if !ok || math.Abs(lat-46.36) > 0.05 {
		t.Fatalf("fixture: got %v %v", lat, ok)
	}
}

func TestGeocodeNameVariantsHyphenatesSpacedName(t *testing.T) {
	got := geocodeNameVariants("Miercurea Ciuc")
	if len(got) != 2 || got[0] != "Miercurea Ciuc" || got[1] != "Miercurea-Ciuc" {
		t.Fatalf("got %#v", got)
	}
}

func TestGeocodeNameVariantsKeepsSingleToken(t *testing.T) {
	got := geocodeNameVariants("  Kolozsvár ")
	if len(got) != 1 || got[0] != "Kolozsvár" {
		t.Fatalf("got %#v", got)
	}
}

func TestCoordKeepsFourDecimals(t *testing.T) {
	for in, want := range map[float64]string{46.35931234: "46.3593", 25.80176: "25.8018", -0.00004: "-0", 46: "46"} {
		if got := coord(in); got != want {
			t.Errorf("coord(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestClampExpires(t *testing.T) {
	now := utc(t, "2026-10-09T10:00:00Z")
	if got := clampExpires(time.Time{}, now); !got.Equal(now.Add(defaultTTL)) {
		t.Fatalf("no header: %v", got)
	}
	if got := clampExpires(now.Add(time.Minute), now); !got.Equal(now.Add(minTTL)) {
		t.Fatalf("too soon: %v", got)
	}
	if got := clampExpires(now.Add(48*time.Hour), now); !got.Equal(now.Add(6 * time.Hour)) {
		t.Fatalf("too late: %v", got)
	}
	if got := clampExpires(now.Add(40*time.Minute), now); !got.Equal(now.Add(40 * time.Minute)) {
		t.Fatalf("normal: %v", got)
	}
}
