package news

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"backend/internal/models"
)

// Székelyföld.ma's feed (sample from 2026-10-09) writes its host twice in item
// links and serves images over http, which the pages' CSP blocks. The parsed
// item gets one host and an https image.
func TestFeedLinkAndImageAreRepaired(t *testing.T) {
	body, err := os.ReadFile("testdata/szekelyfold_feed.xml")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.Write(body) }))
	t.Cleanup(srv.Close)

	items := fetchAndParseFeed(models.NewsFeed{Title: "Székelyfold.ma", FeedURL: srv.URL}, 5)
	if len(items) != 1 {
		t.Fatalf("items = %d, want 1", len(items))
	}
	if want := "https://szekelyfold.ma/szekelyfold/ket-fiu-eletet-vesztette-miutan-motorkerekparjukat-elsodorta-a-vonat-csikverebesen"; items[0].Link != want {
		t.Errorf("link = %q, want %q", items[0].Link, want)
	}
	if want := "https://szekelyfold.ma/api/images/post-header/2026-10/20261009113348-v.jpg"; items[0].Image != want {
		t.Errorf("image = %q, want %q", items[0].Image, want)
	}
}

func TestCleanLinkLeavesOrdinaryLinksAlone(t *testing.T) {
	for _, link := range []string{
		"https://kronika.ro/erdelyi-hirek/valami?utm=rss&x=%C3%A9",
		"https://aa.aa/hir", // a host made of one repeated part, not a doubled host
		"not a url",
		"",
	} {
		if got := cleanLink(link); got != link {
			t.Errorf("cleanLink(%q) = %q, want it unchanged", link, got)
		}
	}
	if got := cleanLink("https://hir.rohir.ro/a"); got != "https://hir.ro/a" {
		t.Errorf("doubled host: got %q", got)
	}
}

func TestSecureImage(t *testing.T) {
	cases := map[string]string{
		"http://example.ro/a.jpg":  "https://example.ro/a.jpg",
		"https://example.ro/a.jpg": "https://example.ro/a.jpg",
		"":                         "",
	}
	for in, want := range cases {
		if got := secureImage(in); got != want {
			t.Errorf("secureImage(%q) = %q, want %q", in, got, want)
		}
	}
}
