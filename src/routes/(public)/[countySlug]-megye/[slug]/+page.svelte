<script>
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";
    import WeatherWidget from "$lib/components/WeatherWidget.svelte";
    import NewsWidget from "$lib/components/NewsWidget.svelte";
    import EventsWidget from "$lib/components/EventsWidget.svelte";
    import { apiFetch } from "$lib/api";
    import Markdown from "$lib/components/Markdown.svelte";

    let settlementData = null;
    let attractionData = null;
    let entries = [];
    let loading = true;
    let entriesError = null;

    let viewMode = "grid";
    let sortMode = "title";
    let visibleCount = 12;
    let sortOpen = false;

    const sortLabels = { title: "Név (A→Z)", newest: "Legújabb" };

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

    async function fetchData() {
        loading = true;
        settlementData = null;
        attractionData = null;
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

    <h1 class="page-title">{attractionData.name}</h1>
    <p class="greeting">
        Látnivaló {attractionData.county_name} megyében.
    </p>

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

    {#if attractionData.featured_image}
        <div class="attraction-featured">
            <img
                src={`${import.meta.env.VITE_API_BASE_URL || "http://localhost:3000"}/api/proxy?url=${encodeURIComponent(attractionData.featured_image)}`}
                alt={attractionData.name}
            />
        </div>
    {/if}

    {#if attractionData.description}
        <p class="attraction-desc">{attractionData.description}</p>
    {/if}

    {#if attractionData.content}
        <div class="attraction-content">
            <Markdown source={attractionData.content} />
        </div>
    {/if}

    {#if attractionData.images && attractionData.images.length > 0}
        <h3 class="widget-title">Galéria</h3>
        <div class="attraction-gallery">
            {#each attractionData.images as url}
                <a
                    href={`${import.meta.env.VITE_API_BASE_URL || "http://localhost:3000"}/api/proxy?url=${encodeURIComponent(url)}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    class="gallery-item"
                >
                    <img
                        src={`${import.meta.env.VITE_API_BASE_URL || "http://localhost:3000"}/api/proxy?url=${encodeURIComponent(url)}`}
                        alt=""
                    />
                </a>
            {/each}
        </div>
    {/if}
{:else if settlementData}
    <Breadcrumbs
        label={settlementData.name}
        settlementType={settlementData.type}
        countyName={settlementData.county}
        countySlug={$page.params.countySlug}
    />

    <h1 class="page-title">
        {settlementData.name}
        {settlementData.type} és környéke
    </h1>
    <p class="greeting">
        Helyi hírek, időjárás és címtár {settlementData.name} területén.
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
            {#if settlementData.crest && settlementData.crest !== "–" && settlementData.crest.length > 5}
                <div class="widget-header">
                    <h3 class="widget-title">{settlementData.name} címere</h3>
                </div>
                <div class="crest-container">
                    <img
                        src={`${import.meta.env.VITE_API_BASE_URL || "http://localhost:3000"}/api/proxy?url=${encodeURIComponent(settlementData.crest)}`}
                        alt="{settlementData.name} címere"
                        class="crest-img"
                    />
                </div>
            {/if}
        </div>

        <WeatherWidget settlementSlug={town} advanced={true} />
    </div>

    <EventsWidget settlementSlug={town} locationName={settlementData.name} />

    <NewsWidget settlementSlug={town} ticker={true} />

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
                <EntryCard {entry} showBadge={false} />
            {/each}
        </div>

        {#if visibleCount < totalCount}
            <div class="load-more">
                <button class="nav-btn" on:click={loadMore}
                    >Több betöltése ↓</button
                >
            </div>
        {/if}
    {/if}
{:else if loading}
    <p class="greeting">Betöltés...</p>
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
        font-size: 0.75rem;
        color: var(--text-faint);
    }
    .entry-placeholder-title {
        font-size: 0.95rem;
        color: var(--text-faint);
        margin-top: 0.5rem;
    }

    .attraction-featured {
        margin: 1.5rem 0;
        border-radius: 0.5rem;
        overflow: hidden;
    }
    .attraction-featured img {
        width: 100%;
        max-height: 300px;
        object-fit: cover;
    }
    .attraction-desc {
        margin: 1rem 0;
        color: var(--text-faint);
    }
    .attraction-content {
        margin: 1.5rem 0;
    }
    .attraction-gallery {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
        gap: 1rem;
        margin: 1rem 0;
    }
    .attraction-gallery .gallery-item img {
        width: 100%;
        height: 120px;
        object-fit: cover;
        border-radius: 0.25rem;
    }
</style>
