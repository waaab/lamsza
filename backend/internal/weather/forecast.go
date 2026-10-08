package weather

import (
	"math"
	"sort"
	"time"
)

// Step is one forecast time step, the same shape for every provider.
//
// Instant values (Temp, Wind, ...) hold at Time. Period values (Symbol,
// PrecipMm, PrecipProb, TMax, TMin) cover [Time, Time+Hours). MET Norway
// gives 1-hour steps for about two and a half days and 6-hour steps after
// that; WeatherAPI.com gives 1-hour steps, OpenWeatherMap 3-hour ones.
// A nil pointer means the provider did not send that value.
type Step struct {
	Time        time.Time `json:"time"`
	Hours       int       `json:"hours"`
	Symbol      string    `json:"symbol"`
	Temp        *float64  `json:"temp,omitempty"`
	Feels       *float64  `json:"feels,omitempty"`
	TMax        *float64  `json:"tmax,omitempty"`
	TMin        *float64  `json:"tmin,omitempty"`
	PrecipMm    *float64  `json:"precip_mm,omitempty"`
	PrecipProb  *float64  `json:"precip_prob,omitempty"`
	WindKph     *float64  `json:"wind_kph,omitempty"`
	WindDir     *float64  `json:"wind_dir,omitempty"`
	GustKph     *float64  `json:"gust_kph,omitempty"`
	Humidity    *float64  `json:"humidity,omitempty"`
	PressureHpa *float64  `json:"pressure_hpa,omitempty"`
	CloudPct    *float64  `json:"cloud_pct,omitempty"`
	UV          *float64  `json:"uv,omitempty"`
	DewPoint    *float64  `json:"dew_point,omitempty"`
	FogPct      *float64  `json:"fog_pct,omitempty"`
}

// Forecast is what one provider said, normalized. It is what the cache
// stores; days and "now" are worked out when it is read, so "today" is
// always the reader's Bucharest day.
type Forecast struct {
	Source string `json:"source"`
	Steps  []Step `json:"steps"`
}

// Day is one Bucharest calendar day of a forecast.
type Day struct {
	Date       string   `json:"date"`
	Symbol     string   `json:"symbol"`
	TMin       *float64 `json:"tmin,omitempty"`
	TMax       *float64 `json:"tmax,omitempty"`
	PrecipMm   float64  `json:"precip_mm"`
	PrecipProb *float64 `json:"precip_prob,omitempty"`
	WindMaxKph *float64 `json:"wind_max_kph,omitempty"`
	GustMaxKph *float64 `json:"gust_max_kph,omitempty"`
	UVMax      *float64 `json:"uv_max,omitempty"`
	// Hours is how many hours of the day the forecast covers; today has
	// fewer, and the last day can too.
	Hours int `json:"hours"`
}

func fp(v float64) *float64 { return &v }

// round1 keeps one decimal, which is all any provider means.
func round1(v float64) float64 { return math.Round(v*10) / 10 }

func maxp(cur *float64, v float64) *float64 {
	if cur == nil || v > *cur {
		return fp(v)
	}
	return cur
}

func minp(cur *float64, v float64) *float64 {
	if cur == nil || v < *cur {
		return fp(v)
	}
	return cur
}

// symbolRank orders symbols by how much they matter to someone deciding
// what to wear: a day with a shower is a showery day.
func symbolRank(code string) int {
	base, _ := splitSymbol(code)
	switch base {
	case "clearsky":
		return 0
	case "fair":
		return 1
	case "partlycloudy":
		return 2
	case "cloudy":
		return 3
	case "fog":
		return 4
	}
	r := 10
	switch {
	case len(base) > 5 && base[:5] == "heavy":
		r += 4
	case len(base) > 5 && base[:5] == "light":
		r += 0
	default:
		r += 2
	}
	if hasThunder(base) {
		r += 10
	}
	return r
}

func hasThunder(base string) bool {
	return len(base) > 10 && base[len(base)-10:] == "andthunder"
}

// stepDuration returns a step's period, at least an hour.
func stepDuration(s Step) time.Duration {
	if s.Hours <= 0 {
		return time.Hour
	}
	return time.Duration(s.Hours) * time.Hour
}

// BuildDays folds steps into Bucharest calendar days.
//
// Temperatures: every instant reading counts for its own day; a period's
// max and min count for the day its middle falls on. Precipitation: each
// period's amount is spread evenly over its hours and each hour goes to its
// own day, so a 6-hour window across midnight is split, and an hour already
// covered by a shorter step is not counted twice. The day's symbol is the
// one that covers most daytime hours (07-19), the more severe on a tie.
func BuildDays(steps []Step, loc *time.Location) []Day {
	type acc struct {
		day       Day
		dayCounts map[string]int
		anyCounts map[string]int
		hours     map[int]bool
	}
	days := map[string]*acc{}
	get := func(t time.Time) *acc {
		d := t.In(loc).Format("2006-01-02")
		a := days[d]
		if a == nil {
			a = &acc{day: Day{Date: d}, dayCounts: map[string]int{}, anyCounts: map[string]int{}, hours: map[int]bool{}}
			days[d] = a
		}
		return a
	}

	sorted := append([]Step(nil), steps...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Time.Before(sorted[j].Time) })

	var covered time.Time
	for _, s := range sorted {
		a := get(s.Time)
		if s.Temp != nil {
			a.day.TMax = maxp(a.day.TMax, *s.Temp)
			a.day.TMin = minp(a.day.TMin, *s.Temp)
		}
		mid := s.Time.Add(stepDuration(s) / 2)
		if s.TMax != nil {
			m := get(mid)
			m.day.TMax = maxp(m.day.TMax, *s.TMax)
		}
		if s.TMin != nil {
			m := get(mid)
			m.day.TMin = minp(m.day.TMin, *s.TMin)
		}
		if s.WindKph != nil {
			a.day.WindMaxKph = maxp(a.day.WindMaxKph, *s.WindKph)
		}
		if s.GustKph != nil {
			a.day.GustMaxKph = maxp(a.day.GustMaxKph, *s.GustKph)
		}
		if s.UV != nil {
			a.day.UVMax = maxp(a.day.UVMax, *s.UV)
		}

		hours := s.Hours
		if hours <= 0 {
			hours = 1
		}
		for h := 0; h < hours; h++ {
			t := s.Time.Add(time.Duration(h) * time.Hour)
			if t.Before(covered) {
				continue
			}
			ha := get(t)
			local := t.In(loc)
			ha.hours[local.Hour()] = true
			if s.PrecipMm != nil {
				ha.day.PrecipMm += *s.PrecipMm / float64(hours)
			}
			if s.PrecipProb != nil {
				ha.day.PrecipProb = maxp(ha.day.PrecipProb, *s.PrecipProb)
			}
			if s.Symbol != "" {
				base, _ := splitSymbol(s.Symbol)
				ha.anyCounts[base]++
				if local.Hour() >= 7 && local.Hour() < 19 {
					ha.dayCounts[base]++
				}
			}
		}
		if end := s.Time.Add(stepDuration(s)); end.After(covered) {
			covered = end
		}
	}

	pick := func(counts map[string]int) string {
		best, bestN := "", -1
		for base, n := range counts {
			if n > bestN || (n == bestN && symbolRank(base) > symbolRank(best)) {
				best, bestN = base, n
			}
		}
		return best
	}

	out := make([]Day, 0, len(days))
	for _, a := range days {
		base := pick(a.dayCounts)
		if base == "" {
			base = pick(a.anyCounts)
		}
		if base == "" {
			base = "cloudy"
		}
		a.day.Symbol = withVariant(base, true)
		a.day.PrecipMm = round1(a.day.PrecipMm)
		a.day.Hours = len(a.hours)
		out = append(out, a.day)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out
}

// CurrentStep is the step that holds at now: the last one starting at or
// before now, or the first one if the forecast starts later.
func CurrentStep(steps []Step, now time.Time) (Step, bool) {
	if len(steps) == 0 {
		return Step{}, false
	}
	cur := steps[0]
	for _, s := range steps {
		if s.Time.After(now) {
			break
		}
		cur = s
	}
	return cur, true
}

// StepsBetween returns the steps starting in [from, to).
func StepsBetween(steps []Step, from, to time.Time) []Step {
	var out []Step
	for _, s := range steps {
		if !s.Time.Before(from) && s.Time.Before(to) {
			out = append(out, s)
		}
	}
	return out
}
