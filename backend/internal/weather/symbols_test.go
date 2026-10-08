package weather

import "testing"

func TestNormalizeSymbol(t *testing.T) {
	for in, want := range map[string]string{
		"fair_night":                      "fair_night",
		"partlycloudy_polartwilight":      "partlycloudy_day",
		"lightssnowshowersandthunder_day": "lightsnowshowersandthunder_day",
		"lightssleetshowersandthunder":    "lightsleetshowersandthunder_day",
		"cloudy_night":                    "cloudy",
		"heavyrain":                       "heavyrain",
		"rainshowers":                     "rainshowers_day",
		"  ClearSky_Day ":                 "clearsky_day",
		"tornado":                         "cloudy",
		"":                                "cloudy",
	} {
		if got := normalizeSymbol(in); got != want {
			t.Errorf("normalizeSymbol(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEverySymbolHasHungarianText(t *testing.T) {
	if len(symbolBases) != 41 {
		t.Fatalf("MET has 41 base codes, the list has %d", len(symbolBases))
	}
	for _, b := range symbolBases {
		if symbolHU[b] == "" {
			t.Errorf("%s has no Hungarian text", b)
		}
	}
	if len(symbolHU) != len(symbolBases) {
		t.Errorf("symbolHU has %d entries for %d codes", len(symbolHU), len(symbolBases))
	}
}

func TestLegacyIcon(t *testing.T) {
	for in, want := range map[string]string{
		"clearsky_day":                   "01d",
		"clearsky_night":                 "01n",
		"fair_day":                       "02d",
		"partlycloudy_night":             "03n",
		"cloudy":                         "04d",
		"fog":                            "50d",
		"lightrainshowers_night":         "09n",
		"rain":                           "10d",
		"sleet":                          "10d",
		"snowshowers_day":                "13d",
		"heavyrainshowersandthunder_day": "11d",
	} {
		if got := legacyIcon(in); got != want {
			t.Errorf("legacyIcon(%q) = %q, want %q", in, got, want)
		}
	}
}

// Every provider code lands on a code we can draw and describe.
func TestProviderCodesMapToKnownSymbols(t *testing.T) {
	for code := 1000; code <= 1282; code++ {
		if b := weatherAPIComSymbol(code); !knownBase[b] {
			t.Fatalf("weatherapi %d -> %q", code, b)
		}
	}
	for id := 200; id <= 804; id++ {
		if b := owmSymbol(id); !knownBase[b] {
			t.Fatalf("owm %d -> %q", id, b)
		}
	}
	checks := map[int]string{1000: "clearsky", 1189: "rain", 1225: "heavysnow", 1276: "heavyrainandthunder"}
	for code, want := range checks {
		if got := weatherAPIComSymbol(code); got != want {
			t.Errorf("weatherapi %d = %q, want %q", code, got, want)
		}
	}
	for id, want := range map[int]string{800: "clearsky", 801: "fair", 804: "cloudy", 741: "fog", 502: "heavyrain", 211: "rainshowersandthunder"} {
		if got := owmSymbol(id); got != want {
			t.Errorf("owm %d = %q, want %q", id, got, want)
		}
	}
}
