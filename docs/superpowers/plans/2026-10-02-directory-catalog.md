# Directory catalog implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the flat directory catalog with the agreed two-level tree and three types, wipe directory rows so seed ids start at 1, and file websites in that same tree.

**Architecture:** `entry_categories` becomes a two-level tree (`parent_id` null on a main category). An entry and a website both point at a subcategory. A listing with no settlement is a website. Types stay a closed list of three. A one-shot migrate, guarded by `site_settings.directory_catalog_v2`, deletes directory rows and inserts the seed. It does not run again, and it does not touch towns, users, events, news, or website submission rows.

**Tech Stack:** PostgreSQL 16, Go net/http, SvelteKit, existing admin handlers in `backend/internal/handlers/admin_entries.go`.

## Global Constraints

- Hungarian UI copy. No em dash. Use ` - ` when a dash is required.
- Directory wipe only: entries, entry tags, entry reviews, entry suggestions, entry members, categories, types, and the `tags` rows. Restart those sequences. Leave settlements, users, events, news, and `websites` rows. `websites.entry_id` becomes null because of `ON DELETE SET NULL`.
- An entry has one subcategory. A main category is not a valid `entries.category_id`.
- A website uses the same subcategory. It is not a fourth type. A webshop listing is type Vállalkozás and has `location_id` null.
- Tags never repeat the category or parent name. Tags are not seeded.
- Public panels hide a subcategory that has no published listing and no approved unlinked website. Admin sees the whole tree, including empty shelves.
- Delete never removes entries or websites. A row with children, entries, or websites cannot be deleted until those are moved. An empty row can be deleted.
- Types cannot be created or deleted. Names stay Személy, Vállalkozás, Intézmény.
- Do not reintroduce the old flat names (Egészségügy, Vendéglő, Bolt, Egyéb, Szolgáltatás, Cég) as the catalog.

### Seed types

| id | name |
|---|---|
| 1 | Személy |
| 2 | Vállalkozás |
| 3 | Intézmény |

### Seed categories

Parents are ids 1-10. Children follow, in this order. `sort_order` is the position inside the parent, starting at 1. Parent `sort_order` is the menu order 1-10.

| id | parent | name |
|---|---|---|
| 1 | | Étkezés |
| 2 | | Szállás |
| 3 | | Egészség és szépség |
| 4 | | Vásárlás |
| 5 | | Autó |
| 6 | | Mesteremberek |
| 7 | | Oktatás |
| 8 | | Hivatalok |
| 9 | | Sport és szabadidő |
| 10 | | Pénzügy |
| 11-15 | 1 | Étterem, Kávézó, Cukrászda, Pékség, Söröző |
| 16-20 | 2 | Szálloda, Motel, Panzió, Apartman, Kemping |
| 21-35 | 3 | Orvos, Fogászat, Bőrgyógyászat, Optika, Csontkovács, Lábgyógyászat, Gyógytorna, Masszázs, Gyógyszertár, Kórház, Állatorvos, Fodrász, Borbély, Körömszalon, Spa |
| 36-40 | 4 | Élelmiszer, Ruházat, Műszaki bolt, Bútor, Piac |
| 41-53 | 5 | Autószerviz, Karosszéria, Olajcsere, Gumiszerviz, Turbószerviz, Autómentés, Autómosó, Autókozmetika, Parkoló, Autókereskedés, Autóbontó, Autóalkatrész, Benzinkút |
| 54-58 | 6 | Villanyszerelő, Vízvezeték-szerelő, Asztalos, Takarítás, Építkezés |
| 59-61 | 7 | Óvoda, Iskola, Egyetem |
| 62-64 | 8 | Polgármesteri hivatal, Megyei intézmény, Posta |
| 65-66 | 9 | Sportegyesület, Sportpálya |
| 67-68 | 10 | Bank, Biztosító |

---

## File structure

- Create `backend/internal/utils/directory_catalog.go`: the seed tree and slug helper calls.
- Create `backend/internal/handlers/migrate_directory_catalog.go`: one-shot wipe, schema, seed.
- Create `backend/directory_catalog_test.go`: wipe flag, leaf rule, delete/move, website category, null town.
- Modify `backend/main.go` and `backend/handlers_test.go` `init`: call `handlers.MigrateDirectoryCatalog()` after `account.MigrateWebsites()`.
- Modify `backend/internal/handlers/migrate_entry_categories.go` and `migrate_entry_types.go`: return immediately. They must not insert the old eight names or prune categories that are not in that list.
- Modify `backend/internal/utils/entry_category.go` and `entry_type.go`: constants match the new names. `DefaultEntryCategory` is no longer Egyéb.
- Modify `backend/internal/models/models.go`: `EntryCategory` gains `ParentID *int`, `Slug string`, `SortOrder int`.
- Modify `backend/internal/handlers/admin_entries.go`: category parent, move-then-delete, closed types.
- Modify `backend/internal/account/websites.go` and `listings.go`: subcategory on submit, town optional.
- Modify `src/lib/entryType.js`, `src/lib/entryCategory.js`, `AddWebsiteForm.svelte`, `ListingFormDialog.svelte`, `IndexTagAside.svelte`, `src/routes/(public)/index/+page.svelte`, `src/routes/admin/+page.svelte`.

---

### Task 1: Seed data and one-shot wipe

**Files:**
- Create: `backend/internal/utils/directory_catalog.go`
- Create: `backend/internal/handlers/migrate_directory_catalog.go`
- Modify: `backend/main.go`
- Modify: `backend/handlers_test.go` (init only)
- Test: `backend/directory_catalog_test.go`

**Interfaces:**
- Consumes: `db.DB`, `utils.Slugify`, `site_settings`.
- Produces: `utils.DirectoryParents() []DirectoryNode`, `utils.DirectoryChildren() []DirectoryNode`, `handlers.MigrateDirectoryCatalog()`. After it runs, `site_settings.key = directory_catalog_v2` and `value = 1`. Category ids 1-68 and type ids 1-3 exist. `entries` is empty.

`DirectoryNode` fields: `ID int`, `ParentID *int`, `Name string`, `SortOrder int`.

- [ ] **Step 1: Write the failing test**

```go
func TestDirectoryCatalogSeedIds(t *testing.T) {
	handlers.MigrateDirectoryCatalog()
	var flag string
	if err := db.DB.QueryRow(`SELECT value FROM site_settings WHERE key = 'directory_catalog_v2'`).Scan(&flag); err != nil {
		t.Fatal(err)
	}
	if flag != "1" {
		t.Fatalf("flag = %q", flag)
	}
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entries`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("entries = %d, want 0", n)
	}
	var name string
	if err := db.DB.QueryRow(`SELECT name FROM entry_types WHERE id = 2`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Vállalkozás" {
		t.Fatalf("type 2 = %q", name)
	}
	if err := db.DB.QueryRow(`SELECT name FROM entry_categories WHERE id = 39`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Bútor" {
		t.Fatalf("category 39 = %q", name)
	}
	var parent int
	if err := db.DB.QueryRow(`SELECT parent_id FROM entry_categories WHERE id = 45`).Scan(&parent); err != nil {
		t.Fatal(err)
	}
	if parent != 5 {
		t.Fatalf("Turbószerviz parent = %d, want 5 Autó", parent)
	}
	var towns int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM settlements`).Scan(&towns); err != nil {
		t.Fatal(err)
	}
	if towns == 0 {
		t.Fatal("settlements were wiped")
	}
}
```

- [ ] **Step 2: Run the test and confirm it fails**

Run: `cd backend && go test -count=1 -run TestDirectoryCatalogSeedIds .`

Expected: FAIL because `MigrateDirectoryCatalog` is undefined.

- [ ] **Step 3: Implement the catalog and the one-shot migrate**

`directory_catalog.go` lists every row in the tables above. `MigrateDirectoryCatalog`:

1. `CREATE TABLE` is already present. Add columns if missing:

```sql
ALTER TABLE entry_categories ADD COLUMN IF NOT EXISTS parent_id INTEGER REFERENCES entry_categories(id);
ALTER TABLE entry_categories ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0;
ALTER TABLE websites ADD COLUMN IF NOT EXISTS category_id INTEGER REFERENCES entry_categories(id);
```

2. If `site_settings.directory_catalog_v2` is already `1`, return.
3. Otherwise, in one transaction:

```sql
DELETE FROM entries;
DELETE FROM tags;
DELETE FROM entry_categories;
DELETE FROM entry_types;
```

`DELETE FROM entries` cascades `entry_tags`, `entry_members`, `entry_reviews`, and `entry_suggestions`. `websites.entry_id` is set null. Do not `DELETE FROM websites`.

Restart sequences after the inserts, not before, with `setval(..., max(id), true)`.

Insert types with explicit ids 1, 2, 3. Insert categories with explicit ids 1-68, `parent_id`, `sort_order`, and `slug = utils.Slugify(name)`. Slug examples: `egeszseg-es-szepseg`, `turbo-szerviz`, `vizvezetek-szerelo`, `polgarmesteri-hivatal`.

```sql
INSERT INTO site_settings (key, value) VALUES ('directory_catalog_v2', '1')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;
```

- [ ] **Step 4: Call it on startup**

In `backend/main.go`, after `account.MigrateWebsites()`:

```go
handlers.MigrateDirectoryCatalog()
```

Add the same call in `backend/handlers_test.go` `init`, after `account.MigrateWebsites()`.

- [ ] **Step 5: Run the test**

Run: `cd backend && go test -count=1 -run TestDirectoryCatalogSeedIds .`

Expected: PASS. Second call in the same process does not wipe settlements and does not change ids.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/utils/directory_catalog.go backend/internal/handlers/migrate_directory_catalog.go backend/directory_catalog_test.go backend/main.go backend/handlers_test.go
git commit -m "Seed the directory tree and reset directory ids once."
```

---

### Task 2: Stop the old catalog migrators

**Files:**
- Modify: `backend/internal/handlers/migrate_entry_categories.go`
- Modify: `backend/internal/handlers/migrate_entry_types.go`
- Modify: `backend/internal/utils/entry_category.go`
- Modify: `backend/internal/utils/entry_type.go`
- Modify: `src/lib/entryType.js`
- Modify: `src/lib/entryCategory.js`

**Interfaces:**
- Consumes: Task 1 seed.
- Produces: `utils.EntryTypeSzemely`, `utils.EntryTypeVallalkozas`, `utils.EntryTypeIntezmeny`. Frontend `ENTRY_TYPE_SZEMELY`, `ENTRY_TYPE_VALLALKOZAS`, `ENTRY_TYPE_INTEZMENY`. `isServiceEntry` is removed. Callers that filtered the index to Szolgáltatás show every published entry instead.

- [ ] **Step 1: Write the failing test**

```go
func TestLegacyCategoryMigrateDoesNotPruneTree(t *testing.T) {
	handlers.MigrateDirectoryCatalog()
	handlers.MigrateEntryCategories()
	handlers.MigrateEntryTypes()
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entry_categories`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 68 {
		t.Fatalf("categories = %d, want 68", n)
	}
	var types int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entry_types`).Scan(&types); err != nil {
		t.Fatal(err)
	}
	if types != 3 {
		t.Fatalf("types = %d, want 3", types)
	}
}
```

- [ ] **Step 2: Run it**

Run: `cd backend && go test -count=1 -run TestLegacyCategoryMigrateDoesNotPruneTree .`

Expected: FAIL if the old prune still deletes empty seed rows. If it already passes because those functions are not destructive until invoked, still replace their bodies so a later call cannot restore Egyéb or Szolgáltatás.

- [ ] **Step 3: Make the old migrators no-ops**

`MigrateEntryCategories` and `MigrateEntryTypes` log `skipped: directory_catalog_v2` and return. Delete the seed loops and the prune `DELETE`.

Replace `entry_type.go` constants:

```go
const (
	EntryTypeSzemely      = "Személy"
	EntryTypeVallalkozas  = "Vállalkozás"
	EntryTypeIntezmeny    = "Intézmény"
)
```

`CanonicalEntryType` maps only those three names. Unknown input returns an empty string, not Szolgáltatás.

`SeedEntryCategories` returns the 68 names from `DirectoryParents` and `DirectoryChildren`. `DefaultEntryCategory` returns `""`. `resolveEntryCategoryID` returns an error when it cannot find the requested child. It must not insert Egyéb.

`src/lib/entryType.js` exports the three new constants. Remove `filterServiceEntries` and `isServiceEntry`.

`src/lib/entryCategory.js`: `canonicalEntryCategory` returns the stored name. It does not map Vendéglő, Bolt, or Egyéb onto a new shelf. `entryMatchesCategory` matches the entry category slug or its parent slug.

- [ ] **Step 4: Run the test**

Run: `cd backend && go test -count=1 -run 'TestLegacyCategoryMigrateDoesNotPruneTree|TestDirectoryCatalogSeedIds' .`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handlers/migrate_entry_categories.go backend/internal/handlers/migrate_entry_types.go backend/internal/utils/entry_category.go backend/internal/utils/entry_type.go src/lib/entryType.js src/lib/entryCategory.js
git commit -m "Stop the old directory seed from replacing the new tree."
```

---

### Task 3: Category create, rename, and move-then-delete

**Files:**
- Modify: `backend/internal/models/models.go`
- Modify: `backend/internal/handlers/admin_entries.go`
- Test: `backend/directory_catalog_test.go`

**Interfaces:**
- Consumes: `entry_categories.parent_id`, `sort_order`.
- Produces: GET `/api/admin/entry_categories` returns `{id, name, slug, parent_id, sort_order}`. POST body `{name, parent_id}` creates a parent when `parent_id` is null or omitted, or a child when `parent_id` is a parent id. PUT body `{id, name, parent_id}` renames and can move a child under another parent. DELETE `?id=` deletes an empty row. DELETE `?id=&move_to=` moves entries and websites, or moves child categories when the deleted row is a parent, then deletes.

- [ ] **Step 1: Write the failing test**

```go
func TestDeleteCategoryRequiresMove(t *testing.T) {
	handlers.MigrateDirectoryCatalog()
	// Étterem is 11. Attach nothing. Delete succeeds.
	rr := doRequest(t, "DELETE", "/api/admin/entry_categories?id=11", nil)
	if rr.Code != 200 {
		t.Fatalf("empty delete: %d %s", rr.Code, rr.Body.String())
	}
	// Recreate Étterem under Étkezés (1) so later tests still have a leaf.
	rr = doRequest(t, "POST", "/api/admin/entry_categories", strings.NewReader(`{"name":"Étterem","parent_id":1}`))
	if rr.Code != 200 {
		t.Fatalf("recreate: %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`
		INSERT INTO entries (name, category_id, type_id, languages)
		VALUES ('Próba étterem', $1, 2, '{HU}')`, created.ID); err != nil {
		t.Fatal(err)
	}
	rr = doRequest(t, "DELETE", "/api/admin/entry_categories?id="+strconv.Itoa(created.ID), nil)
	if rr.Code != 409 {
		t.Fatalf("blocked delete: %d", rr.Code)
	}
	rr = doRequest(t, "DELETE", "/api/admin/entry_categories?id="+strconv.Itoa(created.ID)+"&move_to=12", nil)
	if rr.Code != 200 {
		t.Fatalf("move delete: %d %s", rr.Code, rr.Body.String())
	}
	var cat int
	if err := db.DB.QueryRow(`SELECT category_id FROM entries WHERE name = 'Próba étterem'`).Scan(&cat); err != nil {
		t.Fatal(err)
	}
	if cat != 12 {
		t.Fatalf("moved category = %d, want 12 Kávézó", cat)
	}
	_, _ = db.DB.Exec(`DELETE FROM entries WHERE name = 'Próba étterem'`)
}
```

`entries.location_id` is nullable. The insert omits it. If a NOT NULL constraint exists, this test fails until Task 5 drops it. Run Task 5's column change in this task if the insert fails for that reason: `ALTER TABLE entries ALTER COLUMN location_id DROP NOT NULL`.

- [ ] **Step 2: Run it**

Run: `cd backend && go test -count=1 -run TestDeleteCategoryRequiresMove .`

Expected: FAIL. Current DELETE has no `move_to`, and POST ignores `parent_id`.

- [ ] **Step 3: Implement the rules in `HandleAdminEntryCategories`**

- GET selects `id, name, slug, parent_id, sort_order` and orders by `COALESCE(parent_id, id), sort_order, id`.
- POST: trim name. Reject empty. If `parent_id` is set, the parent row must exist and itself have `parent_id` null. Reject a third level. Slug from `utils.Slugify`. `sort_order` is `MAX(sort_order)+1` among siblings. Return the row.
- PUT: same parent check. A parent that has children cannot gain a `parent_id`. Renaming updates `entries.cat_name` for rows pointing at this id.
- DELETE without `move_to`:
  - Parent with children: 409 `Előbb helyezd át vagy töröld az alkategóriákat.`
  - Category with entries or websites: 409 `Előbb helyezd át a bejegyzéseket.`
  - Empty: delete.
- DELETE with `move_to`:
  - Deleting a parent: `move_to` must be another parent. `UPDATE entry_categories SET parent_id = move_to WHERE parent_id = id`, then delete the empty parent.
  - Deleting a child: `move_to` must be a child. `UPDATE entries SET category_id = move_to, cat_name = (SELECT name FROM entry_categories WHERE id = move_to) WHERE category_id = id`. Same update on `websites.category_id`. Then delete.
- Reject `move_to` equal to the deleted id.

- [ ] **Step 4: Run the test**

Run: `cd backend && go test -count=1 -run TestDeleteCategoryRequiresMove .`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/models/models.go backend/internal/handlers/admin_entries.go backend/directory_catalog_test.go
git commit -m "Let admins add categories and move entries before a delete."
```

---

### Task 4: Close the type list

**Files:**
- Modify: `backend/internal/handlers/admin_entries.go` (`HandleAdminEntryTypes`)
- Test: `backend/directory_catalog_test.go`

**Interfaces:**
- Consumes: type ids 1-3.
- Produces: POST and DELETE `/api/admin/entry_types` return 403 with `A típuslista zárt.` PUT that changes a name returns 403. GET still returns the three rows.

- [ ] **Step 1: Write the failing test**

```go
func TestEntryTypesStayClosed(t *testing.T) {
	handlers.MigrateDirectoryCatalog()
	rr := doRequest(t, "POST", "/api/admin/entry_types", strings.NewReader(`{"name":"Weboldal"}`))
	if rr.Code != 403 {
		t.Fatalf("create type: %d", rr.Code)
	}
	rr = doRequest(t, "DELETE", "/api/admin/entry_types?id=1", nil)
	if rr.Code != 403 {
		t.Fatalf("delete type: %d", rr.Code)
	}
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entry_types`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("types = %d", n)
	}
}
```

- [ ] **Step 2: Run it**

Run: `cd backend && go test -count=1 -run TestEntryTypesStayClosed .`

Expected: FAIL with 200 on POST.

- [ ] **Step 3: Reject create, rename, and delete**

In `HandleAdminEntryTypes`, POST, PUT, and DELETE write 403 and the sentence `A típuslista zárt.` GET stays.

- [ ] **Step 4: Run it**

Run: `cd backend && go test -count=1 -run TestEntryTypesStayClosed .`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handlers/admin_entries.go backend/directory_catalog_test.go
git commit -m "Keep the directory type list closed."
```

---

### Task 5: Websites use a subcategory, and a listing may have no town

**Files:**
- Modify: `backend/internal/account/websites.go`
- Modify: `backend/internal/account/listings.go`
- Modify: `src/lib/components/AddWebsiteForm.svelte`
- Modify: `src/lib/components/ListingFormDialog.svelte`
- Test: `backend/directory_catalog_test.go`

**Interfaces:**
- Consumes: leaf `entry_categories`, type id 2.
- Produces: POST `/api/websites` body `{domain, title, description, category_id}`. `category_id` must be a child. Approved website list items include `category_id` and `category`. `handleCreateListing` accepts `location_id` 0 and stores NULL. It rejects a parent `category_id`. It rejects a tag whose slug equals the category slug or the parent slug.

- [ ] **Step 1: Write the failing test**

Use the existing website test helpers in `backend/website_test.go` for auth. Add:

```go
func TestWebsiteRequiresLeafCategory(t *testing.T) {
	handlers.MigrateDirectoryCatalog()
	// category 4 is Vásárlás, a parent. category 39 is Bútor.
	rr := postWebsite(t, `{"domain":"mobonline.ro","title":"Mobonline","description":"Bútor webshop","category_id":4}`)
	if rr.Code != 400 {
		t.Fatalf("parent category: %d %s", rr.Code, rr.Body.String())
	}
	rr = postWebsite(t, `{"domain":"mobonline.ro","title":"Mobonline","description":"Bútor webshop","category_id":39}`)
	if rr.Code != 201 {
		t.Fatalf("leaf category: %d %s", rr.Code, rr.Body.String())
	}
}
```

`postWebsite` is a small helper in the test file that logs in a non-banned user the same way `backend/website_test.go` already does, then POSTs `/api/websites`. Read that file and copy its user setup. Do not invent a second auth path.

Second test, through `handleCreateListing` or POST `/api/account/listings`: a body with `category_id: 39`, `type_id: 2`, `location_id: 0`, `name: "Mobonline"`, `url: "https://mobonline.ro"`, `website_id` of the approved row. Expect 201 and `entries.location_id` null.

If approval is required before link, call the admin approve action first, as `website_test.go` does.

- [ ] **Step 2: Run it**

Run: `cd backend && go test -count=1 -run 'TestWebsiteRequiresLeafCategory|TestWebshopListingWithoutTown' .`

Expected: FAIL. The website handler ignores `category_id`, and listing create rejects `location_id <= 0`.

- [ ] **Step 3: Implement**

`submitWebsiteBody` gains `CategoryID int`. `handleWebsiteSubmit` checks:

```sql
SELECT parent_id FROM entry_categories WHERE id = $1
```

Missing row or null `parent_id`: 400 `{"error":"invalid","field":"category_id"}`. Insert `category_id` into `websites`.

`handleAdminWebsiteAction` approve keeps the stored category. Add optional `category_id` on the admin body so the admin can correct the shelf before approve. The corrected id must also be a child.

`handleCreateListing`: change the required check to `name`, `category_id`, and `type_id`. `location_id` may be 0. Pass NULL to the INSERT when it is 0. Reject a parent category and an unknown type id. When `website_id` is set, default `category_id` from `websites.category_id` if the body omits it.

Tag check on create and update: for each tag, `utils.Slugify(tag)` must differ from the category slug and the parent slug. On conflict, 400 `{"error":"invalid","field":"tags"}`.

`AddWebsiteForm.svelte` loads `/api/account/listings/catalog` (or the public category list from Task 6) and adds a required subcategory select grouped by parent. The POST body includes `category_id`.

`ListingFormDialog.svelte` sends `location_id: 0` when the town field is empty. The category select lists children, grouped by parent, and does not offer a parent as a value. Type select options are Személy, Vállalkozás, Intézmény.

- [ ] **Step 4: Run the tests**

Run: `cd backend && go test -count=1 -run 'TestWebsiteRequiresLeafCategory|TestWebshopListingWithoutTown' .`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/account/websites.go backend/internal/account/listings.go backend/directory_catalog_test.go src/lib/components/AddWebsiteForm.svelte src/lib/components/ListingFormDialog.svelte
git commit -m "File websites in a subcategory and allow a listing without a town."
```

---

### Task 6: Public browse uses parents, children, and the website list

**Files:**
- Modify: `backend/internal/handlers/public.go`
- Modify: `backend/internal/account/websites.go` (`handleWebsitesList`)
- Modify: `src/lib/entryCategory.js`
- Modify: `src/routes/(public)/index/+page.svelte`
- Modify: `src/routes/(public)/index/[category]/+page.svelte`
- Modify: `src/lib/components/IndexTagAside.svelte`
- Modify: `src/lib/components/SearchEngine.svelte`

**Interfaces:**
- Consumes: entries with `category_id` pointing at a child, `location_id` null or set, websites with `category_id`.
- Produces: public category payload `{id, name, slug, parent_id, sort_order}`. Index route `/index/{parentSlug}` shows every published entry in that parent's children, plus approved websites in those children that have no entry yet. `/index/{childSlug}` shows that child only. A town page includes an entry only when `location_id` is set. The index no longer drops rows that are not Szolgáltatás.

- [ ] **Step 1: Write the failing test**

Insert two published entries: one Bútor (39) with a settlement id that exists, one Bútor with null `location_id`. Insert an approved website on Bútor with `entry_id` null. GET `/api/entries?category=butor` returns both entries. GET `/api/entries?category=vasarlas` returns both. GET `/api/entries?location_id={town}` returns only the row with that town. GET `/api/websites` includes the unlinked site with `category` `Bútor`.

Match the existing public entries query parameter names in `handlers/public.go`. If the parameter is a slug segment rather than `category`, test the parameter that the handler already accepts and extend it.

- [ ] **Step 2: Run it**

Run: `cd backend && go test -count=1 -run TestCategoryBrowseIncludesWebshop .`

Expected: FAIL until the filter walks from child to parent and keeps null locations.

- [ ] **Step 3: Implement browse**

Public entry query joins `entry_categories c` and `entry_categories parent ON parent.id = c.parent_id`. A category filter matches `c.slug` or `parent.slug`. A location filter adds `e.location_id = $n`.

`handleWebsitesList` returns `category_id` and the category name. It still lists approved sites. Search hides a website that already has a published entry, which is the current `SearchWebsites` behavior. Keep that.

`entryCategory.js` `directoryCategoryTabs` builds parents from the catalog response, not from whatever category strings happen to be on loaded cards. Hide a parent when none of its children have a visible row. Hide a child panel item on the same rule.

`index/+page.svelte`: remove `filterServiceEntries`. The services view is all published entries with a town. The websites view is published entries with no town, plus approved websites that are not linked to an entry.

`IndexTagAside.svelte`: remove the Bejegyzés típus block. Keep the keyword tag cloud.

`SearchEngine.svelte`: stop filtering results through `isServiceEntry`.

- [ ] **Step 4: Run the test**

Run: `cd backend && go test -count=1 -run TestCategoryBrowseIncludesWebshop .`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handlers/public.go backend/internal/account/websites.go backend/directory_catalog_test.go src/lib/entryCategory.js src/routes/\(public\)/index/+page.svelte src/routes/\(public\)/index/\[category\]/+page.svelte src/lib/components/IndexTagAside.svelte src/lib/components/SearchEngine.svelte
git commit -m "Browse the directory by parent and subcategory."
```

---

### Task 7: Admin screens for the tree and the closed types

**Files:**
- Modify: `src/routes/admin/+page.svelte`

**Interfaces:**
- Consumes: category JSON `{id, name, slug, parent_id, sort_order}` and type 403 on create.
- Produces: Bejegyzés kategóriák can add a main category or a subcategory under a chosen parent. The table shows the parent name. Delete of a used row opens a picker of other subcategories (or other parents, when deleting a parent) and sends `move_to`. Bejegyzés típusok hides Új típus, Szerk., and Törlés.

- [ ] **Step 1: Manual check before editing**

Open `http://localhost:5173/admin`, Bejegyzés kategóriák. Confirm the screen still lists a flat name column. That is the behavior this task replaces.

- [ ] **Step 2: Update the category screen**

The new-category form has a name field and a parent select. The first option is `Főkategória` and sends no `parent_id`. Other options are the ten parents. POST `{name, parent_id}`.

The table columns are ID, Név, Főkategória, Szerk., Törlés. A parent row shows `-` in Főkategória.

Delete: if the API returns 409, show the message and a select of legal move targets, then repeat DELETE with `move_to`. Do not delete entries.

- [ ] **Step 3: Update the type screen**

Remove the Új típus form and the Szerk. and Törlés buttons. The table lists Személy, Vállalkozás, Intézmény.

The website admin row editor includes the same subcategory select and sends `category_id` with approve.

- [ ] **Step 4: Check in the browser**

Reload admin. Add a temporary subcategory under Étkezés, confirm it appears under that parent, delete it while it is empty, and confirm the type screen has no add button.

- [ ] **Step 5: Commit**

```bash
git add src/routes/admin/+page.svelte
git commit -m "Edit the directory tree in admin and lock the type list."
```

---

## Self-review

- Seed ids, wipe boundary, leaf category, closed types, website subcategory, null town, move-then-delete, hidden empty public shelves, and tags that do not copy a category name each have a task.
- Old `MigrateEntryCategories` prune is disabled before it can delete the empty seed.
- Settlements, users, events, news, and website rows are outside the wipe. The seed test checks settlements stay.
- `postWebsite` must follow `backend/website_test.go` rather than a new login invent.

## Execution note

Run the backend tests against the dev database only after Task 1, because that task deletes directory entries once. Towns and accounts stay.
