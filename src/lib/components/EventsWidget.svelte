<script>
    import { browser } from "$app/environment";
    import { onMount, onDestroy } from "svelte";
    import { apiFetch } from "$lib/api";
    import { formatDateShort } from "$lib/utils";
    import EventDateBadge from "$lib/components/EventDateBadge.svelte";

    export let settlementSlug = null;
    export let countySlug = null;
    export let organizerName = null;
    export let locationName = null;
    export let limit = 3;
    export let ticker = false;

    const EVENT_TYPES = {
        cultural: "Kulturális",
        sports: "Sport",
        festival: "Fesztivál",
        religious: "Vallási",
        other: "Egyéb",
    };

    const LOCATION_TYPES = {
        "város": "Városok",
        "község": "Községek",
        "falu": "Falvak",
    };

    let items = [];
    let loading = true;
    let error = false;
    let activeType = null;
    let activeLocType = null;

    let tickerIndex = 0;
    let tickerInterval = null;
    let tickerResumeTimeout = null;
    const TICKER_RESUME_DELAY_MS = 4000;

    let pageStart = 0;

    async function fetchEvents() {
        try {
            let url = `/api/events`;
            if (organizerName) {
                url += `?organizer=${encodeURIComponent(organizerName)}`;
            } else if (settlementSlug) {
                url += `?location_slug=${encodeURIComponent(settlementSlug)}`;
            } else if (countySlug) {
                url += `?county_slug=${encodeURIComponent(countySlug)}`;
            }

            loading = true;
            error = false;
            const data = await apiFetch(url);
            items = data.events || [];
        } catch (err) {
            error = true;
        } finally {
            loading = false;
            startTicker();
        }
    }

    $: if (browser && (settlementSlug != null || countySlug != null || organizerName != null)) {
        fetchEvents();
    }

    function startTicker() {
        if (!ticker || tickerInterval) return;
        tickerInterval = setInterval(() => {
            if (filteredItems.length <= 1) return;
            tickerIndex = (tickerIndex + 1) % filteredItems.length;
        }, 5000);
    }

    function stopTicker() {
        if (tickerInterval) {
            clearInterval(tickerInterval);
            tickerInterval = null;
        }
        if (tickerResumeTimeout) {
            clearTimeout(tickerResumeTimeout);
            tickerResumeTimeout = null;
        }
    }

    function stopTickerAndResumeLater() {
        stopTicker();
        tickerResumeTimeout = setTimeout(() => {
            tickerResumeTimeout = null;
            startTicker();
        }, TICKER_RESUME_DELAY_MS);
    }

    function handleArrowClick(delta) {
        tickerIndex = (tickerIndex + delta + filteredItems.length) % filteredItems.length;
        stopTickerAndResumeLater();
    }

    onDestroy(() => {
        stopTicker();
    });

    $: filteredItems = (() => {
        let result = items;
        if (activeType) result = result.filter(e => e.event_type === activeType);
        if (activeLocType) result = result.filter(e => e.location_type === activeLocType);
        return result;
    })();
    $: availableTypes = [...new Set(items.map(e => e.event_type).filter(Boolean))];
    $: availableLocTypes = [...new Set(items.map(e => e.location_type).filter(Boolean))];
    $: visibleEvents = filteredItems.slice(pageStart, pageStart + limit);
    $: canPrev = pageStart > 0;
    $: canNext = pageStart + limit < filteredItems.length;
    $: showArrows = filteredItems.length > limit;

    function setTypeFilter(type) {
        activeType = activeType === type ? null : type;
        pageStart = 0;
        tickerIndex = 0;
        stopTicker();
        startTicker();
    }

    function setLocTypeFilter(type) {
        activeLocType = activeLocType === type ? null : type;
        pageStart = 0;
        tickerIndex = 0;
        stopTicker();
        startTicker();
    }

    $: typeLabel = locationName
        ? locationName
        : organizerName
            ? organizerName
            : null;

    $: hasAnyFilter = activeType || activeLocType;
    function clearAllFilters() {
        activeType = null;
        activeLocType = null;
        pageStart = 0;
        tickerIndex = 0;
        stopTicker();
        startTicker();
    }
</script>

<section id="esemenyek">
    <div class="event-widget component-box widget">
        <div class="widget-header">
            <h3 class="widget-title">Események{#if !loading} <span class="widget-title-count">({filteredItems.length})</span>{/if}{#if typeLabel} <span class="type-label">· {typeLabel}</span>{/if}</h3>
            {#if !loading && (availableTypes.length > 1 || availableLocTypes.length > 1)}
                <div class="event-type-badges">
                    {#if hasAnyFilter}
                        <button class="btn btn-xs event-type-badge--clear" on:click={clearAllFilters}>✕</button>
                    {/if}
                    {#each availableTypes as type}
                        <button
                            class="btn btn-xs event-type-badge"
                            class:active={activeType === type}
                            on:click={() => setTypeFilter(type)}
                        >{EVENT_TYPES[type] || type}</button>
                    {/each}
                    {#if availableLocTypes.length > 1}
                        <span class="badge-separator">·</span>
                        {#each availableLocTypes as lt}
                            <button
                                class="btn btn-xs event-type-badge event-loc-badge"
                                class:active={activeLocType === lt}
                                on:click={() => setLocTypeFilter(lt)}
                            >{LOCATION_TYPES[lt] || lt}</button>
                        {/each}
                    {/if}
                </div>
            {/if}
        </div>
        
        <div class="widget-content">
            {#if loading && ticker}
                <div class="event-ticker">
                    <span class="event-ticker-item">
                        <span class="event-ticker-title">...</span>
                        <span class="event-ticker-meta">...</span>
                    </span>
                </div>
                <div class="widget-nav">
                    <a href="/esemenyek" class="nav-btn">Összes esemény</a>
                </div>
            {:else if loading}
                <div class="event-cards-row">
                    {#each Array(limit) as _}
                        <article class="card sm">
                            <span class="event-card-date">...</span>
                            <span class="event-card-title">...</span>
                            <span class="event-card-meta">...</span>
                        </article>
                    {/each}
                </div>
                <div class="widget-nav">
                    <a href="/esemenyek" class="nav-btn">Összes esemény</a>
                </div>
            {:else if error || items.length === 0}
                <span class="info-box"><p>Nincsenek közeli események.</p></span>
            {:else if ticker}
                <div class="event-ticker">
                    {#key tickerIndex}
                        {@const event = filteredItems[tickerIndex]}
                        {#if event}
                            <div
                                class="event-ticker-item"
                                role="group"
                                on:mouseenter={stopTicker}
                                on:mouseleave={startTicker}
                            >
                                <a href="/esemenyek/{event.id}" class="event-ticker-title">
                                    {#if event.event_type}<span class="event-type-inline">{EVENT_TYPES[event.event_type] || event.event_type}</span>{/if}
                                    {event.title}
                                </a>
                                <span class="event-ticker-meta">
                                    <span class="datetime-text">
                                        <span class="start-date">
                                            <span class="start-date-label sr-only">Esemény kezdete:</span>
                                            {formatDateShort(event.start_date)}
                                            {#if event.start_time}
                                                <span class="start-time-label sr-only">Időpont:</span>{event.start_time.slice(0, 5)}
                                            {/if}
                                            <EventDateBadge event={event} live={true} />
                                        </span>
                                        <span class="event-time-separator"> - </span>
                                        <span class="end-date">
                                            <span class="end-date-label sr-only">Esemény vége:</span>
                                            {#if event.end_date && event.end_date !== event.start_date}
                                                {formatDateShort(event.end_date)}
                                                {#if event.end_time} {event.end_time.slice(0, 5)}{/if}
                                            {:else if event.end_time}
                                                - {event.end_time.slice(0, 5)}
                                            {/if}
                                        </span>
                                    </span>
                                    
                                    <span class="event-location-container">
                                        <span class="event-location-name"><span class="event-location-name-label sr-only">Település:</span> {#if event.location_name}<a href="/{event.county_slug}-megye/{event.location_slug}">{event.location_name}</a>{/if}</span>
                                        <span class="event-location-separator"> · </span>
                                        <span class="event-location-county"><span class="event-location-county-label sr-only">Megye:</span> {#if event.county}<a href="/{event.county_slug}-megye">{event.county} megye</a>{/if}</span>
                                    </span>
                                </span>
                            </div>
                        {/if}
                    {/key}
                    <div class="widget-nav">
                        <div class="arrows-container">
                            <button class="scroll-arrow left" on:click={() => handleArrowClick(-1)} aria-label="Előző esemény">&#8249;</button>
                            <button class="scroll-arrow right" on:click={() => handleArrowClick(1)} aria-label="Következő esemény">&#8250;</button>
                        </div>
                        <a href="/esemenyek" class="nav-btn">Összes esemény</a>
                    </div>
                </div>
            {:else}
                <div class="event-cards-row">
                    {#each visibleEvents as event}
                        <article href="/esemenyek/{event.id}" class="card sm">
                            <span class="event-card-date">
                                {#if event.event_type}<span class="event-type-inline">{EVENT_TYPES[event.event_type] || event.event_type}</span>{/if}
                                {formatDateShort(event.start_date)}
                                {#if event.start_time} · {event.start_time.slice(0, 5)}{/if}
                                <EventDateBadge event={event} live={true} />
                            </span>
                            <a href="/esemenyek/{event.id}" title="Esemény részletei" class="event-card-title">{event.title}</a>
                            <span class="event-card-meta">
                                {#if event.end_date && event.end_date !== event.start_date}
                                    {formatDateShort(event.end_date)}
                                    {#if event.end_time} {event.end_time.slice(0, 5)}{/if}
                                {:else if event.end_time}
                                    - {event.end_time.slice(0, 5)}
                                {/if}
                                {#if event.location_name} · {event.location_name}{/if}
                            </span>
                        </article>
                    {/each}
                    </div>
                <div class="widget-nav">
                    {#if showArrows}
                        <div class="arrows-container">
                        <button
                            class="scroll-arrow left"
                            disabled={!canPrev}
                            on:click={() => { pageStart = Math.max(0, pageStart - limit); }}
                            aria-label="Előző események"
                        >&#8249;</button>
                        <button
                            class="scroll-arrow right"
                            disabled={!canNext}
                            on:click={() => { pageStart = pageStart + limit; }}
                            aria-label="Következő események"
                        >&#8250;</button>
                        </div>
                    {/if}
                    <a href="/esemenyek" class="nav-btn">Összes esemény</a>
                </div>
            {/if}
        </div>
    </div>
</section>

<style>
    .event-widget {
        display: flex;
        flex-direction: column;
        flex: 1;
    }

    .event-cards-row {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 0.75rem;
        margin-bottom: 0.5rem;
    }
    .event-card-date {
        font-size: 0.75rem;
        font-weight: 600;
        color: var(--text-faint);
        text-transform: uppercase;
    }
    .event-card-title {
        font-weight: 600;
        font-size: 0.9rem;
        color: var(--text-color);
        line-height: 1.3;
        margin: 0.2rem 0;
    }
    .event-card-meta {
        font-size: 0.75rem;
        color: var(--text-faint);
    }

    .event-ticker {
        display: flex;
        justify-content: space-between;
    }

    @media (max-width: 992px) {
        .event-ticker {
            flex-direction: column;
            justify-content: flex-start;
        }
    }
    .event-ticker-item {
        display: flex;
        flex-direction: column;
        text-decoration: none;
        color: inherit;
        animation: event-ticker-slide 0.35s ease-out;
    }
    .event-ticker-item:hover .event-ticker-title {
        color: var(--szekely-red, #c0392b);
    }
    @keyframes event-ticker-slide {
        from {
            opacity: 0;
            transform: translateY(8px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
    .event-ticker-title {
        font-weight: 500;
        font-size: 1.1rem;
        transition: color 0.15s;
    }
    .event-ticker-meta {
        display: flex;
        flex-direction: column;
        gap: 0.2rem;
        font-size: 0.9em;
        color: var(--text-faint);
        margin-top: 0.2rem;
    }

    .event-type-badges {
        display: flex;
        gap: 0.4rem;
    }
    .event-type-badge {
        padding: 0.15rem 0.5rem;
        border: 1px solid var(--border-color, #ddd);
        background: transparent;
        color: var(--text-secondary, #666);
        transition: background 0.15s, color 0.15s;
        margin-bottom: 0;
        text-transform: uppercase;
    }
    .event-type-badge:hover {
        background: var(--hover-bg, #f5f5f5);
        color: var(--badge-text);
    }
    .event-type-badge.active {
        background: var(--szekely-red, #c0392b);
        color: #fff;
        border-color: var(--szekely-red, #c0392b);
    }
    .event-type-badge--clear {
        padding: 0.15rem 0.35rem;
        font-size: 0.4rem;
    }
    .badge-separator {
        color: var(--text-faint);
        font-size: 0.8rem;
        align-self: center;
    }
    .event-loc-badge {
        border-style: dashed;
    }
    .event-type-inline {
        display: inline-block;
        font-size: 0.65rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.03em;
        color: var(--szekely-red, #c0392b);
        margin-right: 0.3rem;
    }

    @media (max-width: 992px) {
        .event-type-badges {
            gap: 0.2rem;
        }
        .event-type-badge {
            width: 100%;
        }
        .event-loc-badge {
            width: 100%;
        }

        .event-cards-row {
            grid-template-columns: 1fr;
        }
        
    }
</style>
