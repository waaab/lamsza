<script>
    import { onMount } from "svelte";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";

    const getBase = () =>
        import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

    let events = [];
    let loading = true;
    let error = null;
    let total = 0;
    let currentPage = 1;
    const pageSize = 12;

    async function loadPage(page) {
        loading = true;
        error = null;
        try {
            const offset = (page - 1) * pageSize;
            const res = await fetch(
                `${getBase()}/api/events?limit=${pageSize}&offset=${offset}`,
            );
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

    onMount(() => loadPage(1));

    $: totalPages = Math.max(1, Math.ceil(total / pageSize));

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

{#if loading}
    <span class="info-box"><p>adat betöltés...</p></span>
{:else if error}
    <span class="info-box"><p>{error}</p></span>
{:else if events.length === 0}
    <span class="info-box"
        ><p>Jelenleg nincsenek meghirdetett események.</p></span
    >
{:else}
    <div class="list grid" id="esemenyek-lista">
        {#each events as event}
            
                <article class="card">
                    <div class="badge event">
                        {EVENT_TYPE_LABELS[event.event_type] || event.event_type}
                    </div>
                    <h2 class="event-title"><a href="/esemenyek/{event.id}" class="event-card-link">{event.title}</a></h2>

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
                            <span class="event-date-time">{formatEventDateTime(event)}</span>
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
                                <a href="/{event.county_slug}-megye/{event.location_slug}" class="event-location-link">{event.location_name}</a>, <a href="/{event.county_slug}-megye" class="event-county-link">{event.county} megye</a>
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
                disabled={currentPage <= 1}
                on:click={() => loadPage(currentPage - 1)}
            >
                &#8249; Előző
            </button>
            <span class="pagination-info">{currentPage} / {totalPages}</span>
            <button
                class="pagination-btn"
                disabled={currentPage >= totalPages}
                on:click={() => loadPage(currentPage + 1)}
            >
                Következő &#8250;
            </button>
        </div>
    {/if}
{/if}

<style>
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

    .event-location, .event-date {
        display: flex;
        gap: 0.5rem;
        align-items: center;
        flex-direction: row;
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
