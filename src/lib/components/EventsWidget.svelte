<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import { formatDate } from "$lib/utils";

    export let settlementSlug = null;
    export let limit = 3;

    let items = [];
    let loading = true;
    let error = false;

    onMount(async () => {
        try {
            const url = settlementSlug
                ? `/api/events?location_slug=${encodeURIComponent(settlementSlug)}`
                : `/api/events`;

            const data = await apiFetch(url);
            items = data || [];
        } catch (err) {
            error = true;
        } finally {
            loading = false;
        }
    });
</script>

<article id="esemenyek" class="event-widget">
    <h3 class="widget-title">Események</h3>
    {#if loading}
        <div class="skeleton-box">
            <div class="skeleton skeleton-text skeleton-full"></div>
            <div class="skeleton skeleton-text skeleton-60"></div>
        </div>
    {:else if error || items.length === 0}
        <span class="info-box"><p>Nincsenek közeli események.</p></span>
    {:else}
        <ul class="mini-event-list">
            {#each items.slice(0, limit) as event}
                <li>
                    <div class="mini-event-date">
                        {new Date(event.start_date).toLocaleDateString(
                            "hu-HU",
                            { month: "short", day: "numeric" },
                        )}
                    </div>
                    <div class="mini-event-info">
                        <span class="mini-event-title">{event.title}</span>
                        <span class="mini-event-time">
                            {#if event.start_time}{event.start_time.slice(
                                    0,
                                    5,
                                )}{/if}
                            {#if event.end_date && event.end_date !== event.start_date}
                                - {formatDate(event.end_date)}
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
        <a href="/esemenyek" class="widget-more-link">Összes esemény →</a>
    {/if}
</article>

<style>
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
    }
    .mini-event-time {
        font-size: 0.8rem;
        color: var(--text-faint);
    }
    .widget-more-link {
        margin-top: 1rem;
        font-size: 0.9rem;
        color: var(--primary-color);
        text-decoration: none;
        font-weight: 500;
    }
</style>
