package weather

import "strings"

// One icon vocabulary for every provider: MET Norway's symbol codes
// (https://github.com/metno/weathericons), e.g. "partlycloudy_day",
// "heavyrainshowersandthunder_night". WeatherAPI.com and OpenWeatherMap
// conditions are mapped onto it, so the frontend draws one set of icons and
// the Hungarian text lives in one map.
//
// Only clearsky, fair, partlycloudy and the *showers* codes have a _day or
// _night variant. MET's _polartwilight never happens in Székelyföld.

// metTypos are two misspelt codes MET really sends; we store the spelling
// the rest of the family uses.
var metTypos = map[string]string{
	"lightssleetshowersandthunder": "lightsleetshowersandthunder",
	"lightssnowshowersandthunder":  "lightsnowshowersandthunder",
}

// symbolBases is every base code, in MET's legend order.
var symbolBases = []string{
	"clearsky", "fair", "partlycloudy", "cloudy", "fog",
	"lightrainshowers", "rainshowers", "heavyrainshowers",
	"lightrainshowersandthunder", "rainshowersandthunder", "heavyrainshowersandthunder",
	"lightsleetshowers", "sleetshowers", "heavysleetshowers",
	"lightsleetshowersandthunder", "sleetshowersandthunder", "heavysleetshowersandthunder",
	"lightsnowshowers", "snowshowers", "heavysnowshowers",
	"lightsnowshowersandthunder", "snowshowersandthunder", "heavysnowshowersandthunder",
	"lightrain", "rain", "heavyrain",
	"lightrainandthunder", "rainandthunder", "heavyrainandthunder",
	"lightsleet", "sleet", "heavysleet",
	"lightsleetandthunder", "sleetandthunder", "heavysleetandthunder",
	"lightsnow", "snow", "heavysnow",
	"lightsnowandthunder", "snowandthunder", "heavysnowandthunder",
}

var knownBase = func() map[string]bool {
	m := make(map[string]bool, len(symbolBases))
	for _, b := range symbolBases {
		m[b] = true
	}
	return m
}()

// hasDayNight reports whether a base code is drawn differently by day and
// by night (the sun or the moon shows through).
func hasDayNight(base string) bool {
	switch base {
	case "clearsky", "fair", "partlycloudy":
		return true
	}
	return strings.Contains(base, "showers")
}

// splitSymbol returns the base code and "day", "night" or "".
func splitSymbol(code string) (base, variant string) {
	code = strings.ToLower(strings.TrimSpace(code))
	base = code
	if i := strings.LastIndexByte(code, '_'); i >= 0 {
		base, variant = code[:i], code[i+1:]
	}
	if fixed, ok := metTypos[base]; ok {
		base = fixed
	}
	if variant == "polartwilight" {
		variant = "day"
	}
	return base, variant
}

// normalizeSymbol cleans a provider code into our vocabulary: typos fixed,
// polar twilight dropped, the variant added or removed as the base needs.
// An unknown code becomes "cloudy", which is never wrong by much.
func normalizeSymbol(code string) string {
	base, variant := splitSymbol(code)
	if !knownBase[base] {
		return "cloudy"
	}
	return withVariant(base, variant != "night")
}

// withVariant puts the day or night suffix on a base code that has one.
func withVariant(base string, day bool) string {
	if !hasDayNight(base) {
		return base
	}
	if day {
		return base + "_day"
	}
	return base + "_night"
}

// symbolHU is the Hungarian text for each base code. Admin overrides go in
// weather_desc_translations with the base code as source_text.
var symbolHU = map[string]string{
	"clearsky":                    "derült",
	"fair":                        "túlnyomóan derült",
	"partlycloudy":                "részben felhős",
	"cloudy":                      "borult",
	"fog":                         "köd",
	"lightrainshowers":            "gyenge zápor",
	"rainshowers":                 "zápor",
	"heavyrainshowers":            "erős zápor",
	"lightrainshowersandthunder":  "gyenge zápor, zivatar",
	"rainshowersandthunder":       "zápor, zivatar",
	"heavyrainshowersandthunder":  "erős zápor, zivatar",
	"lightsleetshowers":           "gyenge havas zápor",
	"sleetshowers":                "havas zápor",
	"heavysleetshowers":           "erős havas zápor",
	"lightsleetshowersandthunder": "gyenge havas zápor, zivatar",
	"sleetshowersandthunder":      "havas zápor, zivatar",
	"heavysleetshowersandthunder": "erős havas zápor, zivatar",
	"lightsnowshowers":            "gyenge hózápor",
	"snowshowers":                 "hózápor",
	"heavysnowshowers":            "erős hózápor",
	"lightsnowshowersandthunder":  "gyenge hózápor, zivatar",
	"snowshowersandthunder":       "hózápor, zivatar",
	"heavysnowshowersandthunder":  "erős hózápor, zivatar",
	"lightrain":                   "gyenge eső",
	"rain":                        "eső",
	"heavyrain":                   "erős eső",
	"lightrainandthunder":         "gyenge eső, zivatar",
	"rainandthunder":              "eső, zivatar",
	"heavyrainandthunder":         "erős eső, zivatar",
	"lightsleet":                  "gyenge havas eső",
	"sleet":                       "havas eső",
	"heavysleet":                  "erős havas eső",
	"lightsleetandthunder":        "gyenge havas eső, zivatar",
	"sleetandthunder":             "havas eső, zivatar",
	"heavysleetandthunder":        "erős havas eső, zivatar",
	"lightsnow":                   "gyenge havazás",
	"snow":                        "havazás",
	"heavysnow":                   "erős havazás",
	"lightsnowandthunder":         "gyenge havazás, zivatar",
	"snowandthunder":              "havazás, zivatar",
	"heavysnowandthunder":         "erős havazás, zivatar",
}

// symbolDescHU is the default Hungarian text for a code (no DB lookup).
func symbolDescHU(code string) string {
	base, _ := splitSymbol(code)
	if hu, ok := symbolHU[base]; ok {
		return hu
	}
	return symbolHU["cloudy"]
}

// legacyIcon turns a symbol into the OpenWeatherMap-style icon code the
// emoji widgets and older clients still read ("01d", "10n", ...).
func legacyIcon(code string) string {
	base, variant := splitSymbol(code)
	suffix := "d"
	if variant == "night" {
		suffix = "n"
	}
	var n string
	switch {
	case strings.Contains(base, "thunder"):
		n = "11"
	case strings.Contains(base, "snow"):
		n = "13"
	case strings.Contains(base, "showers"):
		n = "09"
	case strings.Contains(base, "rain"), strings.Contains(base, "sleet"):
		n = "10"
	case base == "clearsky":
		n = "01"
	case base == "fair":
		n = "02"
	case base == "partlycloudy":
		n = "03"
	case base == "fog":
		n = "50"
	default:
		n = "04"
	}
	return n + suffix
}

// weatherAPIComSymbol maps a WeatherAPI.com condition code
// (https://www.weatherapi.com/docs/weather_conditions.json) to a base code.
func weatherAPIComSymbol(code int) string {
	switch code {
	case 1000:
		return "clearsky"
	case 1003:
		return "partlycloudy"
	case 1006, 1009:
		return "cloudy"
	case 1030, 1135, 1147:
		return "fog"
	case 1063, 1180:
		return "lightrainshowers"
	case 1066, 1210, 1255:
		return "lightsnowshowers"
	case 1069, 1249, 1261:
		return "lightsleetshowers"
	case 1072, 1168, 1198, 1204:
		return "lightsleet"
	case 1087, 1273:
		return "lightrainshowersandthunder"
	case 1114, 1219:
		return "snow"
	case 1117, 1225:
		return "heavysnow"
	case 1150, 1153, 1183:
		return "lightrain"
	case 1171, 1201, 1207, 1237:
		return "sleet"
	case 1186, 1240, 1243:
		return "rainshowers"
	case 1189:
		return "rain"
	case 1192, 1246:
		return "heavyrainshowers"
	case 1195:
		return "heavyrain"
	case 1213:
		return "lightsnow"
	case 1216, 1258:
		return "snowshowers"
	case 1222:
		return "heavysnowshowers"
	case 1252, 1264:
		return "sleetshowers"
	case 1276:
		return "heavyrainandthunder"
	case 1279:
		return "lightsnowshowersandthunder"
	case 1282:
		return "heavysnowandthunder"
	}
	return "cloudy"
}

// owmSymbol maps an OpenWeatherMap condition id
// (https://openweathermap.org/weather-conditions) to a base code.
func owmSymbol(id int) string {
	switch {
	case id == 200 || (id >= 230 && id <= 232):
		return "lightrainandthunder"
	case id == 201:
		return "rainandthunder"
	case id == 202:
		return "heavyrainandthunder"
	case id == 210:
		return "lightrainshowersandthunder"
	case id == 211 || id == 221:
		return "rainshowersandthunder"
	case id == 212:
		return "heavyrainshowersandthunder"
	case id == 302 || id == 312 || id == 314:
		return "rain"
	case id == 313 || id == 321:
		return "lightrainshowers"
	case id >= 300 && id < 400:
		return "lightrain"
	case id == 500:
		return "lightrain"
	case id == 501:
		return "rain"
	case id >= 502 && id <= 504:
		return "heavyrain"
	case id == 511 || id == 611 || id == 616:
		return "sleet"
	case id == 520:
		return "lightrainshowers"
	case id == 521:
		return "rainshowers"
	case id == 522 || id == 531:
		return "heavyrainshowers"
	case id == 600:
		return "lightsnow"
	case id == 601:
		return "snow"
	case id == 602:
		return "heavysnow"
	case id == 612:
		return "lightsleetshowers"
	case id == 613:
		return "sleetshowers"
	case id == 615:
		return "lightsleet"
	case id == 620:
		return "lightsnowshowers"
	case id == 621:
		return "snowshowers"
	case id == 622:
		return "heavysnowshowers"
	case id == 781:
		return "heavyrainshowersandthunder"
	case id >= 700 && id < 800:
		return "fog"
	case id == 800:
		return "clearsky"
	case id == 801:
		return "fair"
	case id == 802:
		return "partlycloudy"
	}
	return "cloudy"
}
