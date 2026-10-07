package news

import (
	"testing"
	"time"
)

func TestPeekNewsItemsNeverLoadsFeeds(t *testing.T) {
	resetNewsCacheForTest()
	t.Cleanup(resetNewsCacheForTest)

	previous := loadFeedItems
	calls := 0
	loadFeedItems = func() []newsItem {
		calls++
		return []newsItem{{Title: "live", Link: "https://example.test/live"}}
	}
	t.Cleanup(func() { loadFeedItems = previous })

	if items := PeekNewsItems(10); len(items) != 0 {
		t.Fatalf("empty cache = %#v, want none", items)
	}

	newsMu.Lock()
	newsItemsCached = []newsItem{{Title: "stale", Link: "https://example.test/stale"}}
	newsCachedAt = time.Now().Add(-time.Hour)
	newsMu.Unlock()

	items := PeekNewsItems(10)
	if len(items) != 1 || items[0].Title != "stale" {
		t.Fatalf("stale cache = %#v", items)
	}
	time.Sleep(30 * time.Millisecond)
	if calls != 0 {
		t.Fatalf("search loaded live feeds %d times", calls)
	}
}

func TestPeekNewsItemsUsesFreshCache(t *testing.T) {
	resetNewsCacheForTest()
	t.Cleanup(resetNewsCacheForTest)

	previous := loadFeedItems
	calls := 0
	loadFeedItems = func() []newsItem {
		calls++
		return []newsItem{{Title: "cached", Link: "https://example.test/cached"}}
	}
	t.Cleanup(func() { loadFeedItems = previous })

	first := FetchNewsItems(10)
	if len(first) != 1 || first[0].Title != "cached" {
		t.Fatalf("blocking fetch = %#v", first)
	}
	if calls != 1 {
		t.Fatalf("loads = %d, want 1", calls)
	}

	second := PeekNewsItems(10)
	if len(second) != 1 || second[0].Title != "cached" {
		t.Fatalf("peek = %#v", second)
	}
	if calls != 1 {
		t.Fatalf("fresh cache refetched feeds, loads = %d", calls)
	}
}

// A stale cache is served at once and refreshed in the background; only a
// cold cache makes the caller wait for the feeds.
func TestStaleNewsIsServedWhileRefreshing(t *testing.T) {
	resetNewsCacheForTest()
	t.Cleanup(resetNewsCacheForTest)
	previous := loadFeedItems
	release := make(chan struct{})
	loaded := make(chan struct{}, 1)
	loadFeedItems = func() []newsItem {
		<-release
		loaded <- struct{}{}
		return []newsItem{{Title: "new", Link: "https://example.test/new"}}
	}
	t.Cleanup(func() { loadFeedItems = previous })

	newsMu.Lock()
	newsItemsCached = []newsItem{{Title: "stale", Link: "https://example.test/stale"}}
	newsCachedAt = time.Now().Add(-time.Hour)
	newsMu.Unlock()

	done := make(chan []newsItem, 1)
	go func() { done <- ensureNews() }()
	select {
	case items := <-done:
		if len(items) != 1 || items[0].Title != "stale" {
			t.Fatalf("stale read = %#v, want the stale item", items)
		}
	case <-time.After(time.Second):
		t.Fatal("ensureNews waited for the feeds although a stale cache was there")
	}

	close(release)
	select {
	case <-loaded:
	case <-time.After(time.Second):
		t.Fatal("no background refresh was started")
	}
	deadline := time.Now().Add(time.Second)
	for {
		items := cachedNewsItems()
		if len(items) == 1 && items[0].Title == "new" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("cache after refresh = %#v, want the new item", items)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
