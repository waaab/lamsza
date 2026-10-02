package utils

import "strings"

const (
	EntryCategoryEgeszsegugy    = "Egészségügy"
	EntryCategoryOktatas        = "Oktatás"
	EntryCategoryMesteremberek  = "Mesteremberek"
	EntryCategoryHivatalok      = "Hivatalok"
	EntryCategoryVendeglo       = "Vendéglő"
	EntryCategoryEtterem        = "Étterem"
	EntryCategoryBolt           = "Bolt"
	EntryCategorySportegyesulet = "Sportegyesület"
	EntryCategoryEgyeb          = "Egyéb"
)

// SeedEntryCategories returns every name in the v2 directory catalog.
func SeedEntryCategories() []string {
	names := make([]string, 0, 68)
	for _, node := range DirectoryParents() {
		names = append(names, node.Name)
	}
	for _, node := range DirectoryChildren() {
		names = append(names, node.Name)
	}
	return names
}

// CanonicalEntryCategory returns the stored category name.
func CanonicalEntryCategory(raw string) string {
	return strings.TrimSpace(raw)
}

// DefaultEntryCategory is used when an entry has no category.
func DefaultEntryCategory() string {
	return ""
}
