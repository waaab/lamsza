<script>
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import AddWebsiteForm from "$lib/components/AddWebsiteForm.svelte";
    import ListingFormDialog from "$lib/components/ListingFormDialog.svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import WebsiteCard from "$lib/components/WebsiteCard.svelte";
    import ClaimStatusFilter from "$lib/components/ClaimStatusFilter.svelte";
    import IndexTagAside from "$lib/components/IndexTagAside.svelte";
    import { locationMenuTowns, searchPreferredLocation, sortDirectoryEntries, sortWebsites } from "$lib/directoryListingOrder.js";
    import { auth } from "$lib/stores/auth";
    import {
        entryMatchesAsideFilters,
        normalizeTagKey,
        displayTagLabel,
    } from "$lib/directoryTagCloud.js";
    import {
        canonicalEntryType,
        canonicalEntryTypeKey,
        filterServiceEntries,
    } from "$lib/entryType.js";
    import {
        canonicalEntryCategory,
        directoryCategoryTabs,
        entryMatchesCategory,
    } from "$lib/entryCategory.js";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";
    import { apiFetch } from "$lib/api.js";
    import { openLogin } from "$lib/openLogin.js";
    import AppIcon from "$lib/icons/AppIcon.svelte";

    let pageHeader = initialPageHeader("index");
    let servicesHeader = initialPageHeader("index/szolgaltatasok");
    let websitesHeader = initialPageHeader("index/weboldalak");
    let pageHeaderLoading = false;

    let dynamicCategories = [{ id: "osszes", label: "Összes", url: "/index" }];
    let entries = [];
    /** @type {Array<{ id: number, title: string, description: string, domain: string, url: string, claimed?: boolean }>} */
    let websites = [];
    let loading = true;
    let error = null;
    let addWebsiteOpen = false;
    let createListingOpen = false;

    async function openCreateListing() {
        await auth.init();
        if (!$auth.loggedIn) {
            openLogin();
            return;
        }
        createListingOpen = true;
    }

    /** @type {string | null} */
    let selectedTypeKey = null;
    /** @type {string | null} */
    let selectedTagKey = null;
    /** @type {"claimed" | "unclaimed" | null} */
    let serviceClaimFilter = null;
    /** @type {"claimed" | "unclaimed" | null} */
    let websiteClaimFilter = null;

    /** @param {Record<string, any>} item @param {"claimed" | "unclaimed" | null} filter */
    function matchesClaim(item, filter) {
        if (filter === "claimed") return Boolean(item?.claimed);
        if (filter === "unclaimed") return !item?.claimed;
        return true;
    }

    let serviceViewMode = "grid";
    let websiteViewMode = "flex";
    let currentCategory = "osszes";
    let visibleServiceCount = 12;
    let visibleWebsiteCount = 12;
    let serviceSortMode = "title";
    let websiteSortMode = "title";
    /** @type {"services" | "websites" | null} */
    let sortMenuKey = null;
    /** @type {{ slug: string, name: string, county_slug: string } | null} */
    let selectedPlace = null;
    /** @type {string | null} */
    let locationMenuKey = null;
    /** @type {Array<Record<string, any>>} */
    let locations = [];

    const sortLabels = { title: "Név (A→Z)", newest: "Legújabb" };

    $: preferredLocation = searchPreferredLocation($auth.preferredLocation, $auth.loggedIn);
    $: townChoices = locationMenuTowns(locations, preferredLocation);
    $: indexView =
        $page.url.pathname === "/index/weboldalak"
            ? "websites"
            : $page.url.pathname === "/index/szolgaltatasok"
              ? "services"
              : "all";
    $: viewEntries = filterServiceEntries(entries);

    /**
     * @param {"services" | "websites"} kind
     * @param {"title" | "newest"} mode
     */
    function setSortMode(kind, mode) {
        if (kind === "websites") websiteSortMode = mode;
        else serviceSortMode = mode;
        sortMenuKey = null;
        locationMenuKey = null;
    }

    /** @param {"services" | "websites"} kind */
    function toggleSortMenu(kind) {
        sortMenuKey = sortMenuKey === kind ? null : kind;
        locationMenuKey = null;
    }

    /**
     * @param {"services" | "websites"} kind
     * @param {"grid" | "flex"} mode
     */
    function setViewMode(kind, mode) {
        if (kind === "websites") websiteViewMode = mode;
        else serviceViewMode = mode;
    }

    /** @param {string} key */
    function toggleLocationMenu(key) {
        locationMenuKey = locationMenuKey === key ? null : key;
        sortMenuKey = null;
    }

    /** @param {Record<string, any>} loc */
    function choosePlace(loc) {
        const slug = String(loc?.slug || "").trim();
        if (!slug) return;
        selectedPlace = {
            slug,
            name: String(loc?.name || "").trim() || slug,
            county_slug: String(loc?.county_slug || "").trim(),
        };
        locationMenuKey = null;
    }

    /** @param {PointerEvent} event */
    function handleWindowPointerDown(event) {
        const target = event.target;
        if (!locationMenuKey) return;
        if (target instanceof Element && target.closest(".index-location")) return;
        locationMenuKey = null;
    }

    /** @param {Record<string, any>} entry @param {{ slug?: string } | null} place */
    function matchesPlace(entry, place) {
        if (!place?.slug) return true;
        return String(entry?.location_slug || "") === place.slug;
    }

    function scrollToTop() {
        if (typeof window !== "undefined") {
            window.scrollTo({ top: 0, behavior: "smooth" });
        }
    }

    /** Kategória (nem Összes) vagy oldalsáv típus/címke szűrő */
    $: hasActiveFilters =
        currentCategory !== "osszes" ||
        selectedTypeKey != null ||
        selectedTagKey != null ||
        serviceClaimFilter != null ||
        selectedPlace != null;

    function clearAllFilters() {
        currentCategory = "osszes";
        selectedTypeKey = null;
        selectedTagKey = null;
        serviceClaimFilter = null;
        websiteClaimFilter = null;
        selectedPlace = null;
        locationMenuKey = null;
        scrollToTop();
    }

    $: filteredEntries = viewEntries.filter(
        (e) =>
            (currentCategory === "osszes" ||
                entryMatchesCategory(e, currentCategory)) &&
            entryMatchesAsideFilters(e, selectedTypeKey, selectedTagKey) &&
            matchesPlace(e, selectedPlace) &&
            matchesClaim(e, serviceClaimFilter),
    );
    $: serviceClaimCounts = {
        claimed: viewEntries.filter((e) =>
            (currentCategory === "osszes" || entryMatchesCategory(e, currentCategory)) &&
            entryMatchesAsideFilters(e, selectedTypeKey, selectedTagKey) &&
            matchesPlace(e, selectedPlace) &&
            e.claimed,
        ).length,
        unclaimed: viewEntries.filter((e) =>
            (currentCategory === "osszes" || entryMatchesCategory(e, currentCategory)) &&
            entryMatchesAsideFilters(e, selectedTypeKey, selectedTagKey) &&
            matchesPlace(e, selectedPlace) &&
            !e.claimed,
        ).length,
    };

    $: categoryFilterLabel =
        currentCategory === "osszes"
            ? null
            : dynamicCategories.find((c) => c.id === currentCategory)?.label ||
              canonicalEntryCategory(currentCategory);

    $: typeFilterLabel =
        selectedTypeKey &&
        canonicalEntryType(
            entries.find(
                (e) => canonicalEntryTypeKey(e.type) === selectedTypeKey,
            )?.type,
        );

    $: tagFilterLabel = (() => {
        if (!selectedTagKey) return null;
        for (const e of entries) {
            for (const raw of e.tags || []) {
                if (normalizeTagKey(raw) === selectedTagKey) {
                    return displayTagLabel(raw);
                }
            }
        }
        return selectedTagKey;
    })();
    $: sortedEntries = sortDirectoryEntries(filteredEntries, { sortMode: serviceSortMode });
    $: totalCount = sortedEntries.length;
    $: displayItems = sortedEntries.slice(0, visibleServiceCount);
    $: filteredWebsites = websites.filter((site) => matchesClaim(site, websiteClaimFilter));
    $: sortedWebsites = sortWebsites(filteredWebsites, websiteSortMode);
    $: websiteClaimCounts = {
        claimed: websites.filter((site) => site.claimed).length,
        unclaimed: websites.filter((site) => !site.claimed).length,
    };
    $: displayWebsites = sortedWebsites.slice(0, visibleWebsiteCount);

    function loadMoreServices() {
        visibleServiceCount += 12;
    }

    function loadMoreWebsites() {
        visibleWebsiteCount += 12;
    }

    $: {
        currentCategory;
        selectedTypeKey;
        selectedTagKey;
        serviceClaimFilter;
        websiteClaimFilter;
        selectedPlace;
        indexView;
        visibleServiceCount = 12;
        visibleWebsiteCount = 12;
    }

    $: activeHeader =
        indexView === "services"
            ? servicesHeader
            : indexView === "websites"
              ? websitesHeader
              : pageHeader;
    $: indexGreeting = activeHeader.greeting;

    onMount(() => {
        loadPageMeta("index").then((p) => {
            pageHeader = p;
        });
        loadPageMeta("index/szolgaltatasok").then((p) => {
            servicesHeader = p;
        });
        loadPageMeta("index/weboldalak").then((p) => {
            websitesHeader = p;
        });
        (async () => {
            try {
                const [directory, locs, websitesData] = await Promise.all([
                    apiFetch("/api/directory"),
                    apiFetch("/api/locations"),
                    apiFetch("/api/websites"),
                ]);
                entries = directory || [];
                websites = websitesData?.websites || [];
                dynamicCategories = directoryCategoryTabs(entries);
                locations = Array.isArray(locs) ? locs : [];
            } catch (err) {
                console.error(err);
                error = "Hiba történt az adatok betöltésekor.";
            } finally {
                loading = false;
            }
        })();
    });
</script>

<PublicPageHero
    title="Index"
    greeting={indexGreeting}
    showGreeting={false}
    loading={pageHeaderLoading}
    breadcrumbLabel={indexView === "websites"
        ? "Weboldalak"
        : indexView === "services"
          ? "Szolgáltatások"
          : "Index"}
    breadcrumbParentLabel={indexView === "all" ? "" : "Index"}
    breadcrumbParentUrl={indexView === "all" ? "" : "/index"}
    documentTitleSuffix=" - Székely Gugel"
>
    <div slot="title" class="index-heading">
        <div class="index-heading__main">
            <div class="index-heading__titles">
                <h1 class="page-title">
                    {#if indexView === "all"}
                        Index
                    {:else}
                        <a href="/index">Index</a>
                    {/if}
                </h1>
                <a
                    href="/index/szolgaltatasok"
                    class="index-heading__link"
                    class:active={indexView === "services"}
                    aria-current={indexView === "services" ? "page" : undefined}
                >Szolgáltatások</a>
                <a
                    href="/index/weboldalak"
                    class="index-heading__link"
                    class:active={indexView === "websites"}
                    aria-current={indexView === "websites" ? "page" : undefined}
                >Weboldalak</a>
            </div>
            {#if !pageHeaderLoading && indexGreeting}
                <p class="greeting index-heading__greeting">{indexGreeting}</p>
            {/if}
        </div>
        {#if indexView === "websites"}
            <div class="index-heading__add">
                <button
                    type="button"
                    class="btn btn-primary btn-lg"
                    on:click={() => (addWebsiteOpen = true)}
                >Add hozzá a weboldalad</button>
                <p>
                    Ingyenes. A webcím, a cím és egy rövid leírás kell. Az admin jóváhagyása után a weboldal megjelenik az indexen.
                </p>
            </div>
        {:else}
            <div class="index-heading__add">
                <button
                    type="button"
                    class="btn btn-primary btn-lg"
                    on:click={openCreateListing}
                >Új bejegyzés</button>
            </div>
        {/if}
    </div>
</PublicPageHero>

{#if addWebsiteOpen}
    <AddWebsiteForm onClose={() => (addWebsiteOpen = false)} />
{/if}

{#if createListingOpen}
    <ListingFormDialog mode="create" onClose={() => (createListingOpen = false)} />
{/if}

<svelte:window on:pointerdown={handleWindowPointerDown} />

{#snippet indexFilterBar(kind, bar)}
    <div class="filter-actions">
        <span class="info-box">
            <p>
                {#if kind === "websites" && !websiteClaimFilter}
                    💡 Leszűrve: <span class="active">Összes</span>
                {:else if kind === "websites"}
                    🔍 Szűrők:
                    <span class="active">{websiteClaimFilter === "claimed" ? "Átvéve" : "Gazdátlan"}</span>
                    <button
                        type="button"
                        class="clear-filters btn btn-xs"
                        aria-label="Szűrők törlése"
                        title="Szűrők törlése"
                        on:click={() => (websiteClaimFilter = null)}>Szűrő törlése</button
                    >
                {:else if !hasActiveFilters}
                    💡 Leszűrve: <span class="active">Összes</span>
                {:else}
                    🔍 Szűrők:
                    {#if currentCategory !== "osszes"}
                        <span class="active">{categoryFilterLabel}</span>
                    {/if}
                    {#if typeFilterLabel}
                        {#if currentCategory !== "osszes"}<span class="filter-sep">·</span
                            >{/if}
                        <span class="active">{typeFilterLabel}</span>
                    {/if}
                    {#if tagFilterLabel}
                        {#if currentCategory !== "osszes" || typeFilterLabel}<span
                                class="filter-sep">·</span
                            >{/if}
                        <span class="active">{tagFilterLabel}</span>
                    {/if}
                    {#if serviceClaimFilter}
                        {#if currentCategory !== "osszes" || typeFilterLabel || tagFilterLabel}<span
                                class="filter-sep">·</span
                            >{/if}
                        <span class="active">{serviceClaimFilter === "claimed" ? "Átvéve" : "Gazdátlan"}</span>
                    {/if}
                    {#if selectedPlace}
                        {#if currentCategory !== "osszes" || typeFilterLabel || tagFilterLabel || serviceClaimFilter}<span
                                class="filter-sep">·</span
                            >{/if}
                        <span class="active">{selectedPlace.name}</span>
                    {/if}
                    <button
                        type="button"
                        class="clear-filters btn btn-xs"
                        aria-label="Szűrők törlése"
                        title="Szűrők törlése"
                        on:click={clearAllFilters}>Szűrő törlése</button
                    >
                {/if}
            </p>
            <p><span>({kind === "websites" ? displayWebsites.length : displayItems.length}/{kind === "websites" ? sortedWebsites.length : totalCount})</span></p>
        </span>

        <div class="view-mode-toggle">
            <div class="index-location">
                <button
                    type="button"
                    class="btn btn-sm"
                    class:active={kind !== "websites" && !!selectedPlace}
                    disabled={kind === "websites"}
                    aria-expanded={kind !== "websites" && locationMenuKey === `${kind}:${bar}`}
                    aria-haspopup={kind === "websites" ? undefined : "listbox"}
                    title={kind === "websites"
                        ? "A weboldalak település szerint még nem szűrhetők."
                        : selectedPlace
                          ? selectedPlace.name
                          : "Település"}
                    on:click={() => toggleLocationMenu(`${kind}:${bar}`)}
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        aria-hidden="true"
                        ><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path><circle
                            cx="12"
                            cy="10"
                            r="3"
                        ></circle></svg
                    >
                    <span>{kind !== "websites" && selectedPlace ? selectedPlace.name : "Település"}</span>
                </button>
                {#if kind !== "websites" && locationMenuKey === `${kind}:${bar}`}
                    <ul class="index-location-menu" role="listbox">
                        {#if preferredLocation}
                            <li>
                                <button
                                    type="button"
                                    role="option"
                                    aria-selected={selectedPlace?.slug === preferredLocation.slug}
                                    on:click={() => choosePlace(preferredLocation)}
                                >
                                    <span class="index-location-option-label">Településem</span>
                                    <span>{preferredLocation.name}</span>
                                </button>
                            </li>
                        {/if}
                        {#each townChoices as town (town.slug + town.county_slug)}
                            <li>
                                <button
                                    type="button"
                                    role="option"
                                    aria-selected={selectedPlace?.slug === town.slug}
                                    on:click={() => choosePlace(town)}
                                >
                                    <span>{town.name}</span>
                                    {#if town.county}
                                        <span class="index-location-option-meta">{town.county}</span>
                                    {/if}
                                </button>
                            </li>
                        {/each}
                    </ul>
                {/if}
            </div>
            <div class="sort-toggle">
                <button class="btn btn-sm" on:click={() => toggleSortMenu(kind)}>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        ><line x1="4" y1="6" x2="16" y2="6"></line><line
                            x1="4"
                            y1="12"
                            x2="12"
                            y2="12"
                        ></line><line x1="4" y1="18" x2="8" y2="18"></line><polyline
                            points="15 15 18 18 21 15"
                        ></polyline><line x1="18" y1="10" x2="18" y2="18"
                        ></line></svg
                    >
                    <span>{sortLabels[kind === "websites" ? websiteSortMode : serviceSortMode]}</span>
                </button>
                {#if sortMenuKey === kind}
                    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
                    <div class="sort-toggle-menu" on:click|stopPropagation>
                        <button
                            class:active={(kind === "websites" ? websiteSortMode : serviceSortMode) === "title"}
                            on:click={() => setSortMode(kind, "title")}>Név (A→Z)</button
                        >
                        <button
                            class:active={(kind === "websites" ? websiteSortMode : serviceSortMode) === "newest"}
                            on:click={() => setSortMode(kind, "newest")}>Legújabb</button
                        >
                    </div>
                {/if}
            </div>

            <button
                class="btn btn-sm {(kind === 'websites' ? websiteViewMode : serviceViewMode) === 'grid' ? 'active' : ''}"
                on:click={() => setViewMode(kind, "grid")}
                title="Rács nézet"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><rect x="3" y="3" width="7" height="7"></rect><rect
                        x="14"
                        y="3"
                        width="7"
                        height="7"
                    ></rect><rect x="14" y="14" width="7" height="7"></rect><rect
                        x="3"
                        y="14"
                        width="7"
                        height="7"
                    ></rect></svg
                >
                <span>Rács</span>
            </button>
            <button
                class="btn btn-sm {(kind === 'websites' ? websiteViewMode : serviceViewMode) === 'flex' ? 'active' : ''}"
                on:click={() => setViewMode(kind, "flex")}
                title="Lista nézet"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><line x1="8" y1="6" x2="21" y2="6"></line><line
                        x1="8"
                        y1="12"
                        x2="21"
                        y2="12"
                    ></line><line x1="8" y1="18" x2="21" y2="18"></line><line
                        x1="3"
                        y1="6"
                        x2="3.01"
                        y2="6"
                    ></line><line x1="3" y1="12" x2="3.01" y2="12"></line><line
                        x1="3"
                        y1="18"
                        x2="3.01"
                        y2="18"
                    ></line></svg
                >
                <span>Lista</span>
            </button>
        </div>
    </div>
{/snippet}

{#snippet directoryBlock(kind)}
<section class="index-directory" aria-label={kind === "websites" ? websitesHeader.title : servicesHeader.title}>
    <h2 class="index-directory__title">{kind === "websites" ? websitesHeader.title : servicesHeader.title}</h2>
    <div class="header-tabs">
        <span class="header-tabs-label" aria-label="Kiemelt Kategóriák">Kiemelt Kategóriák:</span>
        {#if kind === "websites"}
            <div class="header-tabs-filters-row"></div>
        {:else if loading}
            <span class="btn btn--loading">Szűrők betöltése…</span>
        {:else}
            <div class="header-tabs-filters-row">
                {#each dynamicCategories as cat}
                    <button
                        class="btn btn-md {cat.id === currentCategory ? 'active' : ''}"
                        on:click={() => (currentCategory = cat.id)}>{cat.label}</button
                    >
                {/each}
            </div>
        {/if}
    </div>
    {@render indexFilterBar(kind, "start")}
    <div class="list-page-layout">
        <section class="list">
            {#if kind === "websites"}
                {#if loading}
                    <div class="list {(kind === 'websites' ? websiteViewMode : serviceViewMode) === 'grid' ? 'grid' : 'flex'}">
                        {#each Array(6) as _, i (i)}
                            <EntryCard placeholder layout={(kind === "websites" ? websiteViewMode : serviceViewMode) === "grid" ? "grid" : "list"} />
                        {/each}
                    </div>
                {:else if error}
                    <span class="info-box error"><p>{error}</p></span>
                {:else if displayWebsites.length === 0}
                    <span class="info-box info"><p>{websites.length === 0 ? "Még nincs jóváhagyott weboldal." : "Nincs találat a szűrésre."}</p></span>
                {:else}
                    <div class="list {(kind === 'websites' ? websiteViewMode : serviceViewMode) === 'grid' ? 'grid' : 'flex'}">
                        {#each displayWebsites as website (website.id)}
                            <WebsiteCard {website} />
                        {/each}
                    </div>
                    {#if visibleWebsiteCount < sortedWebsites.length}
                        <div class="load-more">
                            <button class="btn nav-btn" on:click={loadMoreWebsites}>Több betöltése ↓</button>
                        </div>
                    {/if}
                {/if}
            {:else if loading}
                <div class="list {(kind === 'websites' ? websiteViewMode : serviceViewMode) === 'grid' ? 'grid' : 'flex'}">
                    {#each Array(6) as _, i (i)}
                        <EntryCard placeholder layout={(kind === "websites" ? websiteViewMode : serviceViewMode) === "grid" ? "grid" : "list"} />
                    {/each}
                </div>
            {:else if error}
                <span class="info-box error"><p>{error}</p></span>
            {:else if displayItems.length === 0}
                <span class="info-box info"><p>Nincs megjeleníthető szolgáltatás.</p></span>
            {:else}
                <div class="list {(kind === 'websites' ? websiteViewMode : serviceViewMode) === 'grid' ? 'grid' : 'flex'}">
                    {#each displayItems as entry}
                        <EntryCard {entry} layout={(kind === "websites" ? websiteViewMode : serviceViewMode) === "grid" ? "grid" : "list"} />
                    {/each}
                </div>
                {#if visibleServiceCount < totalCount}
                    <div class="load-more">
                        <button class="btn nav-btn" on:click={loadMoreServices}>Több betöltése ↓</button>
                    </div>
                {/if}
            {/if}
        </section>
        <aside class="sidebar index-tags-sidebar" aria-label={kind === "websites" ? "Weboldalak" : "Szolgáltatások"}>
            <div class="sidebar-box">
                <div class="sidebar-header">
                    <h4 class="sidebar-heading">{kind === "websites" ? "Weboldalak" : "Szolgáltatások"}</h4>
                </div>
                {#if kind === "websites"}
                    <ClaimStatusFilter
                        bind:value={websiteClaimFilter}
                        claimedCount={websiteClaimCounts.claimed}
                        unclaimedCount={websiteClaimCounts.unclaimed}
                    />
                {:else if loading}
                    <div class="index-tags-aside-skeleton" aria-busy="true" aria-label="Címkék betöltése">
                        <div class="index-tags-aside-skeleton__row">
                            {#each Array(8) as _}
                                <span class="skeleton index-tags-aside-skeleton__chip"></span>
                            {/each}
                        </div>
                        <div class="index-tags-aside-skeleton__row">
                            {#each Array(6) as _}
                                <span class="skeleton index-tags-aside-skeleton__chip"></span>
                            {/each}
                        </div>
                    </div>
                {:else if error}
                    <p class="index-tags-aside__empty">Nem sikerült betölteni a címkéket.</p>
                {:else}
                    <IndexTagAside
                        bind:selectedTypeKey
                        bind:selectedTagKey
                        bind:claimFilter={serviceClaimFilter}
                        claimedCount={serviceClaimCounts.claimed}
                        unclaimedCount={serviceClaimCounts.unclaimed}
                        entries={viewEntries}
                    />
                {/if}
            </div>
        </aside>
    </div>
    {@render indexFilterBar(kind, "end")}
</section>
{/snippet}

{#if indexView !== "websites"}
    {@render directoryBlock("services")}
{/if}
{#if indexView !== "services"}
    {@render directoryBlock("websites")}
{/if}


{#if indexView !== "websites"}
<section class="index-stats-section">
    <h2 class="index-stats-section-title">Index statisztikák:</h2>
    <div class="index-stats-items">
        <div class="card index-stats-item">
            <div class="index-stats-item-content">
                <div class="index-stats-item-icon"><svg xmlns="http://www.w3.org/2000/svg" height="50px" viewBox="0 -960 960 960" width="50px" fill="currentColor"><path d="m599-538 138-138-35-34-103 104-53-54-35 35 88 87ZM143-192v-72h432v72H143Zm345-296q-56-56-56-136t56-136q56-56 136-56t136 56q56 56 56 136t-56 136q-56 56-136 56t-136-56Zm-345-4v-72h224q5 20 12 37.5t17 34.5H143Zm0 150v-72h322q24 18 51.5 30.5T575-365v23H143Z"/></svg></div>
                <span class="index-stats-item-value">{totalCount}</span>
                <h3 class="index-stats-item-title">Ellenőrzött Bejegyzés</h3>
            </div>
        </div>
        <div class="card index-stats-item">
            <div class="index-stats-item-content">
                <div class="index-stats-item-icon"><AppIcon name="entry_categories" size={50} /></div>
                <span class="index-stats-item-value">10</span>
                <h3 class="index-stats-item-title">Kategória</h3>
            </div>
        </div>
        <div class="card index-stats-item">
            <div class="index-stats-item-content">
                <div class="index-stats-item-icon">
                    <svg xmlns="http://www.w3.org/2000/svg" height="50px" viewBox="0 -960 960 960" width="50px" fill="currentColor"><path d="M216-144q-29.7 0-50.85-21.15Q144-186.3 144-216v-528q0-29.7 21.15-50.85Q186.3-816 216-816h528q29.7 0 50.85 21.15Q816-773.7 816-744v258q-17.1-5.76-35.1-9.92T744-502v-242H216v528h241q1.88 19.52 5.94 37.26Q467-161 473-144H216Zm0-96v24-528 242-2 264Zm72-48h172q4-19 10.19-36.97Q476.38-342.93 484-360H288v72Zm0-156h264q26-20 56-34.5t64-20.5v-17H288v72Zm0-156h384v-72H288v72ZM719.77-48Q640-48 584-104.23q-56-56.22-56-136Q528-320 584.23-376q56.22-56 136-56Q800-432 856-375.77q56 56.22 56 136Q912-160 855.77-104q-56.22 56-136 56ZM696-144h48v-72h72v-48h-72v-72h-48v72h-72v48h72v72Z"/></svg>
                </div>
                <span class="index-stats-item-value">50000</span>
                <h3 class="index-stats-item-title">Havi felhasználó</h3>
            </div>
        </div>
    </div>
</section>
{/if}
<style>
    .index-directory + .index-directory {
        margin-top: 2.5rem;
    }

    .index-directory__title {
        margin: 0 0 0.75rem;
        font-size: var(--text-lg);
        color: var(--text-secondary);
    }

    .index-heading {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 1.5rem 2rem;
    }

    .index-heading__main {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: 0.35rem;
        min-width: 0;
        flex: 1 1 auto;
    }

    .index-heading__titles {
        display: flex;
        align-items: baseline;
        flex-wrap: wrap;
        gap: 0.35rem 1.25rem;
        min-width: 0;
    }

    .index-heading__titles .page-title {
        margin: 0;
    }

    .index-heading__titles .page-title a {
        color: inherit;
        text-decoration: none;
    }

    .index-heading__link {
        font-size: var(--text-xl);
        font-weight: 600;
        color: var(--text-muted);
        text-decoration: none;
    }

    .index-heading__link.active {
        color: var(--szekely-green);
    }

    .index-heading__greeting {
        margin: 0.35rem 0 0;
    }

    .index-heading__add {
        flex: 0 0 auto;
        margin-left: auto;
        display: flex;
        flex-direction: column;
        align-items: flex-end;
        gap: 0.45rem;
        max-width: 28rem;
        text-align: right;
    }

    .index-heading__add p {
        margin: 0;
        color: var(--text-muted);
        font-size: var(--text-sm);
        line-height: 1.4;
    }

.index-stats-items{
    display: flex;
    gap: 1rem;
    flex-direction: row;
    justify-content: space-around;
}

.index-stats-item{
    border: none;
    box-shadow: none;
    background: none;
}

.index-stats-section-title{
    text-align: center;
}

.index-stats-item-content{
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
}

.index-stats-item-value{
    font-weight: 700;
    color: var(--text-primary);
    line-height: 1;
    margin: 1rem 0 0;
}
    .index-tags-aside-skeleton {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        padding: 0.25rem 0 0.5rem;
    }

    .index-tags-aside-skeleton__row {
        display: flex;
        flex-wrap: wrap;
        gap: 0.4rem;
    }

    .index-tags-aside-skeleton__chip {
        display: inline-block;
        height: 1.85rem;
        border-radius: 999px;
        min-width: 3.5rem;
    }

    .index-tags-aside-skeleton__row :nth-child(3n) {
        min-width: 4.75rem;
    }

    .index-tags-aside-skeleton__row :nth-child(4n) {
        min-width: 5.25rem;
    }
</style>