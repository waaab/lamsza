<script>
    import { onMount } from "svelte";
    import EventDateBadge from "$lib/components/EventDateBadge.svelte";
    import {
        venuePageUrl,
        formatDateShort,
        formatMonthShortLikeCard,
        formatDateWithOptionalTime,
    } from "$lib/utils";
    import { eventFeaturedImageUrl } from "$lib/eventImage.js";
    import { eventEntryPriceText } from "$lib/eventEntryPrice.js";
    import { apiFetch, getApiBase } from "$lib/api.js";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";
    import { kindLabel } from "$lib/venueKindLabels.js";

    /** @param {string | null | undefined} iso */
    function calendarChipFromYMD(iso) {
        if (!iso) return { month: "—", day: "—" };
        const part = String(iso).split("T")[0];
        const [y, m, d] = part.split("-").map((x) => parseInt(x, 10));
        if (!y || !m || !d) return { month: "—", day: "—" };
        const mo = formatMonthShortLikeCard(y, m).replace(/\./g, "").trim();
        const monthUpper = mo.toLocaleUpperCase("hu-HU");
        return { month: monthUpper.slice(0, 3), day: String(d) };
    }

    /** Plain text preview for card lede (strip tags, collapse whitespace). */
    /** @param {unknown} raw */
    function eventDescriptionPlain(raw) {
        if (raw == null) return "";
        return String(raw)
            .replace(/<[^>]+>/g, " ")
            .replace(/\s+/g, " ")
            .trim();
    }

    /** Short lede for list cards (ellipsis when longer). */
    /** @param {unknown} raw @param {number} max */
    function truncateEventCardDescription(raw, max = 160) {
        const s = eventDescriptionPlain(raw);
        if (!s) return "";
        if (s.length <= max) return s;
        return s.slice(0, max).trimEnd() + "…";
    }

    let pageHeader = initialPageHeader("esemenyek");
    let pageHeaderLoading = false;

    let events = [];
    let loading = true;
    let error = null;
    let total = 0;
    /** Same batch size as index / hirek lists */
    const PAGE_BATCH = 12;
    /** One grid row (desktop). Avoid a full page of placeholders for a short list. */
    const SKELETON_COUNT = 3;
    let visibleCount = PAGE_BATCH;

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

    /** Facebook-style top filter dropdowns */
    let locationDropdownOpen = false;
    let typeDropdownOpen = false;
    let locationSearchQuery = "";
    let typeSearchQuery = "";

    const sortLabels = {
        start_date: "Dátum (1 → 2)",
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

    /** 12 short month labels (hu-HU) — static, never fetched. */
    const calendarMonthLabels = Array.from({ length: 12 }, (_, i) =>
        formatMonthShortLikeCard(2000, i + 1),
    );

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

    /**
     * YYYY-MM keys that have upcoming events (for enabling cells).
     * Month *names* are local (`calendarMonthLabels`); the API only marks occupancy.
     */
    $: eventMonthsSet = (() => {
        const set = new Set(filterOptions.months || []);
        for (const ev of events) {
            if (!ev?.start_date) continue;
            const start = String(ev.start_date).slice(0, 10);
            const end =
                ev.end_date && String(ev.end_date).trim() !== ""
                    ? String(ev.end_date).slice(0, 10)
                    : start;
            const y0 = parseInt(start.slice(0, 4), 10);
            const m0 = parseInt(start.slice(5, 7), 10);
            const y1 = parseInt(end.slice(0, 4), 10);
            const m1 = parseInt(end.slice(5, 7), 10);
            if (!y0 || !m0 || !y1 || !m1) continue;
            if (y1 < y0 || (y1 === y0 && m1 < m0)) continue;
            let cy = y0;
            let cm = m0;
            for (;;) {
                set.add(`${cy}-${String(cm).padStart(2, "0")}`);
                if (cy === y1 && cm === m1) break;
                cm++;
                if (cm > 12) {
                    cm = 1;
                    cy++;
                }
            }
        }
        return set;
    })();

    $: yearsWithEvents = (() => {
        const ys = new Set();
        for (const raw of eventMonthsSet) {
            const y = parseInt(String(raw).slice(0, 4), 10);
            if (!Number.isNaN(y)) ys.add(y);
        }
        return Array.from(ys).sort((a, b) => a - b);
    })();
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

    function buildQueryParams() {
        const p = new URLSearchParams();
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
            const data = await apiFetch("/api/events/filter-options");
            filterOptions = {
                event_types: data.event_types || [],
                locations: data.locations || [],
                months: data.months || [],
                event_days: data.event_days || [],
                schedule_event_ids: data.schedule_event_ids || [],
            };
        } catch (e) {
            console.error(e);
        } finally {
            filtersLoading = false;
        }
    }

    async function loadEvents() {
        loading = true;
        error = null;
        events = [];
        total = 0;
        try {
            const qs = buildQueryParams();
            const res = await fetch(`${getApiBase()}/api/events?${qs}`);
            if (res.ok) {
                const data = await res.json();
                events = data.events || [];
                total = data.total || 0;
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

    function loadMore() {
        visibleCount += PAGE_BATCH;
    }

    function applyFilters() {
        scrollToTop();
        loadEvents();
    }

    $: {
        filterType;
        filterLocationSlug;
        filterMonthKey;
        filterDayKey;
        sortMode;
        visibleCount = PAGE_BATCH;
    }

    $: displayEvents = events.slice(0, visibleCount);

    function clearAllFilters() {
        filterType = null;
        filterLocationSlug = null;
        filterMonthKey = null;
        filterDayKey = null;
        closeMonthDrill();
        applyFilters();
    }

    function clearMonthOnly() {
        if (filterMonthKey === null && filterDayKey === null) return;
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
        loadPageMeta("esemenyek").then((p) => {
            pageHeader = p;
            pageHeaderLoading = false;
        });
        loadFilterOptions();
        loadEvents();
        const closeFloatingMenus = () => {
            sortOpen = false;
            locationDropdownOpen = false;
            typeDropdownOpen = false;
        };
        window.addEventListener("click", closeFloatingMenus);
        return () => window.removeEventListener("click", closeFloatingMenus);
    });

    $: hasActiveFilters =
        filterType != null ||
        filterLocationSlug != null ||
        filterMonthKey != null ||
        filterDayKey != null;

    /** API filter-options + unique rows from loaded events (fallback if filter-options fails). */
    $: mergedLocationsForFilter = (() => {
        const bySlug = new Map();
        for (const loc of filterOptions.locations || []) {
            if (loc?.location_slug) {
                bySlug.set(loc.location_slug, {
                    name: loc.name,
                    location_slug: loc.location_slug,
                    county_slug: loc.county_slug || "",
                    county_name: loc.county_name || "",
                });
            }
        }
        for (const ev of events) {
            const slug = ev?.location_slug;
            if (!slug || bySlug.has(slug)) continue;
            bySlug.set(slug, {
                name: ev.location_name || slug,
                location_slug: slug,
                county_slug: ev.county_slug || "",
                county_name: ev.county || "",
            });
        }
        return Array.from(bySlug.values()).sort((a, b) =>
            a.name.localeCompare(b.name, "hu"),
        );
    })();

    /** Distinct event type slugs from API + loaded events. */
    $: mergedEventTypeSlugsForFilter = (() => {
        const set = new Set(filterOptions.event_types || []);
        for (const ev of events) {
            const t = ev?.event_type;
            if (t && String(t).trim() !== "") set.add(t);
        }
        return Array.from(set).sort((a, b) => a.localeCompare(b));
    })();

    $: locationFilterLabel =
        mergedLocationsForFilter.find((l) => l.location_slug === filterLocationSlug)
            ?.name || filterLocationSlug;

    const EVENT_TYPE_LABELS = {
        cultural: "Kulturális",
        sports: "Sport",
        festival: "Fesztivál",
        religious: "Vallási",
        other: "Egyéb",
    };

    /** Részvétel / hozzáférés (API: access_type) */
    const ACCESS_LABELS = {
        public: "Nyitott",
        members_only: "Zártkörű",
        invitation_only: "Meghívóval",
    };

    $: filteredLocationsForFilter = mergedLocationsForFilter.filter((loc) => {
        const q = locationSearchQuery.trim().toLowerCase();
        if (!q) return true;
        return (
            loc.name.toLowerCase().includes(q) ||
            (loc.county_name &&
                String(loc.county_name).toLowerCase().includes(q))
        );
    });

    $: filteredEventTypesForFilter = mergedEventTypeSlugsForFilter.filter((t) => {
        const q = typeSearchQuery.trim().toLowerCase();
        if (!q) return true;
        const label = EVENT_TYPE_LABELS[t] || t;
        return (
            label.toLowerCase().includes(q) || String(t).toLowerCase().includes(q)
        );
    });

    $: locationPillText =
        filterLocationSlug == null
            ? "Összes település"
            : locationFilterLabel || "Település";

    $: typePillText =
        filterType == null
            ? "Minden típus"
            : EVENT_TYPE_LABELS[filterType] || filterType;
</script>

<PublicPageHero
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    loading={pageHeaderLoading}
    breadcrumbLabel="Események"
    documentTitleSuffix=" - Székely Gugel"
/>

<div class="header-tabs">
    <span class="header-tabs-label" aria-label="Szűrés településre">Szűrés településre:</span>
    {#if filtersLoading}
        <span class="btn btn--loading">Szűrők betöltése…</span>
    {:else}
    <div class="header-tabs-filters-row">
        <div class="dropdown">
            <button
                type="button"
                class="btn"
                aria-haspopup="listbox"
                aria-expanded={locationDropdownOpen}
                on:click|stopPropagation={() => {
                    typeDropdownOpen = false;
                    locationDropdownOpen = !locationDropdownOpen;
                }}
            >
                <span class="btn-icon" aria-hidden="true">
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        ><path
                            d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"
                        ></path><circle cx="12" cy="10" r="3"></circle></svg>
                </span>
                <span class="btn-label">{locationPillText}</span>
                <span
                    class="btn-chevron"
                    class:btn-chevron--open={locationDropdownOpen}
                    aria-hidden="true"
                    ><svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        ><polyline points="6 9 12 15 18 9"></polyline></svg>
                </span>
            </button>
            {#if locationDropdownOpen}
                <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
                <div
                    class="events-fb-menu"
                    role="listbox"
                    tabindex="0"
                    aria-label="Település választása"
                    on:click|stopPropagation
                >
                    <input
                        class="events-fb-menu-search"
                        type="search"
                        name="location-search"
                        id="location-search"
                        placeholder="Keresés településre…"
                        bind:value={locationSearchQuery}
                        autocomplete="off"
                    />
                    <ul class="events-fb-menu-list">
                        <li class="events-fb-menu-item">
                            <button
                                type="button"
                                class="events-fb-menu-row"
                                class:events-fb-menu-row--active={filterLocationSlug ===
                                    null}
                                role="option"
                                aria-selected={filterLocationSlug === null}
                                on:click={() => {
                                    filterLocationSlug = null;
                                    locationDropdownOpen = false;
                                    locationSearchQuery = "";
                                    applyFilters();
                                }}
                            >
                                <span class="events-fb-menu-row-label"
                                    >Összes település</span
                                >
                                <span class="events-fb-menu-radio" aria-hidden="true"
                                ></span>
                            </button>
                        </li>
                        {#each filteredLocationsForFilter as loc (loc.location_slug + loc.county_slug)}
                            <li class="events-fb-menu-item">
                                <button
                                    type="button"
                                    class="events-fb-menu-row"
                                    class:events-fb-menu-row--active={filterLocationSlug ===
                                        loc.location_slug}
                                    role="option"
                                    aria-selected={filterLocationSlug ===
                                        loc.location_slug}
                                    title="{loc.name}, {loc.county_name} megye"
                                    on:click={() => {
                                        filterLocationSlug = loc.location_slug;
                                        locationDropdownOpen = false;
                                        locationSearchQuery = "";
                                        applyFilters();
                                    }}
                                >
                                    <span class="events-fb-menu-row-label">
                                        {loc.name}
                                        <span class="events-fb-menu-row-sub"
                                            >{loc.county_name} megye</span
                                        >
                                    </span>
                                    <span class="events-fb-menu-radio" aria-hidden="true"
                                    ></span>
                                </button>
                            </li>
                        {/each}
                    </ul>
                </div>
            {/if}
        </div>

        <div class="dropdown">
            <button
                type="button"
                class="btn"
                aria-haspopup="listbox"
                aria-expanded={typeDropdownOpen}
                on:click|stopPropagation={() => {
                    locationDropdownOpen = false;
                    typeDropdownOpen = !typeDropdownOpen;
                }}
            >
                <span class="btn-icon" aria-hidden="true">
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        ><rect x="3" y="4" width="18" height="18" rx="2" ry="2"
                        ></rect><line x1="16" y1="2" x2="16" y2="6"></line><line
                            x1="8"
                            y1="2"
                            x2="8"
                            y2="6"
                        ></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                </span>
                <span class="btn-label">{typePillText}</span>
                <span
                    class="btn-chevron"
                    class:btn-chevron--open={typeDropdownOpen}
                    aria-hidden="true"
                    ><svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        ><polyline points="6 9 12 15 18 9"></polyline></svg>
                </span>
            </button>
            {#if typeDropdownOpen}
                <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
                <div
                    class="events-fb-menu"
                    role="listbox"
                    tabindex="0"
                    aria-label="Eseménytípus választása"
                    on:click|stopPropagation
                >
                    <input
                        class="events-fb-menu-search"
                        type="search"
                        name="type-search"
                        id="type-search"
                        placeholder="Keresés típusra…"
                        bind:value={typeSearchQuery}
                        autocomplete="off"
                    />
                    <ul class="events-fb-menu-list">
                        <li class="events-fb-menu-item">
                            <button
                                type="button"
                                class="events-fb-menu-row"
                                class:events-fb-menu-row--active={filterType === null}
                                role="option"
                                aria-selected={filterType === null}
                                on:click={() => {
                                    filterType = null;
                                    typeDropdownOpen = false;
                                    typeSearchQuery = "";
                                    applyFilters();
                                }}
                            >
                                <span class="events-fb-menu-row-label"
                                    >Minden típus</span
                                >
                                <span class="events-fb-menu-radio" aria-hidden="true"
                                ></span>
                            </button>
                        </li>
                        {#each filteredEventTypesForFilter as t (t)}
                            <li class="events-fb-menu-item">
                                <button
                                    type="button"
                                    class="events-fb-menu-row"
                                    class:events-fb-menu-row--active={filterType === t}
                                    role="option"
                                    aria-selected={filterType === t}
                                    on:click={() => {
                                        filterType = t;
                                        typeDropdownOpen = false;
                                        typeSearchQuery = "";
                                        applyFilters();
                                    }}
                                >
                                    <span class="events-fb-menu-row-label"
                                        >{EVENT_TYPE_LABELS[t] || t}</span
                                    >
                                    <span class="events-fb-menu-radio" aria-hidden="true"
                                    ></span>
                                </button>
                            </li>
                        {/each}
                    </ul>
                </div>
            {/if}
        </div>
    </div>
    {/if}
</div>

{#snippet eventsFilterBar()}
    <div class="filter-actions">
        <span class="info-box">
            <p>
                {#if !hasActiveFilters}
                    💡 Leszűrve: <span class="active">Összes</span>
                {:else}
                    🔍 Aktív szűrők:
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
            <p><span>({displayEvents.length}/{total || 0})</span></p>
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
                            >Dátum (1 → 2)</button
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
{/snippet}

{@render eventsFilterBar()}

<div class="list-page-layout">
    <section class="list">
        {#if loading}
            <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
                {#each Array(SKELETON_COUNT) as _, i (i)}
                    <article
                        class="card event event--skeleton"
                        class:event--grid={viewMode === "grid"}
                        class:event--flex={viewMode === "flex"}
                    >
                        <div class="event-card-media event-card-media--skeleton skeleton"></div>
                        <header class="event-card-block event-card-header">
                            <div class="event-card-header-row">
                                <div class="event-calendar-chip event-calendar-chip--skeleton skeleton"></div>
                                <div class="event-card-header-main">
                                    <div class="skeleton skeleton-text event-skeleton-title"></div>
                                    <div class="skeleton skeleton-text event-skeleton-line"></div>
                                </div>
                            </div>
                        </header>
                        <div class="event-card-block event-card-body">
                            <div class="skeleton skeleton-text event-skeleton-line"></div>
                            <div class="skeleton skeleton-text event-skeleton-line event-skeleton-line--short"></div>
                        </div>
                        <footer class="event-card-block event-card-footer" aria-hidden="true">
                            <div class="event-card-footer__inner">
                                <div class="event-card-footer__meta">
                                    <div class="event-card-footer__row">
                                        <span class="skeleton skeleton-text event-skeleton-footer-line"></span>
                                    </div>
                                    <div class="event-card-footer__row">
                                        <span class="skeleton skeleton-text event-skeleton-footer-line"></span>
                                    </div>
                                    <div class="event-card-footer__row">
                                        <span class="skeleton skeleton-text event-skeleton-footer-line"></span>
                                    </div>
                                </div>
                                <p class="event-card-footer__lede event-card-footer__lede--skeleton">
                                    <span class="skeleton skeleton-text event-skeleton-lede"></span>
                                </p>
                            </div>
                        </footer>
                    </article>
                {/each}
            </div>
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
                {#each displayEvents as event}
                    {@const endDateForCard =
                        event.end_date && String(event.end_date).trim() !== ""
                            ? event.end_date
                            : event.start_date}
                    {@const showProgramLink =
                        event.has_schedule === true ||
                        scheduleEventIdsSet.has(event.id)}
                    {@const chip = calendarChipFromYMD(event.start_date)}
                    {@const eventCardLede = truncateEventCardDescription(
                        event.description,
                    )}
                    <article class="card event">
                        <p class="sr-only">
                            {event.title}. Kezdés:
                            {formatDateWithOptionalTime(
                                event.start_date,
                                event.start_time,
                            )}.
                            {#if endDateForCard !== event.start_date}
                                Befejezés:
                                {formatDateWithOptionalTime(
                                    endDateForCard,
                                    event.end_time,
                                )}.
                            {/if}
                        </p>

                        <div class="event-card-media">
                            <a
                                href="/esemenyek/{event.id}"
                                class="event-card-media-link"
                                tabindex="-1"
                                aria-hidden="true"
                            >
                                <div class="img-wrap">
                                    <img
                                        class="img"
                                        src={eventFeaturedImageUrl(event)}
                                        alt=""
                                        loading="lazy"
                                        decoding="async"
                                    />
                                </div>
                            </a>
                            {#if event.event_type || event.event_subtype_label || event.event_subtype}
                                <div class="event-card-media-type-row">
                                    {#if event.event_type}
                                        <span class="event-card-media-type"
                                            >{event.event_type_label ||
                                                EVENT_TYPE_LABELS[event.event_type] ||
                                                event.event_type}</span
                                        >
                                    {/if}
                                    {#if event.event_subtype_label || event.event_subtype}
                                        <span class="event-card-media-subtype"
                                            >{event.event_subtype_label ||
                                                event.event_subtype}</span
                                        >
                                    {/if}
                                </div>
                            {/if}
                            <div class="event-card-media-badge">
                                <EventDateBadge
                                    event={event}
                                    live={true}
                                    corner={true}
                                />
                            </div>
                        </div>

                        <header class="event-card-block event-card-header">
                            <div class="event-card-header-row">
                                <div
                                    class="event-calendar-chip"
                                    aria-hidden="true"
                                    title="Kezdés napja"
                                >
                                    <span class="event-calendar-chip-month"
                                        >{chip.month}</span
                                    >
                                    <span class="event-calendar-chip-day"
                                        >{chip.day}</span
                                    >
                                </div>
                                <div class="event-card-header-main">
                                    <h2 class="event-card-title">
                                        <a
                                            href="/esemenyek/{event.id}"
                                            class="event-card-link"
                                            title={viewMode === "grid"
                                                ? event.title
                                                : undefined}>{event.title}</a
                                        >
                                    </h2>
                                    {#if event.location_name}
                                        <p class="event-card-lede">
                                            <a
                                                href="/{event.county_slug}-megye/{event.location_slug}"
                                                class="event-card-lede-link"
                                                >{event.location_name}</a
                                            >
                                        </p>
                                    {/if}
                                </div>
                            </div>
                        </header>

                        <div class="event-card-block event-card-body">
                            <div class="event-meta">
                                <span class="event-meta-row event-meta-row--date" aria-label="Kezdés">
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
                                        ><circle cx="12" cy="12" r="10"></circle><polyline
                                            points="12 6 12 12 16 14"
                                        ></polyline></svg
                                    >
                                    <span class="event-meta-date-text"
                                        ><span class="sr-only">Kezdés: </span>{formatDateWithOptionalTime(
                                            event.start_date,
                                            event.start_time,
                                        )}</span
                                    >
                                </span>
                                <span class="event-meta-row event-meta-row--date" aria-label="Befejezés">
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
                                        ><circle cx="12" cy="12" r="10"></circle><polyline
                                            points="12 6 12 12 16 14"
                                        ></polyline></svg
                                    >
                                    <span class="event-meta-date-text"
                                        ><span class="sr-only">Befejezés: </span>{formatDateWithOptionalTime(
                                            endDateForCard,
                                            event.end_time,
                                        )}</span
                                    >
                                </span>
                            </div>
                        </div>

                        <footer class="event-card-block event-card-footer" aria-label="Helyszín, belépő, program és rövid leírás">
                            <div class="event-card-footer__inner">
                                <div class="event-card-footer__meta">
                                    <div class="event-card-footer__row">
                                        <span class="event-card-footer__icon" aria-hidden="true">
                                            <svg
                                                xmlns="http://www.w3.org/2000/svg"
                                                width="16"
                                                height="16"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                stroke-width="2"
                                                ><path
                                                    d="M3 21h18M5 21V7l8-4v18M19 21V11l-6-4"
                                                ></path><path d="M9 9v0M9 12v0M9 15v0M9 18v0"
                                                ></path></svg
                                            >
                                        </span>
                                        <div class="event-card-footer__text">
                                            {#if event.default_venue_name}
                                                {#if venuePageUrl(event, event.default_venue_slug)}
                                                    <a
                                                        href={venuePageUrl(event, event.default_venue_slug)}
                                                        class="event-venue-link"
                                                        title="Helyszín részletei"
                                                    >
                                                        {event.default_venue_name}
                                                        {#if kindLabel(event.default_venue_kind, event.default_venue_kind_label)}
                                                            {' '}{kindLabel(event.default_venue_kind, event.default_venue_kind_label)}
                                                        {/if}
                                                    </a>
                                                {:else}
                                                    {event.default_venue_name}
                                                    {#if kindLabel(event.default_venue_kind, event.default_venue_kind_label)}
                                                        {' '}{kindLabel(event.default_venue_kind, event.default_venue_kind_label)}
                                                    {/if}
                                                {/if}
                                            {:else}
                                                <span class="event-card-footer__muted">-</span>
                                            {/if}
                                        </div>
                                    </div>
                                    <div class="event-card-footer__row">
                                        <span class="event-card-footer__icon" aria-hidden="true">
                                            <svg xmlns="http://www.w3.org/2000/svg" height="16px" viewBox="0 -960 960 960" width="16px" fill="currentColor"><path d="M600-120q-118 0-210-67T260-360H120v-80h122q-3-24-2.5-44.5T242-520H120v-80h140q38-106 130-173t210-67q69 0 130.5 24.5T840-748l-57 56q-37-32-83.5-50T600-760q-85 0-152 44.5T347-600h253v80H323q-4 27-3 47.5t3 32.5h277v80H347q34 71 101 115.5T600-200q53 0 99.5-18t83.5-50l57 56q-48 43-109.5 67.5T600-120Z"/></svg>
                                        </span>
                                        <div class="event-card-footer__text">
                                            <span
                                                class="event-card-footer__price-text"
                                                title={event.entry_price &&
                                                String(event.entry_price).trim()
                                                    ? `Belépő: ${String(event.entry_price).trim()}`
                                                    : "Belépő: nincs megadva"}
                                                >{eventEntryPriceText(event.entry_price)}</span
                                            >
                                        </div>
                                    </div>
                                    <div class="event-card-footer__row">
                                        <span class="event-card-footer__icon" aria-hidden="true">
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
                                        <div class="event-card-footer__text">
                                            {#if showProgramLink}
                                                <a
                                                    href="/esemenyek/{event.id}#program"
                                                    class="event-program-link"
                                                    title="{event.title} program"
                                                    aria-label="Ugrás a(z) {event.title} című esemény napi programjához"
                                                    >Program</a
                                                >
                                            {:else}
                                                <span class="event-card-footer__muted">-</span>
                                            {/if}
                                        </div>
                                    </div>
                                    <div class="event-card-footer__row">
                                        <span class="event-card-footer__icon" aria-hidden="true">
                                            <svg
                                                xmlns="http://www.w3.org/2000/svg"
                                                width="16"
                                                height="16"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                stroke-width="1.75"
                                            >
                                                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
                                            </svg>
                                            
                                        </span>
                                        <div class="event-card-footer__text">
                                            <span class="event-meta-row event-meta-row--access" aria-label="Hozzáférés">
                                                <span
                                                    title="Ki vehet részt"
                                                    >{ACCESS_LABELS[event.access_type] ||
                                                        "Nyitott"}</span
                                                >
                                            </span>
                                        </div>
                                    </div>
                                </div>
                                {#if eventCardLede}
                                    <p
                                        class="event-card-footer__lede"
                                        title={eventDescriptionPlain(event.description)}
                                    >
                                        {eventCardLede}
                                    </p>
                                {/if}
                            </div>
                        </footer>
                    </article>
                {/each}
            </div>

            {#if visibleCount < events.length}
                <div class="load-more">
                    <button type="button" class="btn nav-btn" on:click={loadMore}>
                        Több betöltése ↓
                    </button>
                </div>
            {/if}
        {/if}
    </section>

    <aside class="sidebar events-sidebar" aria-label="Események hónap szerint">
        <div class="sidebar-box events-sidebar-months-box">
            <div class="sidebar-header events-calendar-header">
                {#if sidebarCalendarMode === "month" && drillMonthYM}
                    <h4 class="sidebar-heading">
                        {drillMonthTitleShort(drillMonthYM)}
                    </h4>
                    {#if drillYearMonth}
                            <span class="events-calendar-year-static">{drillYearMonth.y}</span>
                        {/if}
                {:else}
                    <h4
                        class="sidebar-heading"
                        title="A Naptár nézet: Olyan hónapok, amelyekben van közelgő esemény. Válassz hónapot a napok megjelenítéséhez."
                    >
                        Naptár nézet
                    </h4>
                    {#if yearsWithEvents.length > 1}
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
                    {:else}
                        <span class="events-calendar-year-static">{calendarViewYear}</span>
                    {/if}
                {/if}
            </div>
            {#if sidebarCalendarMode === "month" && drillMonthYM && drillYearMonth}
                <button
                    type="button"
                    class="btn btn-sm events-calendar-clear"
                    class:active={filterMonthKey === null && filterDayKey === null}
                    aria-pressed={filterMonthKey === null && filterDayKey === null}
                    disabled={filterMonthKey === null && filterDayKey === null}
                    on:click={clearMonthOnly}>
                    <svg xmlns="http://www.w3.org/2000/svg" height="16px" viewBox="0 -960 960 960" width="16px" fill="currentColor"><path d="M216-96q-29.7 0-50.85-21.5Q144-139 144-168v-528q0-29 21.15-50.5T216-768h72v-96h72v96h240v-96h72v96h72q29.7 0 50.85 21.5Q816-725 816-696v528q0 29-21.15 50.5T744-96H216Zm0-72h528v-360H216v360Zm0-432h528v-96H216v96Zm0 0v-96 96Zm264.21 216q-15.21 0-25.71-10.29t-10.5-25.5q0-15.21 10.29-25.71t25.5-10.5q15.21 0 25.71 10.29t10.5 25.5q0 15.21-10.29 25.71t-25.5 10.5ZM298.5-394.29q-10.5-10.29-10.5-25.5t10.29-25.71q10.29-10.5 25.5-10.5t25.71 10.29q10.5 10.29 10.5 25.5t-10.29 25.71q-10.29 10.5-25.5 10.5t-25.71-10.29ZM636.21-384q-15.21 0-25.71-10.29t-10.5-25.5q0-15.21 10.29-25.71t25.5-10.5q15.21 0 25.71 10.29t10.5 25.5q0 15.21-10.29 25.71t-25.5 10.5Zm-156 144q-15.21 0-25.71-10.29t-10.5-25.5q0-15.21 10.29-25.71t25.5-10.5q15.21 0 25.71 10.29t10.5 25.5q0 15.21-10.29 25.71t-25.5 10.5ZM298.5-250.29q-10.5-10.29-10.5-25.5t10.29-25.71q10.29-10.5 25.5-10.5t25.71 10.29q10.5 10.29 10.5 25.5t-10.29 25.71q-10.29 10.5-25.5 10.5t-25.71-10.29ZM636.21-240q-15.21 0-25.71-10.29t-10.5-25.5q0-15.21 10.29-25.71t25.5-10.5q15.21 0 25.71 10.29t10.5 25.5q0 15.21-10.29 25.71t-25.5 10.5Z"/></svg>
                    <span class="btn-label">Összes hónap</span></button
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
                                        class="btn btn-sm events-day-cell"
                                        class:active={filterDayKey === cell.iso}
                                        aria-pressed={filterDayKey === cell.iso}
                                        aria-label="{formatDayLabelHu(cell.iso)}, eseményekkel"
                                        on:click={() => pickCalendarDay(cell.iso)}>{cell.day}</button
                                    >
                                {:else}
                                    <div
                                        class="btn btn-sm events-day-cell muted"
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
                    disabled={filterMonthKey === null && filterDayKey === null}
                    on:click={clearMonthOnly}>
                    <svg xmlns="http://www.w3.org/2000/svg" height="16px" viewBox="0 -960 960 960" width="16px" fill="currentColor"><path d="M216-96q-29.7 0-50.85-21.5Q144-139 144-168v-528q0-29 21.15-50.5T216-768h72v-96h72v96h240v-96h72v96h72q29.7 0 50.85 21.5Q816-725 816-696v528q0 29-21.15 50.5T744-96H216Zm0-72h528v-360H216v360Zm0-432h528v-96H216v96Zm0 0v-96 96Zm264.21 216q-15.21 0-25.71-10.29t-10.5-25.5q0-15.21 10.29-25.71t25.5-10.5q15.21 0 25.71 10.29t10.5 25.5q0 15.21-10.29 25.71t-25.5 10.5ZM298.5-394.29q-10.5-10.29-10.5-25.5t10.29-25.71q10.29-10.5 25.5-10.5t25.71 10.29q10.5 10.29 10.5 25.5t-10.29 25.71q-10.29 10.5-25.5 10.5t-25.71-10.29ZM636.21-384q-15.21 0-25.71-10.29t-10.5-25.5q0-15.21 10.29-25.71t25.5-10.5q15.21 0 25.71 10.29t10.5 25.5q0 15.21-10.29 25.71t-25.5 10.5Zm-156 144q-15.21 0-25.71-10.29t-10.5-25.5q0-15.21 10.29-25.71t25.5-10.5q15.21 0 25.71 10.29t10.5 25.5q0 15.21-10.29 25.71t-25.5 10.5ZM298.5-250.29q-10.5-10.29-10.5-25.5t10.29-25.71q10.29-10.5 25.5-10.5t25.71 10.29q10.5 10.29 10.5 25.5t-10.29 25.71q-10.29 10.5-25.5 10.5t-25.71-10.29ZM636.21-240q-15.21 0-25.71-10.29t-10.5-25.5q0-15.21 10.29-25.71t25.5-10.5q15.21 0 25.71 10.29t10.5 25.5q0 15.21-10.29 25.71t-25.5 10.5Z"/></svg>
                    <span class="btn-label">Összes hónap</span></button
                >
                <div
                    class="events-year-calendar"
                    role="grid"
                    aria-busy={filtersLoading}
                    aria-label="Hónapok — {calendarViewYear}"
                >
                    {#each calendarMonthLabels as mlabel, i (mlabel)}
                        {@const monthNum = i + 1}
                        {@const cellKey = ymKey(calendarViewYear, monthNum)}
                        {@const hasEvents = eventMonthsSet.has(cellKey)}
                        <button
                            type="button"
                            class="btn btn-sm events-calendar-cell"
                            class:events-calendar-cell--active={hasEvents &&
                                monthCellIsActive(cellKey)}
                            class:events-calendar-cell--disabled={!hasEvents}
                            class:events-calendar-cell--skeleton={filtersLoading &&
                                !hasEvents}
                            disabled={!hasEvents}
                            aria-pressed={hasEvents
                                ? monthCellIsActive(cellKey)
                                : undefined}
                            aria-label={filtersLoading && !hasEvents
                                ? `${mlabel} ${calendarViewYear}`
                                : hasEvents
                                  ? `${mlabel} ${calendarViewYear}, eseményekkel megtekinthető`
                                  : `${mlabel} ${calendarViewYear}, nincs közelgő esemény`}
                            on:click={() => openMonthDrill(monthNum)}
                        >
                            <span class="events-calendar-cell-label">{mlabel}</span>
                        </button>
                    {/each}
                </div>
                {#if !filtersLoading && eventMonthsSet.size === 0}
                    <p class="events-sidebar-months-empty">
                        Nincs közelgő esemény egy hónapban sem.
                    </p>
                {/if}
            {/if}
        </div>
    </aside>
</div>

{@render eventsFilterBar()}

<style>
    .btn--loading {
        opacity: 0.75;
        cursor: default;
        justify-content: center;
    }

    .btn-icon {
        display: flex;
        flex-shrink: 0;
        opacity: 0.95;
    }

    .btn-label {
        flex: 1;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .btn-chevron {
        display: flex;
        flex-shrink: 0;
        opacity: 0.85;
        transition: transform 0.2s ease;
    }

    .btn-chevron--open {
        transform: rotate(180deg);
    }

    .events-fb-menu {
        position: absolute;
        top: calc(100% + 0.35rem);
        left: 0;
        right: 0;
        z-index: 60;
        display: flex;
        flex-direction: column;
        padding: 0.5rem;
        border-radius: 12px;
        border: 1px solid var(--border-color);
        background: var(--card-bg);
        color: var(--text-primary);
        box-shadow:
            0 4px 12px var(--shadow-md),
            0 12px 28px var(--shadow-lg);
        max-height: min(22rem, 70vh);
    }

    .events-fb-menu-search {
        width: 100%;
        box-sizing: border-box;
        padding: 0.45rem 0.75rem;
        margin-bottom: 0.35rem;
        border-radius: 999px;
        border: 1px solid var(--border-color);
        background: var(--bg-body);
        color: var(--text-primary);
    }

    .events-fb-menu-search::placeholder {
        color: var(--text-muted);
    }

    /* Scroll only when the list is long; uses global scrollbar styling */
    .events-fb-menu-list {
        list-style: none;
        margin: 0;
        padding: 0;
        overflow-y: auto;
        min-height: 0;
        overscroll-behavior: contain;
    }

    .events-fb-menu-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 0.5rem;
        width: 100%;
        padding: 0.5rem 0.35rem;
        border: none;
        border-radius: 8px;
        background: transparent;
        color: inherit;
        cursor: pointer;
        text-align: left;
    }

    .events-fb-menu-row:hover {
        background: var(--tab-hover-bg);
    }

    .events-fb-menu-row--active {
        background: var(--info-note-bg);
    }

    .events-fb-menu-row-label {
        display: flex;
        flex-direction: column;
        gap: 0.1rem;
        min-width: 0;
        line-height: 1.25;
    }

    .events-fb-menu-row-sub {
        font-weight: 400;
        color: var(--text-muted);
    }

    .events-fb-menu-radio {
        width: 1.05rem;
        height: 1.05rem;
        border-radius: 50%;
        border: 2px solid var(--text-muted);
        flex-shrink: 0;
    }

    .events-fb-menu-row--active .events-fb-menu-radio {
        border-color: var(--szekely-blue);
        background: var(--szekely-blue);
        box-shadow: inset 0 0 0 3px var(--card-bg);
    }

    .filter-actions {
        min-width: 0;
    }

    .filter-actions .info-box {
        flex: 1 1 auto;
        min-width: 0;
    }

    .filter-actions .info-box p {
        margin: 0;
        word-break: break-word;
    }

    .events-filter-sep {
        color: var(--text-faint);
        margin: 0 0.15rem;
    }

    .events-calendar-header {
        flex-wrap: wrap;
        gap: 0.5rem;
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
        font-weight: 700;
        min-width: 3.25rem;
        text-align: center;
    }

    .events-calendar-clear {
        width: 100%;
        margin: 0.5rem 0 0.65rem;
        justify-content: center;
    }

    .events-calendar-clear:disabled {
        cursor: default;
        opacity: 1;
    }

    .events-year-calendar {
        display: grid;
        grid-template-columns: repeat(3, minmax(0, 1fr));
        grid-auto-rows: 3rem;
        gap: 0.35rem;
        align-items: stretch;
    }

    .events-calendar-cell--active {
        background: var(--info-note-bg);
        border-color: transparent;
    }

    .events-calendar-cell--active .events-calendar-cell-label {
        color: #fff;
    }

    .events-calendar-cell--disabled,
    .events-calendar-cell:disabled {
        opacity: 0.4;
        cursor: default;
        pointer-events: none;
        background: color-mix(in srgb, var(--border-color) 35%, var(--card-bg));
        color: var(--text-faint);
    }

    .events-calendar-cell--skeleton,
    .events-calendar-cell--skeleton:disabled {
        height: 100%;
        min-height: 3rem;
        border-radius: 8px;
        background: var(--skeleton-bg, #eee);
        animation: pulse 1.2s ease-in-out infinite;
        opacity: 1;
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
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.02em;
        color: inherit;
    }

    .events-sidebar-months-empty {
        margin: 0.65rem 0 0;
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
        font-weight: 700;
        color: var(--text-faint);
        text-transform: none;
    }

    .events-day-grid {
        display: grid;
        grid-template-columns: repeat(7, minmax(0, 1fr));
        gap: 0.2rem;
    }

    .events-day-cell--pad {
        visibility: hidden;
        min-height: 1.85rem;
    }

    .events-day-cell.muted {
        border: 1px dashed color-mix(in srgb, var(--border-color) 55%, transparent);
        color: var(--text-faint);
        opacity: 0.45;
        cursor: default;
    }

    .event-card-media {
        position: relative;
        width: 100%;
        aspect-ratio: 16 / 9;
        background: var(--skeleton-bg, #e8eaef);
    }

    .event-card-media-link {
        display: block;
        width: 100%;
        height: 100%;
    }

    .event-card-media-type-row {
        position: absolute;
        top: 0.65rem;
        left: 0.65rem;
        z-index: 2;
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.35rem;
        max-width: calc(100% - 5.5rem);
    }

    .event-card-media-type {
        padding: 0.15rem 0.45rem;
        border-radius: 4px;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.04em;
        color: #fff;
        background: rgba(0, 0, 0, 0.52);
        border: 1px solid var(--border-color);
        line-height: 1.2;
    }    

    .event-card-media-subtype {
        padding: 0.22rem 0.5rem;
        border-radius: 6px;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.04em;
        line-height: 1.2;
        color: #fff;
        background: rgba(55, 95, 140, 0.82);
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.25);
    }

    .event-card-media-badge {
        position: absolute;
        top: 0.55rem;
        right: 0.55rem;
        z-index: 2;
    }
    .event-card-block {
        padding: 0.85rem 1rem;
        min-height: 58px;
    }

    .event-card-body {
        border-top: 1px solid var(--border-color);
    }

    .event-card-footer {
        margin-top: auto;
        border-top: 1px solid var(--border-color);
        padding: 0;
    }

    .event-card-footer__inner {
        padding: 0.65rem 1rem 0.85rem;
    }

    .event-card-footer__meta {
        display: flex;
        flex-direction: column;
        gap: 0.45rem;
    }

    .event-card-footer__row {
        display: flex;
        align-items: stretch;
        gap: 0.5rem;
        padding: 0;
        line-height: 1.45;
        color: var(--text-faint);
        min-width: 0;
        text-align: left;
        justify-content: flex-start;
    }

    .event-card-footer__lede {
        margin: 0.65rem 0 0;
        padding-top: 0.55rem;
        border-top: 1px solid color-mix(in srgb, var(--border-color) 88%, transparent);
        line-height: 1.45;
        color: var(--text-faint);
        text-align: left;
        display: -webkit-box;
        -webkit-box-orient: vertical;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        overflow: hidden;
        min-height: 37.5px;
    }

    .event-card-footer__lede--skeleton {
        margin-top: 0.65rem;
        padding-top: 0.55rem;
        border-top: 1px solid color-mix(in srgb, var(--border-color) 88%, transparent);
    }

    .event-skeleton-lede {
        display: block;
        width: 100%;
        height: 2.6rem;
        border-radius: 4px;
    }

    .event-card-footer__icon {
        flex-shrink: 0;
        width: 16px;
        min-width: 16px;
        display: flex;
        align-items: center;
        justify-content: flex-start;
        color: var(--text-muted);
    }

    .event-card-footer__icon :global(svg) {
        display: block;
    }

    .event-card-footer__text {
        flex: 1;
        min-width: 0;
        text-align: left;
    }

    .event-card-footer__price-text {
        display: inline-block;
        width: 100%;
        text-align: left;
    }

    .event-card-footer__muted {
        color: var(--text-muted);
    }

    .event-card-header-row {
        display: flex;
        align-items: flex-start;
        gap: 0.75rem;
        min-width: 0;
    }

    .event-calendar-chip {
        flex-shrink: 0;
        width: 3.1rem;
        border-radius: 8px;
        overflow: hidden;
        border: 1px solid var(--border-color);
        text-align: center;
        line-height: 1.15;
        box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
    }

    .event-calendar-chip-month {
        display: block;
        padding: 0.2rem 0.2rem 0.15rem;
        font-weight: 700;
        letter-spacing: 0.06em;
        color: #fff;
        background: var(--szekely-red, #c0392b);
    }

    .event-calendar-chip-day {
        display: block;
        padding: 0.3rem 0.2rem 0.35rem;
        font-weight: 700;
        background: var(--card-bg);
    }

    .event-card-header-main {
        flex: 1;
        min-width: 0;
    }

    .event-card-title {
        margin: 0 0 0.25rem;
        font-weight: 700;
    }

    /* Grid: at most 2 lines for title (never 3) */
    .event-card-link {
        display: -webkit-box;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
        text-decoration: none;
        color: inherit;
    }

    .event-card-lede {
        margin: 0;
        line-height: 1.45;
        color: var(--text-faint);
    }

    .event-meta {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        color: var(--text-faint);
        margin: 0;
    }

    .event-meta-row {
        display: flex;
        gap: 0.5rem;
        align-items: center;
        flex-direction: row;
        flex-wrap: wrap;
    }

    .event-meta-row svg {
        flex-shrink: 0;
    }

    .event-meta-date-text {
        line-height: 1.45;
    }

    .event-meta-row--access {
        gap: 0.4rem;
    }

    .event-calendar-chip--skeleton {
        min-height: 3.25rem;
        align-self: stretch;
        border: none;
        box-shadow: none;
    }

    .event-skeleton-title {
        height: 1.05rem;
        max-width: 100%;
        margin-bottom: 0.4rem;
        border-radius: 4px;
    }

    .event-skeleton-line {
        height: 0.7rem;
        margin-bottom: 0.35rem;
        border-radius: 4px;
    }

    .event-skeleton-line--short {
        max-width: 55%;
    }

    .event-skeleton-footer-line {
        display: block;
        width: 100%;
        height: 0.72rem;
        border-radius: 4px;
    }

    .list-page-layout :global(.list.grid) {
        box-sizing: border-box;
    }

    @media (max-width: 1100px) {
        .list-page-layout :global(.list.grid) {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }

    @media (max-width: 560px) {
        .list-page-layout :global(.list.grid) {
            grid-template-columns: 1fr;
        }
    }

    @media (max-width: 560px) {
        .events-sidebar .sidebar-box {
            box-sizing: border-box;
        }
    }
</style>