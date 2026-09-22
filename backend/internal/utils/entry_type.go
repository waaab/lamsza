package utils

import "strings"

const (
	EntryTypeService = "Szolgáltatás"
	EntryTypeCeg     = "Cég"
	EntryTypeEgyeb   = "Egyéb"
)

// CanonicalEntryType maps stored / legacy entries.type values onto catalog labels.
func CanonicalEntryType(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	switch Slugify(trimmed) {
	case "service", "szolgaltatas":
		return EntryTypeService
	case "ceg", "company":
		return EntryTypeCeg
	case "entry", "egyeb", "other":
		return EntryTypeEgyeb
	default:
		return trimmed
	}
}

// DefaultEntryType is used when admin creates/updates an entry with an empty type.
func DefaultEntryType() string {
	return EntryTypeService
}
