<script>
    import { onMount } from "svelte";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import EventDateBadge from "$lib/components/EventDateBadge.svelte";

    const getBase = () =>
        import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

    let events = [];
    let loading = true;
    let error = null;
    let total = 0;
    let currentPage = 1;
    const pageSize = 12;

    /** @type {{ event_types: string[], locations: { name: string, location_slug: string, county_slug: string, county_name: string }[] }} */
    let filterOptions = { event_types: [], locations: [] };
    let filtersLoading = true;

    /** @type {string | null} */
    let filterType = null;
    /** @type {string | null} */
    let filterLocationSlug = null;
    let filterDateFrom = "";
    let filterDateTo = "";

    function scrollToTop() {
        if (typeof window !== "undefined") {
            window.scrollTo({ top: 0, behavior: "smooth" });
        }
    }

    function buildQueryParams(page) {
        const offset = (page - 1) * pageSize;
        const p = new URLSearchParams();
        p.set("limit", String(pageSize));
        p.set("offset", String(offset));
        if (filterType) p.set("event_type", filterType);
        if (filterLocationSlug) p.set("location_slug", filterLocationSlug);
        if (filterDateFrom) p.set("date_from", filterDateFrom);
        if (filterDateTo) p.set("date_to", filterDateTo);
        return p;
    }

    async function loadFilterOptions() {
        filtersLoading = true;
        try {
            const res = await fetch(`${getBase()}/api/events/filter-options`);
            if (res.ok) {
                const data = await res.json();
                filterOptions = {
                    event_types: data.event_types || [],
                    locations: data.locations || [],
                };
            }
        } catch (e) {
            console.error(e);
        } finally {
            filtersLoading = false;
        }
    }

    async function loadPage(page) {
        loading = true;
        error = null;
        try {
            const qs = buildQueryParams(page);
            const res = await fetch(`${getBase()}/api/events?${qs}`);
            if (res.ok) {
                const data = await res.json();
                events = data.events || [];
                total = data.total || 0;
                currentPage = page;
            } else {
                error = "Nem sikerült betölteni az eseményeket.";
            }
        } catch (e) {
            console.error(e);
            error = "Hálózati hiba történt.";
        } finally {
            loading = false;
        }
    }

    function applyFilters() {
        scrollToTop();
        loadPage(1);
    }

    function clearAllFilters() {
        filterType = null;
        filterLocationSlug = null;
        filterDateFrom = "";
        filterDateTo = "";
        applyFilters();
    }

    function clearDateRange() {
        filterDateFrom = "";
        filterDateTo = "";
        applyFilters();
    }

    onMount(() => {
        loadFilterOptions();
        loadPage(1);
    });

    $: totalPages = Math.max(1, Math.ceil(total / pageSize));

    $: hasActiveFilters =
        filterType != null ||
        filterLocationSlug != null ||
        filterDateFrom !== "" ||
        filterDateTo !== "";

    function formatDate(dateStr) {
        if (!dateStr) return "";
        const d = new Date(dateStr);
        return d.toLocaleDateString("hu-HU", {
            year: "numeric",
            month: "long",
            day: "numeric",
            weekday: "long",
        });
    }

    function formatEventDateTime(ev) {
        if (!ev) return "";
        let res = formatDate(ev.start_date);
        if (ev.start_time) res += ` ${ev.start_time.slice(0, 5)}`;

        if (ev.end_date && ev.end_date !== ev.start_date) {
            res += ` - ${formatDate(ev.end_date)}`;
            if (ev.end_time) res += ` ${ev.end_time.slice(0, 5)}`;
        } else if (ev.end_time) {
            res += ` - ${ev.end_time.slice(0, 5)}`;
        }
        return res;
    }

    const EVENT_TYPE_LABELS = {
        cultural: "Kulturális",
        sports: "Sport",
        festival: "Fesztivál",
        religious: "Vallási",
        other: "Egyéb",
    };
</script>

<svelte:head>
    <title>Esemény Naptár - Székely Gugel</title>
</svelte:head>

<Breadcrumbs label="Események" />
<h1 class="page-title">Esemény Naptár</h1>
<p class="greeting">Válogass a legfrissebb székelyföldi események közül.</p>

<div class="news-page-layout events-page-layout">
    <section class="news-list">
        {#if loading && events.length === 0}
            <span class="info-box"><p>adat betöltés...</p></span>
        {:else if error}
            <span class="info-box"><p>{error}</p></span>
        {:else if events.length === 0}
            <span class="info-box">
                <p>
                    {#if hasActiveFilters}
                        Nincs a szűrőknek megfelelő esemény.
                    {:else}
                        Jelenleg nincsenek meghirdetett események.
                    {/if}
                </p>
                {#if hasActiveFilters}
                    <button
                        type="button"
                        class="btn btn-sm clear-filters"
                        on:click={clearAllFilters}
                    >
                        Szűrők törlése
                    </button>
                {/if}
            </span>
        {:else}
            {#if hasActiveFilters}
                <div class="events-filter-strip info-box">
                    <p>
                        Szűrve
                        {#if filterType}
                            · <span class="active"
                                >{EVENT_TYPE_LABELS[filterType] ||
                                    filterType}</span
                            >
                        {/if}
                        {#if filterLocationSlug}
                            · <span class="active"
                                >{filterOptions.locations.find(
                                    (l) => l.location_slug === filterLocationSlug,
                                )?.name || filterLocationSlug}</span
                            >
                        {/if}
                        {#if filterDateFrom || filterDateTo}
                            · <span class="active"
                                >{filterDateFrom || "…"} — {filterDateTo ||
                                    "…"}</span
                            >
                        {/if}
                        <button
                            type="button"
                            class="clear-filters btn btn-xs"
                            on:click={clearAllFilters}
                        >
                            Összes törlése
                        </button>
                    </p>
                </div>
            {/if}

            <div class="list grid" id="esemenyek-lista">
                {#each events as event}
                    <article class="card">
                        <div class="badge event">
                            {EVENT_TYPE_LABELS[event.event_type] ||
                                event.event_type}
                        </div>
                        <h2 class="event-title">
                            <a
                                href="/esemenyek/{event.id}"
                                class="event-card-link">{event.title}</a
                            >
                        </h2>

                        <div class="event-meta">
                            <span class="event-date">
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
                                    ><rect
                                        x="3"
                                        y="4"
                                        width="18"
                                        height="18"
                                        rx="2"
                                        ry="2"
                                    ></rect><line x1="16" y1="2" x2="16" y2="6"
                                    ></line><line x1="8" y1="2" x2="8" y2="6"
                                    ></line><line x1="3" y1="10" x2="21" y2="10"
                                    ></line></svg
                                >
                                <span class="event-date-time"
                                    >{formatEventDateTime(event)}
                                    <EventDateBadge event={event} live={true} /></span
                                >
                            </span>
                            <span class="event-location">
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
                                    ><path
                                        d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"
                                    ></path><circle cx="12" cy="10" r="3"></circle></svg
                                >
                                <span class="event-location-names">
                                    <a
                                        href="/{event.county_slug}-megye/{event.location_slug}"
                                        class="event-location-link"
                                        >{event.location_name}</a
                                    >,
                                    <a
                                        href="/{event.county_slug}-megye"
                                        class="event-county-link"
                                        >{event.county} megye</a
                                    >
                                </span>
                            </span>
                        </div>

                        {#if event.description}
                            <p class="event-description">{event.description}</p>
                        {/if}

                        {#if event.organizer}
                            <div class="event-organizer">
                                <strong>Szervező:</strong>
                                {event.organizer}
                            </div>
                        {/if}
                    </article>
                {/each}
            </div>

            {#if totalPages > 1}
                <div class="pagination">
                    <button
                        class="pagination-btn"
                        disabled={currentPage <= 1 || loading}
                        on:click={() => loadPage(currentPage - 1)}
                    >
                        &#8249; Előző
                    </button>
                    <span class="pagination-info"
                        >{currentPage} / {totalPages}</span
                    >
                    <button
                        class="pagination-btn"
                        disabled={currentPage >= totalPages || loading}
                        on:click={() => loadPage(currentPage + 1)}
                    >
                        Következő &#8250;
                    </button>
                </div>
            {/if}
        {/if}
    </section>

    <aside class="news-sidebar events-sidebar" aria-label="Esemény szűrők">
        <div class="news-sidebar-box">
            <div class="news-sidebar-header">
                <h4 class="news-sidebar-heading">Esemény típusa</h4>
            </div>
            {#if filtersLoading}
                <ul class="news-sidebar-sources">
                    {#each Array(4) as _}
                        <div
                            class="news-sidebar-source-item sidebar-loader-item"
                        >
                            <span
                                class="news-source-dot"
                                style:background="var(--border-color)"
                            ></span>
                            <span>…</span>
                        </div>
                    {/each}
                </ul>
            {:else}
                <ul class="news-sidebar-sources">
                    <button
                        type="button"
                        class="news-sidebar-source-item news-sidebar-source-all"
                        class:active={filterType === null}
                        aria-pressed={filterType === null}
                        aria-label="Összes eseménytípus"
                        title="Összes típus"
                        on:click={() => {
                            filterType = null;
                            applyFilters();
                        }}
                    >
                        <span class="news-source-dot dot-all-sources"></span>
                        Összes típus
                    </button>
                    {#each filterOptions.event_types as t}
                        <button
                            type="button"
                            class="news-sidebar-source-item"
                            class:active={filterType === t}
                            aria-pressed={filterType === t}
                            title={EVENT_TYPE_LABELS[t] || t}
                            aria-label="Szűrés: {EVENT_TYPE_LABELS[t] || t}"
                            on:click={() => {
                                filterType = filterType === t ? null : t;
                                applyFilters();
                            }}
                        >
                            <span class="news-source-dot type-dot"></span>
                            {EVENT_TYPE_LABELS[t] || t}
                        </button>
                    {/each}
                </ul>
            {/if}
        </div>

        <div class="news-sidebar-box">
            <div class="news-sidebar-header">
                <h4 class="news-sidebar-heading">Helyszín</h4>
            </div>
            {#if filtersLoading}
                <ul class="news-sidebar-sources">
                    <div class="news-sidebar-source-item sidebar-loader-item">
                        <span>…</span>
                    </div>
                </ul>
            {:else}
                <ul
                    class="news-sidebar-sources events-sidebar-locations-scroll"
                >
                    <button
                        type="button"
                        class="news-sidebar-source-item news-sidebar-source-all"
                        class:active={filterLocationSlug === null}
                        aria-pressed={filterLocationSlug === null}
                        aria-label="Összes helyszín"
                        title="Összes település"
                        on:click={() => {
                            filterLocationSlug = null;
                            applyFilters();
                        }}
                    >
                        <span class="news-source-dot dot-all-sources"></span>
                        Összes helyszín
                    </button>
                    {#each filterOptions.locations as loc}
                        <button
                            type="button"
                            class="news-sidebar-source-item"
                            class:active={filterLocationSlug === loc.location_slug}
                            aria-pressed={filterLocationSlug === loc.location_slug}
                            title="{loc.name}, {loc.county_name} megye"
                            aria-label="Szűrés helyszínre: {loc.name}, {loc.county_name} megye"
                            on:click={() => {
                                filterLocationSlug =
                                    filterLocationSlug === loc.location_slug
                                        ? null
                                        : loc.location_slug;
                                applyFilters();
                            }}
                        >
                            <span class="news-source-dot loc-dot"></span>
                            {loc.name}
                            <span class="events-loc-county"
                                >({loc.county_name})</span
                            >
                        </button>
                    {/each}
                </ul>
            {/if}
        </div>

        <div class="news-sidebar-box">
            <div class="news-sidebar-header">
                <h4 class="news-sidebar-heading">Kezdő dátum</h4>
            </div>
            <div class="events-date-fields">
                <label class="events-date-label" for="ev-filter-date-from"
                    >Ettől</label
                >
                <input
                    id="ev-filter-date-from"
                    name="date_from"
                    class="events-date-input"
                    type="date"
                    bind:value={filterDateFrom}
                    on:change={applyFilters}
                />
                <label class="events-date-label" for="ev-filter-date-to"
                    >Eddig</label
                >
                <input
                    id="ev-filter-date-to"
                    name="date_to"
                    class="events-date-input"
                    type="date"
                    bind:value={filterDateTo}
                    on:change={applyFilters}
                />
                <button
                    type="button"
                    class="btn btn-xs events-date-clear"
                    disabled={!filterDateFrom && !filterDateTo}
                    on:click={clearDateRange}
                >
                    Időszak törlése
                </button>
            </div>
        </div>
    </aside>
</div>

<style>
    .events-page-layout {
        margin-top: 1rem;
    }

    .events-filter-strip {
        margin-bottom: 1rem;
    }
    .events-filter-strip p {
        margin: 0;
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.35rem 0.5rem;
    }

    .events-sidebar .news-sidebar-box + .news-sidebar-box {
        margin-top: 0.75rem;
    }

    .events-sidebar-locations-scroll {
        max-height: 14rem;
        overflow-y: auto;
        margin-top: 0;
        padding-right: 0.15rem;
    }

    .events-loc-county {
        font-size: 0.78rem;
        color: var(--text-faint);
        font-weight: 500;
    }

    .type-dot {
        background: linear-gradient(
            135deg,
            var(--szekely-red, #c8102e),
            var(--szekely-brown, #7a3c10)
        );
    }
    .loc-dot {
        background: var(--szekely-green, #2f4f4f);
    }

    .events-date-fields {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        margin-top: 0.25rem;
    }
    .events-date-label {
        font-size: 0.8rem;
        font-weight: 600;
        color: var(--text-muted);
    }
    .events-date-input {
        width: 100%;
        padding: 0.4rem 0.5rem;
        border: 1px solid var(--border-color);
        border-radius: 8px;
        background: var(--card-bg);
        color: var(--text-primary);
        font-size: 0.85rem;
    }
    .events-date-clear {
        align-self: flex-start;
        margin-top: 0.25rem;
    }

    .event-card-link {
        text-decoration: none;
        color: inherit;
    }

    .badge.event {
        position: absolute;
        top: 1rem;
        right: 1rem;
    }

    .event-title {
        margin: 0 0 0.5rem;
    }

    .event-meta {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        font-size: 0.9rem;
        color: var(--text-faint);
        margin: 1rem 0;
    }

    .event-location,
    .event-date {
        display: flex;
        gap: 0.5rem;
        align-items: center;
        flex-direction: row;
        flex-wrap: wrap;
    }

    .event-date-time {
        display: inline-flex;
        align-items: center;
        flex-wrap: wrap;
        gap: 0.25rem;
    }

    .event-description {
        margin: 0;
        line-height: 1.6;
        color: var(--text-color);
        display: -webkit-box;
        -webkit-line-clamp: 3;
        line-clamp: 3;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }

    .event-organizer {
        font-size: 0.85rem;
        color: var(--text-faint);
        margin-top: auto;
        padding-top: 1rem;
        border-top: 1px solid var(--border-color);
    }

    .pagination {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 1rem;
        margin-top: 2rem;
        padding: 1rem 0;
    }
    .pagination-btn {
        padding: 0.5rem 1rem;
        border: 1px solid var(--border-color, #d1d5db);
        border-radius: 6px;
        background: var(--card-bg, #fff);
        color: var(--text-primary, #333);
        cursor: pointer;
        font-size: 0.9rem;
        transition: background 0.15s;
    }
    .pagination-btn:hover:not(:disabled) {
        background: var(--skeleton-bg, #f3f4f6);
        color: var(--szekely-red, #c0392b);
    }
    .pagination-btn:disabled {
        opacity: 0.4;
        cursor: default;
    }
    .pagination-info {
        font-size: 0.9rem;
        color: var(--text-faint);
    }
</style>
