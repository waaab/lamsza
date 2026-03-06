<script>
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import { onMount } from "svelte";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";

    let locType = ""; // Dynamically populated from locations API
    let town = "";
    let displayTown = "";
    let displayTownRo = "";
    let displayTownDe = "";
    let countyName = "";
    let displayZipCode = "";
    let displayTownCoordinates = "";
    let displayTownPopulation = "";
    let displayTownArea = "";
    let displayTownCrest = "";
    let displayParentId = null;
    let displayParentName = "";
    let displayParentSlug = "";
    let displayParentCountySlug = "";

    let services = [];
    let loadingServices = true;
    let servicesError = null;

    let viewMode = "grid";
    let sortMode = "title";
    let visibleCount = 12;
    let sortOpen = false;

    const sortLabels = { title: "Név (A→Z)", newest: "Legújabb" };

    function setSortMode(mode) {
        sortMode = mode;
        sortOpen = false;
    }

    $: sortedServices = [...services].sort((a, b) => {
        if (sortMode === "newest") return b.id - a.id;
        return a.name.localeCompare(b.name);
    });
    $: totalCount = sortedServices.length;
    $: displayItems = sortedServices.slice(0, visibleCount);

    function loadMore() {
        visibleCount += 12;
    }

    let weatherData = null;
    let weatherLoading = true;

    let newsItems = [];
    let newsLoading = true;

    let localEvents = [];
    let eventsLoading = true;

    $: if (browser && $page.params.slug) {
        town = $page.params.slug.toLowerCase();
        displayTown = town.charAt(0).toUpperCase() + town.slice(1);
        displayTownRo = "";
        displayTownDe = "";
        fetchData();
    }

    async function fetchData() {
        loadingServices = true;
        weatherLoading = true;
        newsLoading = true;
        eventsLoading = true;

        const apiBase =
            import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

        // 1. Fetch Directory Services
        try {
            const locRes = await fetch(`${apiBase}/api/locations`);
            if (locRes.ok) {
                const locations = await locRes.json();
                const locData = locations.find((l) => l.slug === town);
                if (locData) {
                    displayTown = locData.name;
                    displayTownRo = locData.name_ro;
                    displayTownDe = locData.name_de;
                    locType = locData.type || "";
                    countyName = locData.county || "";
                    displayZipCode = locData.post_code || "";
                    displayTownCoordinates = locData.coordinates || "";
                    displayTownPopulation = locData.population || "";
                    displayTownArea = locData.area || "";
                    displayTownCrest = locData.crest || "";
                    displayParentId = locData.parent_id || null;
                    if (displayParentId) {
                        const parent = locations.find(
                            (l) => l.id === displayParentId,
                        );
                        if (parent) {
                            displayParentName = parent.name;
                            displayParentSlug = parent.slug;
                            displayParentCountySlug = parent.county_slug;
                        }
                    }
                }
            }
            const res = await fetch(
                `${apiBase}/api/directory?location_slug=${encodeURIComponent(town)}`,
            );
            if (!res.ok) throw new Error("Hiba a címtár betöltésekor");
            services = (await res.json()) || [];
        } catch (err) {
            console.error(err);
            servicesError = "Nem sikerült betölteni az adatokat.";
        } finally {
            loadingServices = false;
        }

        // 2. Fetch Weather
        const API_KEY = import.meta.env.VITE_WEATHER_API_KEY;
        if (API_KEY) {
            try {
                const wRes = await fetch(
                    `${apiBase}/api/weather?slug=${encodeURIComponent(town)}&appid=${API_KEY}`,
                );
                if (wRes.ok) {
                    const wData = await wRes.json();
                    weatherData = {
                        temp: Math.round(wData.main.temp),
                        tempMin: Math.round(wData.main.temp_min),
                        desc: wData.weather[0].description,
                        icon: wData.weather[0].icon,
                        timestamp: Date.now(),
                    };
                }
            } catch (err) {
                console.error("Weather fetch error:", err);
            } finally {
                weatherLoading = false;
            }
        } else {
            weatherLoading = false;
        }

        // 3. Fetch News filtering by town
        try {
            const nRes = await fetch(`${apiBase}/api/admin/news_feeds`);
            const dbFeeds = await nRes.json();
            let allNews = [];

            if (dbFeeds && dbFeeds.length > 0) {
                for (const feed of dbFeeds) {
                    if (!feed.feed_url) continue;
                    try {
                        const proxiedUrl =
                            `${apiBase}/api/proxy?url=` +
                            encodeURIComponent(feed.feed_url);
                        const response = await fetch(proxiedUrl);
                        const xmlText = await response.text();
                        const parser = new DOMParser();
                        const xmlDoc = parser.parseFromString(
                            xmlText,
                            "text/xml",
                        );
                        const nodes = Array.from(
                            xmlDoc.querySelectorAll("item"),
                        );

                        nodes.forEach((node) => {
                            const title =
                                node.querySelector("title")?.textContent || "";
                            const desc =
                                node.querySelector("description")
                                    ?.textContent || "";

                            // Check if town is mentioned
                            if (
                                title.toLowerCase().includes(town) ||
                                desc.toLowerCase().includes(town)
                            ) {
                                allNews.push({
                                    title: title,
                                    link:
                                        node.querySelector("link")
                                            ?.textContent || "#",
                                    pubDate:
                                        new Date(
                                            node.querySelector(
                                                "pubDate",
                                            )?.textContent,
                                        ).getTime() || 0,
                                    source: feed.title || "Hír",
                                });
                            }
                        });
                    } catch (e) {
                        console.error("RSS feed error:", feed.feed_url, e);
                    }
                }
            }
            allNews.sort((a, b) => b.pubDate - a.pubDate);
            newsItems = allNews.slice(0, 6); // Top 6 news
        } catch (err) {
            console.error("News fetch error:", err);
        } finally {
            newsLoading = false;
        }

        // 4. Fetch Local Events
        try {
            const eRes = await fetch(
                `${apiBase}/api/events?location_slug=${encodeURIComponent(town)}`,
            );
            if (eRes.ok) {
                localEvents = await eRes.json();
            }
        } catch (err) {
            console.error("Events fetch error:", err);
        } finally {
            eventsLoading = false;
        }
    }

    function iconEmoji(code) {
        const map = {
            "01d": "☀️",
            "01n": "🌙",
            "02d": "⛅",
            "02n": "☁️",
            "03d": "☁️",
            "03n": "☁️",
            "04d": "☁️",
            "04n": "☁️",
            "09d": "🌧️",
            "09n": "🌧️",
            "10d": "🌦️",
            "10n": "🌧️",
            "11d": "⛈️",
            "11n": "⛈️",
            "13d": "❄️",
            "13n": "❄️",
            "50d": "🌫️",
            "50n": "🌫️",
        };
        return map[code] || "🌡️";
    }
</script>

<svelte:head>
    <title>{displayTown} - Index</title>
</svelte:head>

<div class="container main-content">
    <Breadcrumbs
        label={displayTown}
        settlementType={locType}
        {countyName}
        countySlug={$page.params.countySlug}
    />

    <h1 class="page-title">{displayTown} {locType} és környéke</h1>
    <p class="greeting no-top-margin">
        Helyi hírek, időjárás és címtár {displayTown} területén.
    </p>

    <!-- Widgets Row (Weather + News Preview) -->
    <div class="widgets-box">
        <!-- Settlement info -->
        <article id="attekintes">
            <h3 class="widget-title">Áttekintés</h3>
            <div class="more-info">
                {#if displayTownRo}
                    <span title="Román neve: {displayTownRo}"
                        >Románul: {displayTownRo}</span
                    >
                {/if}
                {#if displayTownDe}
                    <span title="Német neve: {displayTownDe}"
                        >Németül: {displayTownDe}</span
                    >
                {/if}
                <span title="Posta kód">
                    Irányítószám: <span>{displayZipCode || "–"}</span>
                </span>
                <span title="Koordináták">
                    Koordináták: <span>{displayTownCoordinates || "–"}</span>
                </span>
                <span title="Lakosság">
                    Lakosság: <span>{displayTownPopulation || "–"} fő</span>
                </span>
                <span title="Terület (négyzetkilométer)">
                    Terület: <span>{displayTownArea || "–"} km²</span>
                </span>
                <span title="Közigazgatási forma">
                    Közigazgatási forma: <span class="capitalize"
                        >{locType || "–"}</span
                    >
                </span>
                {#if displayParentId && displayParentName}
                    <span title="Kapcsolódó település:">
                        Kapcsolódó település: <a
                            href="/{displayParentCountySlug}-megye/{displayParentSlug}"
                            class="parent-city-link">{displayParentName}</a
                        >
                    </span>
                {/if}
            </div>
        </article>

        <!-- Coat of Arms -->
        <article id="cimer" class="crest-card">
            <h3 class="widget-title">{displayTown} címere</h3>
            <div class="crest-container">
                {#if displayTownCrest && displayTownCrest !== "–" && displayTownCrest.length > 5}
                    <img
                        src={`${import.meta.env.VITE_API_BASE_URL || "http://localhost:3000"}/api/proxy?url=${encodeURIComponent(displayTownCrest)}`}
                        alt="{displayTown} címere"
                        class="crest-img"
                    />
                {:else}
                    <span class="no-crest">Nincs elérhető címer</span>
                {/if}
            </div>
        </article>

        <!-- Weather Widget -->
        <article id="idojaras" class="weather-card">
            {#if weatherLoading}
                <div class="weather-left">
                    <span class="widget-title">Időjárás</span>
                    <div class="weather-temp-row">
                        <div class="skeleton weather-skeleton-temp"></div>
                    </div>
                    <div class="weather-desc">
                        <div class="skeleton weather-skeleton-desc"></div>
                    </div>
                    <div class="weather-footer">
                        <div class="skeleton skeleton-footer-1"></div>
                        <div class="skeleton skeleton-footer-2"></div>
                    </div>
                </div>
                <div class="weather-right">
                    <span class="weather-icon">⛅</span>
                </div>
            {:else if weatherData}
                <div class="weather-left">
                    <span class="widget-title">Időjárás</span>
                    <div class="weather-temp-row">
                        <span class="weather-temp">{weatherData.temp}</span
                        ><span class="weather-temp-unit">°C</span>
                        {#if weatherData.tempMin != null}
                            <span class="weather-temp-min"
                                >/ {weatherData.tempMin}°C</span
                            >
                        {/if}
                    </div>
                    <div class="weather-desc capitalize">
                        {weatherData.desc}
                    </div>
                    <div class="weather-footer">
                        {#if weatherData.timestamp}
                            <small class="weather-timestamp"
                                >Utoljára frissítve: {new Date(
                                    weatherData.timestamp,
                                ).toLocaleTimeString("hu-HU", {
                                    hour: "2-digit",
                                    minute: "2-digit",
                                })}</small
                            >
                        {/if}
                        <small class="weather-source"
                            >Forrás: OpenWeatherMap</small
                        >
                    </div>
                </div>
                <div class="weather-right">
                    <span class="weather-icon"
                        >{iconEmoji(weatherData.icon)}</span
                    >
                </div>
            {:else}
                <div class="weather-left">
                    <span class="widget-title">Időjárás</span>
                    <p class="weather-error">Időjárás adat nem elérhető.</p>
                </div>
                <div class="weather-right">
                    <span class="weather-icon">⛅</span>
                </div>
            {/if}
        </article>

        <!-- Events Widget -->
        <article id="esemenyek" class="event-widget">
            <h3 class="widget-title">Események</h3>
            {#if eventsLoading}
                <div class="skeleton-box">
                    <div class="skeleton skeleton-text skeleton-full"></div>
                    <div class="skeleton skeleton-text skeleton-60"></div>
                </div>
            {:else if localEvents.length > 0}
                <ul class="mini-event-list">
                    {#each localEvents.slice(0, 3) as event}
                        <li>
                            <div class="mini-event-date">
                                {new Date(event.start_date).toLocaleDateString(
                                    "hu-HU",
                                    { month: "short", day: "numeric" },
                                )}
                            </div>
                            <div class="mini-event-info">
                                <span class="mini-event-title"
                                    >{event.title}</span
                                >
                                <span class="mini-event-time">
                                    {#if event.start_time}{event.start_time.slice(
                                            0,
                                            5,
                                        )}{/if}
                                    {#if event.end_date && event.end_date !== event.start_date}
                                        - {new Date(
                                            event.end_date,
                                        ).toLocaleDateString("hu-HU", {
                                            month: "short",
                                            day: "numeric",
                                        })}
                                        {#if event.end_time}
                                            {event.end_time.slice(0, 5)}{/if}
                                    {:else if event.end_time}
                                        - {event.end_time.slice(0, 5)}
                                    {/if}
                                </span>
                            </div>
                        </li>
                    {/each}
                </ul>
                <a href="/esemenyek" class="widget-more-link"
                    >Összes esemény →</a
                >
            {:else}
                <span class="info-box">
                    <p>Nincsenek közeli események.</p>
                </span>
            {/if}
        </article>

        <!-- Local News Widget -->
        <article id="hirek" class="news news-widget">
            <h3 class="widget-title">Helyi hírek</h3>
            {#if newsLoading}
                <div class="news-loading-box">
                    <div class="skeleton skeleton-text skeleton-full"></div>
                    <div class="skeleton skeleton-text skeleton-80"></div>
                </div>
            {:else if newsItems.length > 0}
                <ul class="news-list">
                    {#each newsItems.slice(0, 4) as item}
                        <li>
                            <a
                                href={item.link}
                                target="_blank"
                                rel="nofollow noopener"
                                class="news-title"
                            >
                                📰 {item.title}
                            </a>
                            <div class="news-meta">
                                {item.source} - {new Date(
                                    item.pubDate,
                                ).toLocaleDateString("hu-HU")}
                            </div>
                        </li>
                    {/each}
                </ul>
            {:else}
                <span class="info-box"><p>Helyi hírek nem elérhetőek.</p></span>
            {/if}
        </article>
    </div>

    <!-- Directory Section -->
    <h2>
        {displayTown}i cimtár - Helyi Index
    </h2>

    {#if loadingServices}
        <div class="list grid">
            {#each Array(6) as _}
                <article class="card service--skeleton">
                    <div class="skeleton skeleton-text skeleton-30"></div>
                    <div class="skeleton skeleton-text skeleton-80-top"></div>
                    <div
                        class="skeleton skeleton-text skeleton-60-bottom"
                    ></div>
                </article>
            {/each}
        </div>
    {:else if servicesError}
        <div class="note error">{servicesError}</div>
    {:else if services.length === 0}
        <div class="note error">
            Nincs megjeleníthető bejegyzés {displayTown} területén.
        </div>
    {:else}
        <div class="filter-actions">
            <div class="info-box">
                <p>💡 Összesen:</p>
                <p><span>({displayItems.length}/{totalCount})</span></p>
            </div>

            <div class="view-mode-toggle">
                <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
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
            {#each displayItems as service}
                <article class="card service">
                    <span class="badge">{service.category}</span>
                    <h3 class="service-name">
                        <a href="/bejegyzes/{service.slug}">{service.name}</a>
                    </h3>
                    {#if service.url}
                        <div class="service-info service-info-url">
                            <span>🔗</span>
                            <a
                                href={service.url}
                                target="_blank"
                                rel="nofollow noopener">{service.url}</a
                            >
                        </div>
                    {/if}
                    <div class="service-info">
                        📍 {service.location} - {service.address}
                    </div>
                    <div class="service-info">📞 {service.phone}</div>
                    {#if service.notes}
                        <div class="service-notes">{service.notes}</div>
                    {/if}
                </article>
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
</div>

<style>
    .main-content {
        min-height: calc(100vh - 120px);
    }
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
    .skeleton-footer-1 {
        width: 120px;
        height: 0.75rem;
    }
    .skeleton-footer-2 {
        width: 90px;
        height: 0.75rem;
    }
    .weather-error {
        color: var(--text-faint);
        margin: 0.5rem 0 0;
    }
    .news-widget {
        display: flex;
        flex-direction: column;
    }
    .news-loading-box {
        padding: 1rem 0;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .skeleton-full {
        height: 1.2rem;
    }
    .skeleton-80 {
        height: 1.2rem;
        width: 80%;
    }
    .news-list {
        list-style: none;
        padding: 0;
        margin: 0;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .news-title {
        text-decoration: none;
        color: inherit;
        font-weight: 500;
        font-size: 1.05rem;
    }
    .news-meta {
        font-size: 0.8em;
        color: var(--text-faint);
        margin-top: 0.2rem;
    }
    .service--skeleton {
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
    .service-name a {
        color: inherit;
    }
    .service-info-url {
        margin-bottom: 0.5rem;
    }
    .service-info-url span {
        color: var(--text-faint);
        margin-right: 0.3rem;
    }
    .service-info-url a {
        color: var(--primary-color);
    }

    .widgets-box {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 2rem;
        margin-bottom: 2rem;
    }
    .news-widget,
    .event-widget {
        grid-column: span 3;
    }
    @media (max-width: 992px) {
        .widgets-box {
            grid-template-columns: 1fr;
        }
        .news-widget,
        .event-widget {
            grid-column: span 1;
        }
    }
    .weather-card {
        display: flex;
        align-items: flex-start;
        justify-content: flex-end;
        background: none;
        border-radius: 0;
        padding: 0;
        box-shadow: none;
        gap: 1rem;
    }
    .crest-card {
        display: flex;
        flex-direction: column;
        align-items: center;
    }
    /* Event widget */
    .event-widget {
        display: flex;
        flex-direction: column;
    }
    .mini-event-list {
        list-style: none;
        padding: 0;
        margin: 0;
        display: flex;
        flex-direction: column;
        gap: 0.6rem;
    }
    .mini-event-list li {
        display: flex;
        gap: 1rem;
        align-items: center;
        font-size: 0.9rem;
    }
    .mini-event-date {
        background: var(--primary-color);
        color: var(--text-color);
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
        font-weight: 600;
        font-size: 0.8rem;
        min-width: 3.5rem;
        text-align: center;
    }
    .mini-event-info {
        display: flex;
        flex-direction: column;
    }
    .mini-event-title {
        font-weight: 500;
        color: var(--text-color);
    }
    .mini-event-time {
        font-size: 0.8rem;
        color: var(--text-faint);
    }
    .widget-more-link {
        font-size: 0.85rem;
        color: var(--primary-color);
        text-decoration: none;
        margin-top: auto;
        font-weight: 500;
    }
    .widget-more-link:hover {
        text-decoration: underline;
    }
    /* Crest card */
    .crest-container {
        min-height: 120px;
    }
    .crest-img {
        max-width: 100%;
        max-height: 180px;
        object-fit: contain;
    }
    .no-crest {
        color: var(--text-faint);
        font-size: 0.9rem;
        font-style: italic;
    }
</style>
