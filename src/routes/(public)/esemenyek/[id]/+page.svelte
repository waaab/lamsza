<script>
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { apiFetch } from "$lib/api";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import EventDateBadge from "$lib/components/EventDateBadge.svelte";
    import {
        getReferenceNow,
        isScheduleActivityHappeningNow,
        referenceTodayYMD,
    } from "$lib/eventStatus";
    import { SCHEDULE_ACTIVITY_TYPE_LABELS } from "$lib/scheduleActivityTypes.js";
    import { venuePageUrl } from "$lib/utils";

    let event = null;
    let loading = true;
    let error = null;
    /** `null` = show all schedule days; otherwise YYYY-MM-DD */
    let scheduleFilterDay = /** @type {string | null} */ (null);

    $: eventId = $page.params.id;

    $: if (eventId) {
        loadEvent(eventId);
    }

    async function loadEvent(id) {
        scheduleFilterDay = null;
        loading = true;
        error = null;
        try {
            event = await apiFetch(`/api/events/detail?id=${id}`);
        } catch (err) {
            error = "Az esemény nem található.";
        } finally {
            loading = false;
        }
    }

    /** @param {string} dateStr */
    function normalizeScheduleDateKey(dateStr) {
        if (!dateStr) return "";
        const s = String(dateStr).trim();
        return s.length >= 10 ? s.slice(0, 10) : s;
    }

    /** Bumps on interval so “Ma” chips and “Most” badges follow the homepage clock. */
    let scheduleClockTick = 0;

    onMount(() => {
        scheduleClockTick = 1;
        const id = setInterval(() => {
            scheduleClockTick++;
        }, 30000);
        return () => clearInterval(id);
    });

    /** @type {string} YYYY-MM-DD for “today” per #datetime */
    $: todayYmd = (scheduleClockTick, referenceTodayYMD());

    /** @type {Date} */
    $: referenceNow = (scheduleClockTick, getReferenceNow());

    /** @param {string} dateStr */
    function isScheduleDayToday(dateStr) {
        const key = normalizeScheduleDateKey(dateStr);
        return key !== "" && key === todayYmd;
    }

    /** Visible filter chip: weekday only (e.g. vasárnap). */
    /** @param {string} dateStr */
    function formatScheduleFilterWeekday(dateStr) {
        if (!dateStr) return "";
        const key = normalizeScheduleDateKey(dateStr);
        const d = new Date(key + "T12:00:00");
        if (Number.isNaN(d.getTime())) return key;
        return d.toLocaleDateString("hu-HU", { weekday: "long" });
    }

    /** @param {string} dateStr @param {boolean} isToday */
    function scheduleFilterChipLabel(dateStr, isToday) {
        const w = formatScheduleFilterWeekday(dateStr);
        if (isToday) {
            return w ? `Ma ${w}` : "Ma";
        }
        return w;
    }

    // Inline filter so `scheduleFilterDay` is a reactive dependency (nested fn missed updates in Svelte 5).
    $: filteredScheduleDays =
        event?.schedule?.length && scheduleFilterDay != null
            ? event.schedule.filter(
                  (d) =>
                      normalizeScheduleDateKey(d.schedule_date) ===
                      scheduleFilterDay,
              )
            : event?.schedule?.length
              ? event.schedule
              : [];

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
        if (ev.start_time) res += `, ${ev.start_time.slice(0, 5)}`;

        if (ev.end_date && ev.end_date !== ev.start_date) {
            res += ` — ${formatDate(ev.end_date)}`;
            if (ev.end_time) res += `, ${ev.end_time.slice(0, 5)}`;
        } else if (ev.end_time) {
            res += ` — ${ev.end_time.slice(0, 5)}`;
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

    /** Time part for display: start only (typical for matches), or range if end known. */
    function formatScheduleTimePart(a) {
        const s = a.starts_at ? String(a.starts_at).slice(0, 5) : "";
        const e = a.ends_at ? String(a.ends_at).slice(0, 5) : "";
        if (s && e) return `${s}–${e}`;
        if (s) return s;
        if (e) return e;
        return "";
    }

    /** @param {Record<string, unknown>} ev @param {Record<string, unknown>} act */
    function activityVenueLabel(ev, act) {
        const n = act.venue_name ? String(act.venue_name) : "";
        if (n) return n;
        return ev.default_venue_name ? String(ev.default_venue_name) : "";
    }

    /** @param {Record<string, unknown>} ev @param {Record<string, unknown>} act */
    function activityVenueSlugForLink(ev, act) {
        const vs = act.venue_slug ? String(act.venue_slug).trim() : "";
        if (vs) return vs;
        const vn = act.venue_name ? String(act.venue_name).trim() : "";
        if (!vn && ev.default_venue_slug)
            return String(ev.default_venue_slug).trim();
        return "";
    }

    /**
     * Unique venues: order follows the napi program (activities), then default venue if not already listed.
     * @param {Record<string, unknown>} ev
     * @returns {{ name: string, slug: string }[]}
     */
    function uniqueEventVenuesForMeta(ev) {
        /** @type {{ name: string, slug: string }[]} */
        const out = [];
        const seen = new Set();
        /** @param {string} name @param {string} slug */
        const keyOf = (name, slug) => {
            const s = String(slug || "").trim();
            if (s) return `s:${s}`;
            return `n:${String(name || "").trim().toLowerCase()}`;
        };
        /** @param {string} name @param {string} slug */
        const push = (name, slug) => {
            const n = String(name || "").trim();
            if (!n) return;
            const k = keyOf(n, slug);
            if (seen.has(k)) return;
            seen.add(k);
            out.push({ name: n, slug: String(slug || "").trim() });
        };
        const sched = ev.schedule;
        if (Array.isArray(sched)) {
            for (const day of sched) {
                const acts = day?.activities;
                if (!Array.isArray(acts)) continue;
                for (const act of acts) {
                    push(activityVenueLabel(ev, act), activityVenueSlugForLink(ev, act));
                }
            }
        }
        const defN = String(ev.default_venue_name || "").trim();
        const defS = String(ev.default_venue_slug || "").trim();
        if (defN) {
            push(defN, defS);
        }
        return out;
    }

    $: eventMetaVenues =
        event && typeof event === "object"
            ? uniqueEventVenuesForMeta(/** @type {Record<string, unknown>} */ (event))
            : [];
</script>

<svelte:head>
    <title>{event ? event.title : "Esemény"} - Na Lámsza!</title>
</svelte:head>

{#if loading}
    <span class="info-box"><p>adat betöltés...</p></span>
{:else if error || !event}
    <Breadcrumbs label="Esemény" parentLabel="Események" parentUrl="/esemenyek" />
    <span class="info-box"><p>{error || "Az esemény nem található."}</p></span>
    <a href="/esemenyek" class="back-link">← Vissza az eseményekhez</a>
{:else}
    <Breadcrumbs label={event.title} parentLabel="Események" parentUrl="/esemenyek" />

        <article class="event-detail">
            <div class="event-detail-header">
                <div class="badge event">
                    {EVENT_TYPE_LABELS[event.event_type] || event.event_type}
                </div>
                <h1 class="page-title">{event.title}</h1>
            </div>

            <div class="event-detail-meta">
                <div class="meta-row meta-row--wrap">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                    <span class="event-datetime-with-badge"
                        >{formatEventDateTime(event)}
                        <EventDateBadge event={event} live={true} /></span
                    >
                </div>

                <div class="meta-row">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path><circle cx="12" cy="10" r="3"></circle></svg>
                    <span>
                        <a href="/{event.county_slug}-megye/{event.location_slug}">{event.location_name}</a>,
                        <a href="/{event.county_slug}-megye" class="county-link">{event.county} megye</a>
                    </span>
                </div>

                {#if eventMetaVenues.length > 0}
                    <div class="meta-row meta-row--venue">
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M3 21h18M5 21V7l8-4v18M19 21V11l-6-4"></path><path d="M9 9v0M9 12v0M9 15v0M9 18v0"></path></svg>
                        <span class="meta-row-venue-list">
                            <strong>Helyszín:</strong>
                            {#each eventMetaVenues as v, vi (v.slug ? v.slug : v.name + String(vi))}
                                {#if vi > 0}<span class="meta-venue-sep"> · </span>{/if}
                                {#if v.slug && venuePageUrl(event, v.slug)}
                                    <a href={venuePageUrl(event, v.slug)} title="Helyszín részletei">{v.name}</a>
                                {:else}
                                    {v.name}
                                {/if}
                            {/each}
                        </span>
                    </div>
                {/if}

                {#if event.organizer}
                    <div class="meta-row">
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle></svg>
                        <span><strong>Szervező:</strong> {event.organizer}</span>
                    </div>
                {/if}
            </div>

            {#if event.description}
                <div class="event-detail-body">
                    <p>{event.description}</p>
                </div>
            {/if}
        </article>
    {#if event.schedule && event.schedule.length > 0}
    <section id="program" class="event-schedule" aria-label="Napi program">
        <div class="event-schedule-head">
            <h2 class="event-schedule-title">Napi program</h2>
            <div
                class="schedule-filter-bar"
                role="group"
                aria-label="Szűrés nap szerint"
            >
                <button
                    type="button"
                    class="schedule-filter-btn"
                    class:schedule-filter-btn--active={scheduleFilterDay ===
                        null}
                    aria-pressed={scheduleFilterDay === null}
                    title="Teljes program — minden nap"
                    aria-label="Teljes napi program megjelenítése, minden nap"
                    on:click={() => (scheduleFilterDay = null)}
                >
                    Összes
                </button>
                {#each event.schedule as day}
                    {@const dayKey = normalizeScheduleDateKey(
                        day.schedule_date,
                    )}
                    {@const isToday = isScheduleDayToday(
                        day.schedule_date,
                    )}
                    {@const fullDateLine = formatDate(
                        day.schedule_date,
                    )}
                    <button
                        type="button"
                        class="schedule-filter-btn"
                        class:schedule-filter-btn--today={isToday}
                        class:schedule-filter-btn--active={scheduleFilterDay ===
                            dayKey}
                        aria-pressed={scheduleFilterDay === dayKey}
                        title={fullDateLine +
                            (isToday ? " — ma" : "")}
                        aria-label={"Napi program szűrése: " +
                            fullDateLine +
                            (isToday ? ", mai nap" : "")}
                        on:click={() =>
                            (scheduleFilterDay = dayKey)}
                    >
                        {scheduleFilterChipLabel(
                            day.schedule_date,
                            isToday,
                        )}
                    </button>
                {/each}
            </div>
        </div>
        {#each filteredScheduleDays as day}
            {@const dayKey = normalizeScheduleDateKey(day.schedule_date)}
            <div class="schedule-day">
                <h3 class="schedule-day-title">
                    {formatDate(day.schedule_date)}
                </h3>
                <ul class="schedule-activities">
                    {#each day.activities as act}
                        {@const actVenue = activityVenueLabel(event, act)}
                        {@const timePart = formatScheduleTimePart(act)}
                        {@const showNowBadge =
                            timePart &&
                            isScheduleActivityHappeningNow(
                                act,
                                dayKey,
                                referenceNow,
                            )}
                        {@const hasTypePill =
                            act.activity_type &&
                            act.activity_type !== "other" &&
                            SCHEDULE_ACTIVITY_TYPE_LABELS[act.activity_type]}
                        {@const hasTypeCol =
                            !!hasTypePill || !!timePart || !!actVenue}
                        <li
                            class="schedule-activity schedule-activity--{act.activity_type || 'other'}"
                            class:schedule-activity--no-typecol={!hasTypeCol}
                        >
                            {#if hasTypeCol}
                                <div class="schedule-activity-type-col">
                                    {#if hasTypePill}
                                        <span class="schedule-type-pill"
                                            >{SCHEDULE_ACTIVITY_TYPE_LABELS[
                                                act.activity_type
                                            ]}</span
                                        >
                                    {/if}
                                    {#if timePart}
                                        <div class="schedule-activity-time-row">
                                            <span class="schedule-activity-time">{timePart}</span>
                                            {#if showNowBadge}
                                                <span
                                                    class="schedule-activity-now-badge"
                                                    title="A főoldali óra szerint épp most zajlik"
                                                    >Most</span
                                                >
                                            {/if}
                                        </div>
                                    {/if}
                                    {#if actVenue}
                                        <span class="schedule-activity-venue schedule-activity-venue--meta"
                                            >{actVenue}</span
                                        >
                                    {/if}
                                </div>
                            {/if}
                            <div class="schedule-activity-main">
                                <h4 class="schedule-activity-title">{act.title}</h4>
                                {#if act.description}
                                    <p class="schedule-activity-desc">{act.description}</p>
                                {/if}
                            </div>
                        </li>
                    {/each}
                </ul>
            </div>
        {/each}
    </section>
{/if}
{/if}

<style>
    .back-link {
        display: inline-block;
        margin-bottom: 1.5rem;
        color: var(--primary-color);
        text-decoration: none;
        font-size: 0.9rem;
        font-weight: 500;
    }
    .back-link:hover {
        text-decoration: underline;
    }

    .event-detail {
        background: var(--card-bg);
        border: 1px solid var(--border-color);
        border-radius: 12px;
        padding: 2rem;
    }

    .event-detail-header {
        margin-bottom: 1.5rem;
    }
    .event-detail-header .badge {
        margin-bottom: 0.75rem;
        display: inline-block;
    }
    .event-detail-header .page-title {
        margin: 0;
    }

    .event-detail-meta {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        margin-bottom: 2rem;
        padding-bottom: 1.5rem;
        border-bottom: 1px solid var(--border-color);
    }
    .meta-row {
        display: flex;
        align-items: center;
        gap: 0.6rem;
        font-size: 0.95rem;
        color: var(--text-faint);
    }
    .meta-row--wrap {
        flex-wrap: wrap;
    }
    .event-datetime-with-badge {
        display: inline-flex;
        align-items: center;
        flex-wrap: wrap;
        gap: 0.35rem;
    }
    .meta-row svg {
        flex-shrink: 0;
    }
    .meta-row a {
        color: var(--primary-color);
        text-decoration: none;
    }
    .meta-row a:hover {
        text-decoration: underline;
    }

    .event-detail-body {
        line-height: 1.7;
        color: var(--text-color);
        font-size: 1rem;
    }
    .event-detail-body p {
        margin: 0;
    }

    .event-schedule {
        margin-top: 2rem;
        padding-top: 1.5rem;
        border-top: 1px solid var(--border-color);
    }
    .event-schedule-head {
        display: flex;
        flex-wrap: wrap;
        align-items: flex-end;
        justify-content: space-between;
        gap: 0.75rem 1rem;
        margin-bottom: 1rem;
    }
    @media (max-width: 640px) {
        .event-schedule-head {
            flex-direction: column;
            align-items: stretch;
        }
        .event-schedule-title {
            min-width: 0;
        }
    }
    .event-schedule-title {
        font-size: 1.15rem;
        margin: 0;
        color: var(--text-color);
        flex: 1 1 auto;
        min-width: 9rem;
    }
    .schedule-filter-bar {
        display: flex;
        flex-wrap: wrap;
        gap: 0.4rem;
        align-items: center;
        justify-content: flex-end;
        max-width: 100%;
    }
    @media (max-width: 640px) {
        .schedule-filter-bar {
            flex-wrap: nowrap;
            overflow-x: auto;
            overflow-y: hidden;
            justify-content: flex-start;
            padding-bottom: 0.35rem;
            margin: 0 -0.25rem;
            padding-left: 0.25rem;
            padding-right: 0.25rem;
            -webkit-overflow-scrolling: touch;
            scrollbar-width: thin;
        }
    }
    .schedule-filter-btn {
        font-size: 0.85rem;
        font-weight: 600;
        padding: 0.4rem 0.7rem;
        border-radius: 999px;
        border: 1px solid var(--border-color);
        background: var(--card-bg);
        color: var(--text-color);
        cursor: pointer;
        line-height: 1.2;
        transition:
            background 0.15s ease,
            color 0.15s ease,
            border-color 0.15s ease;
    }
    .schedule-filter-btn:hover:not(.schedule-filter-btn--active) {
        border-color: var(--primary-color, var(--primary, #5c6bc0));
        color: var(--primary-color, var(--primary, #5c6bc0));
    }
    /* Today: distinct from default (not selected) */
    .schedule-filter-btn--today:not(.schedule-filter-btn--active) {
        border-color: var(--szekely-green, #357a6f);
        color: var(--szekely-green, #357a6f);
        background: color-mix(
            in srgb,
            var(--szekely-green, #357a6f) 14%,
            var(--card-bg)
        );
    }
    .schedule-filter-btn--active {
        background: var(--primary-color, var(--primary, #5c6bc0));
        color: #fff;
        border-color: transparent;
    }
    .schedule-filter-btn--active:hover {
        filter: brightness(1.05);
    }
    .schedule-day {
        margin: 1.5rem 0;
    }
    .schedule-day-title {
        font-size: 1rem;
    }
    .schedule-day-notes {
        margin: 0 0 0.75rem;
        font-size: 0.9rem;
        color: var(--text-faint);
    }
    .schedule-activities {
        margin: 0;
        padding-left: 1.2rem;
        list-style: disc;
    }
    .schedule-activity {
        margin: 0.75rem 0;
        display: grid;
        grid-template-columns: auto 1fr;
        align-items: start;
        gap: 0.35rem 0.75rem;
    }
    .schedule-activity--no-typecol {
        grid-template-columns: 1fr;
    }
    .schedule-activity-type-col {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: 0.35rem;
        max-width: 11rem;
        flex-shrink: 0;
    }
    .schedule-activity-time-row {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.35rem 0.45rem;
    }
    .schedule-type-pill {
        font-size: 0.68rem;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.03em;
        color: var(--szekely-red, #c0392b);
        flex-shrink: 0;
        line-height: 1.25;
    }
    .schedule-activity-main {
        min-width: 0;
    }
    .schedule-activity-time {
        font-size: 0.85rem;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.03em;
        line-height: 1.25;
        color: var(--text-color);
        flex-shrink: 0;
        white-space: nowrap;
    }
    .schedule-activity-now-badge {
        display: inline-block;
        font-size: 0.62rem;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.04em;
        padding: 0.15rem 0.4rem;
        border-radius: 4px;
        vertical-align: middle;
        line-height: 1.2;
        background: #dcfce7;
        color: #166534;
        border: 1px solid #86efac;
        flex-shrink: 0;
    }
    .schedule-activity-venue--meta {
        display: block;
        width: 100%;
        font-size: 0.68rem;
        color: var(--text-faint);
        line-height: 1.35;
        word-break: break-word;
    }
    .meta-row--venue strong {
        font-weight: 600;
        color: var(--text-faint);
        margin-right: 0.25rem;
    }
    .meta-row-venue-list {
        flex: 1;
        min-width: 0;
    }
    .meta-venue-sep {
        color: var(--text-faint);
        font-weight: 400;
    }
    .schedule-activity-title {
        display: block;
        margin: 0;
    }
    .schedule-activity-desc {
        margin: 0.35rem 0 0;
        font-size: 0.9rem;
        line-height: 1.5;
        color: var(--text-faint);
    }
    @media (max-width: 560px) {
        .schedule-activity-desc {
            margin-top: 0.4rem;
        }
        .schedule-activity:not(.schedule-activity--no-typecol) {
            grid-template-columns: 1fr;
        }
        .schedule-activity-type-col {
            max-width: none;
        }
    }

    .event-detail {
        box-sizing: border-box;
    }
    @media (max-width: 560px) {
        .event-detail {
            padding: 1.25rem;
        }
        .page-title {
            font-size: clamp(1.35rem, 5vw, 1.75rem);
            word-break: break-word;
        }
    }
</style>
