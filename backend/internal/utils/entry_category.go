package utils

import "strings"

const (
	EntryCategoryEgeszsegugy    = "Egészségügy"
	EntryCategoryOktatas        = "Oktatás"
	EntryCategoryMesteremberek  = "Mesteremberek"
	EntryCategoryHivatalok      = "Hivatalok"
	EntryCategoryVendeglo       = "Vendéglő"
	EntryCategoryBolt           = "Bolt"
	EntryCategorySportegyesulet = "Sportegyesület"
	EntryCategoryEgyeb          = "Egyéb"
)

// SeedEntryCategories is the short directory catalog (Index slugs + in-use sectors).
func SeedEntryCategories() []string {
	return []string{
		EntryCategoryEgeszsegugy,
		EntryCategoryOktatas,
		EntryCategoryMesteremberek,
		EntryCategoryHivatalok,
		EntryCategoryVendeglo,
		EntryCategoryBolt,
		EntryCategorySportegyesulet,
		EntryCategoryEgyeb,
	}
}

// CanonicalEntryCategory maps stored / legacy category strings onto catalog labels.
func CanonicalEntryCategory(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	switch Slugify(trimmed) {
	case "egeszsegugy", "orvosi-rendelok", "orvosi-rendelo":
		return EntryCategoryEgeszsegugy
	case "oktatas":
		return EntryCategoryOktatas
	case "mesteremberek", "mesterember":
		return EntryCategoryMesteremberek
	case "hivatalok", "varoshaza":
		return EntryCategoryHivatalok
	case "vendeglo", "vendeglatas":
		return EntryCategoryVendeglo
	case "bolt", "kereskedelem":
		return EntryCategoryBolt
	case "sportegyesulet", "sport":
		return EntryCategorySportegyesulet
	case "egyeb", "other":
		return EntryCategoryEgyeb
	default:
		return trimmed
	}
}

// DefaultEntryCategory is used when an entry has no category.
func DefaultEntryCategory() string {
	return EntryCategoryEgyeb
}
