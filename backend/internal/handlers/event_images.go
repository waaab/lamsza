package handlers

import (
	"os"
)

const eventImageURLPrefix = "/api/media/event-images/"
const maxEventImageBytes = 8 << 20

// EventImagesDir returns the filesystem directory for uploaded event images (default: data/event-images under cwd).
func EventImagesDir() string {
	d := os.Getenv("EVENT_IMAGES_DIR")
	if d == "" {
		d = "data/event-images"
	}
	return d
}
