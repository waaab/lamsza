package main

import (
	"backend/internal/utils"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"Árvíztűrő tükörfúrógép", "arvizturo-tukorfurogep"},
		{"Simple-Test", "simple-test"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"Special!@#$%^&*()Chars", "special-chars"},
		{"--leading-trailing--", "leading-trailing"},
	}

	for _, tt := range tests {
		actual := utils.Slugify(tt.input)
		if actual != tt.expected {
			t.Errorf("Slugify(%q) = %q; want %q", tt.input, actual, tt.expected)
		}
	}
}
