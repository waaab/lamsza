package utils

import "testing"

func TestCanonicalEntryCategory(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"egeszsegugy", "Egészségügy"},
		{"Egészségügy", "Egészségügy"},
		{"  egészségügy  ", "Egészségügy"},
		{"Orvosi rendelők", "Egészségügy"},
		{"oktatas", "Oktatás"},
		{"Oktatás", "Oktatás"},
		{"mesteremberek", "Mesteremberek"},
		{"Mesterember", "Mesteremberek"},
		{"hivatalok", "Hivatalok"},
		{"Hivatalok", "Hivatalok"},
		{"városháza", "Hivatalok"},
		{"Vendéglő", "Vendéglő"},
		{"vendeglo", "Vendéglő"},
		{"bolt", "Bolt"},
		{"sportegyesület", "Sportegyesület"},
		{"egyeb", "Egyéb"},
		{"Egyéb", "Egyéb"},
		{"", ""},
		{"  ", ""},
	}
	for _, c := range cases {
		got := CanonicalEntryCategory(c.in)
		if got != c.want {
			t.Errorf("CanonicalEntryCategory(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
