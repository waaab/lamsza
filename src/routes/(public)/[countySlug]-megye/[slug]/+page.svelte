<script>
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import { onMount } from "svelte";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";
    import WeatherWidget from "$lib/components/WeatherWidget.svelte";
    import NewsWidget from "$lib/components/NewsWidget.svelte";
    import EventsWidget from "$lib/components/EventsWidget.svelte";
    import { apiFetch } from "$lib/api";

    let settlementData = null;
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

    onMount(async () => {
        if (browser && town) {
            fetchData();
        }
    });

    async function fetchData() {
        loading = true;
        try {
            // 1. Fetch Settlement Info
            const locations = await apiFetch("/api/locations");
            const locData = locations.find(
                (l) => l.slug === town.toLowerCase(),
            );
            if (locData) {
                settlementData = locData;
                if (locData.parent_id) {
                    const parent = locations.find(
                        (l) => l.id === locData.parent_id,
                    );
                    if (parent) {
                        settlementData.parent = parent;
                    }
                }
            }

            // 2. Fetch Directory Entries
            const res = await apiFetch(
                `/api/directory?location_slug=${encodeURIComponent(town)}`,
            );
            entries = res || [];
        } catch (err) {
            console.error(err);
            entriesError = "Nem sikerült betölteni az adatokat.";
        } finally {
            loading = false;
        }
    }
</script>

<svelte:head>
    <title>{settlementData?.name || town} - Index</title>
</svelte:head>

{#if settlementData}
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
    <p class="greeting no-top-margin">
        Helyi hírek, időjárás és címtár {settlementData.name} területén.
    </p>

    <div class="widgets-box">
        <!-- Settlement info -->
        <article id="attekintes">
            <h3 class="widget-title">Áttekintés</h3>
            <div class="more-info">
                {#if settlementData.name_ro}<span
                        >Románul: {settlementData.name_ro}</span
                    >{/if}
                {#if settlementData.name_de}<span
                        >Németül: {settlementData.name_de}</span
                    >{/if}
                <span
                    >Irányítószám: <span>{settlementData.post_code || "–"}</span
                    ></span
                >
                <span
                    >Koordináták: <span
                        >{settlementData.coordinates || "–"}</span
                    ></span
                >
                <span
                    >Lakosság: <span>{settlementData.population || "–"} fő</span
                    ></span
                >
                <span
                    >Terület: <span>{settlementData.area || "–"} km²</span
                    ></span
                >
                <span
                    >Közigazgatási forma: <span class="capitalize"
                        >{settlementData.type || "–"}</span
                    ></span
                >
                {#if settlementData.parent}
                    <span
                        >Kapcsolódó település: <a
                            href="/{settlementData.parent
                                .county_slug}-megye/{settlementData.parent
                                .slug}"
                            class="parent-city-link"
                            >{settlementData.parent.name}</a
                        ></span
                    >
                {/if}
            </div>
        </article>

        <!-- Coat of Arms -->
        <article id="cimer" class="crest-card">
            <h3 class="widget-title">{settlementData.name} címere</h3>
            <div class="crest-container">
                {#if settlementData.crest && settlementData.crest !== "–" && settlementData.crest.length > 5}
                    <img
                        src={`${import.meta.env.VITE_API_BASE_URL || "http://localhost:3000"}/api/proxy?url=${encodeURIComponent(settlementData.crest)}`}
                        alt="{settlementData.name} címere"
                        class="crest-img"
                    />
                {:else}
                    <span class="no-crest">Nincs elérhető címer</span>
                {/if}
            </div>
        </article>

        <WeatherWidget settlementSlug={town} />

        <EventsWidget settlementSlug={town} />

        <NewsWidget settlementSlug={town} />
    </div>

    <h2>{settlementData.name}i címtár - Helyi Index</h2>

    {#if loading}
        <div class="list grid">
            {#each Array(6) as _}
                <article class="card entry--skeleton">
                    <div class="skeleton skeleton-text skeleton-30"></div>
                    <div class="skeleton skeleton-text skeleton-80-top"></div>
                    <div
                        class="skeleton skeleton-text skeleton-60-bottom"
                    ></div>
                </article>
            {/each}
        </div>
    {:else if entriesError}
        <div class="note error">{entriesError}</div>
    {:else if entries.length === 0}
        <div class="note error">
            Nincs megjeleníthető bejegyzés {settlementData.name} területén.
        </div>
    {:else}
        <div class="filter-actions">
            <div class="info-box">
                <p>💡 Összesen:</p>
                <p><span>({displayItems.length}/{totalCount})</span></p>
            </div>

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
{/if}

<style>
    .no-top-margin {
        margin-top: 0;
    }
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

    .entry--skeleton {
        height: 150px;
        display: flex;
        flex-direction: column;
        padding: 1rem;
        gap: 0.5rem;
    }
    .skeleton-30 {
        width: 30%;
    }
    .skeleton-80-top {
        width: 80%;
        margin-top: 0.5rem;
        height: 1.2rem;
    }
    .skeleton-60-bottom {
        width: 60%;
        margin-top: auto;
    }
</style>
