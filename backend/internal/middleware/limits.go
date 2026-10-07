package middleware

import (
	"net/http"
	"strings"
)

// Request body limits.
//
// Every handler in this backend decodes r.Body without a size check. A single
// large POST to an unauthenticated route (/api/auth/google) would read the
// whole body into memory and can kill the process. http.MaxBytesReader caps
// the read and makes the decoder fail with a normal error instead, so the
// handlers need no change.
//
// Apply LimitBody once, around the whole mux.
const (
	// DefaultMaxBodyBytes covers every JSON route. The largest real JSON
	// payload is a listing or account body, far under 1 MiB.
	DefaultMaxBodyBytes = 1 << 20 // 1 MiB

	// ImportMaxBodyBytes covers /api/account/import, which posts a whole
	// browser-local favourites and history dump in one request.
	ImportMaxBodyBytes = 4 << 20 // 4 MiB
)

// No upload limit here: this backend has no multipart route. Entry and event
// image uploads moved to lamsza-admin with the admin handlers (BOG-42).

// bodyLimitByPrefix holds the routes that need more than the default. Longest
// matching prefix wins, so an exact path and a subtree can both be listed.
var bodyLimitByPrefix = map[string]int64{
	"/api/account/import": ImportMaxBodyBytes,
}

// MaxBodyBytesFor returns the body limit that applies to a request path.
func MaxBodyBytesFor(path string) int64 {
	var (
		best    int64 = DefaultMaxBodyBytes
		bestLen       = -1
	)
	for prefix, limit := range bodyLimitByPrefix {
		if len(prefix) > bestLen && strings.HasPrefix(path, prefix) {
			best = limit
			bestLen = len(prefix)
		}
	}
	return best
}

// LimitBody caps how much of the request body a handler can read.
func LimitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil && r.Method != http.MethodGet && r.Method != http.MethodHead {
			r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytesFor(r.URL.Path))
		}
		next.ServeHTTP(w, r)
	})
}
