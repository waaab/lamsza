<script>
    import { onMount } from "svelte";
    import { get } from "svelte/store";
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import { auth } from "$lib/stores/auth";
    import { openLogin } from "$lib/openLogin.js";
    import FavoriteButton from "$lib/components/FavoriteButton.svelte";
    import {
        addFavorite,
        favoriteListWithToggle,
        isFavorite,
        loadFavoriteList,
        removeFavorite,
    } from "$lib/favorites.js";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";
    import WeatherWidget from "$lib/components/WeatherWidget.svelte";
    import NewsWidget from "$lib/components/NewsWidget.svelte";
    import EventsWidget from "$lib/components/EventsWidget.svelte";
    import { apiFetch, getApiBase } from "$lib/api";
    import Markdown from "$lib/components/Markdown.svelte";
    import CrestShieldPlaceholder from "$lib/components/CrestShieldPlaceholder.svelte";
    import EntryPhotoGallery from "$lib/components/EntryPhotoGallery.svelte";
    import { attractionGallerySlides } from "$lib/entryPhotos.js";
    import { kindLabel } from "$lib/venueKindLabels.js";

    let settlementData = null;
    let attractionData = null;
    let countyAttractions = [];
    /** @type {Record<string, unknown>[]} */
    let settlementVenues = [];
    let entries = [];
    let loading = true;
    let entriesError = null;
    /** @type {{ type: string, id: number }[]} */
    let favoriteList = [];

    async function initFavorites() {
        await auth.init();
        if (!get(auth).loggedIn) {
            favoriteList = [];
            return;
        }
        try {
            favoriteList = await loadFavoriteList();
        } catch {
            favoriteList = [];
        }
    }

    /** @param {string} type @param {number} id @param {boolean} currentlyActive */
    async function handleFavoriteToggle(type, id, currentlyActive) {
        if (!get(auth).loggedIn) {
            openLogin();
            return;
        }
        try {
            if (currentlyActive) {
                await removeFavorite(type, id);
                favoriteList = favoriteListWithToggle(
                    favoriteList,
                    type,
                    id,
                    false,
                );
            } else {
                await addFavorite(type, id);
                favoriteList = favoriteListWithToggle(
                    favoriteList,
                    type,
                    id,
                    true,
                );
            }
        } catch (err) {
            console.error(err);
        }
    }

    onMount(() => {
        initFavorites();
    });

    let viewMode = "grid";
    let sortMode = "title";
    let visibleCount = 12;
    let sortOpen = false;

    const sortLabels = { title: "Név (A→Z)", newest: "Legújabb" };

    /** @param {unknown} crest */
    function hasCrestUrl(crest) {
        const s = String(crest ?? "").trim();
        return s.length > 5 && s !== "–";
    }

    function setSortMode(mode) {
        sortMode = mode;
        sortOpen = false;
    }

    $: town = $page.params.slug;
    $: sortedEntries = [...entries].sort((a, b) => {
        if (sortMode === "newest") return b.id - a.id;
        return a.name.localeCompare(b.name);
    });
    $: totalCount = sortedEntries.length;
    $: displayItems = sortedEntries.slice(0, visibleCount);

    function loadMore() {
        visibleCount += 12;
    }

    $: if (browser && $page.params.slug) {
        fetchData();
    }

    $: pageTitle = settlementData?.name || attractionData?.name || town;
    $: isAttraction = !!attractionData;
    $: attractionSlides = attractionData
        ? attractionGallerySlides(attractionData, { apiBase: getApiBase() })
        : [];

    async function fetchData() {
        loading = true;
        settlementData = null;
        attractionData = null;
        countyAttractions = [];
        settlementVenues = [];
        entries = [];
        entriesError = null;
        try {
            const countySlug = $page.params.countySlug.toLowerCase();
            const slug = town.toLowerCase();

            // 1. Try settlement first (locations = counties + settlements)
            const locations = await apiFetch(
                `/api/locations?county_slug=${encodeURIComponent(countySlug)}`,
            );
            const locData = locations.find((l) => l.slug === slug && l.type !== "megye");
            if (locData) {
                settlementData = locData;
                if (locData.parent_id) {
                    const parent = locations.find((l) => l.id === locData.parent_id);
                    if (parent) settlementData.parent = parent;
                }
                const res = await apiFetch(
                    `/api/directory?location_slug=${encodeURIComponent(slug)}`,
                );
                entries = res || [];
                try {
                    countyAttractions = await apiFetch(
                        `/api/attractions?county_slug=${encodeURIComponent(countySlug)}`,
                    );
                    if (!Array.isArray(countyAttractions)) countyAttractions = [];
                } catch {
                    countyAttractions = [];
                }
                try {
                    const vlist = await apiFetch(
                        `/api/venues?settlement_id=${locData.id}`,
                    );
                    settlementVenues = Array.isArray(vlist) ? vlist : [];
                } catch {
                    settlementVenues = [];
                }
            } else {
                // 2. Try attraction
                const att = await apiFetch(
                    `/api/attractions?county_slug=${encodeURIComponent(countySlug)}&slug=${encodeURIComponent(slug)}`,
                );
                if (att && att.id) {
                    attractionData = att;
                }
            }
        } catch (err) {
            console.error(err);
            entriesError = "Nem sikerült betölteni az adatokat.";
        } finally {
            loading = false;
        }
    }
</script>

<svelte:head>
    <title>{pageTitle} - Index</title>
</svelte:head>

{#if attractionData}
    <Breadcrumbs
        label={attractionData.name}
        parentLabel={attractionData.county_name}
        parentUrl="/{$page.params.countySlug}-megye"
        countyName={attractionData.county_name}
        countySlug={$page.params.countySlug}
    />

    <div class="location-title-row">
        <h1 class="page-title">{attractionData.name}</h1>
        <FavoriteButton
            type="attraction"
            id={attractionData.id}
            active={isFavorite(favoriteList, "attraction", attractionData.id)}
            ontoggle={() =>
                handleFavoriteToggle(
                    "attraction",
                    attractionData.id,
                    isFavorite(favoriteList, "attraction", attractionData.id),
                )}
        />
    </div>
    <h2 class="greeting">
        Látnivaló {attractionData.county_name} megyében.
    </h2>

    <div class="widgets-box">
        <div id="attekintes" class="widget">
            <div class="widget-header">
                <h3 class="widget-title">Áttekintés</h3>
            </div>
            <div class="more-info">
                {#if attractionData.name_ro}
                    <span>Románul: <span>{attractionData.name_ro}</span></span>
                {/if}
                {#if attractionData.name_de}
                    <span>Németül: <span>{attractionData.name_de}</span></span>
                {/if}
                {#if attractionData.latitude && attractionData.longitude}
                    <span>Koordináták: <span>{attractionData.latitude.toFixed(4)}, {attractionData.longitude.toFixed(4)}</span></span>
                {/if}
                <span>Megye: <span><a href="/{$page.params.countySlug}-megye" class="parent-city-link">{attractionData.county_name}</a></span></span>
            </div>
        </div>

        <WeatherWidget
            settlementSlug={attractionData.latitude && attractionData.longitude ? undefined : town}
            lat={attractionData.latitude}
            lon={attractionData.longitude}
            advanced={true}
        />
    </div>

    {#if attractionSlides.length}
        <section class="attraction-photos" aria-labelledby="attraction-photos-title">
            <h3 id="attraction-photos-title" class="widget-title">Fotók</h3>
            <EntryPhotoGallery
                slides={attractionSlides}
                label={`${attractionData.name} fotói`}
            />
        </section>
    {/if}

    {#if attractionData.description}
        <p class="attraction-desc">{attractionData.description}</p>
    {/if}

    {#if attractionData.content}
        <div class="attraction-content">
            <Markdown source={attractionData.content} />
        </div>
    {/if}
{:else if settlementData}
    <Breadcrumbs
        label={settlementData.name}
        settlementType={settlementData.type}
        countyName={settlementData.county}
        countySlug={$page.params.countySlug}
    />

    <div class="location-title-row">
        <h1 class="page-title">
            {settlementData.name}
            {settlementData.type} és környéke
        </h1>
        <FavoriteButton
            type="settlement"
            id={settlementData.id}
            active={isFavorite(favoriteList, "settlement", settlementData.id)}
            ontoggle={() =>
                handleFavoriteToggle(
                    "settlement",
                    settlementData.id,
                    isFavorite(favoriteList, "settlement", settlementData.id),
                )}
        />
    </div>
    <p class="greeting">
        Helyi események, hírek, időjárás és címtár {settlementData.name} területén.
    </p>

    <div class="widgets-box">
        <div id="attekintes" class="widget">
            <div class="widget-header">
                <h3 class="widget-title">Áttekintés</h3>
            </div>
            <div class="more-info">
                <span>Románul: <span>{settlementData.name_ro || "-"}</span></span>
                <span>Németül: <span>{settlementData.name_de || "-"}</span></span>
                <span>Irányítószám: <span>{settlementData.post_code || "-"}</span></span>
                <span>Koordináták: <span>{settlementData.coordinates || "-"}</span></span>
                <span>Lakosság: <span>{settlementData.population ? settlementData.population + " fő" : "-"}</span></span>
                <span>Terület: <span>{settlementData.area ? settlementData.area + " km²" : "-"}</span></span>
                <span>Közigazgatási forma: <span class="capitalize">{settlementData.type || "-"}</span></span>
                <span>Kapcsolódó település: <span>{#if settlementData.parent}<a href="/{settlementData.parent.county_slug}-megye/{settlementData.parent.slug}" class="parent-city-link">{settlementData.parent.name}</a>{:else}-{/if}</span></span>
                <span>Megye: <span>{#if settlementData.county}<a href="/{settlementData.county_slug}-megye" class="parent-city-link">{settlementData.county}</a>{:else}-{/if}</span></span>
            </div>
        </div>

        <div id="cimer" class="crest-card widget">
            <div class="widget-header">
                <h3 class="widget-title">{settlementData.name} címere</h3>
            </div>
            <div class="crest-container">
                {#if hasCrestUrl(settlementData.crest)}
                    <img
                        src={`${getApiBase()}/api/proxy?url=${encodeURIComponent(settlementData.crest)}`}
                        alt="{settlementData.name} címere"
                        class="crest-img"
                    />
                {:else}
                    <CrestShieldPlaceholder
                        label="{settlementData.name} címere - nincs feltöltött kép, helyőrző pajzs"
                    />
                {/if}
            </div>
        </div>

        <WeatherWidget settlementSlug={town} advanced={true} />
    </div>

    <EventsWidget settlementSlug={town} locationName={settlementData.name} />

    {#if settlementVenues.length > 0}
        <section id="helyszinek" class="widget settlement-venues component-box"
            aria-label="Helyszínek"
        >
            <div class="widget-header" title="{settlementData.name}i helyszínek">
                <h3 class="widget-title">Helyszínek</h3>
            </div>

            <div class="widget-content">
                {#if loading}
                    <span class="info-box"><p>Betöltés...</p></span>
                {:else if settlementVenues.length === 0}
                    <span class="info-box"><p>Nincs megjeleníthető rendezvényhelyszín.</p></span>
                {:else}
                
                <ul class="settlements-grid">
                    {#each settlementVenues as ven (ven.id)}
                        <a
                            href="/{$page.params.countySlug}-megye/{town}/helyszin/{ven.slug}"
                            class="card sm settlement"
                        >
                            {ven.name}
                            {#if kindLabel(ven.kind, ven.kind_label)}
                                {' '}{kindLabel(ven.kind, ven.kind_label)}
                            {/if}
                        </a>
                    {/each}
                </ul>
                {/if}
            </div>
        </section>
    {/if}

    <NewsWidget settlementSlug={town} ticker={true} />

    {#if countyAttractions.length > 0}
        <section class="component-box">
            <h2 class="aside-title">
                Látnivalók {settlementData.county} megyében
            </h2>
            <div class="settlements-grid">
                {#each countyAttractions as att (att.id)}
                    <a
                        href="/{$page.params.countySlug}-megye/{att.slug}"
                        class="card sm settlement"
                    >
                        {att.name}
                    </a>
                {/each}
            </div>
        </section>
    {/if}

    <h2>{settlementData.name}i címtár - Helyi Index</h2>

    {#if loading}
        <div class="list grid">
            {#each Array(6) as _}
                <article class="card entry-placeholder">
                    <span class="entry-placeholder-cat">adat betöltés...</span>
                    <span class="entry-placeholder-title">adat betöltés...</span>
                    <span class="entry-placeholder-loc">adat betöltés...</span>
                </article>
            {/each}
        </div>
    {:else if entriesError}
        <span class="info-box error">
            <p>{entriesError}</p>
        </span>
    {:else if entries.length === 0}
        <span class="info-box error">
            <p>
                Nincs megjeleníthető bejegyzés {settlementData.name} területén.
            </p>
        </span>
    {:else}
        <div class="filter-actions">
            <span class="info-box">
                <p>💡 Összesen:</p>
                <p><span>({displayItems.length}/{totalCount})</span></p>
            </span>

            <div class="view-mode-toggle">
                <div class="sort-toggle">
                    <button
                        class="btn btn-sm"
                        on:click={() => (sortOpen = !sortOpen)}
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
                            ><line x1="4" y1="6" x2="16" y2="6"></line><line
                                x1="4"
                                y1="12"
                                x2="12"
                                y2="12"
                            ></line><line x1="4" y1="18" x2="8" y2="18"
                            ></line><polyline points="15 15 18 18 21 15"
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
                                on:click={() => setSortMode("title")}
                                >Név (A→Z)</button
                            >
                            <button
                                class:active={sortMode === "newest"}
                                on:click={() => setSortMode("newest")}
                                >Legújabb</button
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
                        ></rect><rect x="14" y="14" width="7" height="7"
                        ></rect><rect x="3" y="14" width="7" height="7"
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
                        ></line><line x1="8" y1="18" x2="21" y2="18"
                        ></line><line x1="3" y1="6" x2="3.01" y2="6"
                        ></line><line x1="3" y1="12" x2="3.01" y2="12"
                        ></line><line x1="3" y1="18" x2="3.01" y2="18"
                        ></line></svg
                    >
                    <span>Lista</span>
                </button>
            </div>
        </div>

        <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
            {#each displayItems as entry}
                <EntryCard {entry} showBadge={false} layout={viewMode === "grid" ? "grid" : "list"} />
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
{:else if loading}
    <p class="greeting">...</p>
{:else}
    <p class="greeting">A keresett oldal nem található.</p>
{/if}

<style>
    .more-info {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .capitalize {
        text-transform: capitalize;
    }

    .widgets-box {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 2rem;
        margin-bottom: 2rem;
    }
    :global(.news-widget),
    :global(.event-widget) {
        grid-column: span 3;
    }
    @media (max-width: 992px) {
        .widgets-box {
            grid-template-columns: 1fr;
        }
        :global(.news-widget),
        :global(.event-widget) {
            grid-column: span 1;
        }
    }

    .entry-placeholder {
        display: flex;
        flex-direction: column;
        padding: 1rem;
        gap: 0.5rem;
    }
    .entry-placeholder-cat,
    .entry-placeholder-loc {
        color: var(--text-faint);
    }
    .entry-placeholder-title {
        color: var(--text-faint);
        margin-top: 0.5rem;
    }

    .attraction-photos {
        margin: 1.5rem 0 1rem;
    }
    .attraction-desc {
        margin: 1rem 0;
        color: var(--text-faint);
    }
    .attraction-content {
        margin: 1.5rem 0;
    }

    .component-box {
        padding: 1.5rem;
        background: var(--card-bg);
        border-radius: 12px;
        border: 1px solid var(--border-color);
    }
    .aside-title {
        margin-top: 0;
    }
    .settlements-grid {
        display: flex;
        flex-wrap: wrap;
        gap: 0.8rem;
        margin: 0;
        padding: 0;
    }
    .settlement {
        background: var(--bg-body);
        font-weight: 500;
    }
    .location-title-row {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        gap: 1rem;
    }
    .location-title-row .page-title {
        margin: 0;
        flex: 1 1 auto;
    }
</style>