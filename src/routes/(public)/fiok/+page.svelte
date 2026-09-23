<script>
    import { onMount } from "svelte";
    import { get } from "svelte/store";
    import ListingFormDialog from "$lib/components/ListingFormDialog.svelte";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import { userAccountTabIds } from "$lib/accountPrefs.js";
    import { apiFetch } from "$lib/api.js";
    import { removeFavorite } from "$lib/favorites.js";
    import { openLogin } from "$lib/openLogin.js";
    import { auth } from "$lib/stores/auth";

    const TAB_LABELS = {
        fiok: "Fiók",
        adatlapjaim: "Adatlapjaim",
        weboldalaim: "Weboldalaim",
        kereses: "Keresési előzményeim",
        bongeszes: "Böngészési előzményeim",
        kedvencek: "Kedvenc helyek",
        linkjeim: "Linkjeim",
    };

    function hasField(value) {
        if (value == null) return false;
        if (typeof value === "string") return value.trim() !== "";
        return true;
    }

    /** @param {string | null | undefined} iso */
    function formatTimestamp(iso) {
        if (!iso) return "";
        const d = new Date(iso);
        if (Number.isNaN(d.getTime())) return String(iso);
        return d.toLocaleString("hu-HU");
    }

    let activeTab = $state(userAccountTabIds[0]);
    let displayName = $state("");
    let displayNameError = $state("");
    let displayNameSaving = $state(false);
    let linksError = $state("");
    let historyError = $state("");
    /** @type {Array<{ id: number, title: string, url: string, bg_color?: string, position?: number }>} */
    let accountLinks = $state([]);
    /** @type {Array<{ slug: string, name: string, category?: string, location?: string, photo?: string }>} */
    let accountHistory = $state([]);
    let linksLoading = $state(false);
    let historyLoading = $state(false);
    let favoritesError = $state("");
    /** @type {{ settlements: Array<{ id: number, name: string, slug?: string | null, county_slug?: string | null }>, attractions: Array<{ id: number, name: string, slug?: string | null, county_slug?: string | null }>, entries: Array<{ id: number, name: string, slug?: string | null }>, events: Array<{ id: number, name: string }> }} */
    let accountFavorites = $state({
        settlements: [],
        attractions: [],
        entries: [],
        events: [],
    });
    let favoritesLoading = $state(false);
    let linkDialogOpen = $state(false);
    let linkDialogMode = $state(/** @type {"add" | "edit"} */ ("add"));
    let linkDialogData = $state({
        id: 0,
        title: "",
        url: "",
        bg_color: "#e6f0ff",
    });

    async function loadLinks() {
        linksLoading = true;
        linksError = "";
        try {
            accountLinks = (await apiFetch("/api/account/links")) || [];
        } catch {
            linksError = "A mentés nem sikerült";
        } finally {
            linksLoading = false;
        }
    }

    async function loadHistory() {
        historyLoading = true;
        try {
            accountHistory = (await apiFetch("/api/account/history")) || [];
        } catch {
            accountHistory = [];
        } finally {
            historyLoading = false;
        }
    }

    async function loadFavorites() {
        favoritesLoading = true;
        favoritesError = "";
        try {
            const payload = (await apiFetch("/api/account/favorites")) || {};
            accountFavorites = {
                settlements: Array.isArray(payload.settlements) ? payload.settlements : [],
                attractions: Array.isArray(payload.attractions) ? payload.attractions : [],
                entries: Array.isArray(payload.entries) ? payload.entries : [],
                events: Array.isArray(payload.events) ? payload.events : [],
            };
        } catch {
            accountFavorites = {
                settlements: [],
                attractions: [],
                entries: [],
                events: [],
            };
        } finally {
            favoritesLoading = false;
        }
    }

    /** @param {Record<string, unknown>} item */
    function countyPlaceUrl(item) {
        const countySlug = String(item.county_slug ?? "").trim();
        const slug = String(item.slug ?? "").trim();
        if (!countySlug || !slug) return "";
        return `/${countySlug}-megye/${slug}`;
    }

    /** @param {{ slug?: string | null }} item */
    function entryFavoriteUrl(item) {
        const slug = String(item.slug ?? "").trim();
        if (!slug) return "";
        return `/bejegyzes/${slug}`;
    }

    /** @param {{ id: number }} item */
    function eventFavoriteUrl(item) {
        return `/esemenyek/${item.id}`;
    }

    /** @param {"settlement" | "attraction" | "entry" | "event"} type @param {number} id */
    async function removeFavoriteRow(type, id) {
        favoritesError = "";
        const prev = accountFavorites;
        const key =
            type === "settlement"
                ? "settlements"
                : type === "attraction"
                  ? "attractions"
                  : type === "entry"
                    ? "entries"
                    : "events";
        try {
            await removeFavorite(type, id);
            accountFavorites = {
                ...prev,
                [key]: prev[key].filter((row) => row.id !== id),
            };
        } catch {
            accountFavorites = prev;
            favoritesError = "A mentés nem sikerült";
        }
    }

    function openAddLink() {
        linkDialogMode = "add";
        linkDialogData = { id: 0, title: "", url: "", bg_color: "#e6f0ff" };
        linkDialogOpen = true;
    }

    /** @param {{ id: number, title: string, url: string, bg_color?: string }} link */
    function openEditLink(link) {
        linkDialogMode = "edit";
        linkDialogData = {
            id: link.id,
            title: link.title,
            url: link.url,
            bg_color: link.bg_color || "#e6f0ff",
        };
        linkDialogOpen = true;
    }

    function closeLinkDialog() {
        linkDialogOpen = false;
    }

    async function saveLinkDialog(e) {
        e.preventDefault();
        linksError = "";
        const prevLinks = accountLinks;
        try {
            if (linkDialogMode === "add") {
                await apiFetch("/api/account/links", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        title: linkDialogData.title,
                        url: linkDialogData.url,
                        bg_color: linkDialogData.bg_color,
                    }),
                });
            } else {
                const next = accountLinks.map((l) =>
                    l.id === linkDialogData.id
                        ? {
                              ...l,
                              title: linkDialogData.title,
                              url: linkDialogData.url,
                              bg_color: linkDialogData.bg_color,
                          }
                        : l,
                );
                accountLinks = await apiFetch("/api/account/links", {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ links: next }),
                });
                closeLinkDialog();
                return;
            }
            await loadLinks();
            closeLinkDialog();
        } catch {
            accountLinks = prevLinks;
            linksError = "A mentés nem sikerült";
        }
    }

    async function deleteLink(id) {
        linksError = "";
        const prevLinks = accountLinks;
        try {
            await apiFetch(`/api/account/links?id=${encodeURIComponent(String(id))}`, {
                method: "DELETE",
            });
            accountLinks = prevLinks.filter((l) => l.id !== id);
        } catch {
            accountLinks = prevLinks;
            linksError = "A mentés nem sikerült";
        }
    }

    async function deleteLinkFromDialog(e) {
        e.preventDefault();
        if (!linkDialogData.id) return;
        await deleteLink(linkDialogData.id);
        closeLinkDialog();
    }

    /** @param {number} index @param {number} delta */
    async function moveLink(index, delta) {
        const target = index + delta;
        if (target < 0 || target >= accountLinks.length) return;
        linksError = "";
        const prevLinks = accountLinks;
        const next = [...accountLinks];
        const tmp = next[index];
        next[index] = next[target];
        next[target] = tmp;
        try {
            accountLinks = await apiFetch("/api/account/links", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ links: next }),
            });
        } catch {
            accountLinks = prevLinks;
            linksError = "A mentés nem sikerült";
        }
    }

    /** @param {string} slug */
    async function removeHistoryRow(slug) {
        historyError = "";
        const prev = accountHistory;
        try {
            await apiFetch(
                `/api/account/history?slug=${encodeURIComponent(slug)}`,
                { method: "DELETE" },
            );
            accountHistory = prev.filter((row) => row.slug !== slug);
        } catch {
            accountHistory = prev;
            historyError = "A mentés nem sikerült";
        }
    }

    async function clearAllHistory() {
        historyError = "";
        const prev = accountHistory;
        try {
            await apiFetch("/api/account/history", { method: "DELETE" });
            accountHistory = [];
        } catch {
            accountHistory = prev;
            historyError = "A mentés nem sikerült";
        }
    }

    async function saveDisplayName() {
        displayNameError = "";
        const name = displayName.trim().replace(/\s+/g, " ");
        if ([...name].length > 24) {
            displayNameError = "A megjelenített név legfeljebb 24 karakter lehet.";
            return;
        }
        displayNameSaving = true;
        const prev = displayName;
        displayName = name;
        try {
            await apiFetch("/api/account/preferences", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ display_name: name }),
            });
            await auth.refresh();
        } catch {
            displayName = prev;
            displayNameError = "A mentés nem sikerült";
        } finally {
            displayNameSaving = false;
        }
    }

    let listingsError = $state("");
    /** @type {{ owned: Array<Record<string, unknown>>, member: Array<Record<string, unknown>>, pending: Array<Record<string, unknown>>, unpublished: Array<Record<string, unknown>> }} */
    let accountListings = $state({
        owned: [],
        member: [],
        pending: [],
        unpublished: [],
    });
    /** @type {Record<number, Array<{ user_id: number, email: string }>>} */
    let listingMembersByEntry = $state({});
    let listingsLoading = $state(false);
    let editingListingId = $state(0);

    /** @param {unknown} payload */
    function normalizeListingsPayload(payload) {
        const data = payload && typeof payload === "object" ? payload : {};
        return {
            owned: Array.isArray(data.owned) ? data.owned : [],
            member: Array.isArray(data.member) ? data.member : [],
            pending: Array.isArray(data.pending) ? data.pending : [],
            unpublished: Array.isArray(data.unpublished) ? data.unpublished : [],
        };
    }

    /** @param {number} entryId */
    async function loadListingMembers(entryId) {
        try {
            const rows =
                (await apiFetch(
                    `/api/account/listings/members?entry_id=${encodeURIComponent(String(entryId))}`,
                )) || [];
            listingMembersByEntry = {
                ...listingMembersByEntry,
                [entryId]: Array.isArray(rows) ? rows : [],
            };
        } catch {
            listingMembersByEntry = {
                ...listingMembersByEntry,
                [entryId]: [],
            };
        }
    }

    async function loadListings() {
        listingsLoading = true;
        listingsError = "";
        try {
            const payload = await apiFetch("/api/account/listings");
            accountListings = normalizeListingsPayload(payload);
            const memberLoads = accountListings.owned.map((row) =>
                loadListingMembers(Number(row.id)),
            );
            await Promise.all(memberLoads);
        } catch {
            listingsError = "A mentés nem sikerült";
        } finally {
            listingsLoading = false;
        }
    }

    /** @param {{ slug?: string | null }} item */
    function listingPublicUrl(item) {
        const slug = String(item.slug ?? "").trim();
        if (!slug) return "";
        return `/bejegyzes/${slug}`;
    }

    /** @param {Record<string, unknown>} row */
    function openEditListing(row) {
        listingsError = "";
        editingListingId = Number(row.id);
    }

    /** @param {Record<string, unknown>} row */
    async function deleteListingRow(row) {
        const label = String(row.name ?? "").trim() || "ezt a bejegyzést";
        if (!confirm(`Biztosan törlöd: ${label}?`)) {
            return;
        }
        listingsError = "";
        const prev = accountListings;
        const entryId = Number(row.id);
        try {
            await apiFetch(
                `/api/account/listings?id=${encodeURIComponent(String(entryId))}`,
                { method: "DELETE" },
            );
            await loadListings();
        } catch {
            accountListings = prev;
            listingsError = "A mentés nem sikerült";
        }
    }

    /** @param {number} entryId @param {number} userId */
    async function removeListingMember(entryId, userId) {
        listingsError = "";
        const prevMembers = listingMembersByEntry;
        try {
            await apiFetch(
                `/api/account/listings/members?entry_id=${encodeURIComponent(String(entryId))}&user_id=${encodeURIComponent(String(userId))}`,
                { method: "DELETE" },
            );
            listingMembersByEntry = {
                ...prevMembers,
                [entryId]: (prevMembers[entryId] || []).filter(
                    (row) => row.user_id !== userId,
                ),
            };
        } catch {
            listingMembersByEntry = prevMembers;
            listingsError = "A mentés nem sikerült";
        }
    }



    let websitesError = $state("");
    /** @type {Array<{ id: number, title: string, domain: string, status: string }>} */
    let accountWebsites = $state([]);
    let websitesLoading = $state(false);

    /** @param {string} status */
    function websiteStatusLabel(status) {
        return status === "approved" ? "Jóváhagyva" : "Jóváhagyásra vár";
    }

    async function loadWebsites() {
        websitesLoading = true;
        websitesError = "";
        try {
            const payload = (await apiFetch("/api/account/websites")) || {};
            accountWebsites = Array.isArray(payload.websites) ? payload.websites : [];
        } catch {
            websitesError = "A betöltés nem sikerült";
        } finally {
            websitesLoading = false;
        }
    }

    onMount(() => {
        let accountDataLoaded = false;
        const unsubscribeAuth = auth.subscribe((state) => {
            if (!state.loggedIn) {
                accountDataLoaded = false;
                return;
            }
            if (accountDataLoaded) return;
            accountDataLoaded = true;
            displayName = String(state.displayName || "");
            void loadLinks();
            void loadHistory();
            void loadFavorites();
            void loadListings();
            void loadWebsites();
        });

        void (async () => {
            await auth.init();
            if (!get(auth).loggedIn) {
                openLogin();
            }
        })();

        return () => {
            unsubscribeAuth();
        };
    });
</script>

<section class="page-section profile-page">
    <PublicPageHero
        title="Fiók"
        greeting="A fiókod"
        loading={false}
        breadcrumbLabel="Fiók"
        documentTitleSuffix=" – Lámsza"
    />

    {#if !$auth.loggedIn}
        <div class="info-box profile-login-prompt">
            <p>Jelentkezz be a fiókodhoz</p>
            <button type="button" class="btn" onclick={openLogin}>Belépés</button>
        </div>
    {:else}
        <nav class="header-tabs profile-tabs" aria-label="Fiók">
            {#each userAccountTabIds as tabId (tabId)}
                <button
                    type="button"
                    class="btn"
                    class:active={activeTab === tabId}
                    onclick={() => (activeTab = tabId)}
                >
                    {TAB_LABELS[tabId]}
                </button>
            {/each}
        </nav>

        {#if activeTab === "fiok"}
            <div class="profile-settings">
                <h3>Megjelenített név</h3>
                <p class="profile-hint">
                    Rövid név a kezdőlap üdvözlésében. Üresen a Google-fiók neve marad. Legfeljebb 24 karakter.
                </p>
                <label for="display_name">Név</label>
                <input
                    id="display_name"
                    class="profile-location-select"
                    type="text"
                    maxlength="24"
                    bind:value={displayName}
                    autocomplete="nickname"
                />
                {#if displayNameError}
                    <p class="profile-error">{displayNameError}</p>
                {/if}
                <button
                    type="button"
                    class="btn profile-save"
                    disabled={displayNameSaving}
                    onclick={saveDisplayName}
                >
                    {displayNameSaving ? "Mentés…" : "Mentés"}
                </button>
            </div>
            <dl class="profile-fields">
                {#if hasField($auth.picture)}
                    <dt>Fénykép</dt>
                    <dd><img src={$auth.picture} alt="" class="profile-photo" /></dd>
                {/if}
                {#if hasField($auth.givenName)}
                    <dt>Keresztnév</dt>
                    <dd>{$auth.givenName}</dd>
                {/if}
                {#if hasField($auth.familyName)}
                    <dt>Vezetéknév</dt>
                    <dd>{$auth.familyName}</dd>
                {/if}
                {#if hasField($auth.user)}
                    <dt>Név</dt>
                    <dd>{$auth.user}</dd>
                {/if}
                {#if hasField($auth.email)}
                    <dt>E-mail</dt>
                    <dd>{$auth.email}</dd>
                {/if}
                {#if hasField($auth.locale)}
                    <dt>Nyelv</dt>
                    <dd>{$auth.locale}</dd>
                {/if}
                {#if hasField($auth.googleSub)}
                    <dt>Google azonosító</dt>
                    <dd>{$auth.googleSub}</dd>
                {/if}
                {#if hasField($auth.lastLoginAt)}
                    <dt>Utolsó belépés</dt>
                    <dd>{formatTimestamp($auth.lastLoginAt)}</dd>
                {/if}
                {#if hasField($auth.createdAt)}
                    <dt>Fiók létrehozva</dt>
                    <dd>{formatTimestamp($auth.createdAt)}</dd>
                {/if}
            </dl>
        {:else if activeTab === "adatlapjaim"}
            <div class="profile-listings">
                {#if listingsError}
                    <p class="profile-error">{listingsError}</p>
                {/if}
                {#if listingsLoading}
                    <p>Betöltés…</p>
                {:else}
                    <section class="profile-listings-group">
                        <h3>Saját</h3>
                        {#if accountListings.owned.length === 0}
                            <p class="profile-empty">Nincs saját bejegyzés.</p>
                        {:else}
                            <ul class="profile-listings-list">
                                {#each accountListings.owned as row (row.id)}
                                    <li class="profile-listings-row">
                                        {#if listingPublicUrl(row)}
                                            <a href={listingPublicUrl(row)} class="profile-listings-name">
                                                {row.name}
                                            </a>
                                        {:else}
                                            <span class="profile-listings-name">{row.name}</span>
                                        {/if}
                                        <div class="profile-listings-actions">
                                            <button
                                                type="button"
                                                class="btn btn-xs"
                                                onclick={() => openEditListing(row)}
                                            >Szerkesztés</button>
                                            <button
                                                type="button"
                                                class="btn btn-xs"
                                                onclick={() => deleteListingRow(row)}
                                            >Törlés</button>
                                        </div>
                                        {#if (listingMembersByEntry[Number(row.id)] || []).length > 0}
                                            <ul class="profile-listings-members">
                                                {#each listingMembersByEntry[Number(row.id)] || [] as member (member.user_id)}
                                                    <li class="profile-listings-member-row">
                                                        <span>{member.email || member.user_id}</span>
                                                        <button
                                                            type="button"
                                                            class="btn btn-xs"
                                                            onclick={() =>
                                                                removeListingMember(
                                                                    Number(row.id),
                                                                    Number(member.user_id),
                                                                )}
                                                        >Eltávolítás</button>
                                                    </li>
                                                {/each}
                                            </ul>
                                        {/if}
                                    </li>
                                {/each}
                            </ul>
                        {/if}
                    </section>

                    <section class="profile-listings-group">
                        <h3>Tagság</h3>
                        {#if accountListings.member.length === 0}
                            <p class="profile-empty">Nincs tagsági bejegyzés.</p>
                        {:else}
                            <ul class="profile-listings-list">
                                {#each accountListings.member as row (row.id)}
                                    <li class="profile-listings-row">
                                        {#if listingPublicUrl(row)}
                                            <a href={listingPublicUrl(row)} class="profile-listings-name">
                                                {row.name}
                                            </a>
                                        {:else}
                                            <span class="profile-listings-name">{row.name}</span>
                                        {/if}
                                        <div class="profile-listings-actions">
                                            <button
                                                type="button"
                                                class="btn btn-xs"
                                                onclick={() => openEditListing(row)}
                                            >Szerkesztés</button>
                                        </div>
                                    </li>
                                {/each}
                            </ul>
                        {/if}
                    </section>

                    <section class="profile-listings-group">
                        <h3>Jóváhagyásra vár</h3>
                        {#if accountListings.pending.length === 0}
                            <p class="profile-empty">Nincs függő tagságkérés.</p>
                        {:else}
                            <ul class="profile-listings-list">
                                {#each accountListings.pending as row (row.id)}
                                    <li class="profile-listings-row">
                                        {#if listingPublicUrl(row)}
                                            <a href={listingPublicUrl(row)} class="profile-listings-name">
                                                {row.name}
                                            </a>
                                        {:else}
                                            <span class="profile-listings-name">{row.name}</span>
                                        {/if}
                                    </li>
                                {/each}
                            </ul>
                        {/if}
                    </section>

                    <section class="profile-listings-group">
                        <h3>Közzétételre vár</h3>
                        {#if accountListings.unpublished.length === 0}
                            <p class="profile-empty">Nincs közzétételre váró bejegyzés.</p>
                        {:else}
                            <ul class="profile-listings-list">
                                {#each accountListings.unpublished as row (row.id)}
                                    <li class="profile-listings-row">
                                        <span class="profile-listings-name">{row.name}</span>
                                        <div class="profile-listings-actions">
                                            <button
                                                type="button"
                                                class="btn btn-xs"
                                                onclick={() => openEditListing(row)}
                                            >Szerkesztés</button>
                                            <button
                                                type="button"
                                                class="btn btn-xs"
                                                onclick={() => deleteListingRow(row)}
                                            >Törlés</button>
                                        </div>
                                    </li>
                                {/each}
                            </ul>
                        {/if}
                    </section>
                {/if}
            </div>
        {:else if activeTab === "weboldalaim"}
            <div class="profile-listings">
                {#if websitesError}
                    <p class="profile-error">{websitesError}</p>
                {/if}
                {#if websitesLoading}
                    <p>Betöltés…</p>
                {:else if accountWebsites.length === 0}
                    <p class="profile-empty">Még nincs weboldalad.</p>
                {:else}
                    <ul class="profile-listings-list">
                        {#each accountWebsites as site (site.id)}
                            <li class="profile-listings-row">
                                <span class="profile-listings-name">{site.title}</span>
                                <span class="profile-history-meta">{site.domain}</span>
                                <span class="profile-history-meta">{websiteStatusLabel(site.status)}</span>
                            </li>
                        {/each}
                    </ul>
                {/if}
            </div>
        {:else if activeTab === "kereses"}
            <p class="profile-empty">A keresési előzmények még nincsenek bekötve.</p>
        {:else if activeTab === "bongeszes"}
            <div class="profile-history">
                {#if historyError}
                    <p class="profile-error">{historyError}</p>
                {/if}
                {#if historyLoading}
                    <p>Betöltés…</p>
                {:else if accountHistory.length === 0}
                    <p class="profile-empty">Nincs böngészési előzmény.</p>
                {:else}
                    <button type="button" class="btn profile-clear-all" onclick={clearAllHistory}>
                        Összes törlése
                    </button>
                    <ul class="profile-history-list">
                        {#each accountHistory as row (row.slug)}
                            <li class="profile-history-row">
                                <a href="/bejegyzes/{row.slug}" class="profile-history-name">
                                    {row.name}
                                </a>
                                {#if row.location}
                                    <span class="profile-history-meta">{row.location}</span>
                                {/if}
                                <button
                                    type="button"
                                    class="btn btn-xs"
                                    onclick={() => removeHistoryRow(row.slug)}
                                >Eltávolítás</button>
                            </li>
                        {/each}
                    </ul>
                {/if}
            </div>
        {:else if activeTab === "kedvencek"}
            <div class="profile-favorites">
                {#if favoritesError}
                    <p class="profile-error">{favoritesError}</p>
                {/if}
                {#if favoritesLoading}
                    <p>Betöltés…</p>
                {:else}
                    {#if accountFavorites.settlements.length > 0}
                        <h3>Települések</h3>
                        <ul class="profile-favorites-list">
                            {#each accountFavorites.settlements as row (row.id)}
                                <li class="profile-favorites-row">
                                    {#if countyPlaceUrl(row)}
                                        <a href={countyPlaceUrl(row)} class="profile-favorites-name">
                                            {row.name}
                                        </a>
                                    {:else}
                                        <span class="profile-favorites-name">{row.name}</span>
                                    {/if}
                                    <button
                                        type="button"
                                        class="btn btn-xs"
                                        onclick={() => removeFavoriteRow("settlement", row.id)}
                                    >Eltávolítás</button>
                                </li>
                            {/each}
                        </ul>
                    {/if}
                    {#if accountFavorites.attractions.length > 0}
                        <h3>Látnivalók</h3>
                        <ul class="profile-favorites-list">
                            {#each accountFavorites.attractions as row (row.id)}
                                <li class="profile-favorites-row">
                                    {#if countyPlaceUrl(row)}
                                        <a href={countyPlaceUrl(row)} class="profile-favorites-name">
                                            {row.name}
                                        </a>
                                    {:else}
                                        <span class="profile-favorites-name">{row.name}</span>
                                    {/if}
                                    <button
                                        type="button"
                                        class="btn btn-xs"
                                        onclick={() => removeFavoriteRow("attraction", row.id)}
                                    >Eltávolítás</button>
                                </li>
                            {/each}
                        </ul>
                    {/if}
                    {#if accountFavorites.entries.length > 0}
                        <h3>Bejegyzések</h3>
                        <ul class="profile-favorites-list">
                            {#each accountFavorites.entries as row (row.id)}
                                <li class="profile-favorites-row">
                                    {#if entryFavoriteUrl(row)}
                                        <a href={entryFavoriteUrl(row)} class="profile-favorites-name">
                                            {row.name}
                                        </a>
                                    {:else}
                                        <span class="profile-favorites-name">{row.name}</span>
                                    {/if}
                                    <button
                                        type="button"
                                        class="btn btn-xs"
                                        onclick={() => removeFavoriteRow("entry", row.id)}
                                    >Eltávolítás</button>
                                </li>
                            {/each}
                        </ul>
                    {/if}
                    {#if accountFavorites.events.length > 0}
                        <h3>Események</h3>
                        <ul class="profile-favorites-list">
                            {#each accountFavorites.events as row (row.id)}
                                <li class="profile-favorites-row">
                                    <a href={eventFavoriteUrl(row)} class="profile-favorites-name">
                                        {row.name}
                                    </a>
                                    <button
                                        type="button"
                                        class="btn btn-xs"
                                        onclick={() => removeFavoriteRow("event", row.id)}
                                    >Eltávolítás</button>
                                </li>
                            {/each}
                        </ul>
                    {/if}
                    {#if accountFavorites.settlements.length === 0 && accountFavorites.attractions.length === 0 && accountFavorites.entries.length === 0 && accountFavorites.events.length === 0}
                        <p class="profile-empty">Még nincs kedvenc hely.</p>
                    {/if}
                {/if}
            </div>
        {:else if activeTab === "linkjeim"}
            <div class="profile-links">
                {#if linksError}
                    <p class="profile-error">{linksError}</p>
                {/if}
                <button type="button" class="btn" onclick={openAddLink}>Új link</button>
                {#if linksLoading}
                    <p>Betöltés…</p>
                {:else if accountLinks.length === 0}
                    <p class="profile-empty">Még nincs mentett link.</p>
                {:else}
                    <ul class="profile-link-list">
                        {#each accountLinks as link, index (link.id)}
                            <li class="profile-link-row">
                                <span
                                    class="profile-link-swatch"
                                    style:background={link.bg_color || "#e6f0ff"}
                                ></span>
                                <span class="profile-link-title">{link.title}</span>
                                <a
                                    href={link.url}
                                    class="profile-link-url"
                                    target="_blank"
                                    rel="nofollow noopener"
                                >{link.url}</a>
                                <div class="profile-link-actions">
                                    <button
                                        type="button"
                                        class="btn btn-xs"
                                        disabled={index === 0}
                                        aria-label="Fel"
                                        onclick={() => moveLink(index, -1)}
                                    >↑</button>
                                    <button
                                        type="button"
                                        class="btn btn-xs"
                                        disabled={index === accountLinks.length - 1}
                                        aria-label="Le"
                                        onclick={() => moveLink(index, 1)}
                                    >↓</button>
                                    <button
                                        type="button"
                                        class="btn btn-xs"
                                        onclick={() => openEditLink(link)}
                                    >Szerkesztés</button>
                                    <button
                                        type="button"
                                        class="btn btn-xs"
                                        onclick={() => deleteLink(link.id)}
                                    >Törlés</button>
                                </div>
                            </li>
                        {/each}
                    </ul>
                {/if}
            </div>
        {/if}
    {/if}
</section>

{#if linkDialogOpen}
    <div
        class="link-dialog-overlay"
        role="dialog"
        aria-labelledby="profile-link-dialog-title"
        tabindex="-1"
        onclick={(e) => e.target === e.currentTarget && closeLinkDialog()}
        onkeydown={(e) => e.key === "Escape" && closeLinkDialog()}
    >
        <div class="link-dialog" role="presentation" onclick={(e) => e.stopPropagation()}>
            <h3 id="profile-link-dialog-title">
                {linkDialogMode === "add" ? "Új gyorslink hozzáadása" : "Gyorslink szerkesztése"}
            </h3>
            <form class="link-dialog-form" onsubmit={saveLinkDialog}>
                <label for="profile_link_title">Cím</label>
                <input
                    id="profile_link_title"
                    type="text"
                    bind:value={linkDialogData.title}
                    required
                />

                <label for="profile_link_url">URL</label>
                <input
                    id="profile_link_url"
                    type="url"
                    bind:value={linkDialogData.url}
                    required
                />

                <label for="profile_link_color">Háttérszín (pl. #e6f0ff)</label>
                <input
                    id="profile_link_color"
                    type="text"
                    bind:value={linkDialogData.bg_color}
                    placeholder="#e6f0ff"
                />

                <div class="link-dialog-actions">
                    <button type="submit" class="link-dialog-submit">Mentés</button>
                    {#if linkDialogMode === "edit"}
                        <button
                            type="button"
                            class="link-dialog-delete"
                            onclick={deleteLinkFromDialog}
                        >Törlés</button>
                    {/if}
                    <button type="button" class="link-dialog-cancel" onclick={closeLinkDialog}>
                        Mégse
                    </button>
                </div>
            </form>
        </div>
    </div>
{/if}

{#if editingListingId > 0}
    <ListingFormDialog
        mode="edit"
        entryId={editingListingId}
        onClose={() => (editingListingId = 0)}
        onSaved={loadListings}
        onError={(message) => (listingsError = message)}
    />
{/if}

<style>
    .profile-page {
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }
    .profile-login-prompt {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: 0.75rem;
    }
    .profile-tabs {
        flex-wrap: wrap;
    }
    .profile-fields {
        display: grid;
        grid-template-columns: minmax(8rem, 12rem) 1fr;
        gap: 0.5rem 1rem;
        margin: 0;
    }
    .profile-fields dt {
        margin: 0;
        font-weight: 600;
        color: var(--text-muted, #666);
    }
    .profile-fields dd {
        margin: 0;
    }
    .profile-photo {
        width: 4rem;
        height: 4rem;
        border-radius: 50%;
        object-fit: cover;
    }
    .profile-settings {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        align-items: flex-start;
    }
    .profile-settings h3 {
        margin: 0.5rem 0 0;
        font-size: 1rem;
    }
    .profile-hint {
        margin: 0;
        max-width: 36rem;
        color: var(--text-muted, #666);
    }
    .profile-location-select {
        min-width: min(100%, 22rem);
        max-width: 100%;
        padding: 0.45rem 0.6rem;
        border-radius: 8px;
        border: 1px solid var(--border-color);
        background: var(--card-bg, #fff);
        color: var(--text-primary);
        font: inherit;
    }
    .profile-error {
        margin: 0;
        color: #b00020;
    }
    .profile-empty {
        margin: 0.5rem 0 0;
        color: var(--text-muted, #666);
    }
    .profile-link-list,
    .profile-history-list,
    .profile-favorites-list {
        list-style: none;
        margin: 0.75rem 0 0;
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        width: 100%;
    }
    .profile-link-row,
    .profile-history-row,
    .profile-favorites-row {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 0;
        border-bottom: 1px solid var(--border-color, #ddd);
    }
    .profile-link-swatch {
        width: 1rem;
        height: 1rem;
        border-radius: 3px;
        flex-shrink: 0;
    }
    .profile-link-title {
        font-weight: 600;
    }
    .profile-link-url {
        font-size: 0.9rem;
        word-break: break-all;
    }
    .profile-link-actions {
        display: flex;
        flex-wrap: wrap;
        gap: 0.35rem;
        margin-left: auto;
    }
    .profile-history-name,
    .profile-favorites-name {
        font-weight: 600;
    }
    .profile-favorites h3 {
        margin: 0.75rem 0 0;
        font-size: 1rem;
    }
    .profile-favorites h3:first-child {
        margin-top: 0;
    }
    .profile-history-meta {
        color: var(--text-muted, #666);
        font-size: 0.9rem;
    }
    .profile-clear-all {
        margin-bottom: 0.5rem;
    }
    .profile-listings {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        align-items: flex-start;
        width: 100%;
    }
    .profile-listings-group {
        width: 100%;
    }
    .profile-listings-group h3 {
        margin: 0.75rem 0 0;
        font-size: 1rem;
    }
    .profile-listings-group h3:first-child {
        margin-top: 0;
    }
    .profile-listings-list {
        list-style: none;
        margin: 0.75rem 0 0;
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        width: 100%;
    }
    .profile-listings-row {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 0;
        border-bottom: 1px solid var(--border-color, #ddd);
    }
    .profile-listings-name {
        font-weight: 600;
    }
    .profile-listings-actions {
        display: flex;
        flex-wrap: wrap;
        gap: 0.35rem;
        margin-left: auto;
    }
    .profile-listings-members {
        list-style: none;
        margin: 0.35rem 0 0;
        padding: 0 0 0 1rem;
        width: 100%;
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
    }
    .profile-listings-member-row {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.5rem;
        font-size: 0.9rem;
        color: var(--text-muted, #666);
    }
</style>
