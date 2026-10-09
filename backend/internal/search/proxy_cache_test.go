package search

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fakeHost answers like an image host and counts what it was asked.
type fakeHost struct {
	calls  int
	ua     string
	status int
	ctype  string
}

func (f *fakeHost) do(req *http.Request) (*http.Response, error) {
	f.calls++
	f.ua = req.Header.Get("User-Agent")
	return &http.Response{
		StatusCode: f.status,
		Header:     http.Header{"Content-Type": []string{f.ctype}},
		Body:       io.NopCloser(bytes.NewReader([]byte("crest-bytes"))),
	}, nil
}

func proxied(t *testing.T, target string, host *fakeHost) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	serveProxied(rr, httptest.NewRequest("GET", "/api/proxy", nil), target, host.do)
	return rr
}

func withFreshCache(t *testing.T) *time.Time {
	t.Helper()
	old := proxyCache
	clock := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	proxyCache = newImageCache()
	proxyCache.now = func() time.Time { return clock }
	t.Cleanup(func() { proxyCache = old })
	return &clock
}

// A crest is fetched once a day, with a User-Agent Wikimedia accepts, and the
// browser may keep it for a day.
func TestProxyCachesImagesForADay(t *testing.T) {
	clock := withFreshCache(t)
	host := &fakeHost{status: 200, ctype: "image/png"}
	const crest = "https://upload.wikimedia.org/wikipedia/commons/f/f7/ROU_HR_Miercurea_Ciuc_CoA.PNG"

	rr := proxied(t, crest, host)
	if rr.Code != 200 || rr.Body.String() != "crest-bytes" || rr.Header().Get("X-Proxy-Cache") != "MISS" {
		t.Fatalf("first: %d %q %s", rr.Code, rr.Body.String(), rr.Header().Get("X-Proxy-Cache"))
	}
	if !strings.Contains(host.ua, "lamsza.com") || !strings.HasPrefix(host.ua, "LamszaImageProxy/") {
		t.Errorf("User-Agent %q does not name the site", host.ua)
	}
	if rr.Header().Get("Cache-Control") != browserMaxAge {
		t.Errorf("Cache-Control %q", rr.Header().Get("Cache-Control"))
	}

	*clock = clock.Add(23 * time.Hour)
	rr = proxied(t, crest, host)
	if host.calls != 1 || rr.Header().Get("X-Proxy-Cache") != "HIT" || rr.Body.String() != "crest-bytes" {
		t.Fatalf("within a day: %d calls, %s", host.calls, rr.Header().Get("X-Proxy-Cache"))
	}

	*clock = clock.Add(2 * time.Hour)
	proxied(t, crest, host)
	if host.calls != 2 {
		t.Fatalf("after a day the host is asked again: %d calls", host.calls)
	}
}

// When the host rate-limits or fails, a stale copy (up to a week old) is
// served instead of the error; without one, the error passes through and
// nothing is cached.
func TestProxyServesStaleWhenTheHostRefuses(t *testing.T) {
	clock := withFreshCache(t)
	const crest = "https://upload.wikimedia.org/wikipedia/commons/8/8a/ROU_CV_Sfantu_Gheorghe_CoA.svg"
	host := &fakeHost{status: 429, ctype: "text/html"}
	if rr := proxied(t, crest, host); rr.Code != 429 {
		t.Fatalf("no copy yet: %d, want the host's 429", rr.Code)
	}
	if rr := proxied(t, crest, host); rr.Code != 429 || host.calls != 2 {
		t.Fatalf("a refusal is not cached: %d, %d calls", rr.Code, host.calls)
	}

	host.status, host.ctype = 200, "image/svg+xml"
	proxied(t, crest, host)
	*clock = clock.Add(3 * 24 * time.Hour)
	host.status, host.ctype = 429, "text/html"
	rr := proxied(t, crest, host)
	if rr.Code != 200 || rr.Header().Get("X-Proxy-Cache") != "STALE" || rr.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("stale copy: %d %s %s", rr.Code, rr.Header().Get("X-Proxy-Cache"), rr.Header().Get("Content-Type"))
	}

	*clock = clock.Add(5 * 24 * time.Hour)
	if rr := proxied(t, crest, host); rr.Code != 429 {
		t.Fatalf("older than a week the copy is gone: %d", rr.Code)
	}
}

// A news feed (not an image) is never cached.
func TestProxyDoesNotCacheFeeds(t *testing.T) {
	withFreshCache(t)
	host := &fakeHost{status: 200, ctype: "application/rss+xml"}
	proxied(t, "https://example.com/feed.xml", host)
	proxied(t, "https://example.com/feed.xml", host)
	if host.calls != 2 {
		t.Fatalf("feed fetched %d times, want 2", host.calls)
	}
}

// The cache holds at most proxyCacheMax bytes, dropping the oldest first.
func TestProxyCacheEvictsTheOldest(t *testing.T) {
	withFreshCache(t)
	big := make([]byte, proxyCacheMax/2+1)
	proxyCache.put("a", &cachedImage{body: big, fetched: proxyCache.now()})
	proxyCache.put("b", &cachedImage{body: big, fetched: proxyCache.now()})
	if _, ok := proxyCache.get("a"); ok {
		t.Fatal("the oldest entry should have been dropped")
	}
	if _, ok := proxyCache.get("b"); !ok || proxyCache.size > proxyCacheMax {
		t.Fatalf("newest kept, size %d", proxyCache.size)
	}
}
