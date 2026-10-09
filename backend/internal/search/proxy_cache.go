package search

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// The proxy serves crests and attraction photos from other hosts, Wikimedia
// above all. Fetching them on every page view got the site rate-limited
// (429 Too Many Requests, 2026-10-09), and the page then showed a broken
// image. So successful image answers are kept here: fresh for a day, and up
// to a week as a stale copy that is served when the host refuses or fails.
// News feeds and error answers pass through uncached.

// proxyUserAgent follows Wikimedia's User-Agent policy (a name, a version and
// a way to reach the operator); anonymous-looking clients are throttled first.
const proxyUserAgent = "LamszaImageProxy/1.3 (https://lamsza.com) Go-http-client"

const (
	proxyFreshFor   = 24 * time.Hour
	proxyKeepFor    = 7 * 24 * time.Hour
	proxyCacheMax   = 64 << 20 // bytes of image data held in memory
	browserMaxAge   = "public, max-age=86400"
	proxyCacheHit   = "HIT"
	proxyCacheMiss  = "MISS"
	proxyCacheStale = "STALE"
)

type cachedImage struct {
	body        []byte
	contentType string
	fetched     time.Time
}

type imageCache struct {
	mu      sync.Mutex
	entries map[string]*cachedImage
	order   []string // oldest first, for eviction
	size    int
	now     func() time.Time
}

var proxyCache = newImageCache()

func newImageCache() *imageCache {
	return &imageCache{entries: map[string]*cachedImage{}, now: time.Now}
}

func (c *imageCache) get(key string) (*cachedImage, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if c.now().Sub(e.fetched) > proxyKeepFor {
		c.removeLocked(key)
		return nil, false
	}
	return e, true
}

func (c *imageCache) put(key string, e *cachedImage) {
	if len(e.body) > proxyCacheMax {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.removeLocked(key)
	c.entries[key] = e
	c.order = append(c.order, key)
	c.size += len(e.body)
	for c.size > proxyCacheMax && len(c.order) > 0 {
		c.removeLocked(c.order[0])
	}
}

func (c *imageCache) removeLocked(key string) {
	e, ok := c.entries[key]
	if !ok {
		return
	}
	delete(c.entries, key)
	c.size -= len(e.body)
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}

func (c *imageCache) fresh(e *cachedImage) bool {
	return c.now().Sub(e.fetched) <= proxyFreshFor
}

// serveProxied answers with the target's content: from the cache while it is
// fresh, otherwise from the host (do), keeping a successful image answer. When
// the host refuses or fails and a stale copy exists, the stale copy is served.
func serveProxied(w http.ResponseWriter, r *http.Request, target string, do func(*http.Request) (*http.Response, error)) {
	cached, ok := proxyCache.get(target)
	if ok && proxyCache.fresh(cached) {
		writeCachedImage(w, cached, proxyCacheHit)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	req.Header.Set("User-Agent", proxyUserAgent)
	req.Header.Set("Accept", "*/*")

	resp, err := do(req)
	if err != nil {
		if ok {
			writeCachedImage(w, cached, proxyCacheStale)
			return
		}
		http.Error(w, "upstream failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	isImage := strings.HasPrefix(strings.ToLower(contentType), "image/")
	if resp.StatusCode != http.StatusOK && ok && (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) {
		writeCachedImage(w, cached, proxyCacheStale)
		return
	}
	if resp.StatusCode == http.StatusOK && isImage {
		body, err := io.ReadAll(io.LimitReader(resp.Body, proxyMaxBytes+1))
		if err != nil {
			http.Error(w, "upstream failed", http.StatusBadGateway)
			return
		}
		if len(body) <= proxyMaxBytes {
			e := &cachedImage{body: body, contentType: contentType, fetched: proxyCache.now()}
			proxyCache.put(target, e)
			writeCachedImage(w, e, proxyCacheMiss)
			return
		}
		// Larger than the proxy serves: the first proxyMaxBytes, as before.
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, bytes.NewReader(body[:proxyMaxBytes]))
		return
	}

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "" {
		w.Header().Set("Cache-Control", cc)
	}
	w.Header().Set("X-Proxy-Cache", proxyCacheMiss)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, io.LimitReader(resp.Body, proxyMaxBytes))
}

func writeCachedImage(w http.ResponseWriter, e *cachedImage, state string) {
	w.Header().Set("Content-Type", e.contentType)
	w.Header().Set("Cache-Control", browserMaxAge)
	w.Header().Set("X-Proxy-Cache", state)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(e.body)
}
