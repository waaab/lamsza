package utils

import "testing"

func TestCanonicalEntryCategory(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Bútor", "Bútor"},
		{"  Étterem  ", "Étterem"},
		{"egeszsegugy", "egeszsegugy"},
		{"Vendéglő", "Vendéglő"},
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

func TestSeedEntryCategoriesCount(t *testing.T) {
	if len(SeedEntryCategories()) != 68 {
		t.Fatalf("SeedEntryCategories len = %d, want 68", len(SeedEntryCategories()))
	}
}
