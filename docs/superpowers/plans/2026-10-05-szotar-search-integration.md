# Szótár in Main Search Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Include Székely Szótár headwords in Lámsza unified search results, with a fourth filter pill that shows only those hits, and open each word on the Szótár app in a new tab.

**Architecture:** Lámsza’s `HandleUnifiedSearch` fetches `{SZOTAR_ORIGIN}/api/words?q=…` server-side in parallel with existing sources and returns a `words` array. The frontend maps those hits through `searchResultCardModel("word", …)` (external `szotarUrl('/szo/{id}')`) and adds a **Székely Szótár** result filter after Weboldalak.

**Tech Stack:** Go (Lámsza backend), Svelte (existing `SearchEngine` / `SearchResultCard`), `networkOrigins.js`, `node --test`, Go `testing` + `httptest`.

**Spec:** `docs/superpowers/specs/2026-10-05-szotar-search-integration-design.md`

## Global Constraints

- Hungarian UI label: **Székely Szótár** (accents). Scope text when filtered: `szótárban:`.
- No em dash (U+2014). Use a hyphen with spaces if a dash is needed: ` - `.
- Do not change Szótár’s own app, CORS, or database.
- Do not sync dictionary rows into the Lámsza DB.
- Do not call Szótár from the browser; only from the Lámsza backend.
- Cap dictionary hits at 15. Upstream failure or empty `SZOTAR_ORIGIN` yields `words: []` without failing the rest of search.
- Word rows open in a new tab (`external: true`, `rel="nofollow noopener"`).
- Reuse `SearchResultCard`; no new accent color.
- Do not change `/api/suggest`.
- Follow existing Lámsza network rules (theme, toolbar, wording) when touching shared UI.

## File map

| File | Role |
|---|---|
| `backend/internal/config/config.go` | Load `SZOTAR_ORIGIN` |
| `backend/internal/search/szotar.go` | Fetch + map + cap upstream words |
| `backend/internal/search/szotar_test.go` | Unit tests with `httptest.Server` |
| `backend/internal/search/unified.go` | Add `Words` to result; call fetch in parallel |
| `src/lib/searchResultCard.js` | `kind === "word"` card model |
| `tests/searchResultCard.test.js` | Word card tests |
| `src/lib/components/SearchEngine.svelte` | Filter, section, counts, scope |
| `docs/PRODUCTION_SERVER_SETUP.md` | Document `SZOTAR_ORIGIN` for Lámsza |

---

### Task 1: Backend Szótár fetch helper

**Files:**
- Create: `backend/internal/search/szotar.go`
- Create: `backend/internal/search/szotar_test.go`
- Modify: `backend/internal/config/config.go`

**Interfaces:**
- Consumes: `config.AppConfig.SzotarOrigin` (string, may be empty)
- Produces:
  - `type WordSearchHit struct { ID int; Headword string; DefinitionHu string; SpeechTypes []string }` with JSON tags `id`, `headword`, `definition_hu`, `speech_types`
  - `func fetchSzotarWords(ctx context.Context, origin, q string, client *http.Client) []WordSearchHit`
  - Always returns a non-nil slice. Cap 15. Empty `origin` or `q` → empty slice, no HTTP call.

- [ ] **Step 1: Write the failing tests**

Create `backend/internal/search/szotar_test.go`:

```go
package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchSzotarWordsMapsAndCaps(t *testing.T) {
	var gotQ string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/words" {
			t.Fatalf("path %s", r.URL.Path)
		}
		gotQ = r.URL.Query().Get("q")
		words := make([]map[string]any, 0, 20)
		for i := 1; i <= 20; i++ {
			words = append(words, map[string]any{
				"id":            i,
				"headword":      "szo" + string(rune('a'+i%26)),
				"definition_hu": "def",
				"speech_types":  []string{"főnév"},
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"words": words})
	}))
	defer srv.Close()

	hits := fetchSzotarWords(context.Background(), srv.URL, "gagya", srv.Client())
	if gotQ != "gagya" {
		t.Fatalf("q=%q", gotQ)
	}
	if len(hits) != 15 {
		t.Fatalf("cap got %d", len(hits))
	}
	if hits[0].ID != 1 || hits[0].DefinitionHu != "def" || len(hits[0].SpeechTypes) != 1 {
		t.Fatalf("first hit %#v", hits[0])
	}
}

func TestFetchSzotarWordsEmptyOrigin(t *testing.T) {
	hits := fetchSzotarWords(context.Background(), "", "gagya", http.DefaultClient)
	if hits == nil || len(hits) != 0 {
		t.Fatalf("want empty non-nil, got %#v", hits)
	}
}

func TestFetchSzotarWordsUpstreamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer srv.Close()
	hits := fetchSzotarWords(context.Background(), srv.URL, "gagya", srv.Client())
	if hits == nil || len(hits) != 0 {
		t.Fatalf("want empty on error, got %#v", hits)
	}
}

func TestFetchSzotarWordsSkipsBadRows(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"words": []map[string]any{
				{"id": 0, "headword": "bad"},
				{"id": 2, "headword": "  "},
				{"id": 3, "headword": "jó", "definition_hu": "ok", "speech_types": []string{"ige"}},
			},
		})
	}))
	defer srv.Close()
	hits := fetchSzotarWords(context.Background(), srv.URL, "jó", srv.Client())
	if len(hits) != 1 || hits[0].ID != 3 || hits[0].Headword != "jó" {
		t.Fatalf("got %#v", hits)
	}
}

func TestFetchSzotarWordsTimeoutContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(map[string]any{"words": []any{}})
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	hits := fetchSzotarWords(ctx, srv.URL, "x", srv.Client())
	if hits == nil || len(hits) != 0 {
		t.Fatalf("want empty on cancel, got %#v", hits)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/attila/projects/lamsza/backend && go test ./internal/search/ -run FetchSzotar -count=1`

Expected: FAIL (undefined `fetchSzotarWords` / `WordSearchHit`)

- [ ] **Step 3: Add config field**

In `backend/internal/config/config.go`, add to `Config`:

```go
SzotarOrigin string
```

In `Load()`, after other string env reads:

```go
AppConfig.SzotarOrigin = strings.TrimRight(getEnv("SZOTAR_ORIGIN", ""), "/")
```

- [ ] **Step 4: Implement the fetch helper**

Create `backend/internal/search/szotar.go`:

```go
package search

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const szotarWordCap = 15

type WordSearchHit struct {
	ID           int      `json:"id"`
	Headword     string   `json:"headword"`
	DefinitionHu string   `json:"definition_hu"`
	SpeechTypes  []string `json:"speech_types"`
}

func fetchSzotarWords(ctx context.Context, origin, q string, client *http.Client) []WordSearchHit {
	out := []WordSearchHit{}
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	q = strings.TrimSpace(q)
	if origin == "" || q == "" {
		return out
	}
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Second}
	}
	u, err := url.Parse(origin + "/api/words")
	if err != nil {
		log.Printf("UnifiedSearch szotar origin: %v", err)
		return out
	}
	query := u.Query()
	query.Set("q", q)
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		log.Printf("UnifiedSearch szotar request: %v", err)
		return out
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("UnifiedSearch szotar fetch: %v", err)
		return out
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		log.Printf("UnifiedSearch szotar status: %d", resp.StatusCode)
		return out
	}

	var payload struct {
		Words []struct {
			ID           int      `json:"id"`
			Headword     string   `json:"headword"`
			DefinitionHu string   `json:"definition_hu"`
			SpeechTypes  []string `json:"speech_types"`
		} `json:"words"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		log.Printf("UnifiedSearch szotar decode: %v", err)
		return out
	}
	for _, w := range payload.Words {
		head := strings.TrimSpace(w.Headword)
		if w.ID <= 0 || head == "" {
			continue
		}
		types := w.SpeechTypes
		if types == nil {
			types = []string{}
		}
		out = append(out, WordSearchHit{
			ID:           w.ID,
			Headword:     head,
			DefinitionHu: strings.TrimSpace(w.DefinitionHu),
			SpeechTypes:  types,
		})
		if len(out) >= szotarWordCap {
			break
		}
	}
	return out
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /home/attila/projects/lamsza/backend && go test ./internal/search/ -run FetchSzotar -count=1`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/config/config.go backend/internal/search/szotar.go backend/internal/search/szotar_test.go
git commit -m "$(cat <<'EOF'
Add Szótár word fetch helper for unified search.

EOF
)"
```

---

### Task 2: Wire words into unified search

**Files:**
- Modify: `backend/internal/search/unified.go`

**Interfaces:**
- Consumes: `fetchSzotarWords`, `config.AppConfig.SzotarOrigin`
- Produces: `UnifiedSearchResult.Words []WordSearchHit` JSON key `words` (always present as array)

- [ ] **Step 1: Extend the result struct and empty response**

In `unified.go`, add to `UnifiedSearchResult`:

```go
Words []WordSearchHit `json:"words"`
```

In the empty-`q` early return and the final encode path, include `Words: []WordSearchHit{}` (or the filled slice). After `wg.Wait()`, if `wordHits == nil` set to empty slice like the other groups.

- [ ] **Step 2: Add a parallel goroutine**

Before `wg.Wait()`, with the other `wg.Add` calls:

```go
var wordHits []WordSearchHit
wg.Add(1)
go func() {
	defer wg.Done()
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	wordHits = fetchSzotarWords(ctx, config.AppConfig.SzotarOrigin, q, nil)
}()
```

Add imports: `"context"` and `"time"` if not already present.

- [ ] **Step 3: Encode Words in the response**

```go
Words: wordHits,
```

in the final `json.NewEncoder(w).Encode(...)`.

- [ ] **Step 4: Smoke-check empty origin still returns 200**

Run: `cd /home/attila/projects/lamsza/backend && go test ./internal/search/ -count=1`

Expected: PASS (existing + Task 1 tests)

Optional manual check if the API is up: `curl -s 'http://127.0.0.1:3000/api/search?q=teszt' | jq '.words'` → `[]` when `SZOTAR_ORIGIN` unset.

- [ ] **Step 5: Document the env var**

In `docs/PRODUCTION_SERVER_SETUP.md`, under the Lámsza env table (section for lamsza.com), add:

| `SZOTAR_ORIGIN` | Optional | Base URL for Szótár API, no trailing slash. Local `https://szotar.lamsza.test`, prod `https://szotar.lamsza.com`. Unset skips dictionary hits in `/api/search`. |

- [ ] **Step 6: Commit**

```bash
git add backend/internal/search/unified.go docs/PRODUCTION_SERVER_SETUP.md
git commit -m "$(cat <<'EOF'
Include Szótár words in unified search responses.

EOF
)"
```

---

### Task 3: Word result card model

**Files:**
- Modify: `src/lib/searchResultCard.js`
- Modify: `tests/searchResultCard.test.js`

**Interfaces:**
- Consumes: `szotarUrl` from `src/lib/networkOrigins.js`
- Produces: `searchResultCardModel("word", row)` → card or `null`
  - title: `headword`
  - meta: `speech_types` joined with `, ` (empty string if none)
  - description: `definition_hu`
  - href: `szotarUrl('/szo/' + id)`
  - accent: `"none"`, external: `true`, claimed: `null`
  - Skip when `id` is missing/≤0 or `headword` empty after trim

- [ ] **Step 1: Write the failing tests**

Append to `tests/searchResultCard.test.js`:

```javascript
import { szotarUrl } from "../src/lib/networkOrigins.js";

test("word: headword, speech types, definition, szotar link, new tab", () => {
    const card = searchResultCardModel("word", {
        id: 42,
        headword: "  gagya  ",
        definition_hu: "ügyetlen",
        speech_types: ["melléknév", "főnév"],
    });
    assert.deepEqual(card, {
        href: szotarUrl("/szo/42"),
        title: "gagya",
        meta: "melléknév, főnév",
        description: "ügyetlen",
        accent: "none",
        external: true,
        claimed: null,
    });
});

test("word: no speech types omits meta", () => {
    const card = searchResultCardModel("word", {
        id: 7,
        headword: "csángó",
        definition_hu: "",
        speech_types: [],
    });
    assert.equal(card.meta, "");
    assert.equal(card.description, "");
    assert.equal(card.href, szotarUrl("/szo/7"));
});

test("word: missing id or empty headword is skipped", () => {
    assert.equal(searchResultCardModel("word", { id: 0, headword: "x" }), null);
    assert.equal(searchResultCardModel("word", { id: 1, headword: "  " }), null);
    assert.equal(searchResultCardModel("word", { headword: "x" }), null);
});
```

Also extend the existing `"missing link or empty title is skipped"` test if useful, or keep the dedicated word skips above.

Update the JSDoc `kind` union in `searchResultCard.js` to include `"word"`.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/attila/projects/lamsza && node --test tests/searchResultCard.test.js`

Expected: FAIL on the new word tests (unknown kind → null)

- [ ] **Step 3: Implement the word branch**

In `src/lib/searchResultCard.js`:

```javascript
import { szotarUrl } from "./networkOrigins.js";
```

Before the final `return null;`:

```javascript
if (kind === "word") {
    const id = Number(item.id);
    if (!Number.isFinite(id) || id <= 0) return null;
    const types = Array.isArray(item.speech_types)
        ? item.speech_types.map(text).filter(Boolean)
        : [];
    return card({
        href: szotarUrl(`/szo/${id}`),
        title: item.headword,
        meta: types.join(", "),
        description: item.definition_hu,
        accent: "none",
        external: true,
        claimed: null,
    });
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/attila/projects/lamsza && node --test tests/searchResultCard.test.js`

Expected: PASS (all tests)

- [ ] **Step 5: Commit**

```bash
git add src/lib/searchResultCard.js tests/searchResultCard.test.js
git commit -m "$(cat <<'EOF'
Map Szótár words into search result cards.

EOF
)"
```

---

### Task 4: SearchEngine filter, section, and counts

**Files:**
- Modify: `src/lib/components/SearchEngine.svelte`

**Interfaces:**
- Consumes: `searchResults.words`, `searchResultCardModel("word", …)`, `AppIcon` name `"szotar"`
- Produces: `resultFilter` includes `"szotar"`; pill **Székely Szótár** after Weboldalak; section after Hírek; Index includes words; services/websites hide words; szotar shows only words

- [ ] **Step 1: Extend filter state and derived visibility**

Change the type comment and defaults:

```javascript
/** @type {"index" | "services" | "websites" | "szotar"} */
let resultFilter = "index";
```

Add:

```javascript
$: shownWords =
    resultFilter === "services" || resultFilter === "websites"
        ? []
        : (searchResults?.words || []);
```

Update existing derived values so the **szotar** filter hides other groups:

```javascript
$: shownWebsites =
    resultFilter === "services" || resultFilter === "szotar"
        ? []
        : (searchResults?.websites || []);
$: indexEntries =
    resultFilter === "websites" || resultFilter === "szotar"
        ? []
        : orderedEntries;
$: showBrowseSections = resultFilter === "index";
```

Include words in `hasResults` and `totalCount`:

```javascript
$: hasResults = searchResults && (
    (showBrowseSections && filteredLocations.length > 0) ||
    indexEntries.length > 0 ||
    (showBrowseSections && filteredEvents.length > 0) ||
    (showBrowseSections && (searchResults.news?.length || 0) > 0) ||
    (showBrowseSections && filteredAttractions.length > 0) ||
    (showBrowseSections && filteredVenues.length > 0) ||
    (showBrowseSections && (searchResults.historical_seats?.length || 0) > 0) ||
    shownWebsites.length > 0 ||
    shownWords.length > 0
);
$: totalCount = searchResults
    ? (showBrowseSections ? filteredLocations.length : 0) +
      indexEntries.length +
      (showBrowseSections ? filteredEvents.length : 0) +
      (showBrowseSections ? (searchResults.news?.length || 0) : 0) +
      (showBrowseSections ? filteredAttractions.length : 0) +
      (showBrowseSections ? filteredVenues.length : 0) +
      (showBrowseSections ? (searchResults.historical_seats?.length || 0) : 0) +
      shownWebsites.length +
      shownWords.length
    : 0;
```

Update `searchScope`:

```javascript
$: searchScope = resultFilter === "services"
    ? "szolgáltatásokban:"
    : resultFilter === "websites"
      ? "weboldalakban:"
      : resultFilter === "szotar"
        ? "szótárban:"
        : selectedLocation?.name
          ? `${selectedLocation.name} és környéke:`
          : "mindenhol:";
```

- [ ] **Step 2: Render the Székely Szótár section after Hírek**

After the news `{#if}` block and before the closing of the results column, add:

```svelte
{#if shownWords.length > 0}
    <div class="discover-section">
        <h4 class="discover-section-title">
            <AppIcon name="szotar" size={18} />
            Székely Szótár
        </h4>
        <div class="discover-result-list">
            {#each shownWords as word (word.id)}
                {@const card = searchResultCardModel("word", word)}
                {#if card}
                    <SearchResultCard {...card} />
                {/if}
            {/each}
        </div>
    </div>
{/if}
```

Ensure the comment near `searchResults` mentions `words`.

- [ ] **Step 3: Add the filter pill after Weboldalak**

Inside `.result-filters`, after the Weboldalak button:

```svelte
<button
    type="button"
    class="btn btn-md nav-btn"
    class:active={resultFilter === "szotar"}
    aria-pressed={resultFilter === "szotar"}
    on:click={() => selectResultFilter("szotar")}
>
    Székely Szótár
</button>
```

- [ ] **Step 4: Validate the Svelte component**

Run via Svelte MCP / autofixer on `src/lib/components/SearchEngine.svelte` and fix any reported issues until clean.

Also run: `cd /home/attila/projects/lamsza && node --test tests/searchResultCard.test.js`

Expected: PASS

- [ ] **Step 5: Manual check (when apps are running)**

1. Set `SZOTAR_ORIGIN=https://szotar.lamsza.test` (or the local Szótár API origin) in Lámsza backend env and restart.
2. Search a known headword on Lámsza.
3. Confirm a **Székely Szótár** section appears on Index when there are hits.
4. Click **Székely Szótár** pill: only dictionary rows remain.
5. Click a row: opens `…/szo/{id}` on Szótár in a new tab.
6. Search nonsense: Index still shows no-results when nothing else matches; pill still present when the footer shows.

- [ ] **Step 6: Commit**

```bash
git add src/lib/components/SearchEngine.svelte
git commit -m "$(cat <<'EOF'
Show Szótár hits in search and add the filter pill.

EOF
)"
```

---

## Self-review

1. **Spec coverage:** API `words` + cap + failure → Tasks 1–2. Config `SZOTAR_ORIGIN` → Tasks 1–2. Filter pill + scope + Index/services/websites/szotar visibility → Task 4. Card mapping + external link → Task 3. Section after Hírek + AppIcon → Task 4. Counts → Task 4. Docs env → Task 2. Tests → Tasks 1 and 3.
2. **Placeholders:** None; code and commands are concrete.
3. **Type consistency:** `WordSearchHit` / `words` / `resultFilter: "szotar"` / `searchResultCardModel("word", …)` align across tasks.

## Execution

Plan complete and saved to `docs/superpowers/plans/2026-10-05-szotar-search-integration.md`. Two execution options:

1. **Subagent-Driven (recommended)** - fresh subagent per task, review between tasks
2. **Inline Execution** - execute tasks in this session with checkpoints

Which approach?
