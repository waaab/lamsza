<script>
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import { onMount, onDestroy } from "svelte";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";
    import EventsWidget from "$lib/components/EventsWidget.svelte";
    import WeatherIcon from "$lib/components/WeatherIcon.svelte";

    const locType = "megye"; // Hardcoded type for this route
    let town = "";
    let displayTown = "";
    let displayTownRo = "";
    let displayTownDe = "";
    let displayZipCode = "";
    let displayTownCoordinates = "";
    let displayTownPopulation = "";
    let displayTownArea = "";
    let displayTownType = "";
    let displayTownCrest = "";

    let childSettlements = [];
    let entries = [];
    let loadingEntries = true;
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

    $: sortedEntries = [...entries].sort((a, b) => {
        if (sortMode === "newest") return b.id - a.id;
        return a.name.localeCompare(b.name);
    });
    $: totalCount = sortedEntries.length;
    $: displayItems = sortedEntries.slice(0, visibleCount);

    function loadMore() {
        visibleCount += 12;
    }

    let countyWeather = [];
    let weatherLoading = true;
    let weatherUpdatedAt = null;

    let newsItems = [];
    let newsLoading = true;
    let newsTickerIndex = 0;
    let newsTickerInterval = null;
    /** "emoji" | "svg" from admin setting */
    let weatherIconStyle = "emoji";

    $: countySeat = childSettlements.find((s) => s.is_county_seat);

    $: if (browser && $page.params.countySlug) {
        town = $page.params.countySlug.toLowerCase();
        displayTown = town.charAt(0).toUpperCase() + town.slice(1);
        displayTownRo = "";
        displayTownDe = "";
        fetchData();
    }

    async function fetchData() {
        loadingEntries = true;
        weatherLoading = true;
        newsLoading = true;

        const apiBase =
            import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

        // 0. Public config (weather icon style)
        try {
            const configRes = await fetch(`${apiBase}/api/config/public`);
            if (configRes.ok) {
                const config = await configRes.json();
                if (config.weather_icon_style === "svg") weatherIconStyle = "svg";
                else weatherIconStyle = "emoji";
            }
        } catch (_) {}

        // 1. Fetch Directory Entries
        try {
            const locRes = await fetch(
                `${apiBase}/api/locations?type=${encodeURIComponent(locType)}`,
            );
            if (locRes.ok) {
                const locations = await locRes.json();
                const locData = locations.find((l) => l.slug === town);
                if (locData) {
                    displayTown = locData.name;
                    displayTownRo = locData.name_ro;
                    displayTownDe = locData.name_de;
                    displayZipCode = locData.post_code || "";
                    displayTownCoordinates = locData.coordinates || "";
                    displayTownPopulation = locData.population || "";
                    displayTownArea = locData.area || "";
                    displayTownType = locData.type || "";
                    displayTownCrest = locData.crest || "";
                }
                childSettlements = locations
                    .filter(
                        (l) =>
                            l.county_slug === town &&
                            l.type.toLowerCase() !== "megye",
                    )
                    .sort((a, b) => a.name.localeCompare(b.name));
            }
            const res = await fetch(
                `${apiBase}/api/directory?slug=${encodeURIComponent(town)}&type=megye`,
            );
            if (!res.ok) throw new Error("Hiba a címtár betöltésekor");
            entries = (await res.json()) || [];
        } catch (err) {
            console.error(err);
            entriesError = "Nem sikerült betölteni az adatokat.";
        } finally {
            loadingEntries = false;
        }

        // 2. Fetch Weather for all major cities in the county
        try {
            const wRes = await fetch(
                `${apiBase}/api/weather/county?slug=${encodeURIComponent(town)}`,
            );
            if (wRes.ok) {
                countyWeather = await wRes.json();
                weatherUpdatedAt = new Date();
            }
        } catch (err) {
            console.error("Weather fetch error:", err);
        } finally {
            weatherLoading = false;
        }

        // 3. Fetch News
        try {
            const nRes = await fetch(`${apiBase}/api/news?limit=10`);
            if (nRes.ok) {
                newsItems = (await nRes.json()).slice(0, 6);
            } else {
                newsItems = [];
            }
        } catch (err) {
            console.error("News fetch error:", err);
        } finally {
            newsLoading = false;
            startNewsTicker();
        }

    }

    function startNewsTicker() {
        if (newsTickerInterval) clearInterval(newsTickerInterval);
        newsTickerIndex = 0;
        if (newsItems.length <= 1) return;
        newsTickerInterval = setInterval(() => {
            newsTickerIndex = (newsTickerIndex + 1) % newsItems.length;
        }, 5000);
    }

    onDestroy(() => {
        if (newsTickerInterval) clearInterval(newsTickerInterval);
    });
</script>

<svelte:head>
    <title>{displayTown} Megye - Index</title>
</svelte:head>

<Breadcrumbs label={displayTown} type="Megye" />

<h1 class="page-title">{displayTown} Megye</h1>
<p class="greeting no-top-margin">
    Helyi hírek, időjárás és címtár {displayTown} megye területén.
</p>

<!-- Widgets: 3-column top grid (Áttekintés | Címer | Időjárás), then Events + News full width -->
<div class="widgets-box" id="hasznos-informaciok">
    <div id="attekintes">
        <h3 class="widget-title">Áttekintés</h3>
        <div class="more-info">
            <span title="Román neve">Románul: <span>{displayTownRo || "-"}</span></span>
            <span title="Német neve">Németül: <span>{displayTownDe || "-"}</span></span>
            <span title="Irányítószám">Irányítószám: <span>{displayZipCode || "-"}</span></span>
            <span title="Koordináták">Koordináták: <span>{displayTownCoordinates || "-"}</span></span>
            <span title="Lakosság">Lakosság: <span>{displayTownPopulation ? displayTownPopulation + " fő" : "-"}</span></span>
            <span title="Terület (négyzetkilométer)">Terület: <span>{displayTownArea ? displayTownArea + " km²" : "-"}</span></span>
            <span title="Közigazgatási forma">Közigazgatási forma: <span class="capitalize">{displayTownType || "-"}</span></span>
            <span title="Megyeszékhely">Megyeszékhely: <span>{#if countySeat}<a href="/{town}-megye/{countySeat.slug}" class="parent-city-link">{countySeat.name}</a>{:else}-{/if}</span></span>
        </div>
    </div>

    <div id="cimer" class="crest-card">
        {#if displayTownCrest && displayTownCrest !== "–" && displayTownCrest.length > 5}
            <h3 class="widget-title">{displayTown} címere</h3>
            <div class="crest-container">
                <img
                    src={`${import.meta.env.VITE_API_BASE_URL || "http://localhost:3000"}/api/proxy?url=${encodeURIComponent(displayTownCrest)}`}
                    alt="{displayTown} címere"
                    class="crest-img"
                />
            </div>
        {/if}
    </div>

    <!-- Weather = 3rd grid element (same card structure as homepage) -->
    <div id="idojaras" class="weather-card simple">
        <h3 class="widget-title">Időjárás a megyében</h3>
        <div class="widget-content">
            {#if weatherLoading}
                <div class="weather-left">
                    <div class="county-weather-flex">
                        {#each Array(3) as _}
                            <div class="card sm">
                                <span class="cw-name">adat betöltés...</span>
                                <span class="cw-temp-row">
                                    <span class="cw-temp">...°C</span>
                                    <span class="cw-temp-min">/ ...°C</span>
                                </span>
                                <span class="cw-desc">adat betöltés...</span>
                            </div>
                        {/each}
                    </div>
                </div>
                <div class="weather-right"></div>
            {:else if countyWeather.length > 0}
                <div class="weather-left">
                    <div class="county-weather-flex">
                        {#each countyWeather as cw}
                            <a href="/{town}-megye/{cw.slug}" class="card sm">
                                <span class="cw-name">{cw.city}</span>
                                <span class="cw-temp-row">
                                    <span class="cw-temp">{Math.round(cw.temp)}°C</span>
                                    <span class="cw-temp-min">/ {Math.round(cw.temp_min)}°C</span>
                                    <span class="cw-icon" aria-hidden="true"><WeatherIcon code={cw.icon} style={weatherIconStyle} /></span>
                                </span>
                                <span class="cw-desc capitalize">{cw.desc}</span>
                            </a>
                        {/each}
                    </div>
                </div>
                <div class="weather-right"></div>
            {:else}
                <div class="weather-left">
                    <p class="weather-error">Időjárás adat nem elérhető.</p>
                </div>
                <div class="weather-right"></div>
            {/if}
        </div>
        <div class="weather-footer">
            {#if weatherUpdatedAt}
                <small class="weather-source">Utoljára frissítve: {weatherUpdatedAt.toLocaleTimeString("hu-HU", { hour: "2-digit", minute: "2-digit" })}</small>
            {/if}
            <small class="weather-source" title="Forrás: OpenWeatherMap">OpenWeatherMap</small>
        </div>
    </div>

    <EventsWidget countySlug={town} locationName={displayTown} />

    <article id="hirek" class="news-widget component-box">
        <h3 class="widget-title">Helyi hírek</h3>
        {#if newsLoading}
            <div class="news-loading-placeholder">
                <span class="news-title">adat betöltés...</span>
                <div class="news-meta">adat betöltés...</div>
            </div>
        {:else if newsItems.length > 0}
            <div class="news-ticker">
                {#key newsTickerIndex}
                    {@const item = newsItems[newsTickerIndex]}
                    {#if item}
                        <div class="news-ticker-item">
                            <a
                                href={item.link}
                                target="_blank"
                                rel="nofollow noopener"
                                class="news-title"
                            >
                                📰 {item.title}
                            </a>
                            <div class="news-meta">
                                {item.source} - {new Date(item.pubDate).toLocaleDateString("hu-HU")}
                            </div>
                        </div>
                    {/if}
                {/key}
                <div class="news-ticker-nav">
                    <button class="scroll-arrow left" on:click={() => { newsTickerIndex = (newsTickerIndex - 1 + newsItems.length) % newsItems.length; }} aria-label="Előző hír">&#8249;</button>
                    <button class="scroll-arrow right" on:click={() => { newsTickerIndex = (newsTickerIndex + 1) % newsItems.length; }} aria-label="Következő hír">&#8250;</button>
                    <a href="/hirek" class="nav-btn">Összes hír</a>
                </div>
            </div>
        {:else}
            <span class="info-box"><p>Helyi hírek nem elérhetőek.</p></span>
        {/if}
    </article>
</div>

<!-- County Settlements Aside -->
{#if childSettlements.length > 0}
    <aside class="settlements-aside component-box">
        <h2 class="aside-title">
            Települések {displayTown} megyében
        </h2>
        <div class="settlements-grid">
            {#each childSettlements as child (child.id)}
                <a
                    href="/{$page.params.countySlug}-megye/{child.slug}"
                    class="badge settlement-badge"
                >
                    {child.name}
                </a>
            {/each}
        </div>
    </aside>
{/if}

<!-- Directory Section -->
<h2>
    {displayTown} megyei címtár - Helyi Index
</h2>

{#if loadingEntries}
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
        Nincs megjeleníthető bejegyzés {displayTown} megye területén.
    </p>
    </span>
{:else}
    <div class="filter-actions">
        <span class="info-box">
            <p>💡 Összesen:</p>
            <p><span>({displayItems.length}/{totalCount})</span></p>
        </span>

        <div class="view-mode-toggle">
            <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
            <div class="sort-toggle">
                <button
                    class="btn bnt-sm"
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
                    ></rect><rect x="3" y="14" width="7" height="7"></rect></svg
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

    <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
        {#each displayItems as entry}
            <EntryCard {entry} />
        {/each}
    </div>

    {#if visibleCount < totalCount}
        <div class="load-more">
            <button class="nav-btn" on:click={loadMore}>
                Több betöltése ↓
            </button>
        </div>
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
    .weather-error {
        color: var(--text-faint);
        margin: 0.5rem 0 0;
    }
    .news-widget {
        display: flex;
        flex-direction: column;
    }
    .news-loading-placeholder {
        padding: 0.5rem 0;
    }
    .news-widget {
        grid-column: span 3;
    }
    .news-ticker {
        display: flex;
        justify-content: space-between;
        gap: 0.5rem;
    }
    .news-ticker-item {
        animation: ticker-slide-in 0.35s ease-out;
    }
    .news-ticker-item a:hover {
        color: var(--szekely-red);
    }
    @keyframes ticker-slide-in {
        from { opacity: 0; transform: translateY(8px); }
        to { opacity: 1; transform: translateY(0); }
    }
    .news-ticker-nav {
        display: flex;
        align-items: center;
        gap: 0.4rem;
    }
    .news-ticker-nav :global(.scroll-arrow) {
        position: static;
        transform: none;
        opacity: 1;
        pointer-events: auto;
    }
    .news-title {
        text-decoration: none;
        color: inherit;
        font-weight: 500;
        font-size: 1rem;
    }
    .news-meta {
        font-size: 0.8em;
        color: var(--text-faint);
        margin-top: 0.2rem;
    }
    .component-box {
        padding: 1.5rem;
        background: var(--card-bg);
        border-radius: 12px;
        border: 1px solid var(--border-color);
    }
    .settlements-aside {
        margin-bottom: 2rem;
    }
    .aside-title {
        margin-top: 0;
    }
    .settlements-grid {
        display: flex;
        flex-wrap: wrap;
        gap: 0.8rem;
    }
    .settlement-badge {
        text-decoration: none;
        color: var(--primary-color);
        background: var(--bg-body);
        font-weight: 500;
        padding: 0.5rem 1rem;
        border: 1px solid var(--border-color);
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

    .county-weather-flex {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 0.5rem;
    }
    .cw-name {
        font-size: 0.9rem;
        line-height: 1.2;
    }
    .cw-temp-row {
        display: flex;
        gap: 0.15rem;
        margin: 0.2rem 0;
        align-items: flex-end;
    }
    .cw-temp {
        font-size: 1.15rem;
        font-weight: 700;
        line-height: 1;
    }
    .cw-temp-min {
        font-size: 0.75rem;
        color: var(--text-faint, #999);
        font-weight: 400;
    }
    .cw-icon {
        font-size: 1.4rem;
        line-height: 1;
        margin-left: auto;
    }
    .cw-desc {
        font-size: 0.75rem;
        color: var(--text-faint, #666);
        font-style: italic;
    }

    .crest-card {
        display: flex;
        flex-direction: column;
        align-items: center;
    }
    .crest-container {
        min-height: 120px;
    }
    .crest-img {
        max-width: 100%;
        max-height: 180px;
        object-fit: contain;
    }
</style>
