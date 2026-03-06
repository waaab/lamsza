package utils

import (
	"fmt"
	"strings"
	"unicode"
)

func Slugify(s string) string {
	s = strings.ToLower(s)
	// Simple Hungarian replacement
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ö", "o", "ő", "o", "ú", "u", "ü", "u", "ű", "u",
	)
	s = replacer.Replace(s)

	var res strings.Builder
	lastDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			res.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			res.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(res.String(), "-")
}

func JavaToPg(i int) string {
	return fmt.Sprintf("%d", i)
}
