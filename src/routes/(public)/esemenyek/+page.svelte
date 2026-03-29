<script>
    import { onMount } from "svelte";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import EventDateBadge from "$lib/components/EventDateBadge.svelte";
    import {
        venuePageUrl,
        formatDateShort,
        formatMonthShortLikeCard,
        formatDateWithOptionalTime,
    } from "$lib/utils";

    const getBase = () =>
        import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

    let events = [];
    let loading = true;
    let error = null;
    let total = 0;
    let currentPage = 1;
    const pageSize = 12;

    /** @type {{ event_types: string[], locations: { name: string, location_slug: string, county_slug: string, county_name: string }[], months: string[], event_days: string[], schedule_event_ids: number[] }} */
    let filterOptions = {
        event_types: [],
        locations: [],
        months: [],
        event_days: [],
        schedule_event_ids: [],
    };
    let filtersLoading = true;

    /** @type {string | null} */
    let filterType = null;
    /** @type {string | null} */
    let filterLocationSlug = null;
    /** @type {string | null} YYYY-MM from filter-options, drives date_from / date_to */
    let filterMonthKey = null;
    /** @type {string | null} YYYY-MM-DD — single-day filter (takes precedence over filterMonthKey) */
    let filterDayKey = null;

    /** @type {'year' | 'month'} */
    let sidebarCalendarMode = "year";
    /** @type {string | null} YYYY-MM — month open in sidebar day grid */
    let drillMonthYM = null;

    let viewMode = "grid";
    /** @type {'start_date' | 'title'} */
    let sortMode = "start_date";
    let sortOpen = false;

    const sortLabels = {
        start_date: "Dátum (előbb → később)",
        title: "Név (A→Z)",
    };

    function scrollToTop() {
        if (typeof window !== "undefined") {
            window.scrollTo({ top: 0, behavior: "smooth" });
        }
    }

    /** @param {string} ym */
    function monthRange(ym) {
        const [y, m] = ym.split("-").map(Number);
        const mm = String(m).padStart(2, "0");
        const from = `${y}-${mm}-01`;
        const last = new Date(y, m, 0).getDate();
        const to = `${y}-${mm}-${String(last).padStart(2, "0")}`;
        return { from, to };
    }

    /** @param {string} ym */
    function formatMonthLabelHu(ym) {
        const [y, m] = ym.split("-").map(Number);
        if (!y || !m) return ym;
        return new Date(y, m - 1, 1).toLocaleDateString("hu-HU", {
            year: "numeric",
            month: "long",
        });
    }

    /** @param {string} ym YYYY-MM — short month only (same style as event card dates). */
    function drillMonthTitleShort(ym) {
        const [y, m] = ym.split("-").map(Number);
        if (!y || !m) return ym;
        return formatMonthShortLikeCard(y, m);
    }

    /** Monday-first weekday labels (H–V) */
    const weekdayShortHu = (() => {
        const mon = new Date(2024, 0, 1);
        const out = [];
        for (let i = 0; i < 7; i++) {
            const d = new Date(mon);
            d.setDate(mon.getDate() + i);
            out.push(
                d.toLocaleDateString("hu-HU", { weekday: "narrow" }),
            );
        }
        return out;
    })();

    /** @param {number} y @param {number} month1to12 */
    function buildMonthDayGrid(y, month1to12) {
        const first = new Date(y, month1to12 - 1, 1);
        const lastDay = new Date(y, month1to12, 0).getDate();
        const mondayFirst = (first.getDay() + 6) % 7;
        /** @type {{ type: 'pad' } | { type: 'day', day: number, iso: string }}[] */
        const cells = [];
        for (let i = 0; i < mondayFirst; i++) cells.push({ type: "pad" });
        for (let d = 1; d <= lastDay; d++) {
            const mm = String(month1to12).padStart(2, "0");
            const dd = String(d).padStart(2, "0");
            cells.push({
                type: "day",
                day: d,
                iso: `${y}-${mm}-${dd}`,
            });
        }
        const tail = (7 - (cells.length % 7)) % 7;
        for (let i = 0; i < tail; i++) cells.push({ type: "pad" });
        return cells;
    }

    /** @param {string} iso YYYY-MM-DD */
    function formatDayLabelHu(iso) {
        const [y, m, d] = iso.split("-").map(Number);
        if (!y || !m || !d) return iso;
        return new Date(y, m - 1, d).toLocaleDateString("hu-HU", {
            year: "numeric",
            month: "long",
            day: "numeric",
        });
    }

    /** Every calendar day YYYY-MM-DD from start through end (inclusive), local date math. */
    /** @param {string} startIso @param {string} endIso */
    function ymdRangeInclusive(startIso, endIso) {
        const s = String(startIso).slice(0, 10);
        const e = String(endIso).slice(0, 10);
        const [y1, m1, d1] = s.split("-").map(Number);
        const [y2, m2, d2] = e.split("-").map(Number);
        if (!y1 || !m1 || !d1 || !y2 || !m2 || !d2) return [];
        const out = [];
        const cur = new Date(y1, m1 - 1, d1);
        const last = new Date(y2, m2 - 1, d2);
        if (cur > last) return [];
        while (cur <= last) {
            out.push(
                `${cur.getFullYear()}-${String(cur.getMonth() + 1).padStart(2, "0")}-${String(cur.getDate()).padStart(2, "0")}`,
            );
            cur.setDate(cur.getDate() + 1);
        }
        return out;
    }

    /** @param {string} y @param {number} m */
    function ymKey(y, m) {
        return `${y}-${String(m).padStart(2, "0")}`;
    }

    let calendarViewYear = new Date().getFullYear();

    $: yearsWithEvents = (() => {
        const ys = new Set();
        for (const raw of filterOptions.months || []) {
            const y = parseInt(String(raw).slice(0, 4), 10);
            if (!Number.isNaN(y)) ys.add(y);
        }
        return Array.from(ys).sort((a, b) => a - b);
    })();

    $: eventMonthsSet = new Set(filterOptions.months || []);
    /** Days with ≥1 upcoming event (API) plus every day in each loaded event’s start–end span (multi-day). */
    $: eventDaysSet = (() => {
        const set = new Set(filterOptions.event_days || []);
        for (const ev of events) {
            if (!ev.start_date) continue;
            const end =
                ev.end_date && String(ev.end_date).trim() !== ""
                    ? ev.end_date
                    : ev.start_date;
            for (const k of ymdRangeInclusive(ev.start_date, end)) {
                set.add(k);
            }
        }
        return set;
    })();
    /** IDs of upcoming events that have napi program activities (matches API has_schedule; set is a fallback if the list row omits the flag). */
    $: scheduleEventIdsSet = new Set(filterOptions.schedule_event_ids || []);

    $: if (!filtersLoading && yearsWithEvents.length > 0) {
        if (!yearsWithEvents.includes(calendarViewYear)) {
            calendarViewYear = yearsWithEvents[0];
        }
    }

    $: if (filterMonthKey) {
        const y = parseInt(filterMonthKey.slice(0, 4), 10);
        if (!Number.isNaN(y)) calendarViewYear = y;
    }

    $: if (filterDayKey) {
        const y = parseInt(filterDayKey.slice(0, 4), 10);
        if (!Number.isNaN(y)) calendarViewYear = y;
    }

    $: drillYearMonth =
        drillMonthYM != null
            ? (() => {
                  const [yy, mm] = drillMonthYM.split("-").map(Number);
                  return Number.isFinite(yy) && Number.isFinite(mm)
                      ? { y: yy, m: mm }
                      : null;
              })()
            : null;

    $: monthDayCells =
        drillYearMonth && sidebarCalendarMode === "month"
            ? buildMonthDayGrid(drillYearMonth.y, drillYearMonth.m)
            : [];

    function stepCalendarYear(delta) {
        if (yearsWithEvents.length < 2) return;
        const i = yearsWithEvents.indexOf(calendarViewYear);
        const j = Math.min(
            yearsWithEvents.length - 1,
            Math.max(0, i + delta),
        );
        if (j !== i) calendarViewYear = yearsWithEvents[j];
    }

    /** @param {number} month1to12 */
    function openMonthDrill(month1to12) {
        const key = ymKey(calendarViewYear, month1to12);
        if (!eventMonthsSet.has(key)) return;
        filterMonthKey = key;
        filterDayKey = null;
        drillMonthYM = key;
        sidebarCalendarMode = "month";
        applyFilters();
    }

    function closeMonthDrill() {
        sidebarCalendarMode = "year";
        drillMonthYM = null;
    }

    /** @param {string} iso YYYY-MM-DD */
    function pickCalendarDay(iso) {
        if (!eventDaysSet.has(iso)) return;
        filterDayKey = iso;
        filterMonthKey = null;
        applyFilters();
    }

    /** @param {string} cellKey YYYY-MM */
    function monthCellIsActive(cellKey) {
        if (filterMonthKey === cellKey && !filterDayKey) return true;
        if (filterDayKey) return filterDayKey.startsWith(`${cellKey}-`);
        return false;
    }

    function buildQueryParams(page) {
        const offset = (page - 1) * pageSize;
        const p = new URLSearchParams();
        p.set("limit", String(pageSize));
        p.set("offset", String(offset));
        if (filterType) p.set("event_type", filterType);
        if (filterLocationSlug) p.set("location_slug", filterLocationSlug);
        if (filterDayKey) {
            p.set("date_from", filterDayKey);
            p.set("date_to", filterDayKey);
        } else if (filterMonthKey) {
            const { from, to } = monthRange(filterMonthKey);
            p.set("date_from", from);
            p.set("date_to", to);
        }
        if (sortMode === "title") p.set("sort", "title");
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
                    months: data.months || [],
                    event_days: data.event_days || [],
                    schedule_event_ids: data.schedule_event_ids || [],
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
        filterMonthKey = null;
        filterDayKey = null;
        closeMonthDrill();
        applyFilters();
    }

    function clearMonthOnly() {
        filterMonthKey = null;
        filterDayKey = null;
        closeMonthDrill();
        applyFilters();
    }

    function setSortMode(/** @type {'start_date' | 'title'} */ mode) {
        sortMode = mode;
        sortOpen = false;
        applyFilters();
    }

    onMount(() => {
        loadFilterOptions();
        loadPage(1);
        const closeSort = () => {
            sortOpen = false;
        };
        window.addEventListener("click", closeSort);
        return () => window.removeEventListener("click", closeSort);
    });

    $: totalPages = Math.max(1, Math.ceil(total / pageSize));

    $: hasActiveFilters =
        filterType != null ||
        filterLocationSlug != null ||
        filterMonthKey != null ||
        filterDayKey != null;

    $: locationFilterLabel =
        filterOptions.locations.find((l) => l.location_slug === filterLocationSlug)
            ?.name || filterLocationSlug;

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

<div class="events-top-filters">
    <div class="header-tabs">
        <span class="header-tabs-label events-header-tabs-label">Települések:</span>
        {#if filtersLoading}
            <span class="btn btn-md" style:opacity="0.5">adat betöltés...</span>
        {:else}
            <button
                type="button"
                class="btn btn-md"
                class:active={filterLocationSlug === null}
                aria-pressed={filterLocationSlug === null}
                on:click={() => {
                    filterLocationSlug = null;
                    applyFilters();
                }}>Összes</button
            >
            {#each filterOptions.locations as loc (loc.location_slug + loc.county_slug)}
                <button
                    type="button"
                    class="btn btn-md"
                    class:active={filterLocationSlug === loc.location_slug}
                    aria-pressed={filterLocationSlug === loc.location_slug}
                    title="{loc.name}, {loc.county_name} megye"
                    on:click={() => {
                        filterLocationSlug = loc.location_slug;
                        applyFilters();
                    }}
                >
                    {loc.name}
                </button>
            {/each}
        {/if}
    </div>

    <div class="header-tabs events-filter-row-second">
        <span class="header-tabs-label events-header-tabs-label">Esemény típusa:</span>
        {#if filtersLoading}
            <span class="btn btn-md" style:opacity="0.5">adat betöltés...</span>
        {:else}
            <button
                type="button"
                class="btn btn-md"
                class:active={filterType === null}
                aria-pressed={filterType === null}
                on:click={() => {
                    filterType = null;
                    applyFilters();
                }}>Összes</button
            >
            {#each filterOptions.event_types as t (t)}
                <button
                    type="button"
                    class="btn btn-md"
                    class:active={filterType === t}
                    aria-pressed={filterType === t}
                    title={EVENT_TYPE_LABELS[t] || t}
                    on:click={() => {
                        filterType = t;
                        applyFilters();
                    }}
                >
                    {EVENT_TYPE_LABELS[t] || t}
                </button>
            {/each}
        {/if}
    </div>

    <div class="filter-actions">
        <span class="info-box">
            <p>
                {#if !hasActiveFilters}
                    💡 Leszűrve: <span class="active">Összes</span>
                {:else}
                    🔍 Szűrők:
                    {#if filterLocationSlug}
                        <span class="active">{locationFilterLabel}</span>
                    {/if}
                    {#if filterType}
                        {#if filterLocationSlug}<span class="events-filter-sep">·</span>{/if}
                        <span class="active">{EVENT_TYPE_LABELS[filterType] || filterType}</span>
                    {/if}
                    {#if filterDayKey}
                        {#if filterLocationSlug || filterType}<span class="events-filter-sep">·</span>{/if}
                        <span class="active">{formatDayLabelHu(filterDayKey)}</span>
                    {:else if filterMonthKey}
                        {#if filterLocationSlug || filterType}<span class="events-filter-sep">·</span>{/if}
                        <span class="active">{formatMonthLabelHu(filterMonthKey)}</span>
                    {/if}
                    <button
                        type="button"
                        class="clear-filters btn btn-xs"
                        on:click={clearAllFilters}>Szűrő törlése</button
                    >
                {/if}
            </p>
            <p><span>({events.length}/{total || 0})</span></p>
        </span>

        <div class="view-mode-toggle">
            <div class="sort-toggle">
                <button
                    type="button"
                    class="btn btn-sm"
                    on:click|stopPropagation={() => (sortOpen = !sortOpen)}
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
                            type="button"
                            class:active={sortMode === "start_date"}
                            on:click|stopPropagation={() => setSortMode("start_date")}
                            >Dátum (előbb → később)</button
                        >
                        <button
                            type="button"
                            class:active={sortMode === "title"}
                            on:click|stopPropagation={() => setSortMode("title")}>Név (A→Z)</button
                        >
                    </div>
                {/if}
            </div>

            <button
                type="button"
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
                type="button"
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
</div>

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
            </span>
        {:else}
            <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}" id="esemenyek-lista">
                {#each events as event}
                    {@const endDateForCard =
                        event.end_date && String(event.end_date).trim() !== ""
                            ? event.end_date
                            : event.start_date}
                    {@const showProgramLink =
                        event.has_schedule === true ||
                        scheduleEventIdsSet.has(event.id)}
                    <article class="card event-list-card">
                        <header class="event-card-head">
                            <div class="event-card-head-left">
                                {#if event.event_type}
                                    <div class="event-card-type-row">
                                        <span class="event-type-inline"
                                            >{EVENT_TYPE_LABELS[event.event_type] ||
                                                event.event_type}</span
                                        >
                                    </div>
                                {/if}
                                <span class="event-start-datetime">
                                    <span class="sr-only">
                                        Kezdés: {formatDateWithOptionalTime(
                                            event.start_date,
                                            event.start_time,
                                        )}{#if endDateForCard !== event.start_date}
                                            , befejezés: {formatDateWithOptionalTime(
                                                endDateForCard,
                                                event.end_time,
                                            )}
                                        {/if}
                                    </span>
                                    <span aria-hidden="true">
                                        {formatDateShort(event.start_date)}
                                        {#if endDateForCard !== event.start_date}
                                            <span> – </span>{formatDateShort(endDateForCard)}
                                        {/if}
                                    </span>
                                </span>
                            </div>
                            <div class="event-card-head-right">
                                <EventDateBadge
                                    event={event}
                                    live={true}
                                    corner={true}
                                />
                            </div>
                        </header>
                        <h2 class="event-title">
                            <a
                                href="/esemenyek/{event.id}"
                                class="event-card-link">{event.title}</a
                            >
                        </h2>

                        <div class="event-meta">
                            <span class="event-meta-row event-meta-row--date">
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
                                <span class="event-meta-date-text"
                                    ><span class="sr-only">Kezdés: </span>{formatDateWithOptionalTime(
                                        event.start_date,
                                        event.start_time,
                                    )}</span
                                >
                            </span>
                            <span class="event-meta-row event-meta-row--date">
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
                                <span class="event-meta-date-text"
                                    ><span class="sr-only">Befejezés: </span>{formatDateWithOptionalTime(
                                        endDateForCard,
                                        event.end_time,
                                    )}</span
                                >
                            </span>
                            {#if event.location_name}
                                <span class="event-meta-row event-location">
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
                            {/if}
                            {#if event.default_venue_name}
                                <span class="event-meta-row event-venue-row">
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="16"
                                        height="16"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        aria-hidden="true"
                                        ><path
                                            d="M3 21h18M5 21V7l8-4v18M19 21V11l-6-4"
                                        ></path><path d="M9 9v0M9 12v0M9 15v0M9 18v0"
                                        ></path></svg
                                    >
                                    <span class="event-venue-name">
                                        {#if venuePageUrl(event, event.default_venue_slug)}
                                            <a
                                                href={venuePageUrl(event, event.default_venue_slug)}
                                                class="event-venue-link"
                                                title="Helyszín részletei">{event.default_venue_name}</a
                                            >
                                        {:else}
                                            {event.default_venue_name}
                                        {/if}
                                    </span>
                                </span>
                            {/if}
                            {#if showProgramLink}
                                <span class="event-meta-row event-program-row">
                                    <span class="event-program-icon" aria-hidden="true">
                                        <svg
                                            xmlns="http://www.w3.org/2000/svg"
                                            width="16"
                                            height="16"
                                            viewBox="0 0 24 24"
                                            fill="none"
                                            stroke="currentColor"
                                            stroke-width="1.75"
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            ><rect
                                                x="3"
                                                y="5"
                                                width="14"
                                                height="13"
                                                rx="1.5"
                                            ></rect><line
                                                x1="3"
                                                y1="9"
                                                x2="17"
                                                y2="9"
                                            ></line><line
                                                x1="7"
                                                y1="3"
                                                x2="7"
                                                y2="6"
                                            ></line><line
                                                x1="13"
                                                y1="3"
                                                x2="13"
                                                y2="6"
                                            ></line><circle
                                                cx="17.5"
                                                cy="17"
                                                r="4.25"
                                                fill="var(--card-bg, #fff)"
                                                stroke="currentColor"
                                            ></circle><line
                                                x1="17.5"
                                                y1="17"
                                                x2="17.5"
                                                y2="14.6"
                                            ></line><line
                                                x1="17.5"
                                                y1="17"
                                                x2="19.6"
                                                y2="17"
                                            ></line><circle
                                                cx="17.5"
                                                cy="17"
                                                r="0.45"
                                                fill="currentColor"
                                                stroke="none"
                                            ></circle></svg
                                        >
                                    </span>
                                    <a
                                        href="/esemenyek/{event.id}#program"
                                        class="event-program-link"
                                        title="{event.title} program"
                                        aria-label="Ugrás a(z) {event.title} című esemény napi programjához"
                                        >Program</a
                                    >
                                </span>
                            {/if}
                        </div>

                        {#if event.description}
                            <p class="event-description">{event.description}</p>
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

    <aside class="news-sidebar events-sidebar" aria-label="Események hónap szerint">
        <div class="news-sidebar-box events-sidebar-months-box">
            <div class="news-sidebar-header events-calendar-header">
                {#if sidebarCalendarMode === "month" && drillMonthYM}
                    <h4 class="news-sidebar-heading events-calendar-drill-heading">
                        {drillMonthTitleShort(drillMonthYM)}
                    </h4>
                    {#if drillYearMonth}
                        <span class="events-calendar-year-static">{drillYearMonth.y}</span>
                    {/if}
                {:else}
                    <h4
                        class="news-sidebar-heading"
                        title="Hónap szerinti szűrés: Olyan hónapok, amelyekben van közelgő esemény. Válassz hónapot a napok megjelenítéséhez."
                    >
                        Hónap szerinti szűrés
                    </h4>
                    {#if !filtersLoading && yearsWithEvents.length > 1}
                        <div class="events-calendar-year-nav" aria-label="Év választása">
                            <button
                                type="button"
                                class="events-calendar-year-btn"
                                disabled={calendarViewYear <= yearsWithEvents[0]}
                                aria-label="Előző év"
                                on:click={() => stepCalendarYear(-1)}>‹</button
                            >
                            <span class="events-calendar-year-label">{calendarViewYear}</span>
                            <button
                                type="button"
                                class="events-calendar-year-btn"
                                disabled={calendarViewYear >=
                                    yearsWithEvents[yearsWithEvents.length - 1]}
                                aria-label="Következő év"
                                on:click={() => stepCalendarYear(1)}>›</button
                            >
                        </div>
                    {:else if !filtersLoading}
                        <span class="events-calendar-year-static">{calendarViewYear}</span>
                    {/if}
                {/if}
            </div>
            {#if filtersLoading}
                <div class="events-year-calendar events-year-calendar--skeleton" aria-hidden="true">
                    {#each Array(12) as _, i (i)}
                        <div class="events-calendar-cell events-calendar-cell--skeleton"></div>
                    {/each}
                </div>
            {:else if sidebarCalendarMode === "month" && drillMonthYM && drillYearMonth}
                <button
                    type="button"
                    class="btn btn-sm events-calendar-clear"
                    class:active={filterMonthKey === null && filterDayKey === null}
                    aria-pressed={filterMonthKey === null && filterDayKey === null}
                    on:click={clearMonthOnly}>Összes hónap</button
                >
                <div
                    class="events-month-day-calendar"
                    role="grid"
                    aria-label={"Napok — " +
                        drillMonthTitleShort(drillMonthYM) +
                        (drillYearMonth ? " " + drillYearMonth.y : "")}
                >
                    <div class="events-day-weekdays" aria-hidden="true">
                        {#each weekdayShortHu as w, wi (wi)}
                            <span class="events-day-weekday">{w}</span>
                        {/each}
                    </div>
                    <div class="events-day-grid">
                        {#each monthDayCells as cell, ci (ci)}
                            {#if cell.type === "pad"}
                                <div class="events-day-cell events-day-cell--pad"></div>
                            {:else}
                                {@const hasEv = eventDaysSet.has(cell.iso)}
                                {#if hasEv}
                                    <button
                                        type="button"
                                        class="events-day-cell events-day-cell--btn"
                                        class:events-day-cell--active={filterDayKey === cell.iso}
                                        aria-pressed={filterDayKey === cell.iso}
                                        aria-label="{formatDayLabelHu(cell.iso)}, eseményekkel"
                                        on:click={() => pickCalendarDay(cell.iso)}>{cell.day}</button
                                    >
                                {:else}
                                    <div
                                        class="events-day-cell events-day-cell--muted"
                                        aria-hidden="true"
                                    >
                                        {cell.day}
                                    </div>
                                {/if}
                            {/if}
                        {/each}
                    </div>
                </div>
            {:else}
                <button
                    type="button"
                    class="btn btn-sm events-calendar-clear"
                    class:active={filterMonthKey === null && filterDayKey === null}
                    aria-pressed={filterMonthKey === null && filterDayKey === null}
                    on:click={clearMonthOnly}>Összes hónap</button
                >
                <div
                    class="events-year-calendar"
                    role="grid"
                    aria-label="Hónapok — {calendarViewYear}"
                >
                    {#each Array(12) as _, i (i)}
                        {@const monthNum = i + 1}
                        {@const cellKey = ymKey(calendarViewYear, monthNum)}
                        {@const hasEvents = eventMonthsSet.has(cellKey)}
                        {@const mlabel = formatMonthShortLikeCard(
                            calendarViewYear,
                            monthNum,
                        )}
                        {#if hasEvents}
                            <button
                                type="button"
                                class="events-calendar-cell"
                                class:events-calendar-cell--active={monthCellIsActive(cellKey)}
                                aria-pressed={monthCellIsActive(cellKey)}
                                aria-label="{mlabel} {calendarViewYear}, eseményekkel megtekinthető"
                                on:click={() => openMonthDrill(monthNum)}
                            >
                                <span class="events-calendar-cell-label">{mlabel}</span>
                            </button>
                        {:else}
                            <div
                                class="events-calendar-cell events-calendar-cell--disabled"
                                role="gridcell"
                                aria-disabled="true"
                                aria-label="{mlabel} {calendarViewYear}, nincs közelgő esemény"
                            >
                                <span class="events-calendar-cell-label">{mlabel}</span>
                            </div>
                        {/if}
                    {/each}
                </div>
                {#if filterOptions.months.length === 0}
                    <p class="events-sidebar-months-empty">
                        Nincs közelgő esemény egy hónapban sem.
                    </p>
                {/if}
            {/if}
        </div>
    </aside>
</div>

<style>
    .events-page-layout {
        margin-top: 1rem;
    }

    .events-top-filters {
        margin-top: 0.5rem;
        width: 100%;
        min-width: 0;
        box-sizing: border-box;
    }

    .events-top-filters .filter-actions {
        min-width: 0;
    }

    .events-top-filters .info-box {
        flex: 1 1 auto;
        min-width: 0;
    }

    .events-top-filters .info-box p {
        margin: 0;
        word-break: break-word;
    }

    @media (max-width: 640px) {
        .events-top-filters .header-tabs {
            flex-wrap: nowrap;
            overflow-x: auto;
            overflow-y: hidden;
            -webkit-overflow-scrolling: touch;
            padding-bottom: 0.35rem;
            margin: 0 -0.15rem;
            padding-left: 0.15rem;
            scrollbar-width: thin;
        }
        .events-top-filters .header-tabs .btn {
            flex-shrink: 0;
        }
    }

    .events-filter-row-second {
        margin-top: 0.65rem;
    }

    .events-header-tabs-label {
        max-width: none;
        width: auto;
        min-width: 7.5rem;
    }

    .events-filter-sep {
        color: var(--text-faint);
        margin: 0 0.15rem;
    }

    .events-calendar-header {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        gap: 0.5rem;
    }

    .events-calendar-drill-heading {
        font-size: 0.75rem;
        font-weight: 600;
        color: var(--text-faint);
        text-transform: uppercase;
        letter-spacing: 0.02em;
        line-height: 1.25;
    }

    .events-sidebar-months-box .news-sidebar-heading {
        margin: 0;
    }

    .events-calendar-year-nav {
        display: flex;
        align-items: center;
        gap: 0.25rem;
    }

    .events-calendar-year-btn {
        width: 1.75rem;
        height: 1.75rem;
        padding: 0;
        border: 1px solid var(--border-color);
        border-radius: 8px;
        background: var(--card-bg);
        color: var(--text-color);
        font-size: 1rem;
        line-height: 1;
        cursor: pointer;
        transition:
            background 0.15s,
            color 0.15s;
    }

    .events-calendar-year-btn:hover:not(:disabled) {
        background: var(--skeleton-bg, #f3f4f6);
        color: var(--szekely-red, #c0392b);
    }

    .events-calendar-year-btn:disabled {
        opacity: 0.35;
        cursor: default;
    }

    .events-calendar-year-label,
    .events-calendar-year-static {
        font-size: 0.95rem;
        font-weight: 700;
        color: var(--text-color);
        min-width: 3.25rem;
        text-align: center;
    }

    .events-calendar-clear {
        width: 100%;
        margin: 0.5rem 0 0.65rem;
        justify-content: center;
    }

    .events-calendar-clear.active {
        background: var(--szekely-green, #357a6f);
        color: #fff;
        border-color: transparent;
    }

    .events-year-calendar {
        display: grid;
        grid-template-columns: repeat(3, minmax(0, 1fr));
        grid-auto-rows: 3rem;
        gap: 0.35rem;
        align-items: stretch;
    }

    .events-year-calendar--skeleton {
        margin-top: 0.35rem;
    }

    .events-calendar-cell {
        box-sizing: border-box;
        display: flex;
        align-items: center;
        justify-content: center;
        min-height: 0;
        height: 100%;
        max-height: 100%;
        padding: 0.25rem 0.2rem;
        overflow: hidden;
        border: 1px solid var(--border-color);
        border-radius: 8px;
        background: var(--card-bg);
        color: var(--text-faint);
        cursor: pointer;
        transition:
            background 0.15s,
            border-color 0.15s,
            color 0.15s;
        user-select: none;
    }

    .events-calendar-cell:hover:not(.events-calendar-cell--disabled):not(
            .events-calendar-cell--active
        ) {
        border-color: var(--primary-color, #5c6bc0);
        color: var(--primary-color, #5c6bc0);
    }

    .events-calendar-cell--active {
        background: var(--primary-color, #5c6bc0);
        color: #fff;
        border-color: transparent;
    }

    .events-calendar-cell--active .events-calendar-cell-label {
        color: #fff;
    }

    .events-calendar-cell--disabled {
        opacity: 0.4;
        cursor: default;
        pointer-events: none;
        background: color-mix(in srgb, var(--border-color) 35%, var(--card-bg));
        color: var(--text-faint);
    }

    .events-calendar-cell--skeleton {
        height: 100%;
        min-height: 3rem;
        border-radius: 8px;
        background: var(--skeleton-bg, #eee);
        animation: pulse 1.2s ease-in-out infinite;
    }

    .events-calendar-cell-label {
        display: -webkit-box;
        -webkit-box-orient: vertical;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        overflow: hidden;
        text-align: center;
        line-height: 1.15;
        max-height: 100%;
        width: 100%;
        min-width: 0;
        font-size: 0.75rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.02em;
        color: inherit;
    }

    .events-sidebar-months-empty {
        margin: 0.65rem 0 0;
        font-size: 0.82rem;
        color: var(--text-faint);
        line-height: 1.45;
    }

    .events-month-day-calendar {
        margin-top: 0.15rem;
    }

    .events-day-weekdays {
        display: grid;
        grid-template-columns: repeat(7, minmax(0, 1fr));
        gap: 0.2rem;
        margin-bottom: 0.3rem;
    }

    .events-day-weekday {
        text-align: center;
        font-size: 0.62rem;
        font-weight: 700;
        color: var(--text-faint);
        text-transform: none;
    }

    .events-day-grid {
        display: grid;
        grid-template-columns: repeat(7, minmax(0, 1fr));
        gap: 0.2rem;
    }

    .events-day-cell {
        box-sizing: border-box;
        min-height: 1.85rem;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: 6px;
        font-size: 0.72rem;
        font-weight: 600;
    }

    .events-day-cell--pad {
        visibility: hidden;
        min-height: 1.85rem;
    }

    .events-day-cell--muted {
        border: 1px dashed color-mix(in srgb, var(--border-color) 55%, transparent);
        color: var(--text-faint);
        opacity: 0.45;
        cursor: default;
    }

    button.events-day-cell--btn {
        border: 1px solid var(--border-color);
        background: var(--card-bg);
        color: var(--text-color);
        cursor: pointer;
        transition:
            background 0.15s,
            border-color 0.15s,
            color 0.15s;
    }

    button.events-day-cell--btn:hover:not(.events-day-cell--active) {
        border-color: var(--primary-color, #5c6bc0);
        color: var(--primary-color, #5c6bc0);
    }

    .events-day-cell--active {
        background: var(--primary-color, #5c6bc0) !important;
        color: #fff !important;
        border-color: transparent !important;
    }

    .event-card-link {
        text-decoration: none;
        color: inherit;
    }

    .event-list-card {
        display: flex;
        flex-direction: column;
    }

    .event-card-head {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        gap: 0.5rem;
        width: 100%;
        margin: 0 0 0.25rem;
        min-width: 0;
    }

    .event-card-head-left {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: 0.3rem;
        min-width: 0;
    }

    .event-card-type-row {
        width: 100%;
        min-width: 0;
    }

    .event-card-head-right {
        flex-shrink: 0;
        margin-left: auto;
    }

    .event-type-inline {
        font-size: 0.65rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.03em;
        color: var(--szekely-red, #c0392b);
    }

    .event-start-datetime {
        font-size: 0.75rem;
        font-weight: 600;
        color: var(--text-faint);
        text-transform: uppercase;
        letter-spacing: 0.02em;
    }

    .event-title {
        margin: 0 0 0.35rem;
    }

    .event-meta {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        font-size: 0.9rem;
        color: var(--text-faint);
        margin: 0.85rem 0 1rem;
    }

    .event-meta-row {
        display: flex;
        gap: 0.5rem;
        align-items: flex-start;
        flex-direction: row;
        flex-wrap: wrap;
    }

    .event-meta-row svg {
        flex-shrink: 0;
    }

    .event-program-icon {
        display: flex;
        flex-shrink: 0;
        color: var(--text-faint);
    }

    .event-program-icon svg {
        display: block;
    }

    .event-meta-date-text {
        line-height: 1.45;
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

    .events-page-layout :global(.list.grid) {
        box-sizing: border-box;
    }

    @media (max-width: 1100px) {
        .events-page-layout :global(.list.grid) {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }

    @media (max-width: 560px) {
        .events-page-layout :global(.list.grid) {
            grid-template-columns: 1fr;
        }
    }

    @media (max-width: 560px) {
        .events-page-layout {
            margin-top: 0.5rem;
        }
        .events-sidebar .news-sidebar-box {
            box-sizing: border-box;
        }
    }
</style>
