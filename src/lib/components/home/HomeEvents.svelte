<script>
    import { apiFetch } from "$lib/api";
    import { browser } from "$app/environment";

    /**
     * "Közelgő események" on the home page: the next four events from
     * /api/events (which already answers by Bucharest's day, R19), each with a
     * date block. One row of cards at every width (it scrolls sideways on
     * phones), so loading, a short list or none never changes its height
     * (ld-reserve-space). With a saved place, the events are that place's.
     */
    /** @type {{ settlementSlug?: string | null, locationName?: string | null }} */
    let { settlementSlug = null, locationName = null } = $props();

    const COUNT = 4;
    const MONTHS = ["JAN", "FEB", "MÁR", "ÁPR", "MÁJ", "JÚN", "JÚL", "AUG", "SZEP", "OKT", "NOV", "DEC"];

    /** @type {Array<{ id: number, title: string, start_date: string, start_time?: string, location_name?: string, county?: string, default_venue_name?: string }>} */
    let events = $state([]);
    let loading = $state(true);

    $effect(() => {
        if (!browser) return;
        const slug = settlementSlug;
        loading = true;
        apiFetch(slug ? `/api/events?location_slug=${encodeURIComponent(slug)}` : "/api/events")
            .then((data) => {
                events = (data?.events || []).slice(0, COUNT);
            })
            .catch(() => {
                events = [];
            })
            .finally(() => {
                loading = false;
            });
    });

    /** "2026-11-07" → { month: "NOV", day: 7 }, read as a Bucharest calendar date. */
    /** @param {string} ymd */
    function dateParts(ymd) {
        const [, m, d] = String(ymd || "").split("-").map(Number);
        return { month: MONTHS[(m || 1) - 1], day: d || "" };
    }

    /** @param {{ start_time?: string, default_venue_name?: string, location_name?: string, county?: string }} e */
    function place(e) {
        const where = [e.default_venue_name, e.location_name].filter(Boolean).join(", ") || (e.county ? `${e.county} megye` : "");
        const time = e.start_time ? e.start_time.slice(0, 5) : "";
        return [time, where].filter(Boolean).join(" · ");
    }
</script>

<section class="home-section home-events" aria-labelledby="home-events-title">
    <div class="home-section__head">
        <h2 id="home-events-title" class="widget-title">Közelgő események{#if locationName}<span class="type-label">&nbsp;·&nbsp;{locationName}</span>{/if}</h2>
        <a class="home-section__more" href="/esemenyek">Összes esemény ›</a>
    </div>
    <ul class="home-events__row" aria-busy={loading}>
        {#if loading}
            {#each { length: COUNT } as _, i (i)}
                <li class="card home-event" aria-hidden="true">
                    <span class="home-event__date skeleton"><span>&nbsp;</span><b>&nbsp;</b></span>
                    <span class="home-event__text"><span class="home-event__title skeleton">&nbsp;</span><span class="home-event__place skeleton">&nbsp;</span></span>
                </li>
            {/each}
        {:else if events.length === 0}
            <li class="card home-event home-event--empty">
                <span class="home-event__text"><span class="home-event__title">Nincs közelgő esemény.</span><a class="home-event__place" href="/esemenyek">Az eseménynaptár ›</a></span>
            </li>
        {:else}
            {#each events as e (e.id)}
                {@const d = dateParts(e.start_date)}
                <li>
                    <a class="card home-event" href="/esemenyek/{e.id}">
                        <span class="home-event__date"><span>{d.month}</span><b>{d.day}</b></span>
                        <span class="home-event__text">
                            <span class="home-event__title">{e.title}</span>
                            <span class="home-event__place">{place(e)}</span>
                        </span>
                    </a>
                </li>
            {/each}
        {/if}
    </ul>
</section>

<style>
    .home-events__row {
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
        gap: 1rem;
        list-style: none;
        margin: 0;
        padding: 0;
    }

    .home-events__row > li {
        min-width: 0;
    }

    .home-event {
        display: flex;
        align-items: center;
        gap: 0.9rem;
        height: 100%;
        min-height: 5.5rem;
        box-sizing: border-box;
    }

    .home-event--empty {
        grid-column: 1 / -1;
    }

    .home-event__date {
        display: flex;
        flex-direction: column;
        align-items: center;
        flex: none;
        width: 3.4rem;
        padding: 0.35rem 0;
        border: 1px solid var(--border-color);
        border-radius: 12px;
        line-height: 1.1;
    }

    .home-event__date span {
        font-size: var(--text-xs);
        font-weight: 600;
        color: var(--szekely-red);
    }

    .home-event__date b {
        font-size: var(--text-xl);
    }

    .home-event__text {
        display: flex;
        flex-direction: column;
        gap: 0.2rem;
        min-width: 0;
    }

    .home-event__title {
        font-weight: 700;
        line-height: 1.3;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }

    a.home-event:hover .home-event__title {
        color: var(--szekely-red);
    }

    .home-event__text:has(.skeleton) {
        flex: 1;
    }

    .home-event__title.skeleton {
        display: block;
        width: 80%;
    }

    .home-event__place.skeleton {
        width: 50%;
    }

    .home-event__place {
        color: var(--text-muted);
        font-size: var(--text-sm);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    @media (max-width: 760px) {
        .home-events__row {
            display: flex;
            overflow-x: auto;
            scroll-snap-type: x mandatory;
            scrollbar-width: none;
        }

        .home-events__row > li {
            flex: 0 0 80%;
            scroll-snap-align: start;
        }

        .home-events__row > .home-event--empty {
            flex-basis: 100%;
        }
    }
</style>
