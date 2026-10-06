package handlers

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

const entryImageURLPrefix = "/api/media/entry-images/"
const maxEntryImageBytes = 8 << 20
const defaultPhotoWidth = 1600
const defaultPhotoHeight = 1200

// EntryImagesDir returns the filesystem directory for uploaded entry gallery images.
func EntryImagesDir() string {
	d := os.Getenv("ENTRY_IMAGES_DIR")
	if d == "" {
		d = "data/entry-images"
	}
	return d
}
