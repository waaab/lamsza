<script>
    import { browser } from "$app/environment";
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import { onMount } from "svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";
    import WebsiteCard from "$lib/components/WebsiteCard.svelte";
    import ChipScrollRow from "$lib/components/ChipScrollRow.svelte";
    import PageHeader from "$lib/components/PageHeader.svelte";
    import IndexTagAside from "$lib/components/IndexTagAside.svelte";
    import {
        entryMatchesAsideFilters,
        normalizeTagKey,
        displayTagLabel,
    } from "$lib/directoryTagCloud.js";
    import {
        entryTypeLabelFromKey,
    } from "$lib/entryType.js";
    import {
        canonicalEntryCategory,
        DIRECTORY_CATALOG,
        directoryCatalogFromApi,
        directoryCategoryTabs,
        directoryChildTabs,
        entryMatchesCategory,
        parentCategoryTabActive,
    } from "$lib/entryCategory.js";
    import { apiFetch } from "$lib/api.js";
    import { listingAnchor, sortDirectoryEntries, sortWebsites } from "$lib/directoryListingOrder.js";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";
    import { auth } from "$lib/stores/auth";

    let pageHeader = initialPageHeader("index");
    let pageHeaderLoading = false;

    let dynamicCategories = [{ id: "osszes", label: "Összes", url: "/index" }];
    let catalog = DIRECTORY_CATALOG;
    let childCategories = [];
    let entries = [];
    /** @type {Array<{ id: number, title: string, description: string, domain: string, url: string, category?: string, entry_id?: number }>} */
    let shelfWebsites = [];
    let loading = true;
    let error = null;

    /** @type {string | null} */
    let selectedTypeKey = null;
    /** @type {string | null} */
    let selectedTagKey = null;

    let viewMode = "grid";
    let currentCategory = "";
    let visibleCount = 12;
    let sortMode = "title";
    let sortOpen = false;
    let locationOrderOn = false;
    /** @type {{ slug: string, name: string, county_slug: string } | null} */
    let siteLocation = null;
    /** @type {Array<Record<string, any>>} */
    let locations = [];

    const sortLabels = { title: "Név (A→Z)", newest: "Legújabb" };

    $: listingLocation = listingAnchor(siteLocation, $auth.preferredLocation, $auth.loggedIn);

    function setSortMode(mode) {
        sortMode = mode;
        sortOpen = false;
    }

    function scrollToTop() {
        if (typeof window !== "undefined") {
            window.scrollTo({ top: 0, behavior: "smooth" });
        }
    }

    /** @param {{ id: string, url: string }} cat */
    function openCategory(cat) {
        if (cat.id === currentCategory) {
            goto("/index");
            return;
        }
        goto(cat.url);
    }

    function clearAllFilters() {
        selectedTypeKey = null;
        selectedTagKey = null;
        locationOrderOn = false;
        goto("/index");
    }

    function toggleLocationOrder() {
        locationOrderOn = !locationOrderOn;
        sortOpen = false;
        scrollToTop();
    }

    $: filteredEntries = entries.filter((e) =>
        entryMatchesAsideFilters(e, selectedTypeKey, selectedTagKey),
    );

    $: categoryFilterLabel =
        catalog.find((row) => row.slug === currentCategory)?.name ||
        canonicalEntryCategory(currentCategory) ||
        currentCategory;

    $: typeFilterLabel = entryTypeLabelFromKey(selectedTypeKey) || null;

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
    $: sortedEntries = sortDirectoryEntries(filteredEntries, {
        sortMode,
        location: locationOrderOn ? listingLocation : null,
        locations,
    });
    $: sortedWebsites = sortWebsites(shelfWebsites, sortMode);
    $: totalCount = sortedEntries.length + sortedWebsites.length;
    $: displayItems = sortedEntries.slice(0, visibleCount);
    $: displayWebsites = sortedWebsites.slice(
        0,
        Math.max(0, visibleCount - sortedEntries.length),
    );

    function loadMore() {
        visibleCount += 12;
    }

    $: {
        currentCategory;
        selectedTypeKey;
        selectedTagKey;
        locationOrderOn;
        visibleCount = 12;
    }

    $: if (browser) {
        const categoryId = $page.params.category;
        currentCategory = categoryId;
        fetchData(categoryId);
    }

    onMount(() => {
        loadPageMeta("index").then((p) => {
            pageHeader = p;
            pageHeaderLoading = false;
        });
        apiFetch("/api/config/public")
            .then((config) => {
                if (config?.my_location_slug) {
                    siteLocation = {
                        slug: config.my_location_slug,
                        name: config.my_location_name || config.my_location_slug,
                        county_slug: config.my_location_county_slug || "",
                    };
                }
            })
            .catch(() => {
                siteLocation = null;
            });
        apiFetch("/api/locations")
            .then((locs) => {
                locations = Array.isArray(locs) ? locs : [];
            })
            .catch(() => {
                locations = [];
            });
    });

    async function fetchData(categoryId) {
        loading = true;
        error = null;
        try {
            const [allEntries, websitesData, categoryRows] = await Promise.all([
                apiFetch("/api/directory"),
                apiFetch("/api/websites"),
                apiFetch("/api/entry-categories"),
            ]);
            const websites = websitesData?.websites || [];
            if (Array.isArray(categoryRows) && categoryRows.length) {
                catalog = directoryCatalogFromApi(categoryRows);
            }
            dynamicCategories = directoryCategoryTabs(
                catalog,
                allEntries || [],
                websites,
            );
            childCategories = directoryChildTabs(
                categoryId,
                catalog,
                allEntries || [],
                websites,
            );
            entries = (allEntries || []).filter((e) =>
                entryMatchesCategory(e, categoryId, catalog),
            );
            shelfWebsites = websites.filter(
                (site) => !site.entry_id && entryMatchesCategory(site, categoryId, catalog),
            );
        } catch (err) {
            console.error(err);
            error = "Hiba történt az adatok betöltésekor.";
        } finally {
            loading = false;
        }
    }
</script>

<PageHeader
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    loading={pageHeaderLoading}
    breadcrumbLabel="Index"
    breadcrumbParentLabel=""
    breadcrumbParentUrl=""
    documentTitleSuffix=" - Lámsza"
/>

{#if loading}
    <div class="header-tabs chips">
        <span class="header-tabs-label" aria-label="Kiemelt kategóriák">Kiemelt kategóriák:</span>
        <span class="btn btn-md btn--loading">Szűrők betöltése…</span>
    </div>
{:else}
    <ChipScrollRow label="Kiemelt kategóriák:">
        {#each dynamicCategories as cat}
            <button
                class="btn btn-md {parentCategoryTabActive(cat.id, currentCategory, catalog)
                    ? 'active'
                    : ''}"
                on:click={() => openCategory(cat)}>{cat.label}</button
            >
        {/each}
    </ChipScrollRow>
    {#if childCategories.length > 0}
        <ChipScrollRow>
            {#each childCategories as cat}
                <button
                    class="btn btn-sm {cat.id === currentCategory ? 'active' : ''}"
                    on:click={() => openCategory(cat)}>{cat.label}</button
                >
            {/each}
        </ChipScrollRow>
    {/if}
{/if}

{#snippet categoryFilterBar()}
    <div class="filter-actions">
        <span class="info-box">
            <p>
                🔍 Szűrők:
                <span class="active">{categoryFilterLabel}</span>
                {#if typeFilterLabel}
                    <span class="filter-sep">·</span>
                    <span class="active">{typeFilterLabel}</span>
                {/if}
                {#if tagFilterLabel}
                    <span class="filter-sep">·</span>
                    <span class="active">{tagFilterLabel}</span>
                {/if}
                {#if locationOrderOn && listingLocation}
                    <span class="filter-sep">·</span>
                    <span class="active">{listingLocation.name}</span>
                {/if}
                <button
                    type="button"
                    class="clear-filters btn btn-xs"
                    aria-label="Szűrők törlése"
                    title="Szűrők törlése"
                    on:click={clearAllFilters}>Szűrő törlése</button
                >
            </p>
            <p>({displayItems.length + displayWebsites.length}/{totalCount})</p>
        </span>

        <div class="view-mode-toggle">
            {#if listingLocation}
                <button
                    type="button"
                    class="btn btn-sm"
                    class:active={locationOrderOn}
                    aria-pressed={locationOrderOn}
                    title={locationOrderOn
                        ? `${listingLocation.name}: először itt, aztán a környék, majd távolabb. Név szerint minden sávban.`
                        : `Közelség szerint: ${listingLocation.name}`}
                    on:click={toggleLocationOrder}
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
                    <span>{listingLocation.name}</span>
                </button>
            {/if}
            <div class="sort-toggle">
                <button class="btn btn-sm" on:click={() => (sortOpen = !sortOpen)}>
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
                    <span>{sortLabels[sortMode]}</span>
                </button>
                {#if sortOpen}
                    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
                    <div class="sort-toggle-menu" on:click|stopPropagation>
                        <button
                            class:active={sortMode === "title"}
                            on:click={() => setSortMode("title")}>Név (A→Z)</button
                        >
                        <button
                            class:active={sortMode === "newest"}
                            on:click={() => setSortMode("newest")}>Legújabb</button
                        >
                    </div>
                {/if}
            </div>

            <button
                class="btn btn-sm {viewMode === 'grid' ? 'active' : ''}"
                on:click={() => (viewMode = "grid")}
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
                class="btn btn-sm {viewMode === 'flex' ? 'active' : ''}"
                on:click={() => (viewMode = "flex")}
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

{@render categoryFilterBar()}

<div class="list-page-layout">
    <section class="list">
        {#if loading}
            <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
                {#each Array(6) as _, i (i)}
                    <EntryCard placeholder layout={viewMode === "grid" ? "grid" : "list"} />
                {/each}
            </div>
        {:else if error}
            <span class="info-box error">
                <p>{error}</p>
            </span>
        {:else if entries.length === 0 && shelfWebsites.length === 0}
            <span class="info-box info">
                <p>Nincs megjeleníthető bejegyzés ebben a kategóriában.</p>
            </span>
        {:else if displayItems.length === 0 && displayWebsites.length === 0}
            <span class="info-box info"
                ><p>Nincs a szűrőknek megfelelő bejegyzés.</p></span
            >
        {:else}
            <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
                {#each displayItems as entry}
                    <EntryCard {entry} layout={viewMode === "grid" ? "grid" : "list"} />
                {/each}
                {#each displayWebsites as website (website.id)}
                    <WebsiteCard {website} />
                {/each}
            </div>
            {#if visibleCount < totalCount}
                <div class="load-more">
                    <button class="btn nav-btn" on:click={loadMore}
                        >Több betöltése ↓</button
                    >
                </div>
            {/if}
        {/if}
    </section>

    <aside class="sidebar index-tags-sidebar" aria-label="Címkék">
        <div class="sidebar-box">
            <div class="sidebar-header">
                <h4 class="sidebar-heading">Szolgáltatások</h4>
            </div>
            {#if error}
                <p class="index-tags-aside__empty">Nem sikerült betölteni a címkéket.</p>
            {/if}
            <IndexTagAside
                {loading}
                bind:selectedTypeKey
                bind:selectedTagKey
                {entries}
            />
        </div>
    </aside>
</div>

{@render categoryFilterBar()}

<style>

</style>