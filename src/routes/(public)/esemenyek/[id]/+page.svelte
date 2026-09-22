<script>
    import { onMount } from "svelte";
    import { get } from "svelte/store";
    import { page } from "$app/stores";
    import { apiFetch } from "$lib/api";
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
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import EventDateBadge from "$lib/components/EventDateBadge.svelte";
    import {
        getReferenceNow,
        isScheduleActivityHappeningNow,
        referenceTodayYMD,
    } from "$lib/eventStatus";
    import { SCHEDULE_ACTIVITY_TYPE_LABELS } from "$lib/scheduleActivityTypes.js";
    import { venuePageUrl } from "$lib/utils";
    import { eventFeaturedImageUrl } from "$lib/eventImage.js";
    import EventEntryPrice from "$lib/components/EventEntryPrice.svelte";
    import { kindLabel } from "$lib/venueKindLabels.js";

    let event = null;
    let loading = true;
    let error = null;
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
        initFavorites();
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

    const ACCESS_LABELS = {
        public: "Nyitott",
        members_only: "Zártkörű",
        invitation_only: "Meghívóval",
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
        const defN = String(ev.default_venue_name || "").trim();
        const defS = String(ev.default_venue_slug || "").trim();
        if (defN) {
            const kindLabelText = kindLabel(
                /** @type {string | null | undefined} */ (ev.default_venue_kind),
                /** @type {string | null | undefined} */ (ev.default_venue_kind_label),
            );
            const fullName =
                kindLabelText && kindLabelText.trim() !== ""
                    ? `${defN} ${kindLabelText}`
                    : defN;
            push(fullName, defS);
        }
        const sched = ev.schedule;
        if (Array.isArray(sched)) {
            for (const day of sched) {
                const acts = day?.activities;
                if (!Array.isArray(acts)) continue;
                for (const act of acts) {
                    push(
                        activityVenueLabel(ev, act),
                        activityVenueSlugForLink(ev, act),
                    );
                }
            }
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

<PublicPageHero
    title={event ? event.title : "Esemény"}
    greeting=""
    loading={loading}
    breadcrumbLabel=""
    breadcrumbParentLabel="Index"
    breadcrumbParentUrl="/index"
    breadcrumbExtraLabel="Események"
    breadcrumbExtraUrl="/esemenyek"
    showTitle={false}
    showGreeting={false}
    documentTitleSuffix={event
        ? ` - ${event.title} - Lámsza Index`
        : " - Lámsza Index"}
/>

{#if loading}
    <span class="info-box"><p>esemény adatok betöltése...</p></span>
{:else if error || !event}
    <span class="info-box"><p>{error || "Az esemény nem található."}</p></span>
{:else}

    <article class="event-detail">
        <div class="event-detail-header">
            <div class="event-detail-badges">
                {#if event.event_type}
                    <div class="badge event">
                        {event.event_type_label ||
                            EVENT_TYPE_LABELS[event.event_type] ||
                            event.event_type}
                    </div>
                {/if}
                {#if event.event_subtype_label || event.event_subtype}
                    <div class="badge event event-badge--subtype">
                        {event.event_subtype_label ||
                            event.event_subtype}
                    </div>
                {/if}
            </div>
            <h1 class="page-title">{event.title}</h1>
            <div class="event-detail-hero">
                <img
                    class="featured-img"
                    src={eventFeaturedImageUrl(event)}
                    alt=""
                    loading="eager"
                    decoding="async"
                />
            </div>
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
                    <span class="meta-row-label"><strong>Helyszín:</strong></span>
                    <span class="meta-row-value">
                        {#each eventMetaVenues as v, vi (v.slug ? v.slug : v.name + String(vi))}
                            {#if vi > 0}<span class="meta-venue-sep"> · </span>{/if}
                            {#if v.slug && venuePageUrl(event, v.slug)}
                                <a href={venuePageUrl(event, v.slug)} title="Helyszín részletei IIIIITTTTT">{v.name}</a>
                            {:else}
                                {v.name}
                            {/if}
                        {/each}
                    </span>
                </div>
            {/if}

            <div class="meta-row meta-row--entry-price">
                <EventEntryPrice value={event.entry_price} size="lg" />
            </div>

            {#if event.organizer}
                <div class="meta-row">
                    <svg xmlns="http://www.w3.org/2000/svg" height="18px" viewBox="0 -960 960 960" width="18px" fill="currentColor"><path d="M480-240q-56 0-107 17.5T280-170v10h400v-10q-42-35-93-52.5T480-240Zm-280 34q54-53 125.5-83.5T480-320q83 0 154.5 30.5T760-206v-514H200v514Zm181-235q-41-41-41-99t41-99q41-41 99-41t99 41q41 41 41 99t-41 99q-41 41-99 41t-99-41Zm141.5-56.5Q540-515 540-540t-17.5-42.5Q505-600 480-600t-42.5 17.5Q420-565 420-540t17.5 42.5Q455-480 480-480t42.5-17.5ZM200-80q-33 0-56.5-23.5T120-160v-560q0-33 23.5-56.5T200-800h40v-80h80v80h320v-80h80v80h40q33 0 56.5 23.5T840-720v560q0 33-23.5 56.5T760-80H200Zm280-460Zm0 380h200-400 200Z"/></svg>
                    <span class="meta-row-label">Szervező:</span>
                    <span class="meta-row-value">{event.organizer}</span>
                </div>
            {/if}

            {#if event.access_type !== "public"}
            <div class="meta-row meta-row--wrap">
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path><circle cx="12" cy="10" r="3"></circle></svg>
                <span class="meta-row-label">Hozzáférés:</span>
                <span class="meta-row-value">{ACCESS_LABELS[event.access_type] || event.access_type}</span>
            </div>
            {/if}

            {#if event.broadcast_type}
                <div class="meta-row meta-row--wrap">
                    <svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M320-320h80v-240h70l90 240h80l120-320H660l-60 180-60-180H200v80h120v240ZM200-120q-33 0-56.5-23.5T120-200v-560q0-33 23.5-56.5T200-840h560q33 0 56.5 23.5T840-760v560q0 33-23.5 56.5T760-120H200Zm0-80h560v-560H200v560Zm0-560v560-560Z"/></svg>
                    <span class="meta-row-label"><strong>Közvetítés:</strong></span>
                    <span class="meta-row-value">{event.broadcast_type || event.broadcast_type}</span>
                </div>
            {/if}
        </div>  

        {#if event.description}
            <div class="event-detail-body">
                <p>{event.description}</p>
            </div>
        {/if}

        <div class="page-actions">
            <div class="share-buttons">
                <button class="btn btn-lg btn-share" aria-label="Megosztás">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 2h-3a5 5 0 0 0-5 5v3H7a5 5 0 0 0 0 10h3v3a5 5 0 0 0 10 0v-3h3a5 5 0 0 0 0-10z"></path></svg>
                </button>
            </div>
            <FavoriteButton
                type="event"
                id={event.id}
                active={isFavorite(favoriteList, "event", event.id)}
                ontoggle={() =>
                    handleFavoriteToggle(
                        "event",
                        event.id,
                        isFavorite(favoriteList, "event", event.id),
                    )}
            />
            <div class="contact-button">
                <button class="btn btn-lg btn-contact" aria-label="Kapcsolat">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path><circle cx="12" cy="10" r="3"></circle></svg>
                </button>
            </div>
        </div>
        {#if event.gallery && event.gallery.length > 0}
            <div class="gallery event-gallery">
                {#each event.gallery as image}
                    <div class="gallery-item">
                        <img src={image.url} alt={image.alt} />
                    </div>
                {/each}
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
                    class="btn btn-sm"
                    class:active={scheduleFilterDay ===
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
                        class="btn btn-sm"
                        class:filter-btn--today={isToday}
                        class:active={scheduleFilterDay ===
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
        font-weight: 500;
    }
    .back-link:hover {
        text-decoration: underline;
    }

    .event-detail-hero {
        margin: 1rem 0;
        border-radius: 12px;
        overflow: hidden;
        aspect-ratio: 21 / 9;
        background: var(--skeleton-bg, #e8eaef);
    }
    .featured-img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        display: block;
    }

    .event-detail-header {
        margin-bottom: 2rem;
    }
    .event-detail-badges {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.5rem;
        margin-bottom: 0.75rem;
    }
    .event-detail-header .badge {
        margin-bottom: 0;
        display: inline-block;
    }
    .event-detail-header .event-badge--subtype {
        background: color-mix(in srgb, var(--primary-color, #375f8c) 88%, #fff);
        color: #fff;
    }
    .event-detail-header .page-title {
        margin: 0;
    }

    .event-detail-meta {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        margin-bottom: 2rem;
        padding-bottom: 2rem;
        border-bottom: 1px solid var(--border-color);
    }
    .meta-row {
        display: flex;
        align-items: center;
        gap: 0.5rem;
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

    .event-detail-body {
        line-height: 1.7;
        border-bottom: 1px solid var(--border-color);
        margin-bottom: 2rem;
        padding-bottom: 2rem;
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
        margin: 0;
        
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
    .filter-btn {
        transition:
            background 0.15s ease,
            color 0.15s ease,
            border-color 0.15s ease;
        color: var(--white);
    }
    /* Today: distinct from default (not selected) */
    .filter-btn--today:not(.active) {
        border-color: var(--szekely-red, #c0392b);
        color: var(--white);
        background: color-mix(
            in srgb,
            var(--szekely-red, #c0392b) 14%,
            var(--card-bg)
        );
    }
    .schedule-day {
        margin: 1.5rem 0;
    }
    .schedule-day-title {
    }
    .schedule-day-notes {
        margin: 0 0 0.75rem;
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
        border-bottom: 1px solid var(--border-color);
        padding-bottom: 0.75rem;
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
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.03em;
        color: var(--szekely-blue, #375f8c);
        flex-shrink: 0;
        line-height: 1.25;
    }
    .schedule-activity-main {
        min-width: 0;
    }
    .schedule-activity-time {
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.03em;
        line-height: 1.25;
        
        flex-shrink: 0;
        white-space: nowrap;
    }
    .schedule-activity-now-badge {
        display: inline-block;
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
        color: var(--text-faint);
        line-height: 1.35;
        word-break: break-word;
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
    @media (max-width: 560px) {
        .event-detail {
            padding: 1.25rem;
        }
        .page-title {
            word-break: break-word;
        }
    }
</style>