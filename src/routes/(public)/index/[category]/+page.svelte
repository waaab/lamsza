<script>
    import { browser } from "$app/environment";
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import { onMount } from "svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import IndexTagAside from "$lib/components/IndexTagAside.svelte";
    import {
        entryMatchesAsideFilters,
        normalizeTagKey,
        displayTagLabel,
    } from "$lib/directoryTagCloud.js";
    import {
        canonicalEntryType,
        canonicalEntryTypeKey,
    } from "$lib/entryType.js";
    import {
        canonicalEntryCategory,
        directoryCategoryTabs,
        entryMatchesCategory,
    } from "$lib/entryCategory.js";
    import { apiFetch } from "$lib/api.js";
    import { listingAnchor, sortDirectoryEntries } from "$lib/directoryListingOrder.js";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";
    import { auth } from "$lib/stores/auth";

    let pageHeader = initialPageHeader("index");
    let pageHeaderLoading = false;

    let dynamicCategories = [{ id: "osszes", label: "Összes", url: "/index" }];
    let entries = [];
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
        canonicalEntryCategory(currentCategory) || currentCategory;

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
    $: sortedEntries = sortDirectoryEntries(filteredEntries, {
        sortMode,
        location: locationOrderOn ? listingLocation : null,
        locations,
    });
    $: totalCount = sortedEntries.length;
    $: displayItems = sortedEntries.slice(0, visibleCount);

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
            const allEntries = (await apiFetch("/api/directory")) || [];
            dynamicCategories = directoryCategoryTabs(allEntries);
            entries = allEntries.filter((e) =>
                entryMatchesCategory(e, categoryId),
            );
        } catch (err) {
            console.error(err);
            error = "Hiba történt az adatok betöltésekor.";
        } finally {
            loading = false;
        }
    }
</script>

<PublicPageHero
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    loading={pageHeaderLoading}
    breadcrumbLabel="Index"
    breadcrumbParentLabel=""
    breadcrumbParentUrl=""
    documentTitleSuffix=" - Székely Gugel"
/>

<div class="header-tabs">
    <span class="header-tabs-label">Kiemelt Kategóriák:</span>
    {#if loading}
        <span class="btn btn-md" style="opacity:0.5">adat betöltés...</span>
    {:else}
        {#each dynamicCategories as cat}
            <button
                class="btn btn-md {cat.id === currentCategory ||
                cat.id.toLowerCase() === currentCategory.toLowerCase()
                    ? 'active'
                    : ''}"
                on:click={() => goto(cat.url)}>{cat.label}</button
            >
        {/each}
    {/if}
</div>

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
            <p>({displayItems.length}/{totalCount})</p>
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
        {:else if entries.length === 0}
            <span class="info-box info">
                <p>Nincs megjeleníthető bejegyzés ebben a kategóriában.</p>
            </span>
        {:else if displayItems.length === 0}
            <span class="info-box info"
                ><p>Nincs a szűrőknek megfelelő bejegyzés.</p></span
            >
        {:else}
            <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
                {#each displayItems as entry}
                    <EntryCard {entry} layout={viewMode === "grid" ? "grid" : "list"} />
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
                <h4 class="sidebar-heading">Címkék</h4>
            </div>
            {#if loading}
                <div
                    class="index-tags-aside-skeleton"
                    aria-busy="true"
                    aria-label="Címkék betöltése"
                >
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
                    {entries}
                />
            {/if}
        </div>
    </aside>
</div>

{@render categoryFilterBar()}

<style>
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