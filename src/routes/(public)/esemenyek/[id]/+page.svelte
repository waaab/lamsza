<script>
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { apiFetch } from "$lib/api";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";

    let event = null;
    let loading = true;
    let error = null;

    $: eventId = $page.params.id;

    $: if (eventId) {
        loadEvent(eventId);
    }

    async function loadEvent(id) {
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

    <section id="event-detail">
        <article class="event-detail">
            <div class="event-detail-header">
                <div class="badge event">
                    {EVENT_TYPE_LABELS[event.event_type] || event.event_type}
                </div>
                <h1 class="page-title">{event.title}</h1>
            </div>

            <div class="event-detail-meta">
                <div class="meta-row">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                    <span>{formatEventDateTime(event)}</span>
                </div>

                <div class="meta-row">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path><circle cx="12" cy="10" r="3"></circle></svg>
                    <span>
                        <a href="/{event.county_slug}-megye/{event.location_slug}">{event.location_name}</a>,
                        <a href="/{event.county_slug}-megye" class="county-link">{event.county} megye</a>
                    </span>
                </div>

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
    </section>
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
</style>
