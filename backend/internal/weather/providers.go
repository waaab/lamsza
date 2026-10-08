package weather

import (
	"backend/internal/config"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	providerMETNo          = "metno"
	providerWeatherAPICom  = "weatherapi_com"
	providerOpenWeatherMap = "openweathermap"
)

// Base URLs, variables so tests can point them at an httptest server.
var (
	metnoBaseURL      = "https://api.met.no/weatherapi"
	weatherAPIBaseURL = "https://api.weatherapi.com/v1"
	owmBaseURL        = "https://api.openweathermap.org"
)

// weatherClientTimeout: http.DefaultClient has no timeout, so a provider that
// accepts the connection and never answers held a goroutine and a socket for
// the life of the process.
const weatherClientTimeout = 4 * time.Second

var weatherClient = &http.Client{Timeout: weatherClientTimeout}

func weatherHTTPGet(u string) (*http.Response, error) {
	return weatherClient.Get(u)
}

// errNoKey: the provider needs a key and none is configured. Not an outage.
var errNoKey = errors.New("no api key configured")

// providerDisplayName is the name the attribution line shows.
func providerDisplayName(id string) string {
	switch id {
	case providerMETNo:
		return "MET Norway"
	case providerWeatherAPICom:
		return "WeatherAPI.com"
	case providerOpenWeatherMap:
		return "OpenWeather"
	}
	return id
}

// coord formats a coordinate with at most 4 decimals, as MET's terms ask
// (more precision only splits their cache).
func coord(v float64) string {
	return strconv.FormatFloat(math.Round(v*1e4)/1e4, 'f', -1, 64)
}

func readOK(resp *http.Response, name string) ([]byte, error) {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d", name, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
}

// ---- MET Norway -----------------------------------------------------------

// metnoResult is one Locationforecast answer. NotModified means our copy
// (sent as If-Modified-Since) is still current and Forecast is nil.
type metnoResult struct {
	Forecast     *Forecast
	Expires      time.Time
	LastModified string
	NotModified  bool
}

func metnoRequest(u, ifModifiedSince string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", config.AppConfig.METNoUserAgent)
	if ifModifiedSince != "" {
		req.Header.Set("If-Modified-Since", ifModifiedSince)
	}
	return weatherClient.Do(req)
}

// fetchMETNo gets the Locationforecast "complete" product. elevation (m) is
// optional; without it MET uses its own terrain height.
func fetchMETNo(lat, lon float64, elevation *int, ifModifiedSince string) (*metnoResult, error) {
	u := metnoBaseURL + "/locationforecast/2.0/complete?lat=" + coord(lat) + "&lon=" + coord(lon)
	if elevation != nil {
		u += "&altitude=" + strconv.Itoa(*elevation)
	}
	resp, err := metnoRequest(u, ifModifiedSince)
	if err != nil {
		return nil, err
	}
	res := &metnoResult{LastModified: resp.Header.Get("Last-Modified")}
	if exp, err := http.ParseTime(resp.Header.Get("Expires")); err == nil {
		res.Expires = exp
	}
	if resp.StatusCode == http.StatusNotModified {
		resp.Body.Close()
		res.NotModified = true
		if res.LastModified == "" {
			res.LastModified = ifModifiedSince
		}
		return res, nil
	}
	body, err := readOK(resp, "met.no")
	if err != nil {
		return nil, err
	}
	fc, err := parseMETNo(body)
	if err != nil {
		return nil, err
	}
	res.Forecast = fc
	return res, nil
}

type metnoPeriod struct {
	Summary struct {
		SymbolCode string `json:"symbol_code"`
	} `json:"summary"`
	Details struct {
		PrecipitationAmount        *float64 `json:"precipitation_amount"`
		ProbabilityOfPrecipitation *float64 `json:"probability_of_precipitation"`
		AirTemperatureMax          *float64 `json:"air_temperature_max"`
		AirTemperatureMin          *float64 `json:"air_temperature_min"`
	} `json:"details"`
}

func parseMETNo(body []byte) (*Forecast, error) {
	var doc struct {
		Properties struct {
			Timeseries []struct {
				Time time.Time `json:"time"`
				Data struct {
					Instant struct {
						Details struct {
							AirTemperature         *float64 `json:"air_temperature"`
							ApparentAirTemperature *float64 `json:"apparent_air_temperature"`
							WindSpeed              *float64 `json:"wind_speed"`
							WindSpeedOfGust        *float64 `json:"wind_speed_of_gust"`
							WindFromDirection      *float64 `json:"wind_from_direction"`
							RelativeHumidity       *float64 `json:"relative_humidity"`
							AirPressureAtSeaLevel  *float64 `json:"air_pressure_at_sea_level"`
							CloudAreaFraction      *float64 `json:"cloud_area_fraction"`
							UltravioletIndex       *float64 `json:"ultraviolet_index_clear_sky"`
							DewPointTemperature    *float64 `json:"dew_point_temperature"`
							FogAreaFraction        *float64 `json:"fog_area_fraction"`
						} `json:"details"`
					} `json:"instant"`
					Next1h  *metnoPeriod `json:"next_1_hours"`
					Next6h  *metnoPeriod `json:"next_6_hours"`
					Next12h *metnoPeriod `json:"next_12_hours"`
				} `json:"data"`
			} `json:"timeseries"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	ts := doc.Properties.Timeseries
	if len(ts) == 0 {
		return nil, errors.New("met.no: empty timeseries")
	}
	fc := &Forecast{Source: providerMETNo}
	for _, t := range ts {
		in := t.Data.Instant.Details
		s := Step{
			Time:        t.Time.UTC(),
			Temp:        in.AirTemperature,
			Feels:       in.ApparentAirTemperature,
			WindDir:     in.WindFromDirection,
			Humidity:    in.RelativeHumidity,
			PressureHpa: in.AirPressureAtSeaLevel,
			CloudPct:    in.CloudAreaFraction,
			UV:          in.UltravioletIndex,
			DewPoint:    in.DewPointTemperature,
			FogPct:      in.FogAreaFraction,
		}
		if in.WindSpeed != nil {
			s.WindKph = fp(round1(*in.WindSpeed * 3.6))
		}
		if in.WindSpeedOfGust != nil {
			s.GustKph = fp(round1(*in.WindSpeedOfGust * 3.6))
		}
		// The shortest period sets the step's own symbol and amount; the
		// 6-hour window always gives the max and min when it is there.
		var p *metnoPeriod
		switch {
		case t.Data.Next1h != nil:
			p, s.Hours = t.Data.Next1h, 1
		case t.Data.Next6h != nil:
			p, s.Hours = t.Data.Next6h, 6
		case t.Data.Next12h != nil:
			p, s.Hours = t.Data.Next12h, 12
		}
		if p == nil {
			// The very last steps carry no period; nothing to show for them.
			continue
		}
		s.Symbol = normalizeSymbol(p.Summary.SymbolCode)
		s.PrecipMm = p.Details.PrecipitationAmount
		s.PrecipProb = p.Details.ProbabilityOfPrecipitation
		if six := t.Data.Next6h; six != nil {
			s.TMax = six.Details.AirTemperatureMax
			s.TMin = six.Details.AirTemperatureMin
		}
		if s.Hours == 12 {
			// A 12-hour step has no amount and is beyond any day we show.
			continue
		}
		fc.Steps = append(fc.Steps, s)
	}
	if len(fc.Steps) == 0 {
		return nil, errors.New("met.no: no usable steps")
	}
	return fc, nil
}

// Astro is one day's sun and moon times for a place.
type Astro struct {
	Date      string     `json:"date"`
	Sunrise   *time.Time `json:"sunrise,omitempty"`
	Sunset    *time.Time `json:"sunset,omitempty"`
	Moonrise  *time.Time `json:"moonrise,omitempty"`
	Moonset   *time.Time `json:"moonset,omitempty"`
	MoonPhase *float64   `json:"moon_phase,omitempty"` // degrees: 0 new, 90 first quarter, 180 full, 270 last quarter
}

// fetchMETNoAstro gets sunrise, sunset, moonrise, moonset and the moon phase
// for one local date. offset is the zone's UTC offset that day ("+03:00").
func fetchMETNoAstro(lat, lon float64, date, offset string) (*Astro, error) {
	q := "?lat=" + coord(lat) + "&lon=" + coord(lon) + "&date=" + date + "&offset=" + url.QueryEscape(offset)
	a := &Astro{Date: date}
	resp, err := metnoRequest(metnoBaseURL+"/sunrise/3.0/sun"+q, "")
	if err != nil {
		return nil, err
	}
	body, err := readOK(resp, "met.no sun")
	if err != nil {
		return nil, err
	}
	if err := parseMETNoSun(body, a); err != nil {
		return nil, err
	}
	resp, err = metnoRequest(metnoBaseURL+"/sunrise/3.0/moon"+q, "")
	if err != nil {
		return nil, err
	}
	body, err = readOK(resp, "met.no moon")
	if err != nil {
		return nil, err
	}
	if err := parseMETNoMoon(body, a); err != nil {
		return nil, err
	}
	return a, nil
}

type metnoEvent struct {
	Time *string `json:"time"`
}

// metnoTime parses MET's "2026-10-09T07:25+03:00" (no seconds).
func metnoTime(e *metnoEvent) *time.Time {
	if e == nil || e.Time == nil || *e.Time == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02T15:04Z07:00", time.RFC3339} {
		if t, err := time.Parse(layout, *e.Time); err == nil {
			t = t.UTC()
			return &t
		}
	}
	return nil
}

func parseMETNoSun(body []byte, a *Astro) error {
	var doc struct {
		Properties struct {
			Sunrise *metnoEvent `json:"sunrise"`
			Sunset  *metnoEvent `json:"sunset"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return err
	}
	a.Sunrise = metnoTime(doc.Properties.Sunrise)
	a.Sunset = metnoTime(doc.Properties.Sunset)
	return nil
}

func parseMETNoMoon(body []byte, a *Astro) error {
	var doc struct {
		Properties struct {
			Moonrise  *metnoEvent `json:"moonrise"`
			Moonset   *metnoEvent `json:"moonset"`
			MoonPhase *float64    `json:"moonphase"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return err
	}
	a.Moonrise = metnoTime(doc.Properties.Moonrise)
	a.Moonset = metnoTime(doc.Properties.Moonset)
	a.MoonPhase = doc.Properties.MoonPhase
	return nil
}

// ---- WeatherAPI.com (fallback) --------------------------------------------

// fetchWeatherAPICom gets the free plan's 3-day hourly forecast.
func fetchWeatherAPICom(lat, lon float64) (*Forecast, error) {
	key := config.AppConfig.WeatherAPIComKey
	if key == "" {
		return nil, errNoKey
	}
	u := weatherAPIBaseURL + "/forecast.json?key=" + url.QueryEscape(key) +
		"&q=" + url.QueryEscape(coord(lat)+","+coord(lon)) + "&days=3&aqi=no&alerts=no"
	resp, err := weatherHTTPGet(u)
	if err != nil {
		return nil, err
	}
	body, err := readOK(resp, "weatherapi.com")
	if err != nil {
		return nil, err
	}
	return parseWeatherAPIComForecast(body)
}

func parseWeatherAPIComForecast(body []byte) (*Forecast, error) {
	var doc struct {
		Forecast struct {
			Forecastday []struct {
				Hour []struct {
					TimeEpoch    int64   `json:"time_epoch"`
					TempC        float64 `json:"temp_c"`
					FeelslikeC   float64 `json:"feelslike_c"`
					IsDay        int     `json:"is_day"`
					WindKph      float64 `json:"wind_kph"`
					WindDegree   float64 `json:"wind_degree"`
					GustKph      float64 `json:"gust_kph"`
					PressureMb   float64 `json:"pressure_mb"`
					PrecipMm     float64 `json:"precip_mm"`
					Humidity     float64 `json:"humidity"`
					Cloud        float64 `json:"cloud"`
					DewpointC    float64 `json:"dewpoint_c"`
					ChanceOfRain float64 `json:"chance_of_rain"`
					ChanceOfSnow float64 `json:"chance_of_snow"`
					UV           float64 `json:"uv"`
					Condition    struct {
						Code int `json:"code"`
					} `json:"condition"`
				} `json:"hour"`
			} `json:"forecastday"`
		} `json:"forecast"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	fc := &Forecast{Source: providerWeatherAPICom}
	for _, d := range doc.Forecast.Forecastday {
		for _, h := range d.Hour {
			prob := h.ChanceOfRain
			if h.ChanceOfSnow > prob {
				prob = h.ChanceOfSnow
			}
			fc.Steps = append(fc.Steps, Step{
				Time:        time.Unix(h.TimeEpoch, 0).UTC(),
				Hours:       1,
				Symbol:      withVariant(weatherAPIComSymbol(h.Condition.Code), h.IsDay == 1),
				Temp:        fp(h.TempC),
				Feels:       fp(h.FeelslikeC),
				PrecipMm:    fp(h.PrecipMm),
				PrecipProb:  fp(prob),
				WindKph:     fp(h.WindKph),
				WindDir:     fp(h.WindDegree),
				GustKph:     fp(h.GustKph),
				Humidity:    fp(h.Humidity),
				PressureHpa: fp(h.PressureMb),
				CloudPct:    fp(h.Cloud),
				UV:          fp(h.UV),
				DewPoint:    fp(h.DewpointC),
			})
		}
	}
	if len(fc.Steps) == 0 {
		return nil, errors.New("weatherapi.com: no hours")
	}
	return fc, nil
}

// ---- OpenWeatherMap (fallback) --------------------------------------------

// fetchOpenWeatherMap gets the free plan's 5-day, 3-hour forecast.
func fetchOpenWeatherMap(lat, lon float64) (*Forecast, error) {
	key := config.AppConfig.WeatherAPIKey
	if key == "" {
		return nil, errNoKey
	}
	u := owmBaseURL + "/data/2.5/forecast?lat=" + coord(lat) + "&lon=" + coord(lon) +
		"&units=metric&appid=" + url.QueryEscape(key)
	resp, err := weatherHTTPGet(u)
	if err != nil {
		return nil, err
	}
	body, err := readOK(resp, "openweathermap")
	if err != nil {
		return nil, err
	}
	return parseOpenWeatherMapForecast(body)
}

func parseOpenWeatherMapForecast(body []byte) (*Forecast, error) {
	var doc struct {
		List []struct {
			Dt   int64 `json:"dt"`
			Main struct {
				Temp      float64 `json:"temp"`
				FeelsLike float64 `json:"feels_like"`
				Pressure  float64 `json:"pressure"`
				Humidity  float64 `json:"humidity"`
			} `json:"main"`
			Weather []struct {
				ID int `json:"id"`
			} `json:"weather"`
			Clouds struct {
				All float64 `json:"all"`
			} `json:"clouds"`
			Wind struct {
				Speed float64  `json:"speed"`
				Deg   float64  `json:"deg"`
				Gust  *float64 `json:"gust"`
			} `json:"wind"`
			Pop  float64            `json:"pop"`
			Rain map[string]float64 `json:"rain"`
			Snow map[string]float64 `json:"snow"`
			Sys  struct {
				Pod string `json:"pod"`
			} `json:"sys"`
		} `json:"list"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	fc := &Forecast{Source: providerOpenWeatherMap}
	for _, it := range doc.List {
		if len(it.Weather) == 0 {
			continue
		}
		s := Step{
			Time:        time.Unix(it.Dt, 0).UTC(),
			Hours:       3,
			Symbol:      withVariant(owmSymbol(it.Weather[0].ID), it.Sys.Pod != "n"),
			Temp:        fp(it.Main.Temp),
			Feels:       fp(it.Main.FeelsLike),
			PrecipMm:    fp(round1(it.Rain["3h"] + it.Snow["3h"])),
			PrecipProb:  fp(math100(it.Pop)),
			WindKph:     fp(round1(it.Wind.Speed * 3.6)),
			WindDir:     fp(it.Wind.Deg),
			Humidity:    fp(it.Main.Humidity),
			PressureHpa: fp(it.Main.Pressure),
			CloudPct:    fp(it.Clouds.All),
		}
		if it.Wind.Gust != nil {
			s.GustKph = fp(round1(*it.Wind.Gust * 3.6))
		}
		fc.Steps = append(fc.Steps, s)
	}
	if len(fc.Steps) == 0 {
		return nil, errors.New("openweathermap: empty list")
	}
	return fc, nil
}

func math100(p float64) float64 { return round1(p * 100) }

// geocodeOWM finds a Romanian settlement's coordinates by name with
// OpenWeatherMap's free geocoder. county (Romanian, e.g. "Harghita") picks
// the right one when several villages share a name.
func geocodeOWM(name, county string) (lat, lon float64, err error) {
	key := config.AppConfig.WeatherAPIKey
	if key == "" {
		return 0, 0, errNoKey
	}
	for _, variant := range geocodeNameVariants(name) {
		u := owmBaseURL + "/geo/1.0/direct?q=" + url.QueryEscape(variant+",RO") + "&limit=5&appid=" + url.QueryEscape(key)
		resp, err := weatherHTTPGet(u)
		if err != nil {
			return 0, 0, err
		}
		body, err := readOK(resp, "openweathermap geocode")
		if err != nil {
			return 0, 0, err
		}
		if lat, lon, ok := pickGeocode(body, county); ok {
			return lat, lon, nil
		}
	}
	return 0, 0, fmt.Errorf("geocode %q: no results", name)
}

func pickGeocode(body []byte, county string) (lat, lon float64, ok bool) {
	var hits []struct {
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
		Country string  `json:"country"`
		State   string  `json:"state"`
	}
	if json.Unmarshal(body, &hits) != nil {
		return 0, 0, false
	}
	county = strings.ToLower(strings.TrimSpace(county))
	first := -1
	for i, h := range hits {
		if h.Country != "RO" {
			continue
		}
		if first < 0 {
			first = i
		}
		if county != "" && strings.Contains(strings.ToLower(h.State), county) {
			return h.Lat, h.Lon, true
		}
	}
	if first < 0 {
		return 0, 0, false
	}
	return hits[first].Lat, hits[first].Lon, true
}

// geocodeNameVariants lists search strings for a name: geocoders store
// Miercurea Ciuc as "Miercurea-Ciuc", so a spaced name may find nothing.
func geocodeNameVariants(city string) []string {
	city = strings.TrimSpace(city)
	if city == "" {
		return nil
	}
	variants := []string{city}
	hyphenated := strings.Join(strings.Fields(city), "-")
	if hyphenated != city {
		variants = append(variants, hyphenated)
	}
	return variants
}
