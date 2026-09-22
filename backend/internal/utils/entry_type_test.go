package utils

import "testing"

func TestCanonicalEntryType(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"service", "Szolgáltatás"},
		{"Service", "Szolgáltatás"},
		{"Szolgáltatás", "Szolgáltatás"},
		{"  szolgáltatás  ", "Szolgáltatás"},
		{"Cég", "Cég"},
		{"ceg", "Cég"},
		{"company", "Cég"},
		{"entry", "Egyéb"},
		{"Egyéb", "Egyéb"},
		{"other", "Egyéb"},
		{"", ""},
		{"  ", ""},
	}
	for _, c := range cases {
		got := CanonicalEntryType(c.in)
		if got != c.want {
			t.Errorf("CanonicalEntryType(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
