<script>
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import { apiFetch } from "$lib/api";
    import { formatDateShort } from "$lib/utils";

    const KIND_LABELS = {
        sports_arena: "Sportcsarnok / pálya",
        indoor_hall: "Fedett csarnok",
        outdoor_area: "Szabadtéri terület",
        market_square: "Piac / tér",
        park: "Park",
        street: "Utca / felvonulás",
        mixed: "Több helyszín",
        temporary: "Ideiglenes",
        other: "Egyéb",
    };

    /** @type {Record<string, unknown> | null} */
    let venue = null;
    /** @type {Record<string, unknown>[]} */
    let venueEvents = [];
    /** @type {Record<string, unknown>[]} */
    let venueRelatedEvents = [];
    let loading = true;
    let error = null;

    $: countySlug = String($page.params.countySlug || "").toLowerCase();
    $: settlementSlug = String($page.params.slug || "").toLowerCase();
    $: venueSlug = String($page.params.venueSlug || "").toLowerCase();

    $: if (browser && countySlug && settlementSlug && venueSlug) {
        loadVenue();
    }

    async function loadVenue() {
        loading = true;
        error = null;
        venue = null;
        venueEvents = [];
        venueRelatedEvents = [];
        try {
            const q = new URLSearchParams({
                county_slug: countySlug,
                settlement_slug: settlementSlug,
                venue_slug: venueSlug,
            });
            venue = await apiFetch(`/api/venues?${q}`);
            const rawId = venue?.id;
            const vid =
                typeof rawId === "number"
                    ? rawId
                    : typeof rawId === "string" && /^\d+$/.test(rawId)
                      ? parseInt(rawId, 10)
                      : null;
            if (vid != null) {
                try {
                    const evData = await apiFetch(
                        `/api/events?venue_id=${vid}&limit=50`,
                    );
                    venueEvents = Array.isArray(evData.events)
                        ? evData.events
                        : [];
                } catch {
                    venueEvents = [];
                }
                const kind = String(venue?.kind || "").trim();
                if (kind) {
                    try {
                        const rel = new URLSearchParams({
                            location_slug: settlementSlug,
                            county_slug: countySlug,
                            venue_kind: kind,
                            exclude_venue_id: String(vid),
                            limit: "40",
                        });
                        const relData = await apiFetch(
                            `/api/events?${rel.toString()}`,
                        );
                        const raw = Array.isArray(relData.events)
                            ? relData.events
                            : [];
                        const atThis = new Set(
                            venueEvents.map((e) => e.id).filter(Boolean),
                        );
                        venueRelatedEvents = raw.filter((e) => e.id && !atThis.has(e.id));
                    } catch {
                        venueRelatedEvents = [];
                    }
                }
            }
        } catch {
            error = true;
        } finally {
            loading = false;
        }
    }

    const EVENT_TYPE_LABELS = {
        cultural: "Kulturális",
        sports: "Sport",
        festival: "Fesztivál",
        religious: "Vallási",
        other: "Egyéb",
    };

    function kindLabel(k, labelFromApi) {
        if (labelFromApi && String(labelFromApi).trim()) return String(labelFromApi);
        return KIND_LABELS[/** @type {keyof typeof KIND_LABELS} */ (k)] || k;
    }
</script>

<svelte:head>
    <title
        >{venue
            ? `${String(venue.name)} — ${String(venue.settlement_name || "")}`
            : "Helyszín"} · Na Lámsza!</title
    >
</svelte:head>

{#if loading}
    <span class="info-box"><p>adat betöltés...</p></span>
{:else if error || !venue}
    <span class="info-box"><p>A helyszín nem található.</p></span>
{:else}
    <Breadcrumbs
        label={String(venue.name)}
        countyName={String(venue.county_name || "")}
        countySlug={countySlug}
        settlementSlug={settlementSlug}
        settlementName={String(venue.settlement_name || "")}
    />

    <article class="venue-detail card-like">
        <h1 class="page-title">{venue.name}</h1>
        <p class="venue-sub">
            <a href="/{countySlug}-megye/{settlementSlug}" class="parent-city-link"
                >{venue.settlement_name}</a
            >,
            <a href="/{countySlug}-megye" class="parent-city-link"
                >{venue.county_name} megye</a
            >
        </p>

        <div class="venue-meta-grid">
            {#if venue.name_ro}
                <div class="venue-meta-item">
                    <span class="venue-meta-label">Románul</span>
                    <span>{venue.name_ro}</span>
                </div>
            {/if}
            {#if venue.name_de}
                <div class="venue-meta-item">
                    <span class="venue-meta-label">Németül</span>
                    <span>{venue.name_de}</span>
                </div>
            {/if}
            <div class="venue-meta-item">
                <span class="venue-meta-label">Típus</span>
                <span>{kindLabel(venue.kind, venue.kind_label)}</span>
            </div>
            {#if venue.address}
                <div class="venue-meta-item venue-meta-item--wide">
                    <span class="venue-meta-label">Cím</span>
                    <span>{venue.address}</span>
                </div>
            {/if}
            {#if venue.latitude != null && venue.longitude != null}
                <div class="venue-meta-item venue-meta-item--wide">
                    <span class="venue-meta-label">Koordináták</span>
                    <span
                        >{Number(venue.latitude).toFixed(5)}, {Number(
                            venue.longitude,
                        ).toFixed(5)}</span
                    >
                </div>
            {/if}
            {#if venue.seating_capacity != null}
                <div class="venue-meta-item">
                    <span class="venue-meta-label">Férőhely</span>
                    <span>{venue.seating_capacity}</span>
                </div>
            {/if}
        </div>

        {#if venue.description}
            <div class="venue-description">
                <h2 class="venue-section-title">Leírás</h2>
                <p class="venue-desc-text">{venue.description}</p>
            </div>
        {/if}
    </article>

    {#if venueEvents.length > 0}
        <section class="venue-events card-like" aria-label="Események ezen a helyszínen">
            <h2 class="venue-section-title">Események itt</h2>
            <p class="venue-events-hint">
                Közelgő események, amelyeknek ez a helyszín az alapértelmezett helyszíne, vagy szerepel a napi programban
                (ugyanabban a településben).
            </p>
            <ul class="venue-events-list">
                {#each venueEvents as ev (ev.id)}
                    <li class="venue-events-item">
                        <a href="/esemenyek/{ev.id}" class="venue-events-link">{ev.title}</a>
                        <span class="venue-events-meta">
                            {#if ev.event_type}
                                <span class="venue-events-type"
                                    >{EVENT_TYPE_LABELS[ev.event_type] || ev.event_type}</span
                                >
                            {/if}
                            {#if ev.start_date}
                                <span class="venue-events-date">{formatDateShort(ev.start_date)}</span>
                            {/if}
                        </span>
                    </li>
                {/each}
            </ul>
        </section>
    {/if}

    {#if venueRelatedEvents.length > 0}
        <section
            class="venue-events venue-events--related card-like"
            aria-label="További események a településen hasonló helyszínnel"
        >
            <h2 class="venue-section-title">További események a településen</h2>
            <p class="venue-events-hint">
                Ugyanilyen helyszíntípus ({kindLabel(venue.kind, venue.kind_label)}) más helyszínen ebben a városban — nem
                közvetlenül ennél a címnél.
            </p>
            <ul class="venue-events-list">
                {#each venueRelatedEvents as ev (ev.id)}
                    <li class="venue-events-item">
                        <a href="/esemenyek/{ev.id}" class="venue-events-link">{ev.title}</a>
                        <span class="venue-events-meta">
                            {#if ev.event_type}
                                <span class="venue-events-type"
                                    >{EVENT_TYPE_LABELS[ev.event_type] || ev.event_type}</span
                                >
                            {/if}
                            {#if ev.start_date}
                                <span class="venue-events-date">{formatDateShort(ev.start_date)}</span>
                            {/if}
                            {#if ev.default_venue_name}
                                <span class="venue-events-venue-name">{ev.default_venue_name}</span>
                            {/if}
                        </span>
                    </li>
                {/each}
            </ul>
        </section>
    {/if}
{/if}

<style>
    .card-like {
        background: var(--card-bg);
        border: 1px solid var(--border-color);
        border-radius: 12px;
        padding: 2rem;
        margin-bottom: 1.5rem;
    }
    .venue-sub {
        margin: 0 0 1.5rem;
        font-size: 0.95rem;
        color: var(--text-faint);
    }
    .parent-city-link {
        color: var(--primary-color);
        text-decoration: none;
    }
    .parent-city-link:hover {
        text-decoration: underline;
    }
    .venue-meta-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(14rem, 1fr));
        gap: 1rem;
        margin-bottom: 1rem;
    }
    .venue-meta-item {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
        font-size: 0.95rem;
    }
    .venue-meta-item--wide {
        grid-column: 1 / -1;
    }
    .venue-meta-label {
        font-size: 0.75rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.04em;
        color: var(--text-faint);
    }
    .venue-section-title {
        font-size: 1.1rem;
        margin: 0 0 0.75rem;
        color: var(--primary-color);
    }
    .venue-desc-text {
        margin: 0;
        line-height: 1.65;
        color: var(--text-color);
        white-space: pre-wrap;
    }

    .venue-events-hint {
        margin: 0 0 1rem;
        font-size: 0.88rem;
        color: var(--text-faint);
        line-height: 1.5;
    }
    .venue-events-list {
        margin: 0;
        padding: 0;
        list-style: none;
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
    }
    .venue-events-item {
        display: flex;
        flex-direction: column;
        gap: 0.2rem;
        padding-bottom: 0.75rem;
        border-bottom: 1px solid var(--border-color);
    }
    .venue-events-item:last-child {
        border-bottom: none;
        padding-bottom: 0;
    }
    .venue-events-link {
        font-weight: 600;
        color: var(--primary-color);
        text-decoration: none;
        font-size: 1rem;
    }
    .venue-events-link:hover {
        text-decoration: underline;
    }
    .venue-events-meta {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem 0.75rem;
        font-size: 0.85rem;
        color: var(--text-faint);
    }
    .venue-events-type {
        font-weight: 600;
        text-transform: uppercase;
        font-size: 0.72rem;
        letter-spacing: 0.03em;
        color: var(--szekely-red, #c0392b);
    }
    .venue-events-venue-name {
        font-style: italic;
    }
    .venue-events--related {
        margin-top: 0.25rem;
    }
</style>
